package packaging

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"wails3-manager/core/contracts"
	packagingConfig "wails3-manager/core/packaging/config"
	"wails3-manager/core/project"
	"wails3-manager/core/runlog"
)

func TestBuildSettingsBecomeTaskVarsAndEnv(t *testing.T) {
	cfg := contracts.PackagingConfig{
		Build: contracts.BuildSettings{
			Taskfile:   "Taskfile.yml",
			Task:       "release",
			AppName:    "demo",
			Production: true,
			CGOEnabled: true,
		},
	}

	if got, want := buildTaskVars(cfg), []string{"APP_NAME=demo", "PRODUCTION=true", "CGO_ENABLED=1"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("buildTaskVars() = %#v, want %#v", got, want)
	}

	env := buildEnv(cfg)
	for k, want := range map[string]string{
		"APP_NAME":    "demo",
		"PRODUCTION":  "true",
		"CGO_ENABLED": "1",
	} {
		if env[k] != want {
			t.Fatalf("env[%s] = %q, want %q", k, env[k], want)
		}
	}

	cmd, err := buildCommand(cfg)
	if err != nil {
		t.Fatal(err)
	}
	wantCmd := []string{"wails3", "task", "-taskfile", "Taskfile.yml", "release", "APP_NAME=demo", "PRODUCTION=true", "CGO_ENABLED=1"}
	if !reflect.DeepEqual(cmd, wantCmd) {
		t.Fatalf("buildCommand() = %#v, want %#v", cmd, wantCmd)
	}
}

func TestDefaultPackagingConfigInitializesBothPlatforms(t *testing.T) {
	cfg := defaultPackagingConfig(t.TempDir(), contracts.WailsProjectConfig{
		Info: contracts.WailsAppInfo{ProductName: "Demo"},
	})

	if !cfg.Windows.Enabled || cfg.Windows.InnoScript == "" {
		t.Fatalf("windows config was not initialized: %#v", cfg.Windows)
	}
	if !cfg.MacOS.Enabled || cfg.MacOS.AppBundle == "" || cfg.MacOS.Background == "" {
		t.Fatalf("macos config was not initialized: %#v", cfg.MacOS)
	}
}

func TestRunDMGUsesGoLibrary(t *testing.T) {
	originalBuildDMG := buildDMG
	defer func() { buildDMG = originalBuildDMG }()

	projectDir := t.TempDir()
	cfg := contracts.PackagingConfig{MacOS: contracts.MacOSConfig{Enabled: true}}
	projectConfig := contracts.WailsProjectConfig{
		Info: contracts.WailsAppInfo{ProductName: "Demo", Version: "1.0.0"},
	}
	appBundle := filepath.Join(projectDir, "bin", "Demo.app")
	background := filepath.Join(projectDir, "builder", "macos", "dmg-background.png")
	wantOutput := filepath.Join(projectDir, "builder", "release", "darwin", "Demo.dmg")

	called := false
	buildDMG = func(gotProjectDir string, gotCfg contracts.PackagingConfig, gotProjectConfig contracts.WailsProjectConfig, gotAppBundle, gotBackground string) (string, error) {
		called = true
		if gotProjectDir != projectDir || !reflect.DeepEqual(gotCfg, cfg) || !reflect.DeepEqual(gotProjectConfig, projectConfig) {
			t.Fatalf("unexpected DMG inputs: %q %#v %#v", gotProjectDir, gotCfg, gotProjectConfig)
		}
		if gotAppBundle != appBundle || gotBackground != background {
			t.Fatalf("DMG paths = %q, %q", gotAppBundle, gotBackground)
		}
		return wantOutput, nil
	}

	got, err := NewService(nil).runDMG(projectDir, cfg, projectConfig, appBundle, background, runlog.Transaction{})
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("DMG builder was not called")
	}
	if got != wantOutput {
		t.Fatalf("runDMG() = %q, want %q", got, wantOutput)
	}
}

