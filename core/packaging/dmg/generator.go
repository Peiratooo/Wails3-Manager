package dmg

import (
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"

	"wails3-manager/core/contracts"
	"wails3-manager/core/fsx"
	"wails3-manager/core/packaging/config"
)

func DefaultBackgroundPNG() []byte {
	// 1x1 transparent PNG placeholder; users can replace it from GUI.
	data, _ := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAAEUlEQVR4nGJiYGBgAAQAAP//AA8AA/6P688AAAAASUVORK5CYII=")
	return data
}

func GenerateScript(projectDir string, cfg contracts.PackagingConfig, projectConfig contracts.WailsProjectConfig) (string, error) {
	path := fsx.Resolve(projectDir, cfg.MacOS.DMGScript)
	if path == "" {
		return "", fmt.Errorf("DMG script path is not configured")
	}
	projectName := config.ProjectName(projectConfig)
	if projectName == "" {
		return "", fmt.Errorf("project productName is required for macOS packaging")
	}
	if config.ProjectVersion(projectConfig) == "" {
		return "", fmt.Errorf("project version is required for macOS packaging")
	}
	appBundle := config.MacOSAppBundlePath(cfg, projectConfig)
	outputDir := config.MacOSOutputDir(cfg, projectConfig)
	outputName := config.RenderPlaceholders(cfg.MacOS.OutputName, cfg, projectConfig)
	background, err := prepareBackground(projectDir, config.RenderPlaceholders(cfg.MacOS.Background, cfg, projectConfig), cfg.MacOS.WindowWidth, cfg.MacOS.WindowHeight)
	if err != nil {
		return "", err
	}
	extraFiles := buildExtraFiles(projectDir, cfg, projectConfig, appBundle)
	content := strings.NewReplacer(
		"{{appName}}", projectName,
		"{{appBundle}}", filepath.ToSlash(appBundle),
		"{{outputDir}}", filepath.ToSlash(outputDir),
		"{{outputName}}", outputName,
		"{{background}}", filepath.ToSlash(background),
		"{{windowWidth}}", fmt.Sprint(cfg.MacOS.WindowWidth),
		"{{windowHeight}}", fmt.Sprint(cfg.MacOS.WindowHeight),
		"{{iconSize}}", fmt.Sprint(cfg.MacOS.IconSize),
		"{{appX}}", fmt.Sprint(cfg.MacOS.AppX),
		"{{appY}}", fmt.Sprint(cfg.MacOS.AppY),
		"{{applicationsX}}", fmt.Sprint(cfg.MacOS.ApplicationsX),
		"{{applicationsY}}", fmt.Sprint(cfg.MacOS.ApplicationsY),
		"{{createDmg}}", cfg.MacOS.CreateDMGPath,
		"{{extraFiles}}", extraFiles,
	).Replace(defaultTemplate)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", err
	}
	return path, os.WriteFile(path, []byte(content), 0755)
}

func prepareBackground(projectDir, background string, width, height int) (string, error) {
	background = strings.TrimSpace(background)
	if background == "" {
		return "", nil
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
	output := filepath.Join(projectDir, "builder", "macos", "background.png")
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
	scale := math.Max(float64(width)/float64(sourceWidth), float64(height)/float64(sourceHeight))
	visibleWidth := float64(width) / scale
	visibleHeight := float64(height) / scale
	startX := float64(bounds.Min.X) + (float64(sourceWidth)-visibleWidth)/2
	startY := float64(bounds.Min.Y) + (float64(sourceHeight)-visibleHeight)/2
	output := image.NewNRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			sourceX := startX + (float64(x)+0.5)/scale
			sourceY := startY + (float64(y)+0.5)/scale
			output.SetNRGBA(x, y, sampleBilinear(source, sourceX, sourceY))
		}
	}
	return output
}

