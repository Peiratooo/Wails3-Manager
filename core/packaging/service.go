package packaging

import (
	"context"
	"errors"
	"fmt"
	"os"
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
	projectConfig, err := loadProjectConfig(projectDir)
	if err != nil {
		return contracts.PackagingConfig{}, err
	}
	existing, err := config.LoadPackagingConfig(projectDir)
	if err == nil {
		if err := s.writeTemplates(projectDir, existing, projectConfig); err != nil {
			return contracts.PackagingConfig{}, err
		}
		return existing, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return contracts.PackagingConfig{}, err
	}
	cfg := defaultPackagingConfig(projectDir, projectConfig)
	if err := config.SavePackagingConfig(projectDir, cfg); err != nil {
		return contracts.PackagingConfig{}, err
	}
	if err := s.writeTemplates(projectDir, cfg, projectConfig); err != nil {
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
	projectConfig, err := loadProjectConfig(projectDir)
	if err != nil {
		return contracts.PackagingConfig{}, err
	}
	if err := config.SavePackagingConfig(projectDir, cfg); err != nil {
		return contracts.PackagingConfig{}, err
	}
	if err := s.writeTemplates(projectDir, cfg, projectConfig); err != nil {
		return contracts.PackagingConfig{}, err
	}
	return config.LoadPackagingConfig(projectDir)
}

func (s *PackagingService) GetPackagingRuntimeInfo(projectDir string) (contracts.PackagingRuntimeInfo, error) {
	projectDir, err := fsx.NormalizePath(projectDir)
	if err != nil {
		return contracts.PackagingRuntimeInfo{}, err
	}
	cfg, err := config.LoadPackagingConfig(projectDir)
	if err != nil {
		return contracts.PackagingRuntimeInfo{}, err
	}
	projectConfig, err := loadProjectConfig(projectDir)
	if err != nil {
		return contracts.PackagingRuntimeInfo{}, err
	}
	if runtimePlatform() == contracts.PlatformMacOS {
		return config.ResolveMacOSAppBundlePath(cfg, projectConfig), nil
	}
	return config.ResolveExecutablePath(cfg, projectConfig, runtimePlatform()), nil
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
	projectConfig, err := loadProjectConfig(projectDir)
	if err != nil {
		return contracts.PackageResult{}, err
	}
	platform := req.Platform
	if platform == "" || platform == contracts.PlatformAuto {
		platform = runtimePlatform()
	}
	switch platform {
	case contracts.PlatformWindows, contracts.PlatformMacOS, contracts.PlatformAll:
	default:
		return contracts.PackageResult{}, fmt.Errorf("unsupported packaging platform: %s", platform)
	}
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
		if !cfg.Windows.Enabled || cfg.Windows.InnoScript == "" {
			if platform == contracts.PlatformWindows {
				return contracts.PackageResult{}, errors.New("Windows packaging has not been initialized")
			}
			result.Warnings = append(result.Warnings, "Windows packaging has not been initialized; skipped.")
		} else {
			if _, err := inno.Generate(projectDir, cfg, projectConfig); err != nil {
				return contracts.PackageResult{}, err
			}
			if req.DryRun || runtime.GOOS != "windows" {
				result.Warnings = append(result.Warnings, "Windows packaging generated the Inno script only; ISCC was not executed.")
			} else if err := s.runISCC(ctx, projectDir, cfg); err != nil {
				return contracts.PackageResult{}, err
			}
		}
	}
	if platform == contracts.PlatformMacOS || platform == contracts.PlatformAll {
		if !cfg.MacOS.Enabled || cfg.MacOS.DMGScript == "" {
			if platform == contracts.PlatformMacOS {
				return contracts.PackageResult{}, errors.New("macOS packaging has not been initialized")
			}
			result.Warnings = append(result.Warnings, "macOS packaging has not been initialized; skipped.")
		} else {
			if _, err := dmg.GenerateScript(projectDir, cfg, projectConfig); err != nil {
				return contracts.PackageResult{}, err
			}
			if req.DryRun || runtime.GOOS != "darwin" {
				result.Warnings = append(result.Warnings, "macOS packaging generated the DMG script only; create-dmg was not executed.")
			} else if err := s.runDMG(ctx, projectDir, cfg); err != nil {
				return contracts.PackageResult{}, err
			}
		}
	}
	result.Artifacts = release.ReleaseArtifacts(projectDir)
	result.Message = "Packaging completed."
	return result, nil
}

func (s *PackagingService) Artifacts(projectDir string) []contracts.Artifact {
	projectDir, err := fsx.NormalizePath(projectDir)
	if err != nil {
		return []contracts.Artifact{}
	}
	return release.ReleaseArtifacts(projectDir)
}