func TestInitPackagingSyncsRootTaskfile(t *testing.T) {
	SetPlatformOverride(contracts.Platform("linux"))
	defer SetPlatformOverride("")

	projectDir := t.TempDir()
	wailsConfig := contracts.WailsProjectConfig{
		Info: contracts.WailsAppInfo{
			ProductName:       "Demo",
			ProductIdentifier: "com.example.demo",
			Version:           "1.0.0",
		},
	}
	if err := project.SaveProjectRecord(contracts.ProjectRecord{
		ProjectDir: projectDir,
		Project: contracts.WailsProjectManager{
			ProjectDir:  projectDir,
			WailsConfig: wailsConfig,
		},
	}); err != nil {
		t.Fatal(err)
	}
	taskfile := `version: '3'

vars:
  APP_NAME: "task-app"
  BIN_DIR: "bin"
  CGO_ENABLED: "1"
  PRODUCTION: "false"

includes:
  common: ./build/Taskfile.yml
  windows: ./build/windows/Taskfile.yml

tasks:
  build:
    summary: Builds the application
    cmds:
      - task: "{{OS}}:build"
        vars:
          CGO_ENABLED: "{{.CGO_ENABLED}}"
          PRODUCTION: "{{.PRODUCTION}}"
`
	if err := os.WriteFile(filepath.Join(projectDir, "Taskfile.yml"), []byte(taskfile), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := NewService(nil).InitPackaging(projectDir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Build.AppName != "task-app" {
		t.Fatalf("Build.AppName = %q, want task-app", cfg.Build.AppName)
	}
	if cfg.Build.Task != "release" {
		t.Fatalf("Build.Task = %q, want release", cfg.Build.Task)
	}

	data, err := os.ReadFile(filepath.Join(projectDir, "Taskfile.yml"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range []string{
		`APP_NAME: "task-app"`,
		`CGO_ENABLED: "1"`,
		`PRODUCTION: "false"`,
		`  release:`,
		`    summary: Builds the release application`,
		`          APP_NAME: "{{.APP_NAME}}"`,
		`          CGO_ENABLED: "{{.CGO_ENABLED}}"`,
		`          PRODUCTION: "{{.PRODUCTION}}"`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("synced Taskfile is missing %q:\n%s", want, text)
		}
	}
}

func TestSavePackagingConfigSyncsTaskfileVarsAndReleaseTask(t *testing.T) {
	SetPlatformOverride(contracts.Platform("linux"))
	defer SetPlatformOverride("")

	projectDir := t.TempDir()
	wailsConfig := contracts.WailsProjectConfig{
		Info: contracts.WailsAppInfo{
			ProductName:       "Demo",
			ProductIdentifier: "com.example.demo",
			Version:           "1.0.0",
		},
	}
	if err := project.SaveProjectRecord(contracts.ProjectRecord{
		ProjectDir: projectDir,
		Project: contracts.WailsProjectManager{
			ProjectDir:  projectDir,
			WailsConfig: wailsConfig,
		},
	}); err != nil {
		t.Fatal(err)
	}
	taskfile := `version: '3'

vars:
  APP_NAME: "old"
  CGO_ENABLED: "0"
  PRODUCTION: "false"

tasks:
  build:
    summary: Builds the application
    cmds:
      - task: "{{OS}}:build"

  release:
    summary: Builds the release application
    cmds:
      - task: "{{OS}}:build"
        vars:
          CGO_ENABLED: "{{.CGO_ENABLED}}"
          PRODUCTION: "true"
`
	if err := os.WriteFile(filepath.Join(projectDir, "Taskfile.yml"), []byte(taskfile), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := defaultPackagingConfig(projectDir, wailsConfig)
	cfg.Build.AppName = "next-app"
	cfg.Build.Production = true
	cfg.Build.CGOEnabled = true
	saved, err := NewService(nil).SavePackagingConfig(projectDir, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Build.Task != "release" {
		t.Fatalf("Build.Task = %q, want release", saved.Build.Task)
	}
	if _, err := NewService(nil).SavePackagingConfig(projectDir, saved); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(projectDir, "Taskfile.yml"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range []string{
		`APP_NAME: "next-app"`,
		`CGO_ENABLED: "1"`,
		`PRODUCTION: "true"`,
		`          APP_NAME: "{{.APP_NAME}}"`,
		`          CGO_ENABLED: "{{.CGO_ENABLED}}"`,
		`          PRODUCTION: "{{.PRODUCTION}}"`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("synced Taskfile is missing %q:\n%s", want, text)
		}
	}
	if got := strings.Count(text, "  release:"); got != 1 {
		t.Fatalf("release task count = %d, want 1:\n%s", got, text)
	}
	if got := strings.Count(text, `          APP_NAME: "{{.APP_NAME}}"`); got < 2 {
		t.Fatalf("APP_NAME should be passed by build and release tasks:\n%s", text)
	}
}

func TestSavePackagingConfigRejectsInvalidAssetTargetForCurrentPlatform(t *testing.T) {
	SetPlatformOverride(contracts.PlatformWindows)
	defer SetPlatformOverride("")

	projectDir := t.TempDir()
	wailsConfig := contracts.WailsProjectConfig{
		Info: contracts.WailsAppInfo{
			ProductName:       "Demo",
			ProductIdentifier: "com.example.demo",
			Version:           "1.0.0",
		},
	}
	if err := project.SaveProjectRecord(contracts.ProjectRecord{
		ProjectDir: projectDir,
		Project: contracts.WailsProjectManager{
			ProjectDir:  projectDir,
			WailsConfig: wailsConfig,
		},
	}); err != nil {
		t.Fatal(err)
	}

	cfg := defaultPackagingConfig(projectDir, wailsConfig)
	cfg.Assets = []contracts.PackagingAsset{
		{Src: "README.md", Type: "file", Required: false, Target: "docs"},
	}

	if _, err := NewService(nil).SavePackagingConfig(projectDir, cfg); err == nil {
		t.Fatal("expected invalid current-platform asset target to be rejected")
	}
}

func TestGetPackagingRuntimeInfoUsesDefaultExecutableWhenEntryIsEmpty(t *testing.T) {
	projectDir := t.TempDir()
	wailsConfig := contracts.WailsProjectConfig{
		Info: contracts.WailsAppInfo{
			ProductName: "Demo",
			Version:     "1.0.0",
		},
	}
	if err := project.SaveProjectRecord(contracts.ProjectRecord{
		ProjectDir: projectDir,
		Project: contracts.WailsProjectManager{
			ProjectDir:  projectDir,
			WailsConfig: wailsConfig,
		},
	}); err != nil {
		t.Fatal(err)
	}

	cfg := contracts.PackagingConfig{
		SchemaVersion: 1,
		Build: contracts.BuildSettings{
			Taskfile: "Taskfile.yml",
			Task:     "release",
			AppName:  "demo",
		},
		Windows: contracts.WindowsConfig{
			Enabled:               true,
			InnoScript:            "builder/windows/inno.iss",
			DefaultDirName:        `{autopf}\${project.name}`,
			PrivilegesRequired:    "lowest",
			SetupIcon:             "build/windows/icon.ico",
			OutputBaseName:        packagingConfig.InstallerOutputNameTemplate,
			CreateDesktopShortcut: true,
		},
		MacOS: contracts.MacOSConfig{
			Enabled:       true,
			AppBundle:     "bin/${build.appName}.app",
			Background:    "assets/install-grid.png",
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
	if err := packagingConfig.SavePackagingConfig(projectDir, cfg); err != nil {
		t.Fatal(err)
	}

	info, err := NewService(nil).GetPackagingRuntimeInfo(projectDir)
	if err != nil {
		t.Fatal(err)
	}

	wantDefault := packagingConfig.DefaultExecutablePath(cfg, runtimePlatform())
	if runtimePlatform() == contracts.PlatformMacOS {
		wantDefault = packagingConfig.DefaultMacOSBinaryPath(cfg)
	}
	if !info.UsingDefaultExecutable {
		t.Fatalf("UsingDefaultExecutable = false, want true")
	}
	if info.DefaultExecutablePath != wantDefault || info.EffectiveExecutablePath != wantDefault {
		t.Fatalf("runtime info = %#v, want default/effective %q", info, wantDefault)
	}
}

func TestGetPackagingRuntimeInfoUsesConfiguredExecutableWithPlaceholders(t *testing.T) {
	projectDir := t.TempDir()
	wailsConfig := contracts.WailsProjectConfig{
		Info: contracts.WailsAppInfo{
			ProductName: "Demo",
			Version:     "1.0.0",
		},
	}
	if err := project.SaveProjectRecord(contracts.ProjectRecord{
		ProjectDir: projectDir,
		Project: contracts.WailsProjectManager{
			ProjectDir:  projectDir,
			WailsConfig: wailsConfig,
		},
	}); err != nil {
		t.Fatal(err)
	}

	cfg := contracts.PackagingConfig{
		SchemaVersion: 1,
		Build: contracts.BuildSettings{
			Taskfile: "Taskfile.yml",
			Task:     "release",
			AppName:  "demo",
		},
		Entry: contracts.ProgramEntry{
			ExecutablePath: "dist/${build.appName}",
		},
		Windows: contracts.WindowsConfig{
			Enabled:               true,
			InnoScript:            "builder/windows/inno.iss",
			DefaultDirName:        `{autopf}\${project.name}`,
			PrivilegesRequired:    "lowest",
			SetupIcon:             "build/windows/icon.ico",
			OutputBaseName:        packagingConfig.InstallerOutputNameTemplate,
			CreateDesktopShortcut: true,
		},
		MacOS: contracts.MacOSConfig{
			Enabled:       true,
			AppBundle:     "bin/${build.appName}.app",
			Background:    "assets/install-grid.png",
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
	if err := packagingConfig.SavePackagingConfig(projectDir, cfg); err != nil {
		t.Fatal(err)
	}

	info, err := NewService(nil).GetPackagingRuntimeInfo(projectDir)
	if err != nil {
		t.Fatal(err)
	}
	if info.UsingDefaultExecutable {
		t.Fatalf("UsingDefaultExecutable = true, want false")
	}
	if info.DefaultExecutablePath == "" {
		t.Fatalf("DefaultExecutablePath is empty")
	}
	if info.EffectiveExecutablePath != "dist/demo" {
		t.Fatalf("EffectiveExecutablePath = %q, want %q", info.EffectiveExecutablePath, "dist/demo")
	}
}

func writeManagedTaskfile(t *testing.T, projectDir string) {
	t.Helper()
	taskfile := `version: '3'

vars:
  APP_NAME: "demo"
  CGO_ENABLED: "0"
  PRODUCTION: "false"

tasks:
  build:
    summary: Builds the application
    cmds:
      - task: "{{OS}}:build"

  release:
    summary: Builds the release application
    cmds:
      - task: "{{OS}}:build"
`
	if err := os.WriteFile(filepath.Join(projectDir, "Taskfile.yml"), []byte(taskfile), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestPackageRejectsUnsupportedPlatform(t *testing.T) {
	projectDir := t.TempDir()
	wailsConfig := contracts.WailsProjectConfig{
		Info: contracts.WailsAppInfo{
			ProductName:       "Demo",
			ProductIdentifier: "com.example.demo",
			Version:           "1.0.0",
		},
	}
	if err := project.SaveProjectRecord(contracts.ProjectRecord{
		ProjectDir: projectDir,
		Project: contracts.WailsProjectManager{
			ProjectDir:  projectDir,
			WailsConfig: wailsConfig,
		},
	}); err != nil {
		t.Fatal(err)
	}
	writeManagedTaskfile(t, projectDir)

	cfg := contracts.PackagingConfig{
		SchemaVersion: 1,
		Build: contracts.BuildSettings{
			Taskfile: "Taskfile.yml",
			Task:     "release",
			AppName:  "demo",
		},
		Windows: contracts.WindowsConfig{
			Enabled:               true,
			InnoScript:            "builder/windows/inno.iss",
			DefaultDirName:        `{autopf}\${project.name}`,
			PrivilegesRequired:    "lowest",
			SetupIcon:             "build/windows/icon.ico",
			OutputBaseName:        packagingConfig.InstallerOutputNameTemplate,
			CreateDesktopShortcut: true,
		},
		MacOS: contracts.MacOSConfig{
			Enabled:       true,
			AppBundle:     "bin/${build.appName}.app",
			Background:    "assets/install-grid.png",
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
	if err := packagingConfig.SavePackagingConfig(projectDir, cfg); err != nil {
		t.Fatal(err)
	}

	_, err := NewService(nil).Package(contracts.PackageRequest{
		ProjectDir:    projectDir,
		Platform:      contracts.Platform("linux"),
		TransactionID: "package-test",
	})
	if err == nil || !strings.Contains(err.Error(), "unsupported packaging platform") {
		t.Fatalf("Package() error = %v, want unsupported platform", err)
	}
}

func TestPackageRejectsMissingTransactionID(t *testing.T) {
	_, err := NewService(nil).Package(contracts.PackageRequest{
		ProjectDir: t.TempDir(),
		Platform:   contracts.PlatformWindows,
	})
	if err == nil || !strings.Contains(err.Error(), "transactionId") {
		t.Fatalf("Package() error = %v, want missing transactionId", err)
	}
}

func TestPackageSkipsDisabledPlatformInstaller(t *testing.T) {
	projectDir := t.TempDir()
	projectConfig := contracts.WailsProjectConfig{
		Info: contracts.WailsAppInfo{
			ProductName:       "Demo",
			ProductIdentifier: "com.example.demo",
			Version:           "1.0.0",
		},
	}
	if err := project.SaveProjectRecord(contracts.ProjectRecord{
		ProjectDir: projectDir,
		Project: contracts.WailsProjectManager{
			ProjectDir:  projectDir,
			WailsConfig: projectConfig,
		},
	}); err != nil {
		t.Fatal(err)
	}
	writeManagedTaskfile(t, projectDir)

	cfg := defaultPackagingConfig(projectDir, projectConfig)
	cfg.Windows.Enabled = false
	if err := packagingConfig.SavePackagingConfig(projectDir, cfg); err != nil {
		t.Fatal(err)
	}

	result, err := NewService(nil).Package(contracts.PackageRequest{
		ProjectDir:    projectDir,
		Platform:      contracts.PlatformWindows,
		DryRun:        true,
		TransactionID: "package-test",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Warnings) != 1 || !strings.Contains(result.Warnings[0], "disabled") {
		t.Fatalf("warnings = %#v, want disabled installer warning", result.Warnings)
	}
}

func TestValidateWindowsPackagingInputsRejectsMissingExecutable(t *testing.T) {
	projectDir := t.TempDir()
	projectConfig := contracts.WailsProjectConfig{
		Info: contracts.WailsAppInfo{
			ProductName:       "Demo",
			ProductIdentifier: "com.example.demo",
			Version:           "1.0.0",
		},
	}
	cfg := defaultPackagingConfig(projectDir, projectConfig)
	cfg.Build.AppName = "missing-app"

	err := validateWindowsPackagingInputs(projectDir, cfg, projectConfig)
	if err == nil {
		t.Fatal("expected missing executable error")
	}
	for _, want := range []string{"Windows Wails build output", "missing-app.exe", `build.appName="missing-app"`} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error %q missing %q", err.Error(), want)
		}
	}
}

func TestRunISCCRetriesUntilSuccess(t *testing.T) {
	oldRunISCCCommand := runISCCCommand
	oldRetryDelay := innoCompileRetryDelay
	defer func() {
		runISCCCommand = oldRunISCCCommand
		innoCompileRetryDelay = oldRetryDelay
	}()

	innoCompileRetryDelay = 0
	attempts := 0
	runISCCCommand = func(ctx context.Context, projectDir string, command []string, runner runlog.Runner) error {
		attempts++
		if attempts < 4 {
			return errors.New("file is locked")
		}
		return nil
	}

	projectDir := t.TempDir()
	cfg := packagingConfigWithFakeISCC(t, projectDir)

	if err := NewService(nil).runISCC(context.Background(), projectDir, cfg, runlog.Transaction{}); err != nil {
		t.Fatal(err)
	}
	if attempts != 4 {
		t.Fatalf("ISCC attempts = %d, want 4", attempts)
	}
}

func TestRunISCCRetriesTenTimesBeforeFailing(t *testing.T) {
	oldRunISCCCommand := runISCCCommand
	oldRetryDelay := innoCompileRetryDelay
	defer func() {
		runISCCCommand = oldRunISCCCommand
		innoCompileRetryDelay = oldRetryDelay
	}()

	innoCompileRetryDelay = 0
	attempts := 0
	runISCCCommand = func(ctx context.Context, projectDir string, command []string, runner runlog.Runner) error {
		attempts++
		return errors.New("file is locked")
	}

	projectDir := t.TempDir()
	cfg := packagingConfigWithFakeISCC(t, projectDir)

	err := NewService(nil).runISCC(context.Background(), projectDir, cfg, runlog.Transaction{})
	if err == nil {
		t.Fatal("expected ISCC failure")
	}
	if attempts != innoCompileMaxRetries+1 {
		t.Fatalf("ISCC attempts = %d, want %d", attempts, innoCompileMaxRetries+1)
	}
	if !strings.Contains(err.Error(), "after 10 retries") {
		t.Fatalf("error = %q, want retry count", err.Error())
	}
}

func packagingConfigWithFakeISCC(t *testing.T, projectDir string) contracts.PackagingConfig {
	t.Helper()
	fakeISCC := filepath.Join(projectDir, "ISCC.exe")
	if err := os.WriteFile(fakeISCC, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(projectDir, "builder", "windows", "inno.iss")
	if err := os.MkdirAll(filepath.Dir(script), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(script, []byte("[Setup]\n"), 0644); err != nil {
		t.Fatal(err)
	}
	return contracts.PackagingConfig{
		Windows: contracts.WindowsConfig{
			InnoScript: "builder/windows/inno.iss",
			ISCCPath:   fakeISCC,
		},
	}
}

func TestPackageLogsCarryTransaction(t *testing.T) {
	projectDir := t.TempDir()
	wailsConfig := contracts.WailsProjectConfig{
		Info: contracts.WailsAppInfo{
			ProductName:       "Demo",
			ProductIdentifier: "com.example.demo",
			Version:           "1.0.0",
		},
	}
	if err := project.SaveProjectRecord(contracts.ProjectRecord{
		ProjectDir: projectDir,
		Project: contracts.WailsProjectManager{
			ProjectDir:  projectDir,
			WailsConfig: wailsConfig,
		},
	}); err != nil {
		t.Fatal(err)
	}
	writeManagedTaskfile(t, projectDir)
	cfg := defaultPackagingConfig(projectDir, wailsConfig)
	if err := packagingConfig.SavePackagingConfig(projectDir, cfg); err != nil {
		t.Fatal(err)
	}

	log := runlog.New()
	events := []runlog.Line{}
	log.OnLine = func(line runlog.Line) {
		events = append(events, line)
	}

	_, err := NewService(log).Package(contracts.PackageRequest{
		ProjectDir:    projectDir,
		Platform:      contracts.PlatformWindows,
		DryRun:        true,
		RunBuild:      true,
		TransactionID: "package-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) == 0 {
		t.Fatal("expected package logs")
	}
	for _, event := range events {
		if event.TransactionID != "package-1" || event.TransactionType != "package" || event.TransactionTitle == "" {
			t.Fatalf("log event transaction = %#v", event)
		}
	}
}

func TestPackageRejectsConcurrentRuns(t *testing.T) {
	projectDir := t.TempDir()
	wailsConfig := contracts.WailsProjectConfig{
		Info: contracts.WailsAppInfo{
			ProductName:       "Demo",
			ProductIdentifier: "com.example.demo",
			Version:           "1.0.0",
		},
	}
	if err := project.SaveProjectRecord(contracts.ProjectRecord{
		ProjectDir: projectDir,
		Project: contracts.WailsProjectManager{
			ProjectDir:  projectDir,
			WailsConfig: wailsConfig,
		},
	}); err != nil {
		t.Fatal(err)
	}
	writeManagedTaskfile(t, projectDir)

	cfg := defaultPackagingConfig(projectDir, wailsConfig)
	cfg.Build.Command = sleepCommand()
	cfg.Windows.Enabled = false
	if err := packagingConfig.SavePackagingConfig(projectDir, cfg); err != nil {
		t.Fatal(err)
	}

	svc := NewService(nil)
	var wg sync.WaitGroup
	wg.Add(1)
	firstErr := make(chan error, 1)
	go func() {
		defer wg.Done()
		_, err := svc.Package(contracts.PackageRequest{
			ProjectDir:    projectDir,
			Platform:      contracts.PlatformWindows,
			RunBuild:      true,
			TransactionID: "package-1",
		})
		firstErr <- err
	}()

	deadline := time.Now().Add(time.Second)
	for {
		if !svc.beginPackage() {
			break
		}
		svc.endPackage()
		if time.Now().After(deadline) {
			t.Fatal("first package did not acquire the package lock")
		}
		time.Sleep(10 * time.Millisecond)
	}

	_, err := svc.Package(contracts.PackageRequest{
		ProjectDir:    projectDir,
		Platform:      contracts.PlatformWindows,
		TransactionID: "package-2",
	})
	if err == nil || !strings.Contains(err.Error(), "already running") {
		t.Fatalf("Package() error = %v, want already running", err)
	}

	wg.Wait()
	if err := <-firstErr; err != nil {
		t.Fatal(err)
	}
}

func sleepCommand() []string {
	if runtime.GOOS == "windows" {
		return []string{"powershell", "-NoProfile", "-Command", "Start-Sleep -Milliseconds 250"}
	}
	return []string{"sh", "-c", "sleep 0.25"}
}
