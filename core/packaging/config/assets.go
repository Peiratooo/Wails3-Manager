package config

import (
	"path/filepath"
	"strings"

	"wails3-manager/core/contracts"
)

func WailsBuildOutputPath(cfg contracts.PackagingConfig, project contracts.WailsProjectConfig, platform contracts.Platform) string {
	if platform == contracts.PlatformMacOS {
		appBundle := strings.TrimSpace(cfg.MacOS.AppBundle)
		if appBundle == "" {
			appBundle = DefaultExecutablePath(cfg, platform)
		}
		return RenderPlaceholders(appBundle, cfg, project)
	}
	return RenderPlaceholders(DefaultExecutablePath(cfg, platform), cfg, project)
}

func StartupExecutablePath(cfg contracts.PackagingConfig, project contracts.WailsProjectConfig, platform contracts.Platform) string {
	if platform == contracts.PlatformMacOS {
		return MacOSAppBundlePath(cfg, project)
	}
	return WindowsExecutablePath(cfg, project)
}

func EffectiveAssets(cfg contracts.PackagingConfig, project contracts.WailsProjectConfig, platform contracts.Platform) []contracts.PackagingAsset {
	out := []contracts.PackagingAsset{}
	seen := map[string]bool{}

	add := func(asset contracts.PackagingAsset) {
		asset.Src = strings.TrimSpace(RenderPlaceholders(asset.Src, cfg, project))
		if asset.Src == "" {
			return
		}
		key := assetPathKey(asset.Src, platform)
		if seen[key] {
			return
		}
		seen[key] = true
		out = append(out, asset)
	}

	add(wailsBuildOutputAsset(cfg, project, platform))
	add(startupAsset(cfg, project, platform))
	for _, asset := range cfg.Assets {
		add(asset)
	}
	return out
}

func SameAssetPath(left, right string, platform contracts.Platform) bool {
	left = strings.TrimSpace(left)
	right = strings.TrimSpace(right)
	if left == "" || right == "" {
		return false
	}
	return assetPathKey(left, platform) == assetPathKey(right, platform)
}

func wailsBuildOutputAsset(cfg contracts.PackagingConfig, project contracts.WailsProjectConfig, platform contracts.Platform) contracts.PackagingAsset {
	assetType := "file"
	if platform == contracts.PlatformMacOS {
		assetType = "directory"
	}
	return contracts.PackagingAsset{
		Src:      WailsBuildOutputPath(cfg, project, platform),
		Type:     assetType,
		Required: true,
	}
}

func startupAsset(cfg contracts.PackagingConfig, project contracts.WailsProjectConfig, platform contracts.Platform) contracts.PackagingAsset {
	path := StartupExecutablePath(cfg, project, platform)
	assetType := "file"
	if platform == contracts.PlatformMacOS && macOSStartupIsAppBundle(cfg, path) {
		assetType = "directory"
	}
	return contracts.PackagingAsset{
		Src:      path,
		Type:     assetType,
		Required: true,
	}
}

func macOSStartupIsAppBundle(cfg contracts.PackagingConfig, path string) bool {
	if strings.TrimSpace(cfg.Entry.ExecutablePath) == "" {
		return true
	}
	return strings.HasSuffix(strings.ToLower(strings.TrimRight(path, `/\`)), ".app")
}

func assetPathKey(path string, platform contracts.Platform) string {
	key := filepath.ToSlash(filepath.Clean(strings.TrimSpace(path)))
	if platform == contracts.PlatformWindows {
		key = strings.ToLower(key)
	}
	return key
}
