package dmg

import (
	"encoding/base64"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	upstreamdmg "github.com/leaanthony/dmg/dmg"
	"golang.org/x/image/draw"

	"wails3-manager/core/contracts"
	"wails3-manager/core/fsx"
	"wails3-manager/core/packaging/config"
)

var buildImage = upstreamdmg.Build

func DefaultBackgroundPNG() []byte {
	// 1x1 transparent PNG placeholder; users can replace it from GUI.
	data, _ := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAAEUlEQVR4nGJiYGBgAAQAAP//AA8AA/6P688AAAAASUVORK5CYII=")
	return data
}

// Build creates the configured DMG and returns its absolute output path.
func Build(projectDir string, cfg contracts.PackagingConfig, projectConfig contracts.WailsProjectConfig, appBundle, background string) (string, error) {
	opts, err := BuildOptions(projectDir, cfg, projectConfig, appBundle, background)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(opts.OutputPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create DMG output directory: %w", err)
	}
	if err := os.Remove(opts.OutputPath); err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to remove existing DMG: %w", err)
	}
	if err := buildImage(opts); err != nil {
		return "", fmt.Errorf("failed to build DMG: %w", err)
	}
	return opts.OutputPath, nil
}

// BuildOptions maps Wails3 Manager packaging settings to leaanthony/dmg.
func BuildOptions(projectDir string, cfg contracts.PackagingConfig, projectConfig contracts.WailsProjectConfig, appBundle, background string) (upstreamdmg.Options, error) {
	projectName := strings.TrimSpace(config.ProjectName(projectConfig))
	if projectName == "" {
		return upstreamdmg.Options{}, fmt.Errorf("project productName is required for macOS packaging")
	}
	if strings.TrimSpace(config.ProjectVersion(projectConfig)) == "" {
		return upstreamdmg.Options{}, fmt.Errorf("project version is required for macOS packaging")
	}

	appBundle = fsx.Resolve(projectDir, appBundle)
	if appBundle == "" {
		return upstreamdmg.Options{}, fmt.Errorf("macOS app bundle path is not configured")
	}
	outputDir := fsx.Resolve(projectDir, config.MacOSOutputDir(cfg, projectConfig))
	if outputDir == "" {
		return upstreamdmg.Options{}, fmt.Errorf("macOS output directory is not configured")
	}
	outputPath := filepath.Join(outputDir, config.InstallerOutputName(cfg, projectConfig, contracts.PlatformMacOS)+".dmg")
	appName := strings.TrimSuffix(projectName, ".app") + ".app"

	opts := upstreamdmg.Options{
		VolumeName: projectName,
		OutputPath: outputPath,
		Files: map[string]string{
			appName: appBundle,
		},
		AddApplicationsSymlink: true,
		IconPositions: map[string]upstreamdmg.IconPosition{
			appName: {
				X: cfg.MacOS.AppX,
				Y: cfg.MacOS.AppY,
			},
			"Applications": {
				X: cfg.MacOS.ApplicationsX,
				Y: cfg.MacOS.ApplicationsY,
			},
		},
		Window: upstreamdmg.WindowConfig{
			X:      100,
			Y:      100,
			Width:  cfg.MacOS.WindowWidth,
			Height: cfg.MacOS.WindowHeight,
		},
		Icon: upstreamdmg.IconConfig{
			Size:      cfg.MacOS.IconSize,
			TextSize:  12,
			GridSpace: 100,
		},
		Format:     upstreamdmg.FormatUDZO,
		Filesystem: upstreamdmg.FSHFSPlus,
		Backend:    upstreamdmg.BackendAuto,
	}
	if background = strings.TrimSpace(background); background != "" {
		opts.Background = &upstreamdmg.BackgroundConfig{File: fsx.Resolve(projectDir, background)}
	}
	return opts, nil
}

func PrepareBackground(projectDir, background string, width, height int, output string) (string, error) {
	background = strings.TrimSpace(background)
	if background == "" {
		return "", nil
	}
	if width <= 0 || height <= 0 {
		return "", fmt.Errorf("DMG background dimensions must be greater than zero")
	}
	sourcePath := fsx.Resolve(projectDir, background)
	file, err := os.Open(sourcePath)
	if err != nil {
		return "", fmt.Errorf("failed to read DMG background: %w", err)
	}
	defer file.Close()
	source, _, err := image.Decode(file)
	if err != nil {
		return "", fmt.Errorf("failed to decode DMG background: %w", err)
	}
	if strings.TrimSpace(output) == "" {
		output = filepath.Join(projectDir, "builder", "macos", "dmg-background.png")
	}
	if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
		return "", err
	}
	out, err := os.Create(output)
	if err != nil {
		return "", err
	}
	encodeErr := png.Encode(out, coverImage(source, width, height))
	closeErr := out.Close()
	if encodeErr != nil {
		return "", encodeErr
	}
	if closeErr != nil {
		return "", closeErr
	}
	return output, nil
}

func coverImage(source image.Image, width, height int) *image.NRGBA {
	bounds := source.Bounds()
	sourceWidth := bounds.Dx()
	sourceHeight := bounds.Dy()
	crop := bounds
	if sourceWidth*height > sourceHeight*width {
		cropWidth := max(1, sourceHeight*width/height)
		crop.Min.X += (sourceWidth - cropWidth) / 2
		crop.Max.X = crop.Min.X + cropWidth
	} else {
		cropHeight := max(1, sourceWidth*height/width)
		crop.Min.Y += (sourceHeight - cropHeight) / 2
		crop.Max.Y = crop.Min.Y + cropHeight
	}
	output := image.NewNRGBA(image.Rect(0, 0, width, height))
	draw.BiLinear.Scale(output, output.Bounds(), source, crop, draw.Src, nil)
	return output
}
