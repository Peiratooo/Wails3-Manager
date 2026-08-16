package dmg

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	upstreamdmg "github.com/leaanthony/dmg/dmg"

	"wails3-manager/core/contracts"
)

func TestBuildOptionsUsesConfiguredLayout(t *testing.T) {
	projectDir := t.TempDir()
	cfg := testConfig()
	projectConfig := contracts.WailsProjectConfig{
		Info: contracts.WailsAppInfo{
			ProductName: "Demo Product",
			Version:     "1.0.0",
		},
	}
	appBundle := filepath.Join(projectDir, "bin", "Demo Product.app")
	background := filepath.Join(projectDir, "builder", "macos", "dmg-background.png")

	opts, err := BuildOptions(projectDir, cfg, projectConfig, appBundle, background)
	if err != nil {
		t.Fatal(err)
	}
	if opts.VolumeName != "Demo Product" {
		t.Fatalf("VolumeName = %q, want Demo Product", opts.VolumeName)
	}
	if got := opts.Files["Demo Product.app"]; got != appBundle {
		t.Fatalf("Files[Demo Product.app] = %q, want %q", got, appBundle)
	}
	if !opts.AddApplicationsSymlink {
		t.Fatal("AddApplicationsSymlink = false, want true")
	}
	if opts.Window.Width != 300 || opts.Window.Height != 400 {
		t.Fatalf("Window = %#v, want 300x400", opts.Window)
	}
	if opts.Icon.Size != 96 || opts.Icon.TextSize != 12 || opts.Icon.GridSpace != 100 {
		t.Fatalf("Icon = %#v", opts.Icon)
	}
	if got := opts.IconPositions["Demo Product.app"]; got != (upstreamdmg.IconPosition{X: 90, Y: 200}) {
		t.Fatalf("app icon position = %#v", got)
	}
	if got := opts.IconPositions["Applications"]; got != (upstreamdmg.IconPosition{X: 210, Y: 200}) {
		t.Fatalf("Applications icon position = %#v", got)
	}
	if opts.Background == nil || opts.Background.File != background {
		t.Fatalf("Background = %#v, want %q", opts.Background, background)
	}
	if opts.Format != upstreamdmg.FormatUDZO || opts.Filesystem != upstreamdmg.FSHFSPlus || opts.Backend != upstreamdmg.BackendAuto {
		t.Fatalf("image settings = format %q, filesystem %q, backend %v", opts.Format, opts.Filesystem, opts.Backend)
	}
	if got := filepath.Base(opts.OutputPath); got != "Demo Product-1.0.0-macos-setup.dmg" {
		t.Fatalf("output filename = %q", got)
	}
}

func TestBuildOptionsUsesInstallerOutputName(t *testing.T) {
	projectDir := t.TempDir()
	cfg := testConfig()
	projectConfig := contracts.WailsProjectConfig{
		Info: contracts.WailsAppInfo{
			ProductName: "Wails3.Manager",
			Version:     "1.0.0",
		},
	}

	opts, err := BuildOptions(projectDir, cfg, projectConfig, filepath.Join(projectDir, "bin", "Wails3.Manager.app"), "")
	if err != nil {
		t.Fatal(err)
	}
	if got := filepath.Base(opts.OutputPath); got != "Wails3.Manager-1.0.0-macos-setup.dmg" {
		t.Fatalf("output filename = %q", got)
	}
	if opts.Background != nil {
		t.Fatalf("Background = %#v, want nil", opts.Background)
	}
}

func TestBuildRemovesExistingOutputAndInvokesLibrary(t *testing.T) {
	projectDir := t.TempDir()
	cfg := testConfig()
	projectConfig := contracts.WailsProjectConfig{
		Info: contracts.WailsAppInfo{
			ProductName: "Demo",
			Version:     "1.0.0",
		},
	}
	appBundle := filepath.Join(projectDir, "bin", "Demo.app")
	if err := os.MkdirAll(appBundle, 0755); err != nil {
		t.Fatal(err)
	}

	opts, err := BuildOptions(projectDir, cfg, projectConfig, appBundle, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(opts.OutputPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(opts.OutputPath, []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}

	originalBuildImage := buildImage
	defer func() { buildImage = originalBuildImage }()
	called := false
	buildImage = func(got upstreamdmg.Options) error {
		called = true
		if _, err := os.Stat(got.OutputPath); !os.IsNotExist(err) {
			t.Fatalf("existing output was not removed: %v", err)
		}
		return os.WriteFile(got.OutputPath, []byte("new dmg"), 0644)
	}

	outputPath, err := Build(projectDir, cfg, projectConfig, appBundle, "")
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("upstream Build was not called")
	}
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "new dmg" {
		t.Fatalf("output content = %q", data)
	}
}

func TestPrepareBackgroundCropsToWindowSize(t *testing.T) {
	projectDir := t.TempDir()
	sourcePath := filepath.Join(projectDir, "background.png")
	outputPath := filepath.Join(projectDir, "builder", "macos", "dmg-background.png")
	writePNG(t, sourcePath, 800, 400)

	got, err := PrepareBackground(projectDir, "background.png", 320, 180, outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if got != outputPath {
		t.Fatalf("output = %q, want %q", got, outputPath)
	}
	file, err := os.Open(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(file)
	closeErr := file.Close()
	if err != nil {
		t.Fatal(err)
	}
	if closeErr != nil {
		t.Fatal(closeErr)
	}
	if img.Bounds().Dx() != 320 || img.Bounds().Dy() != 180 {
		t.Fatalf("background size = %dx%d", img.Bounds().Dx(), img.Bounds().Dy())
	}
}

func testConfig() contracts.PackagingConfig {
	return contracts.PackagingConfig{
		Build: contracts.BuildSettings{AppName: "demo"},
		MacOS: contracts.MacOSConfig{
			AppBundle:     "bin/${build.appName}.app",
			Background:    "background.png",
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

func writePNG(t *testing.T, path string, width, height int) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	img.Set(0, 0, color.NRGBA{R: 255, A: 255})
	encodeErr := png.Encode(file, img)
	closeErr := file.Close()
	if encodeErr != nil {
		t.Fatal(encodeErr)
	}
	if closeErr != nil {
		t.Fatal(closeErr)
	}
}
