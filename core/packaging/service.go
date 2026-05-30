package packaging

import (
	"context"
	"errors"
	"path/filepath"
	"runtime"
	"strings"

	"wails3-manager/core/contracts"
	"wails3-manager/core/environment"
	"wails3-manager/core/fsx"
	"wails3-manager/core/packaging/config"
	"wails3-manager/core/packaging/dmg"
	"wails3-manager/core/packaging/inno"
	"wails3-manager/core/packaging/release"
	"wails3-manager/core/project"
	"wails3-manager/core/runlog"
)

type PackagingService struct {
	Log *runlog.Logger
}

func NewService(log *runlog.Logger) *PackagingService {
	return &PackagingService{Log: log}
}

func (s *PackagingService) InitPackaging(projectDir string) (contracts.PackagingConfig, error) {
	projectDir, err := fsx.NormalizePath(projectDir)
	if err != nil {
		return contracts.PackagingConfig{}, err
	}
	if existing, err := config.LoadPackagingConfig(projectDir); err == nil {
		if err := s.writeTemplates(projectDir, existing); err != nil {
			return contracts.PackagingConfig{}, err
		}
		return existing, nil
	}
	cfg := defaultPackagingConfig(projectDir)
	if err := config.SavePackagingConfig(projectDir, cfg); err != nil {
		return contracts.PackagingConfig{}, err
	}
	if err := s.writeTemplates(projectDir, cfg); err != nil {
		return contracts.PackagingConfig{}, err
	}
	return config.LoadPackagingConfig(projectDir)
}

func (s *PackagingService) LoadPackagingConfig(projectDir string) (contracts.PackagingConfig, error) {
	return config.LoadPackagingConfig(projectDir)
}

func (s *PackagingService) SavePackagingConfig(projectDir string, cfg contracts.PackagingConfig) (contracts.PackagingConfig, error) {
	projectDir, err := fsx.NormalizePath(projectDir)
	if err != nil {
		return contracts.PackagingConfig{}, err
	}
	if err := config.SavePackagingConfig(projectDir, cfg); err != nil {
		return contracts.PackagingConfig{}, err
	}
	if err := s.writeTemplates(projectDir, cfg); err != nil {
		return contracts.PackagingConfig{}, err
	}
	return config.LoadPackagingConfig(projectDir)
}

func (s *PackagingService) Package(req contracts.PackageRequest) (contracts.PackageResult, error) {
	projectDir, err := fsx.NormalizePath(req.ProjectDir)
	if err != nil {
		return contracts.PackageResult{}, err
	}
	cfg, err := config.LoadPackagingConfig(projectDir)
	if err != nil {
		return contracts.PackageResult{}, err
	}
	platform := normalizePlatform(req.Platform)
	runID := fsx.Timestamp()
	result := contracts.PackageResult{OK: true, RunID: runID, Warnings: []string{}}
	ctx := context.Background()
	if req.RunBuild {
		if err := s.runBuild(ctx, projectDir, cfg, req.DryRun); err != nil {
			return contracts.PackageResult{}, err
		}
	}
	// Packaging intentionally has no before/after script stages. The only custom
	// execution path is the explicit build command in packaging.json, which keeps
	// package review and dry-run behavior predictable.
	if platform == contracts.PlatformWindows || platform == contracts.PlatformAll {
		if _, err := inno.Generate(projectDir, cfg); err != nil {
			return contracts.PackageResult{}, err
		}
		if req.DryRun || runtime.GOOS != "windows" {
			result.Warnings = append(result.Warnings, "Windows 打包仅生成 Inno 脚本，未执行 ISCC。")
		} else if err := s.runISCC(ctx, projectDir, cfg); err != nil {
			return contracts.PackageResult{}, err
		}
	}
	if platform == contracts.PlatformMacOS || platform == contracts.PlatformAll {
		if _, err := dmg.GenerateScript(projectDir, cfg); err != nil {
			return contracts.PackageResult{}, err
		}
		if req.DryRun || runtime.GOOS != "darwin" {
			result.Warnings = append(result.Warnings, "macOS 打包仅生成 DMG 脚本，未执行 create-dmg。")
		} else if err := s.runDMG(ctx, projectDir, cfg); err != nil {
			return contracts.PackageResult{}, err
		}
	}
	result.Artifacts = release.ReleaseArtifacts(projectDir)
	result.Message = "打包流程完成"
	return result, nil
}

func (s *PackagingService) Artifacts(projectDir string) []contracts.Artifact {
	projectDir, err := fsx.NormalizePath(projectDir)
	if err != nil {
		return []contracts.Artifact{}
	}
	return release.ReleaseArtifacts(projectDir)
}

