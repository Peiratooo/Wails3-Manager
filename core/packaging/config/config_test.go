package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"wails3-manager/core/contracts"
)

func TestSavePackagingConfigDoesNotWriteProjectMetadata(t *testing.T) {
	path := filepath.Join(t.TempDir(), "packaging.json")
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
			OutputBaseName:        InstallerOutputNameTemplate,
			CreateDesktopShortcut: true,
		},
		MacOS: contracts.MacOSConfig{
			Enabled:       true,
			AppBundle:     "bin/${build.appName}.app",
			DMGScript:     "builder/macos/dmg.sh",
			Background:    "assets/install-grid.png",
			OutputName:    InstallerOutputNameTemplate,
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

	if err := SavePackagingConfigFile(path, cfg); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if strings.Contains(text, `"project"`) || strings.Contains(text, `"productName"`) || strings.Contains(text, `"productIdentifier"`) {
		t.Fatalf("packaging config should not contain project metadata:\n%s", text)
	}
}

func TestSavePackagingConfigRejectsMissingBuildTaskfile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "packaging.json")
	cfg := contracts.PackagingConfig{
		SchemaVersion: 1,
		Build: contracts.BuildSettings{
			Task:    "release",
			AppName: "demo",
		},
		Windows: contracts.WindowsConfig{
			Enabled:               true,
			InnoScript:            "builder/windows/inno.iss",
			DefaultDirName:        `{autopf}\${project.name}`,
			PrivilegesRequired:    "lowest",
			SetupIcon:             "build/windows/icon.ico",
			OutputBaseName:        InstallerOutputNameTemplate,
			CreateDesktopShortcut: true,
		},
		MacOS: contracts.MacOSConfig{
			Enabled:       true,
			AppBundle:     "bin/${build.appName}.app",
			DMGScript:     "builder/macos/dmg.sh",
			Background:    "assets/install-grid.png",
			OutputName:    InstallerOutputNameTemplate,
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

	if err := SavePackagingConfigFile(path, cfg); err == nil {
		t.Fatal("expected missing build.taskfile to be rejected")
	}
}

func TestResolveExecutablePathDefaultsByPlatform(t *testing.T) {
	cfg := contracts.PackagingConfig{
		Build: contracts.BuildSettings{AppName: "demo"},
	}
	project := contracts.WailsProjectConfig{}

	windows := ResolveExecutablePath(cfg, project, contracts.PlatformWindows)
	if !windows.UsingDefaultExecutable || windows.EffectiveExecutablePath != "bin/demo.exe" {
		t.Fatalf("windows runtime info = %#v", windows)
	}

	macos := ResolveExecutablePath(cfg, project, contracts.PlatformMacOS)
	if !macos.UsingDefaultExecutable || macos.EffectiveExecutablePath != "bin/demo.app" {
		t.Fatalf("macos runtime info = %#v", macos)
	}
}

func TestInstallerOutputNameUsesOneRuleForAllPlatforms(t *testing.T) {
	cfg := contracts.PackagingConfig{
		Build: contracts.BuildSettings{AppName: "demo"},
	}
	project := contracts.WailsProjectConfig{
		Info: contracts.WailsAppInfo{
			ProductName: "Wails3.Manager",
			Version:     "v1.0.0",
		},
	}

	tests := []struct {
		platform contracts.Platform
		want     string
	}{
		{
			platform: contracts.PlatformWindows,
			want:     "Wails3.Manager-1.0.0-windows-setup",
		},
		{
			platform: contracts.PlatformMacOS,
			want:     "Wails3.Manager-1.0.0-macos-setup",
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.platform), func(t *testing.T) {
			if got := InstallerOutputName(cfg, project, tt.platform); got != tt.want {
				t.Fatalf("InstallerOutputName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveExecutablePathUsesConfiguredPath(t *testing.T) {
	cfg := contracts.PackagingConfig{
		Build: contracts.BuildSettings{AppName: "demo"},
		Entry: contracts.ProgramEntry{ExecutablePath: "out/${build.appName}"},
	}
	info := ResolveExecutablePath(cfg, contracts.WailsProjectConfig{}, contracts.PlatformWindows)
	if info.UsingDefaultExecutable {
		t.Fatalf("UsingDefaultExecutable = true, want false")
	}
	if info.DefaultExecutablePath != "bin/demo.exe" || info.EffectiveExecutablePath != "out/demo" {
		t.Fatalf("runtime info = %#v", info)
	}
}

func TestResolveMacOSAppBundlePathTreatsDefaultBundleAsDefault(t *testing.T) {
	cfg := contracts.PackagingConfig{
		Build: contracts.BuildSettings{AppName: "demo"},
		MacOS: contracts.MacOSConfig{AppBundle: "bin/${build.appName}.app"},
	}
	project := contracts.WailsProjectConfig{Info: contracts.WailsAppInfo{ProductName: "Demo Product"}}
	info := ResolveMacOSAppBundlePath(cfg, project)
	if !info.UsingDefaultExecutable {
		t.Fatalf("UsingDefaultExecutable = false, want true")
	}
	if info.DefaultExecutablePath != "bin/Demo Product.app" || info.EffectiveExecutablePath != "bin/Demo Product.app" {
		t.Fatalf("runtime info = %#v", info)
	}
}

func TestResolveMacOSAppBundlePathUsesCustomBundle(t *testing.T) {
	cfg := contracts.PackagingConfig{
		Build: contracts.BuildSettings{AppName: "demo"},
		MacOS: contracts.MacOSConfig{AppBundle: "release/Demo.app"},
	}
	info := ResolveMacOSAppBundlePath(cfg, contracts.WailsProjectConfig{})
	if info.UsingDefaultExecutable {
		t.Fatalf("UsingDefaultExecutable = true, want false")
	}
	if info.EffectiveExecutablePath != "release/Demo.app" {
		t.Fatalf("EffectiveExecutablePath = %q, want %q", info.EffectiveExecutablePath, "release/Demo.app")
	}
}

func TestEffectiveAssetsIncludesBuildOutputAndLaunchProgram(t *testing.T) {
	cfg := contracts.PackagingConfig{
		Build: contracts.BuildSettings{AppName: "demo"},
		Entry: contracts.ProgramEntry{ExecutablePath: "launcher/start.exe"},
		Assets: []contracts.PackagingAsset{
			{Src: "assets", Type: "directory", Required: false},
		},
	}

	assets := EffectiveAssets(cfg, contracts.WailsProjectConfig{}, contracts.PlatformWindows)
	want := []contracts.PackagingAsset{
		{Src: "bin/demo.exe", Type: "file", Required: true},
		{Src: "launcher/start.exe", Type: "file", Required: true},
		{Src: "assets", Type: "directory", Required: false},
	}
	if len(assets) != len(want) {
		t.Fatalf("EffectiveAssets length = %d, want %d: %#v", len(assets), len(want), assets)
	}
	for i := range want {
		if assets[i] != want[i] {
			t.Fatalf("EffectiveAssets[%d] = %#v, want %#v", i, assets[i], want[i])
		}
	}
}

func TestEffectiveAssetsDeduplicatesImplicitAssets(t *testing.T) {
	cfg := contracts.PackagingConfig{
		Build: contracts.BuildSettings{AppName: "demo"},
		Assets: []contracts.PackagingAsset{
			{Src: "bin/demo.exe", Type: "file", Required: false},
		},
	}

	assets := EffectiveAssets(cfg, contracts.WailsProjectConfig{}, contracts.PlatformWindows)
	if len(assets) != 1 {
		t.Fatalf("EffectiveAssets length = %d, want 1: %#v", len(assets), assets)
	}
	if !assets[0].Required {
		t.Fatalf("implicit build output should stay required: %#v", assets[0])
	}
}

func TestEffectiveAssetsIncludesMacOSAppBundle(t *testing.T) {
	cfg := contracts.PackagingConfig{
		Build: contracts.BuildSettings{AppName: "demo"},
		Entry: contracts.ProgramEntry{ExecutablePath: "launcher/Helper.app"},
		MacOS: contracts.MacOSConfig{
			AppBundle: "bin/${build.appName}.app",
		},
	}

	project := contracts.WailsProjectConfig{Info: contracts.WailsAppInfo{ProductName: "Demo Product"}}
	assets := EffectiveAssets(cfg, project, contracts.PlatformMacOS)
	want := []contracts.PackagingAsset{
		{Src: "bin/Demo Product.app", Type: "directory", Required: true},
	}
	if len(assets) != len(want) {
		t.Fatalf("EffectiveAssets length = %d, want %d: %#v", len(assets), len(want), assets)
	}
	for i := range want {
		if assets[i] != want[i] {
			t.Fatalf("EffectiveAssets[%d] = %#v, want %#v", i, assets[i], want[i])
		}
	}
}

func TestSavePackagingConfigAllowsDisabledPlatformSections(t *testing.T) {
	path := filepath.Join(t.TempDir(), "packaging.json")
	cfg := contracts.PackagingConfig{
		SchemaVersion: 1,
		Build: contracts.BuildSettings{
			Taskfile: "Taskfile.yml",
			Task:     "release",
			AppName:  "demo",
		},
		Artifacts: contracts.ArtifactConfig{OutputRoot: "builder/release"},
		Windows:   contracts.WindowsConfig{Enabled: false},
		MacOS:     contracts.MacOSConfig{Enabled: false},
	}

	if err := SavePackagingConfigFile(path, cfg); err != nil {
		t.Fatal(err)
	}
}
