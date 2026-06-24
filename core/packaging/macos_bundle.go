package packaging

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"wails3-manager/core/contracts"
	"wails3-manager/core/fsx"
	"wails3-manager/core/packaging/config"
	"wails3-manager/core/packaging/dmg"
)

func prepareMacOSAppBundle(projectDir string, cfg contracts.PackagingConfig, projectConfig contracts.WailsProjectConfig) (string, error) {
	appBundle := fsx.Resolve(projectDir, config.MacOSAppBundlePath(cfg, projectConfig))
	if appBundle == "" {
		return "", fmt.Errorf("macOS app bundle path is not configured")
	}
	if !strings.HasSuffix(strings.ToLower(strings.TrimRight(appBundle, `/\`)), ".app") {
		return "", fmt.Errorf("macOS app bundle path must end with .app: %s", appBundle)
	}

	contentsDir := filepath.Join(appBundle, "Contents")
	macOSDir := filepath.Join(contentsDir, "MacOS")
	resourcesDir := filepath.Join(contentsDir, "Resources")
	if err := os.RemoveAll(contentsDir); err != nil {
		return "", err
	}
	for _, dir := range []string{macOSDir, resourcesDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", err
		}
	}

	iconSrc := filepath.Join(projectDir, "build", "darwin", "icons.icns")
	iconDst := filepath.Join(resourcesDir, "icons.icns")
	if err := copyRequiredFile(iconSrc, iconDst, "macOS icon", 0644); err != nil {
		return "", err
	}
	if strings.TrimSpace(cfg.MacOS.Background) != "" {
		background := config.RenderPlaceholders(cfg.MacOS.Background, cfg, projectConfig)
		if _, err := dmg.PrepareBackground(projectDir, background, cfg.MacOS.WindowWidth, cfg.MacOS.WindowHeight, filepath.Join(resourcesDir, "dmg-background.png")); err != nil {
			return "", err
		}
	}

	plist, err := renderMacOSInfoPlist(projectDir, cfg, projectConfig)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(contentsDir, "Info.plist"), []byte(plist), 0644); err != nil {
		return "", err
	}

	if err := copyMacOSPayloads(projectDir, contentsDir, cfg, projectConfig); err != nil {
		return "", err
	}

	binarySrc := fsx.Resolve(projectDir, config.DefaultMacOSBinaryPath(cfg))
	binaryDst := filepath.Join(macOSDir, config.AppName(cfg))
	if err := copyRequiredFile(binarySrc, binaryDst, "macOS executable", 0755); err != nil {
		return "", err
	}

	return appBundle, nil
}

func renderMacOSInfoPlist(projectDir string, cfg contracts.PackagingConfig, projectConfig contracts.WailsProjectConfig) (string, error) {
	path := filepath.Join(projectDir, "build", "darwin", "Info.plist")
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	text := string(data)
	if text == "" {
		text = defaultMacOSInfoPlist()
	}

	info := projectConfig.Info
	updates := map[string]string{
		"CFBundleName":               config.ProjectName(projectConfig),
		"CFBundleDisplayName":        config.ProjectName(projectConfig),
		"CFBundleExecutable":         config.AppName(cfg),
		"CFBundleIdentifier":         config.ProjectBundleID(projectConfig),
		"CFBundleVersion":            config.ProjectVersion(projectConfig),
		"CFBundleShortVersionString": config.ProjectVersion(projectConfig),
		"CFBundleIconFile":           "icons",
	}
	if strings.TrimSpace(info.Copyright) != "" {
		updates["NSHumanReadableCopyright"] = info.Copyright
	}
	if strings.TrimSpace(info.Comments) != "" {
		updates["CFBundleGetInfoString"] = info.Comments
	}
	for key, value := range updates {
		if strings.TrimSpace(value) == "" {
			continue
		}
		text = setPlistString(text, key, value)
	}
	return text, nil
}

func copyMacOSPayloads(projectDir, contentsDir string, cfg contracts.PackagingConfig, projectConfig contracts.WailsProjectConfig) error {
	macOSDir := filepath.Join(contentsDir, "MacOS")
	for i, asset := range cfg.Assets {
		src := strings.TrimSpace(config.RenderPlaceholders(asset.Src, cfg, projectConfig))
		if src == "" {
			continue
		}
		target, err := config.AssetTarget(asset, contracts.PlatformMacOS)
		if err != nil {
			return fmt.Errorf("macOS asset %d: %w", i, err)
		}
		targetDir := filepath.Join(contentsDir, filepath.FromSlash(target))
		if err := copyMacOSPayload(projectDir, targetDir, src, asset.Type, asset.Required); err != nil {
			return fmt.Errorf("macOS asset %d: %w", i, err)
		}
	}
	entry := strings.TrimSpace(config.RenderPlaceholders(cfg.Entry.ExecutablePath, cfg, projectConfig))
	if entry != "" &&
		!config.SameAssetPath(entry, config.DefaultMacOSBinaryPath(cfg), contracts.PlatformMacOS) &&
		!config.SameAssetPath(entry, config.MacOSAppBundlePath(cfg, projectConfig), contracts.PlatformMacOS) {
		if err := copyMacOSPayload(projectDir, macOSDir, entry, "", true); err != nil {
			return fmt.Errorf("macOS launch program: %w", err)
		}
	}
	return nil
}

func copyMacOSPayload(projectDir, macOSDir, src, assetType string, required bool) error {
	from := fsx.Resolve(projectDir, src)
	if from == "" {
		if required {
			return fmt.Errorf("asset path is empty")
		}
		return nil
	}
	if !fsx.Exists(from) {
		if required {
			return fmt.Errorf("asset not found: %s", from)
		}
		return nil
	}
	name := filepath.Base(strings.TrimRight(from, `/\`))
	if name == "." || name == string(filepath.Separator) {
		return fmt.Errorf("asset has no filename: %s", from)
	}
	to := filepath.Join(macOSDir, name)
	if err := os.RemoveAll(to); err != nil {
		return err
	}
	if assetType == "" {
		assetType = "file"
		if fsx.DirExists(from) {
			assetType = "directory"
		}
	}
	switch assetType {
	case "directory":
		if !fsx.DirExists(from) {
			return fmt.Errorf("asset is not a directory: %s", from)
		}
		return fsx.CopyDir(from, to, true)
	case "file":
		if !fsx.FileExists(from) {
			return fmt.Errorf("asset is not a file: %s", from)
		}
		info, err := os.Stat(from)
		if err != nil {
			return err
		}
		return fsx.CopyFile(from, to, info.Mode())
	default:
		return fmt.Errorf("asset type must be file or directory")
	}
}

func copyRequiredFile(src, dst, label string, defaultMode os.FileMode) error {
	if !fsx.FileExists(src) {
		return fmt.Errorf("%s not found: %s", label, src)
	}
	mode := defaultMode
	if info, err := os.Stat(src); err == nil {
		mode = info.Mode()
		if defaultMode&0111 != 0 {
			mode |= 0111
		}
	}
	if err := fsx.CopyFile(src, dst, mode); err != nil {
		return err
	}
	return os.Chmod(dst, mode)
}

func setPlistString(text, key, value string) string {
	re := regexp.MustCompile(`(?s)(<key>` + regexp.QuoteMeta(key) + `</key>\s*<string>)(.*?)(</string>)`)
	if re.MatchString(text) {
		return re.ReplaceAllStringFunc(text, func(match string) string {
			parts := re.FindStringSubmatch(match)
			return parts[1] + xmlEscape(value) + parts[3]
		})
	}
	insertAt := strings.LastIndex(text, "</dict>")
	if insertAt < 0 {
		return text
	}
	insert := fmt.Sprintf("        <key>%s</key>\n        <string>%s</string>\n", key, xmlEscape(value))
	return text[:insertAt] + insert + text[insertAt:]
}

func xmlEscape(value string) string {
	var b strings.Builder
	_ = xml.EscapeText(&b, []byte(value))
	return b.String()
}

func defaultMacOSInfoPlist() string {
	return `<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
    <dict>
        <key>CFBundlePackageType</key>
        <string>APPL</string>
        <key>NSHighResolutionCapable</key>
        <string>true</string>
    </dict>
</plist>
`
}
