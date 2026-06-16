package packaging

import (
	"encoding/json"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"wails3-manager/core/contracts"
	packagingConfig "wails3-manager/core/packaging/config"
	"wails3-manager/core/project"
)

func TestBuildSettingsBecomeTaskVarsAndEnv(t *testing.T) {
	cfg := contracts.PackagingConfig{
		Build: contracts.BuildSettings{
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
}

func TestDefaultPackagingConfigInitializesOnlyCurrentPlatform(t *testing.T) {
	cfg := defaultPackagingConfig(t.TempDir(), contracts.WailsProjectConfig{
		Info: contracts.WailsAppInfo{ProductName: "Demo"},
	})

	switch runtime.GOOS {
	case "windows":
		if !cfg.Windows.Enabled || cfg.Windows.InnoScript == "" {
			t.Fatalf("windows config was not initialized: %#v", cfg.Windows)
		}
		if cfg.MacOS.Enabled || cfg.MacOS.DMGScript != "" || cfg.MacOS.Background != "" {
			t.Fatalf("macos config should not be initialized on windows: %#v", cfg.MacOS)
		}
	case "darwin":
		if !cfg.MacOS.Enabled || cfg.MacOS.DMGScript == "" || cfg.MacOS.Background == "" {
			t.Fatalf("macos config was not initialized: %#v", cfg.MacOS)
		}
		if cfg.Windows.Enabled || cfg.Windows.InnoScript != "" {
			t.Fatalf("windows config should not be initialized on macos: %#v", cfg.Windows)
		}
	default:
		if cfg.Windows.Enabled || cfg.Windows.InnoScript != "" || cfg.MacOS.Enabled || cfg.MacOS.DMGScript != "" {
			t.Fatalf("platform config should not be initialized on %s: windows=%#v macos=%#v", runtime.GOOS, cfg.Windows, cfg.MacOS)
		}
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if runtime.GOOS == "windows" && strings.Contains(text, `"macos"`) {
		t.Fatalf("windows initialization should not write macos config: %s", text)
	}
	if runtime.GOOS == "darwin" && strings.Contains(text, `"windows"`) {
		t.Fatalf("macos initialization should not write windows config: %s", text)
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
			Task:     "builder:release",
			AppName:  "demo",
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
			Task:     "builder:release",
			AppName:  "demo",
		},
		Entry: contracts.ProgramEntry{
			ExecutablePath: "dist/${build.appName}",
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

	cfg := contracts.PackagingConfig{
		SchemaVersion: 1,
		Build: contracts.BuildSettings{
			Taskfile: "Taskfile.yml",
			Task:     "builder:release",
			AppName:  "demo",
		},
		Artifacts: contracts.ArtifactConfig{OutputRoot: "builder/release"},
	}
	if err := packagingConfig.SavePackagingConfig(projectDir, cfg); err != nil {
		t.Fatal(err)
	}

	_, err := NewService(nil).Package(contracts.PackageRequest{
		ProjectDir: projectDir,
		Platform:   contracts.Platform("linux"),
	})
	if err == nil || !strings.Contains(err.Error(), "unsupported packaging platform") {
		t.Fatalf("Package() error = %v, want unsupported platform", err)
	}
}