func sampleBilinear(source image.Image, x, y float64) color.NRGBA {
	bounds := source.Bounds()
	x = math.Max(float64(bounds.Min.X), math.Min(float64(bounds.Max.X-1), x))
	y = math.Max(float64(bounds.Min.Y), math.Min(float64(bounds.Max.Y-1), y))
	x0 := int(math.Floor(x))
	y0 := int(math.Floor(y))
	x1 := minInt(x0+1, bounds.Max.X-1)
	y1 := minInt(y0+1, bounds.Max.Y-1)
	tx := x - float64(x0)
	ty := y - float64(y0)

	c00 := color.NRGBAModel.Convert(source.At(x0, y0)).(color.NRGBA)
	c10 := color.NRGBAModel.Convert(source.At(x1, y0)).(color.NRGBA)
	c01 := color.NRGBAModel.Convert(source.At(x0, y1)).(color.NRGBA)
	c11 := color.NRGBAModel.Convert(source.At(x1, y1)).(color.NRGBA)

	return color.NRGBA{
		R: blendChannel(c00.R, c10.R, c01.R, c11.R, tx, ty),
		G: blendChannel(c00.G, c10.G, c01.G, c11.G, tx, ty),
		B: blendChannel(c00.B, c10.B, c01.B, c11.B, tx, ty),
		A: blendChannel(c00.A, c10.A, c01.A, c11.A, tx, ty),
	}
}

func blendChannel(c00, c10, c01, c11 uint8, tx, ty float64) uint8 {
	top := float64(c00)*(1-tx) + float64(c10)*tx
	bottom := float64(c01)*(1-tx) + float64(c11)*tx
	return uint8(math.Round(top*(1-ty) + bottom*ty))
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func buildExtraFiles(projectDir string, cfg contracts.PackagingConfig, projectConfig contracts.WailsProjectConfig, skipPaths ...string) string {
	var lines []string
	for _, asset := range config.EffectiveAssets(cfg, projectConfig, contracts.PlatformMacOS) {
		if strings.TrimSpace(asset.Src) == "" {
			continue
		}
		if shouldSkipAsset(asset.Src, skipPaths) {
			continue
		}
		source := filepath.ToSlash(asset.Src)
		requiredCheck := "echo \"Optional asset not found: " + source + "\" >&2"
		if asset.Required {
			requiredCheck = "echo \"Required asset not found: " + source + "\" >&2; exit 1"
		}
		// `cp -R src "$TMP_DIR/"` keeps a directory as "$TMP_DIR/<dirname>"
		// and puts a file directly in the package root. That matches the GUI
		// model: users only choose source assets, not custom target paths.
		lines = append(lines,
			fmt.Sprintf(`if [ -e %s ]; then`, shellQuote(source)),
			fmt.Sprintf(`  cp -R %s "$TMP_DIR/"`, shellQuote(source)),
			"else",
			"  "+requiredCheck,
			"fi",
		)
		if !fsx.Exists(fsx.Resolve(projectDir, asset.Src)) && asset.Required {
			lines = append(lines, fmt.Sprintf(`# Required asset currently missing during script generation: %s`, source))
		}
	}
	return strings.Join(lines, "\n")
}

func shouldSkipAsset(src string, skipPaths []string) bool {
	for _, skipPath := range skipPaths {
		if config.SameAssetPath(src, skipPath, contracts.PlatformMacOS) {
			return true
		}
	}
	return false
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}

const defaultTemplate = `#!/usr/bin/env bash
set -euo pipefail

APP_NAME="{{appName}}"
APP_BUNDLE="{{appBundle}}"
OUT_DIR="{{outputDir}}"
DMG_NAME="{{outputName}}.dmg"
BACKGROUND="{{background}}"
CREATE_DMG="{{createDmg}}"
FINAL_DMG="$OUT_DIR/$DMG_NAME"
VOLUME_NAME="$APP_NAME"

mkdir -p "$OUT_DIR"
rm -f "$FINAL_DMG"

if [ ! -d "$APP_BUNDLE" ]; then
  echo "App bundle not found: $APP_BUNDLE" >&2
  exit 1
fi

if ! command -v "$CREATE_DMG" >/dev/null 2>&1; then
  echo "create-dmg not found: $CREATE_DMG" >&2
  echo "Install it with: brew install create-dmg" >&2
  exit 1
fi

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT
cp -R "$APP_BUNDLE" "$TMP_DIR/$APP_NAME.app"
{{extraFiles}}

CREATE_DMG_ARGS=(
  --volname "$VOLUME_NAME"
  --window-size {{windowWidth}} {{windowHeight}}
  --icon-size {{iconSize}}
  --icon "$APP_NAME.app" {{appX}} {{appY}}
  --app-drop-link {{applicationsX}} {{applicationsY}}
)

if [ -f "$BACKGROUND" ]; then
  CREATE_DMG_ARGS+=(--background "$BACKGROUND")
fi

"$CREATE_DMG" "${CREATE_DMG_ARGS[@]}" "$FINAL_DMG" "$TMP_DIR"

echo "DMG created: $FINAL_DMG"
`
