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

	"wails3-manager/core/contracts"
	"wails3-manager/core/fsx"
	"wails3-manager/core/packaging/config"

	xdraw "golang.org/x/image/draw"
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
	content := strings.NewReplacer(
		"{{appName}}", projectName,
		"{{appBundle}}", filepath.ToSlash(appBundle),
		"{{outputDir}}", filepath.ToSlash(outputDir),
		"{{outputName}}", outputName,
		"{{windowWidth}}", fmt.Sprint(cfg.MacOS.WindowWidth),
		"{{windowHeight}}", fmt.Sprint(cfg.MacOS.WindowHeight),
		"{{iconSize}}", fmt.Sprint(cfg.MacOS.IconSize),
		"{{appX}}", fmt.Sprint(cfg.MacOS.AppX),
		"{{appY}}", fmt.Sprint(cfg.MacOS.AppY),
		"{{applicationsX}}", fmt.Sprint(cfg.MacOS.ApplicationsX),
		"{{applicationsY}}", fmt.Sprint(cfg.MacOS.ApplicationsY),
		"{{createDmg}}", cfg.MacOS.CreateDMGPath,
	).Replace(defaultTemplate)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", err
	}
	return path, os.WriteFile(path, []byte(content), 0755)
}

func PrepareBackground(projectDir, background string, width, height int, output string) (string, error) {
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
	if strings.TrimSpace(output) == "" {
		output = filepath.Join(projectDir, "builder", "macos", "background.png")
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
	xdraw.BiLinear.Scale(output, output.Bounds(), source, crop, xdraw.Src, nil)
	return output
}

const defaultTemplate = `#!/usr/bin/env bash
set -euo pipefail

APP_NAME="{{appName}}"
APP_BUNDLE="{{appBundle}}"
OUT_DIR="{{outputDir}}"
DMG_NAME="{{outputName}}.dmg"
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
BACKGROUND="$TMP_DIR/$APP_NAME.app/Contents/Resources/dmg-background.png"

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
