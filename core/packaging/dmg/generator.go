package dmg

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"wails3-manager/core/contracts"
	"wails3-manager/core/fsx"
	"wails3-manager/core/packaging/config"
)

func DefaultBackgroundPNG() []byte {
	// 1x1 transparent PNG placeholder; users can replace it from GUI.
	data, _ := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+/p9sAAAAASUVORK5CYII=")
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
	extraFiles := buildExtraFiles(projectDir, cfg)
	content := strings.NewReplacer(
		"{{appName}}", projectName,
		"{{appBundle}}", filepath.ToSlash(appBundle),
		"{{outputDir}}", filepath.ToSlash(outputDir),
		"{{outputName}}", outputName,
		"{{background}}", filepath.ToSlash(config.RenderPlaceholders(cfg.MacOS.Background, cfg, projectConfig)),
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

func buildExtraFiles(projectDir string, cfg contracts.PackagingConfig) string {
	var lines []string
	for _, asset := range cfg.Assets {
		if strings.TrimSpace(asset.Src) == "" {
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
