package packaging

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

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

type Service struct {
	Log              *runlog.Logger
	DMGBackgroundPNG []byte
	packageMu        sync.Mutex
	packageRunning   bool
}

type ServiceOptions struct {
	DMGBackgroundPNG []byte
}

const innoCompileMaxRetries = 10

var (
	innoCompileRetryDelay = 3 * time.Second
	runISCCCommand        = func(ctx context.Context, projectDir string, command []string, runner runlog.Runner) error {
		return runner.Run(ctx, projectDir, command)
	}
)

var platformOverride contracts.Platform

func SetPlatformOverride(platform contracts.Platform) {
	platformOverride = platform
}

func NewService(log *runlog.Logger) *Service {
	return NewServiceWithOptions(log, ServiceOptions{})
}

func NewServiceWithOptions(log *runlog.Logger, options ServiceOptions) *Service {
	return &Service{Log: log, DMGBackgroundPNG: options.DMGBackgroundPNG}
}

func (s *Service) InitPackaging(projectDir string) (contracts.PackagingConfig, error) {
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
		if err := syncPackagingTaskfile(projectDir, existing); err != nil {
			return contracts.PackagingConfig{}, err
		}
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
	if err := syncPackagingTaskfile(projectDir, cfg); err != nil {
		return contracts.PackagingConfig{}, err
	}
	if err := s.writeTemplates(projectDir, cfg, projectConfig); err != nil {
		return contracts.PackagingConfig{}, err
	}
	return config.LoadPackagingConfig(projectDir)
}

func (s *Service) LoadPackagingConfig(projectDir string) (contracts.PackagingConfig, error) {
	return config.LoadPackagingConfig(projectDir)
}

func (s *Service) SavePackagingConfig(projectDir string, cfg contracts.PackagingConfig) (contracts.PackagingConfig, error) {
	projectDir, err := fsx.NormalizePath(projectDir)
	if err != nil {
		return contracts.PackagingConfig{}, err
	}
	projectConfig, err := loadProjectConfig(projectDir)
	if err != nil {
		return contracts.PackagingConfig{}, err
	}
	if platform := runtimePlatform(); platform == contracts.PlatformWindows || platform == contracts.PlatformMacOS {
		if err := config.ValidateAssetTargets(cfg, platform); err != nil {
			return contracts.PackagingConfig{}, err
		}
	}
	if err := config.SavePackagingConfig(projectDir, cfg); err != nil {
		return contracts.PackagingConfig{}, err
	}
	if err := syncPackagingTaskfile(projectDir, cfg); err != nil {
		return contracts.PackagingConfig{}, err
	}
	if err := s.writeTemplates(projectDir, cfg, projectConfig); err != nil {
		return contracts.PackagingConfig{}, err
	}
	return config.LoadPackagingConfig(projectDir)
}

func (s *Service) GetPackagingRuntimeInfo(projectDir string) (contracts.PackagingRuntimeInfo, error) {
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
		defaultPath := config.DefaultMacOSBinaryPath(cfg)
		effectivePath := macOSLaunchExecutablePath(cfg, projectConfig)
		return contracts.PackagingRuntimeInfo{
			DefaultExecutablePath:   defaultPath,
			EffectiveExecutablePath: effectivePath,
			UsingDefaultExecutable:  config.SameAssetPath(effectivePath, defaultPath, contracts.PlatformMacOS),
		}, nil
	}
	return config.ResolveExecutablePath(cfg, projectConfig, runtimePlatform()), nil
}

