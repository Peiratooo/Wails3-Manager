package dmg

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"wails3-manager/core/contracts"
)

func TestGenerateScriptFitsBackgroundToWindowSize(t *testing.T) {
	projectDir := t.TempDir()
	source := image.NewNRGBA(image.Rect(0, 0, 8, 9))
	for y := 0; y < 9; y++ {
		for x := 0; x < 8; x++ {
			source.SetNRGBA(x, y, color.NRGBA{R: uint8(x * 20), G: uint8(y * 20), B: 120, A: 255})
		}
	}
	sourcePath := filepath.Join(projectDir, "background.png")
	writePNG(t, sourcePath, source)

	cfg := contracts.PackagingConfig{
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
	backgroundPath := filepath.Join(projectDir, "builder", "macos", "background.png")
	if !strings.Contains(string(script), filepath.ToSlash(backgroundPath)) {
		t.Fatalf("script does not use generated background path:\n%s", string(script))
	}

	file, err := os.Open(backgroundPath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	output, err := png.Decode(file)
	if err != nil {
		t.Fatal(err)
	}
	if output.Bounds().Dx() != 300 || output.Bounds().Dy() != 400 {
		t.Fatalf("background size = %dx%d, want 300x400", output.Bounds().Dx(), output.Bounds().Dy())
	}
}

func TestGenerateScriptUsesAppBundleOnly(t *testing.T) {
	projectDir := t.TempDir()
	sourcePath := filepath.Join(projectDir, "background.png")
	writePNG(t, sourcePath, image.NewNRGBA(image.Rect(0, 0, 4, 4)))

	cfg := contracts.PackagingConfig{
		Build: contracts.BuildSettings{AppName: "demo"},
		Entry: contracts.ProgramEntry{ExecutablePath: "launcher/Helper.app"},
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

func writePNG(t *testing.T, path string, img image.Image) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	encodeErr := png.Encode(file, img)
	closeErr := file.Close()
	if encodeErr != nil {
		t.Fatal(encodeErr)
	}
	if closeErr != nil {
		t.Fatal(closeErr)
	}
}