func (s *PackagingService) writeTemplates(projectDir string, cfg contracts.PackagingConfig, projectConfig contracts.WailsProjectConfig) error {
	switch runtime.GOOS {
	case "windows":
		if !cfg.Windows.Enabled || cfg.Windows.InnoScript == "" {
			return nil
		}
		if _, err := inno.Generate(projectDir, cfg, projectConfig); err != nil {
			return err
		}
	case "darwin":
		if !cfg.MacOS.Enabled {
			return nil
		}
		if cfg.MacOS.Background != "" {
			bg := fsx.Resolve(projectDir, cfg.MacOS.Background)
			if _, err := fsx.WriteIfMissing(bg, dmg.DefaultBackgroundPNG(), 0644); err != nil {
				return err
			}
		}
		if cfg.MacOS.DMGScript == "" {
			return nil
		}
		if _, err := dmg.GenerateScript(projectDir, cfg, projectConfig); err != nil {
			return err
		}
	}
	return nil
}

func (s *PackagingService) runBuild(ctx context.Context, projectDir string, cfg contracts.PackagingConfig, dryRun bool) error {
	cmd := cfg.Build.Command
	env := buildEnv(cfg)
	if len(cmd) == 0 {
		if cfg.Build.Task == "" {
			return errors.New("build.task is required when build.command is empty")
		}
		cmd = []string{"wails3", "task", cfg.Build.Task}
		cmd = append(cmd, buildTaskVars(cfg)...)
	}
	return (runlog.Runner{Log: s.Log, DryRun: dryRun, Env: env}).Run(ctx, projectDir, cmd)
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

func defaultPackagingConfig(projectDir string, projectConfig contracts.WailsProjectConfig) contracts.PackagingConfig {
	info := projectConfig.Info
	name := strings.TrimSpace(info.ProductName)
	if name == "" {
		name = filepath.Base(projectDir)
	}
	taskVars := project.LoadTaskVars(projectDir)
	appNameSource := strings.TrimSpace(taskVars.AppName)
	if appNameSource == "" {
		appNameSource = name
	}
	appName := fsx.SafeName(appNameSource)
	cfg := contracts.PackagingConfig{
		SchemaVersion: 1,
		Build: contracts.BuildSettings{
			Taskfile:   "Taskfile.yml",
			Task:       "builder:release",
			Production: taskVars.Production,
			CGOEnabled: project.ResolveBoolean(taskVars.CGOEnabled),
			AppName:    appName,
		},
		Assets:    []contracts.PackagingAsset{},
		Artifacts: contracts.ArtifactConfig{OutputRoot: "builder/release"},
	}
	switch runtime.GOOS {
	case "windows":
		cfg.Windows = contracts.WindowsConfig{
			Enabled:               true,
			InnoScript:            "builder/windows/inno.iss",
			DefaultDirName:        `{autopf}\${project.name}`,
			PrivilegesRequired:    "lowest",
			SetupIcon:             "build/windows/icon.ico",
			OutputBaseName:        "${build.appName}-${project.version}-windows-setup",
			CreateDesktopShortcut: true,
		}
	case "darwin":
		cfg.MacOS = contracts.MacOSConfig{
			Enabled:       true,
			AppBundle:     "bin/${build.appName}.app",
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
		}
	}
	return cfg
}

func loadProjectConfig(projectDir string) (contracts.WailsProjectConfig, error) {
	record, ok := project.LoadProjectRecord(projectDir)
	if !ok {
		return contracts.WailsProjectConfig{}, errors.New("project is not imported; builder/project.json is missing")
	}
	return record.Project.WailsConfig, nil
}

func buildEnv(cfg contracts.PackagingConfig) map[string]string {
	return map[string]string{
		"APP_NAME":    config.AppName(cfg),
		"PRODUCTION":  boolString(cfg.Build.Production),
		"CGO_ENABLED": cgoString(cfg.Build.CGOEnabled),
	}
}

func buildTaskVars(cfg contracts.PackagingConfig) []string {
	return []string{
		"APP_NAME=" + config.AppName(cfg),
		"PRODUCTION=" + boolString(cfg.Build.Production),
		"CGO_ENABLED=" + cgoString(cfg.Build.CGOEnabled),
	}
}

func boolString(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func cgoString(v bool) string {
	if v {
		return "1"
	}
	return "0"
}

func runtimePlatform() contracts.Platform {
	switch runtime.GOOS {
	case "windows":
		return contracts.PlatformWindows
	case "darwin":
		return contracts.PlatformMacOS
	default:
		return contracts.Platform(runtime.GOOS)
	}
}
