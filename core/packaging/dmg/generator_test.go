package dmg

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"wails3-manager/core/contracts"
)

func TestGenerateScriptUsesBundledBackgroundPath(t *testing.T) {
	projectDir := t.TempDir()
	cfg := testConfig()
	projectConfig := contracts.WailsProjectConfig{
		Info: contracts.WailsAppInfo{
			ProductName: "Demo",
			Version:     "1.0.0",
		},
	}

	scriptPath, err := GenerateScript(projectDir, cfg, projectConfig)
	if err != nil {
		t.Fatal(err)
	}
	script, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatal(err)
	}
	if bash, err := exec.LookPath("bash"); err == nil {
		cmd := exec.Command(bash, "-n", "-")
		cmd.Stdin = strings.NewReader(string(script))
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("generated script is not valid bash: %v\n%s", err, output)
		}
	}
	text := string(script)
	for _, want := range []string{
		`BACKGROUND="$TMP_DIR/$APP_NAME.app/Contents/Resources/dmg-background.png"`,
		`CREATE_DMG_ARGS+=(--background "$BACKGROUND")`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("generated script missing %q:\n%s", want, text)
		}
	}
	for _, unwanted := range []string{
		`BACKGROUND="background.png"`,
		filepath.ToSlash(filepath.Join(projectDir, "builder", "macos", "background.png")),
	} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("generated script should not contain %q:\n%s", unwanted, text)
		}
	}
}

func TestGenerateScriptUsesAppBundleOnly(t *testing.T) {
	projectDir := t.TempDir()
	cfg := testConfig()
	cfg.Entry = contracts.ProgramEntry{ExecutablePath: "launcher/Helper.app"}
	projectConfig := contracts.WailsProjectConfig{
		Info: contracts.WailsAppInfo{
			ProductName: "Demo Product",
			Version:     "1.0.0",
		},
	}

	scriptPath, err := GenerateScript(projectDir, cfg, projectConfig)
	if err != nil {
		t.Fatal(err)
	}
	script, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatal(err)
	}
	text := string(script)
	for _, want := range []string{
		`APP_BUNDLE="bin/Demo Product.app"`,
		`cp -R "$APP_BUNDLE" "$TMP_DIR/$APP_NAME.app"`,
		`--window-size 300 400`,
		`--text-size 12`,
		`--icon "$APP_NAME.app" 90 200`,
		`--app-drop-link 210 200`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("generated script missing %q:\n%s", want, text)
		}
	}
	for _, unwanted := range []string{
		`if [ -e 'bin/demo.app' ]; then`,
		`if [ -e 'launcher/Helper.app' ]; then`,
		`{{extraFiles}}`,
	} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("generated script should not contain %q:\n%s", unwanted, text)
		}
	}
}

func TestGenerateScriptWrapsCreateDMGWithTimeoutCleanup(t *testing.T) {
	projectDir := t.TempDir()
	cfg := testConfig()
	projectConfig := contracts.WailsProjectConfig{
		Info: contracts.WailsAppInfo{
			ProductName: "Demo",
			Version:     "1.0.0",
		},
	}

	scriptPath, err := GenerateScript(projectDir, cfg, projectConfig)
	if err != nil {
		t.Fatal(err)
	}
	script, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatal(err)
	}
	text := string(script)
	for _, want := range []string{
		`CREATE_DMG_TIMEOUT_SECONDS="${CREATE_DMG_TIMEOUT_SECONDS:-180}"`,
		`sleep "$CREATE_DMG_TIMEOUT_SECONDS" >/dev/null 2>&1`,
		`kill_tree "$create_dmg_pid" TERM`,
		`kill_tree "$create_dmg_pid" KILL`,
		`kill_tree "$watchdog_pid" TERM`,
		`cleanup_partial_dmg`,
		`return 124`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("generated script missing %q:\n%s", want, text)
		}
	}
}

func testConfig() contracts.PackagingConfig {
	return contracts.PackagingConfig{
		Build: contracts.BuildSettings{AppName: "demo"},
		MacOS: contracts.MacOSConfig{
			AppBundle:     "bin/${build.appName}.app",
			DMGScript:     "builder/macos/dmg.sh",
			Background:    "background.png",
			OutputName:    "${build.appName}-${project.version}",
			CreateDMGPath: "create-dmg",
			WindowWidth:   300,
			WindowHeight:  400,
			IconSize:      96,
			AppX:          90,
			AppY:          200,
			ApplicationsX: 210,
			ApplicationsY: 200,
		},
	}
}