func (s *Service) Package(req contracts.PackageRequest) (result contracts.PackageResult, err error) {
	projectDir, err := fsx.NormalizePath(req.ProjectDir)
	if err != nil {
		return contracts.PackageResult{}, err
	}
	transactionID := strings.TrimSpace(req.TransactionID)
	if transactionID == "" {
		return contracts.PackageResult{}, errors.New("package transactionId is required")
	}
	if !s.beginPackage() {
		return contracts.PackageResult{}, errors.New("another package task is already running")
	}
	defer s.endPackage()
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
	if err := syncPackagingTaskfile(projectDir, cfg); err != nil {
		return contracts.PackageResult{}, err
	}
	tx := runlog.Transaction{
		ID:    transactionID,
		Type:  "package",
		Title: "Packaging " + string(platform),
	}
	if s.Log != nil {
		s.Log.PrintlnWithTransaction(tx, "Packaging started:", platform)
	}
	defer func() {
		if s.Log == nil {
			return
		}
		if err != nil {
			s.Log.PrintlnWithTransaction(tx, "Packaging failed:", err)
			return
		}
		s.Log.PrintlnWithTransaction(tx, "Packaging completed:", platform)
	}()
	runID := fsx.Timestamp()
	result = contracts.PackageResult{
		OK:               true,
		RunID:            runID,
		Warnings:         []string{},
		BuildOutputDir:   fsx.Resolve(projectDir, "bin"),
		PackageOutputDir: packageOutputDir(projectDir, cfg, projectConfig, platform),
	}
	ctx := context.Background()
	if req.RunBuild {
		if err := s.runBuild(ctx, projectDir, cfg, req.DryRun, tx); err != nil {
			return contracts.PackageResult{}, err
		}
	}
	// Packaging intentionally has no before/after script stages. The only custom
	// execution path is the explicit build command in packaging.json, which keeps
	// package review and dry-run behavior predictable.
	if platform == contracts.PlatformWindows || platform == contracts.PlatformAll {
		if !cfg.Windows.Enabled {
			result.Warnings = append(result.Warnings, "Windows installer generation is disabled; skipped.")
		} else if cfg.Windows.InnoScript == "" {
			return contracts.PackageResult{}, errors.New("Windows packaging has not been initialized")
		} else {
			if err := config.ValidateAssetTargets(cfg, contracts.PlatformWindows); err != nil {
				return contracts.PackageResult{}, err
			}
			if _, err := inno.Generate(projectDir, cfg, projectConfig); err != nil {
				return contracts.PackageResult{}, err
			}
			if req.DryRun || runtime.GOOS != "windows" {
				result.Warnings = append(result.Warnings, "Windows packaging generated the Inno script only; ISCC was not executed.")
			} else {
				if err := validateWindowsPackagingInputs(projectDir, cfg, projectConfig); err != nil {
					return contracts.PackageResult{}, err
				}
				if err := s.runISCC(ctx, projectDir, cfg, tx); err != nil {
					return contracts.PackageResult{}, err
				}
			}
		}
	}
	if platform == contracts.PlatformMacOS || platform == contracts.PlatformAll {
		if !cfg.MacOS.Enabled {
			result.Warnings = append(result.Warnings, "macOS DMG generation is disabled; skipped.")
		} else if cfg.MacOS.DMGScript == "" {
			return contracts.PackageResult{}, errors.New("macOS packaging has not been initialized")
		} else {
			if err := config.ValidateAssetTargets(cfg, contracts.PlatformMacOS); err != nil {
				return contracts.PackageResult{}, err
			}
			if _, err := dmg.GenerateScript(projectDir, cfg, projectConfig); err != nil {
				return contracts.PackageResult{}, err
			}
			if req.DryRun || runtime.GOOS != "darwin" {
				result.Warnings = append(result.Warnings, "macOS packaging generated the DMG script only; create-dmg was not executed.")
			} else {
				if _, err := prepareMacOSAppBundle(projectDir, cfg, projectConfig); err != nil {
					return contracts.PackageResult{}, err
				}
				if err := validateMacOSPackagingInputs(projectDir, cfg, projectConfig); err != nil {
					return contracts.PackageResult{}, err
				}
				if err := s.runDMG(ctx, projectDir, cfg, tx); err != nil {
					return contracts.PackageResult{}, err
				}
			}
		}
	}
	result.Artifacts = release.ReleaseArtifacts(projectDir)
	result.Message = "Packaging completed."
	return result, nil
}

func (s *Service) Artifacts(projectDir string) []contracts.Artifact {
	projectDir, err := fsx.NormalizePath(projectDir)
	if err != nil {
		return []contracts.Artifact{}
	}
	return release.ReleaseArtifacts(projectDir)
}

