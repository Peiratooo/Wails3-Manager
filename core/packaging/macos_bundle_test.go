package packaging

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"wails3-manager/core/contracts"
)

func TestPrepareMacOSAppBundleCreatesBundleFromProjectConfig(t *testing.T) {
	projectDir := t.TempDir()
	mustWriteFile(t, filepath.Join(projectDir, "bin", "demo"), "binary")
	mustWriteFile(t, filepath.Join(projectDir, "build", "darwin", "icons.icns"), "icon")
	mustWriteFile(t, filepath.Join(projectDir, "assets", "runtime.dat"), "runtime")
	mustWriteFile(t, filepath.Join(projectDir, "assets", "license.txt"), "license")
	mustWriteFile(t, filepath.Join(projectDir, "extras", "config.json"), "{}")
	mustWriteFile(t, filepath.Join(projectDir, "launcher", "helper"), "helper")
	mustWriteFile(t, filepath.Join(projectDir, "build", "darwin", "Info.plist"), `<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
    <dict>
        <key>CFBundleName</key>
        <string>Old Name</string>
        <key>CFBundleExecutable</key>
        <string>old</string>
        <key>CFBundleIdentifier</key>
        <string>old.id</string>
        <key>CFBundleVersion</key>
        <string>0.0.1</string>
        <key>CFBundleShortVersionString</key>
        <string>0.0.1</string>
        <key>NSHumanReadableCopyright</key>
        <string>old copyright</string>
    </dict>
</plist>
`)

	cfg := contracts.PackagingConfig{
		Build: contracts.BuildSettings{AppName: "demo"},
		Entry: contracts.ProgramEntry{ExecutablePath: "launcher/helper"},
		MacOS: contracts.MacOSConfig{
			AppBundle:    "bin/${project.name}.app",
			Background:   "assets/dmg-bg.png",
			WindowWidth:  320,
			WindowHeight: 180,
		},
		Assets: []contracts.PackagingAsset{
			{Src: "assets/runtime.dat", Type: "file", Required: true},
			{Src: "assets/license.txt", Type: "file", Required: true, Target: "Resources"},
			{Src: "extras", Type: "directory", Required: true, Target: "SharedSupport/data"},
		},
	}
	projectConfig := contracts.WailsProjectConfig{
		Info: contracts.WailsAppInfo{
			ProductName:       "Demo Product",
			ProductIdentifier: "com.example.demo",
			Version:           "v2.3.4",
			Copyright:         "(c) Demo",
			Comments:          "Demo comment",
		},
	}

	appBundle, err := prepareMacOSAppBundle(projectDir, cfg, projectConfig)
	if err != nil {
		t.Fatal(err)
	}

	wantBundle := filepath.Join(projectDir, "bin", "Demo Product.app")
	if appBundle != wantBundle {
		t.Fatalf("appBundle = %q, want %q", appBundle, wantBundle)
	}
	for _, want := range []string{
		filepath.Join(wantBundle, "Contents", "MacOS", "demo"),
		filepath.Join(wantBundle, "Contents", "MacOS", "runtime.dat"),
		filepath.Join(wantBundle, "Contents", "MacOS", "helper"),
		filepath.Join(wantBundle, "Contents", "Resources", "icons.icns"),
		filepath.Join(wantBundle, "Contents", "Resources", "license.txt"),
		filepath.Join(wantBundle, "Contents", "SharedSupport", "data", "extras", "config.json"),
	} {
		if _, err := os.Stat(want); err != nil {
			t.Fatalf("expected bundle file %s: %v", want, err)
		}
	}
	for _, unwanted := range []string{
		filepath.Join(wantBundle, "Contents", "Resources", "runtime.dat"),
		filepath.Join(wantBundle, "Contents", "Resources", "helper"),
		filepath.Join(wantBundle, "Contents", "MacOS", "license.txt"),
		filepath.Join(wantBundle, "Contents", "MacOS", "extras"),
	} {
		if _, err := os.Stat(unwanted); !os.IsNotExist(err) {
			t.Fatalf("unexpected file in Resources: %s", unwanted)
		}
	}
	if _, err := os.Stat(filepath.Join(wantBundle, "Contents", "Resources", "dmg-background.png")); !os.IsNotExist(err) {
		t.Fatalf("DMG background should not be embedded in the app bundle: %v", err)
	}

	plist, err := os.ReadFile(filepath.Join(wantBundle, "Contents", "Info.plist"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(plist)
	for _, want := range []string{
		`<key>CFBundleName</key>
        <string>Demo Product</string>`,
		`<key>CFBundleDisplayName</key>
        <string>Demo Product</string>`,
		`<key>CFBundleExecutable</key>
        <string>helper</string>`,
		`<key>CFBundleIdentifier</key>
        <string>com.example.demo</string>`,
		`<key>CFBundleVersion</key>
        <string>2.3.4</string>`,
		`<key>CFBundleShortVersionString</key>
        <string>2.3.4</string>`,
		`<key>NSHumanReadableCopyright</key>
        <string>(c) Demo</string>`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("Info.plist missing %q:\n%s", want, text)
		}
	}
}

func TestPrepareMacOSAppBundleKeepsSelectedLaunchProgramWhenNamesOverlap(t *testing.T) {
	projectDir := t.TempDir()
	mustWriteFile(t, filepath.Join(projectDir, "bin", "demo"), "wails binary")
	mustWriteFile(t, filepath.Join(projectDir, "build", "darwin", "icons.icns"), "icon")
	mustWriteFile(t, filepath.Join(projectDir, "launcher", "demo"), "selected launcher")

	cfg := contracts.PackagingConfig{
		Build: contracts.BuildSettings{AppName: "demo"},
		Entry: contracts.ProgramEntry{ExecutablePath: "launcher/demo"},
		MacOS: contracts.MacOSConfig{
			AppBundle: "bin/${project.name}.app",
		},
	}
	projectConfig := contracts.WailsProjectConfig{
		Info: contracts.WailsAppInfo{
			ProductName:       "Demo Product",
			ProductIdentifier: "com.example.demo",
			Version:           "1.0.0",
		},
	}

	appBundle, err := prepareMacOSAppBundle(projectDir, cfg, projectConfig)
	if err != nil {
		t.Fatal(err)
	}

	executable, err := os.ReadFile(filepath.Join(appBundle, "Contents", "MacOS", "demo"))
	if err != nil {
		t.Fatal(err)
	}
	if string(executable) != "selected launcher" {
		t.Fatalf("launch executable content = %q, want selected launcher", executable)
	}

	plist, err := os.ReadFile(filepath.Join(appBundle, "Contents", "Info.plist"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(plist), `<key>CFBundleExecutable</key>
        <string>demo</string>`) {
		t.Fatalf("Info.plist should point at selected launch executable:\n%s", plist)
	}
}

func TestPrepareMacOSAppBundleDoesNotUseAppBundleAsExecutable(t *testing.T) {
	projectDir := t.TempDir()
	mustWriteFile(t, filepath.Join(projectDir, "bin", "demo"), "binary")
	mustWriteFile(t, filepath.Join(projectDir, "build", "darwin", "icons.icns"), "icon")

	cfg := contracts.PackagingConfig{
		Build: contracts.BuildSettings{AppName: "demo"},
		Entry: contracts.ProgramEntry{ExecutablePath: "bin/${project.name}.app"},
		MacOS: contracts.MacOSConfig{
			AppBundle: "bin/${project.name}.app",
		},
	}
	projectConfig := contracts.WailsProjectConfig{
		Info: contracts.WailsAppInfo{
			ProductName:       "Demo Product",
			ProductIdentifier: "com.example.demo",
			Version:           "1.0.0",
		},
	}

	appBundle, err := prepareMacOSAppBundle(projectDir, cfg, projectConfig)
	if err != nil {
		t.Fatal(err)
	}

	plist, err := os.ReadFile(filepath.Join(appBundle, "Contents", "Info.plist"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(plist)
	if strings.Contains(text, `<string>Demo Product.app</string>`) {
		t.Fatalf("Info.plist should not use .app bundle as CFBundleExecutable:\n%s", text)
	}
	if !strings.Contains(text, `<key>CFBundleExecutable</key>
        <string>demo</string>`) {
		t.Fatalf("Info.plist should point at default executable inside Contents/MacOS:\n%s", text)
	}
}

func TestPrepareMacOSAppBundleRejectsInvalidAssetTarget(t *testing.T) {
	projectDir := t.TempDir()
	mustWriteFile(t, filepath.Join(projectDir, "bin", "demo"), "binary")
	mustWriteFile(t, filepath.Join(projectDir, "build", "darwin", "icons.icns"), "icon")
	mustWriteFile(t, filepath.Join(projectDir, "assets", "runtime.dat"), "runtime")

	cfg := contracts.PackagingConfig{
		Build: contracts.BuildSettings{AppName: "demo"},
		MacOS: contracts.MacOSConfig{
			AppBundle: "bin/${build.appName}.app",
		},
		Assets: []contracts.PackagingAsset{
			{Src: "assets/runtime.dat", Type: "file", Required: true, Target: "../MacOS"},
		},
	}

	if _, err := prepareMacOSAppBundle(projectDir, cfg, contracts.WailsProjectConfig{}); err == nil {
		t.Fatal("expected invalid macOS asset target to be rejected")
	}
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