func (s *PackagingService) writeTemplates(projectDir string, cfg contracts.PackagingConfig) error {
	if cfg.Windows.InnoScript != "" {
		if _, err := inno.Generate(projectDir, cfg); err != nil {
			return err
		}
	}
	if cfg.MacOS.Background != "" {
		bg := fsx.Resolve(projectDir, cfg.MacOS.Background)
		if _, err := fsx.WriteIfMissing(bg, dmg.DefaultBackgroundPNG(), 0644); err != nil {
			return err
		}
	}
	if cfg.MacOS.DMGScript != "" {
		if _, err := dmg.GenerateScript(projectDir, cfg); err != nil {
			return err
		}
	}
	return nil
}

func (s *PackagingService) runBuild(ctx context.Context, projectDir string, cfg contracts.PackagingConfig, dryRun bool) error {
	cmd := cfg.Build.Command
	if len(cmd) == 0 {
		task := fsx.FirstNonEmpty(cfg.Build.Task, "builder:release")
		cmd = []string{"wails3", "task", task}
	}
	return (runlog.Runner{Log: s.Log, DryRun: dryRun}).Run(ctx, projectDir, cmd)
}

func (s *PackagingService) runISCC(ctx context.Context, projectDir string, cfg contracts.PackagingConfig) error {
	req := environment.InnoRequirement(cfg.Windows.ISCCPath)
	if !req.Found {
		return errors.New(req.Message)
	}
	script := fsx.Resolve(projectDir, cfg.Windows.InnoScript)
	return (runlog.Runner{Log: s.Log}).Run(ctx, projectDir, []string{req.Path, script})
}

func (s *PackagingService) runDMG(ctx context.Context, projectDir string, cfg contracts.PackagingConfig) error {
	req := environment.CreateDMGRequirement(cfg.MacOS.CreateDMGPath)
	if !req.Found {
		return errors.New(req.Message)
	}
	script := fsx.Resolve(projectDir, cfg.MacOS.DMGScript)
	return (runlog.Runner{Log: s.Log}).Run(ctx, projectDir, []string{"bash", script})
}

func defaultPackagingConfig(projectDir string) contracts.PackagingConfig {
	record, _ := project.LoadProjectRecord(projectDir)
	info := record.Project.WailsConfig.Info
	name := fsx.FirstNonEmpty(info.ProductName, filepath.Base(projectDir))
	version := fsx.FirstNonEmpty(info.Version, "0.0.1")
	appName := fsx.SafeName(fsx.FirstNonEmpty(record.Project.TaskVars.AppName, name))
	return contracts.PackagingConfig{
		SchemaVersion: 1,
		Project: contracts.ProjectInfo{
			Name:        name,
			Version:     version,
			BundleID:    info.ProductIdentifier,
			Publisher:   info.CompanyName,
			Copyright:   info.Copyright,
			Description: info.Description,
			Icon:        record.Project.WailsConfig.Icon,
		},
		Build: contracts.BuildSettings{
			Taskfile:   "Taskfile.yml",
			Task:       "builder:release",
			Production: record.Project.TaskVars.Production,
			CGOEnabled: record.Project.TaskVars.CGOEnabled == "1" || strings.EqualFold(record.Project.TaskVars.CGOEnabled, "true"),
			AppName:    appName,
		},
		Assets: []contracts.PackagingAsset{},
		Windows: contracts.WindowsConfig{
			Enabled:               true,
			InnoScript:            "builder/windows/inno.iss",
			DefaultDirName:        `{autopf}\` + name,
			PrivilegesRequired:    "lowest",
			SetupIcon:             "build/windows/icon.ico",
			OutputBaseName:        "${build.appName}-${project.version}-windows-setup",
			CreateDesktopShortcut: true,
		},
		MacOS: contracts.MacOSConfig{
			Enabled:       true,
			AppBundle:     filepath.ToSlash(filepath.Join("bin", name+".app")),
			DMGScript:     "builder/macos/dmg.sh",
			Background:    "builder/macos/background.png",
			OutputName:    "${build.appName}-${project.version}",
			CreateDMGPath: "create-dmg",
			WindowWidth:   640,
			WindowHeight:  420,
			IconSize:      96,
			AppX:          180,
			AppY:          210,
			ApplicationsX: 460,
			ApplicationsY: 210,
		},
		Artifacts: contracts.ArtifactConfig{OutputRoot: "builder/release"},
	}
}

func normalizePlatform(platform contracts.Platform) contracts.Platform {
	switch platform {
	case "", contracts.PlatformAuto:
		if runtime.GOOS == "windows" {
			return contracts.PlatformWindows
		}
		if runtime.GOOS == "darwin" {
			return contracts.PlatformMacOS
		}
		return contracts.Platform(runtime.GOOS)
	case "macos":
		return contracts.PlatformMacOS
	default:
		return platform
	}
}