func (s *Service) writeTemplates(projectDir string, cfg contracts.PackagingConfig, projectConfig contracts.WailsProjectConfig) error {
	switch runtimePlatform() {
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
			if _, err := fsx.WriteIfMissing(bg, s.defaultDMGBackgroundPNG(), 0644); err != nil {
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

func (s *Service) runBuild(ctx context.Context, projectDir string, cfg contracts.PackagingConfig, dryRun bool, tx runlog.Transaction) error {
	cmd, err := buildCommand(cfg)
	if err != nil {
		return err
	}
	env := buildEnv(cfg)
	return (runlog.Runner{Log: s.Log, DryRun: dryRun, Env: env, Transaction: tx}).Run(ctx, projectDir, cmd)
}

func (s *Service) runISCC(ctx context.Context, projectDir string, cfg contracts.PackagingConfig, tx runlog.Transaction) error {
	req := environment.InnoRequirement(cfg.Windows.ISCCPath)
	if !req.Found {
		return errors.New(req.Message)
	}
	script := fsx.Resolve(projectDir, cfg.Windows.InnoScript)
	command := []string{req.Path, script}
	runner := runlog.Runner{Log: s.Log, Transaction: tx}

	var lastErr error
	for retry := 0; retry <= innoCompileMaxRetries; retry++ {
		if retry > 0 && s.Log != nil {
			s.Log.PrintlnWithTransaction(tx, fmt.Sprintf("Retrying Inno Setup compile (%d/%d).", retry, innoCompileMaxRetries))
		}
		if err := runISCCCommand(ctx, projectDir, command, runner); err != nil {
			lastErr = err
			if retry == innoCompileMaxRetries {
				return fmt.Errorf("Inno Setup compile failed after %d retries: %w", innoCompileMaxRetries, lastErr)
			}
			if s.Log != nil {
				s.Log.PrintlnWithTransaction(tx, fmt.Sprintf("Inno Setup compile failed (attempt %d/%d): %v", retry+1, innoCompileMaxRetries+1, err))
				s.Log.PrintlnWithTransaction(tx, fmt.Sprintf("Retrying Inno Setup compile in %s.", innoCompileRetryDelay))
			}
			if err := waitForInnoRetry(ctx, innoCompileRetryDelay); err != nil {
				return fmt.Errorf("Inno Setup compile failed: %v; retry wait interrupted: %w", lastErr, err)
			}
			continue
		}
		return nil
	}
	return lastErr
}

func waitForInnoRetry(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (s *Service) runDMG(ctx context.Context, projectDir string, cfg contracts.PackagingConfig, tx runlog.Transaction) error {
	req := environment.CreateDMGRequirement(cfg.MacOS.CreateDMGPath)
	if !req.Found {
		return errors.New(req.Message)
	}
	script := fsx.Resolve(projectDir, cfg.MacOS.DMGScript)
	return (runlog.Runner{Log: s.Log, Transaction: tx}).Run(ctx, projectDir, []string{"bash", script})
}

func defaultPackagingConfig(projectDir string, projectConfig contracts.WailsProjectConfig) contracts.PackagingConfig {
	info := projectConfig.Info
	name := strings.TrimSpace(info.ProductName)
	if name == "" {
		name = filepath.Base(projectDir)
	}
	taskVars := project.LoadTaskVars(projectDir)
	taskfile := project.RootTaskfileRelPathForProject(projectDir)
	appNameSource := strings.TrimSpace(taskVars.AppName)
	if appNameSource == "" {
		appNameSource = name
	}
	appName := fsx.SafeName(appNameSource)
	return contracts.PackagingConfig{
		SchemaVersion: 1,
		Build: contracts.BuildSettings{
			Taskfile:   taskfile,
			Task:       "release",
			Production: taskVars.Production,
			CGOEnabled: project.ResolveBoolean(taskVars.CGOEnabled),
			AppName:    appName,
		},
		Assets:    []contracts.PackagingAsset{},
		Artifacts: contracts.ArtifactConfig{OutputRoot: "builder/release"},
		Windows: contracts.WindowsConfig{
			Enabled:               true,
			InnoScript:            "builder/windows/inno.iss",
			DefaultDirName:        `{autopf}\${project.name}`,
			PrivilegesRequired:    "lowest",
			SetupIcon:             "build/windows/icon.ico",
			OutputBaseName:        "${project.name}-${project.version}-windows-setup",
			CreateDesktopShortcut: true,
		},
		MacOS: contracts.MacOSConfig{
			Enabled:       true,
			AppBundle:     "bin/${project.name}.app",
			DMGScript:     "builder/macos/dmg.sh",
			Background:    "assets/install-grid.png",
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
	}
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

func buildCommand(cfg contracts.PackagingConfig) ([]string, error) {
	if len(cfg.Build.Command) > 0 {
		return cfg.Build.Command, nil
	}
	if cfg.Build.Task == "" {
		return nil, errors.New("build.task is required when build.command is empty")
	}
	cmd := []string{"wails3", "task"}
	if taskfile := strings.TrimSpace(cfg.Build.Taskfile); taskfile != "" {
		cmd = append(cmd, "-taskfile", filepath.ToSlash(taskfile))
	}
	cmd = append(cmd, cfg.Build.Task)
	cmd = append(cmd, buildTaskVars(cfg)...)
	return cmd, nil
}

func buildTaskVars(cfg contracts.PackagingConfig) []string {
	return []string{
		"APP_NAME=" + config.AppName(cfg),
		"PRODUCTION=" + boolString(cfg.Build.Production),
		"CGO_ENABLED=" + cgoString(cfg.Build.CGOEnabled),
	}
}

func validateWindowsPackagingInputs(projectDir string, cfg contracts.PackagingConfig, projectConfig contracts.WailsProjectConfig) error {
	return validateRequiredAssets(projectDir, cfg, projectConfig, contracts.PlatformWindows)
}

func validateMacOSPackagingInputs(projectDir string, cfg contracts.PackagingConfig, projectConfig contracts.WailsProjectConfig) error {
	return validateRequiredAssets(projectDir, cfg, projectConfig, contracts.PlatformMacOS)
}

func validateRequiredAssets(projectDir string, cfg contracts.PackagingConfig, projectConfig contracts.WailsProjectConfig, platform contracts.Platform) error {
	wailsBuildOutput := config.WailsBuildOutputPath(cfg, projectConfig, platform)
	startupProgram := config.StartupExecutablePath(cfg, projectConfig, platform)
	for i, asset := range config.EffectiveAssets(cfg, projectConfig, platform) {
		if !asset.Required {
			continue
		}
		label := platformRequiredAssetLabel(platform, asset.Src, wailsBuildOutput, startupProgram, i)
		if err := validateRequiredPath(projectDir, asset.Src, asset.Type == "directory", label, "packaging assets", cfg); err != nil {
			return err
		}
	}
	return nil
}

func platformRequiredAssetLabel(platform contracts.Platform, src, wailsBuildOutput, startupProgram string, index int) string {
	platformName := "Windows"
	if platform == contracts.PlatformMacOS {
		platformName = "macOS"
	}
	if config.SameAssetPath(src, wailsBuildOutput, platform) {
		return platformName + " Wails build output"
	}
	if config.SameAssetPath(src, startupProgram, platform) {
		return platformName + " launch program"
	}
	return fmt.Sprintf("required packaging asset %d", index)
}

func validateRequiredPath(projectDir, relOrAbsPath string, wantDir bool, label, source string, cfg contracts.PackagingConfig) error {
	path := fsx.Resolve(projectDir, relOrAbsPath)
	if path == "" {
		return fmt.Errorf("%s path is empty (%s)", label, source)
	}
	exists := fsx.FileExists(path)
	if wantDir {
		exists = fsx.DirExists(path)
	}
	if exists {
		return nil
	}
	kind := "file"
	if wantDir {
		kind = "directory"
	}
	return fmt.Errorf("%s %s does not exist: %s (source: %s, build.appName=%q, build.task=%q, build.taskfile=%q)", label, kind, path, source, config.AppName(cfg), cfg.Build.Task, cfg.Build.Taskfile)
}

func packageOutputDir(projectDir string, cfg contracts.PackagingConfig, projectConfig contracts.WailsProjectConfig, platform contracts.Platform) string {
	switch platform {
	case contracts.PlatformWindows:
		return fsx.Resolve(projectDir, config.WindowsOutputDir(cfg, projectConfig))
	case contracts.PlatformMacOS:
		return fsx.Resolve(projectDir, config.MacOSOutputDir(cfg, projectConfig))
	case contracts.PlatformAll:
		return fsx.Resolve(projectDir, config.ExportRoot(cfg, projectConfig))
	}
	return ""
}

func (s *Service) defaultDMGBackgroundPNG() []byte {
	if len(s.DMGBackgroundPNG) > 0 {
		return s.DMGBackgroundPNG
	}
	return dmg.DefaultBackgroundPNG()
}

func (s *Service) beginPackage() bool {
	s.packageMu.Lock()
	defer s.packageMu.Unlock()
	if s.packageRunning {
		return false
	}
	// ponytail: global package lock, switch to per-project locks if multi-project builds matter.
	s.packageRunning = true
	return true
}

func (s *Service) endPackage() {
	s.packageMu.Lock()
	defer s.packageMu.Unlock()
	s.packageRunning = false
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
	if platformOverride != "" {
		return platformOverride
	}
	switch runtime.GOOS {
	case "windows":
		return contracts.PlatformWindows
	case "darwin":
		return contracts.PlatformMacOS
	default:
		return contracts.Platform(runtime.GOOS)
	}
}
