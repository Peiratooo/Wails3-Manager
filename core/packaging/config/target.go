package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"wails3-manager/core/contracts"
)

const (
	DefaultWindowsAssetTarget = "/"
	DefaultMacOSAssetTarget   = "MacOS"
)

func AssetTarget(asset contracts.PackagingAsset, platform contracts.Platform) (string, error) {
	switch platform {
	case contracts.PlatformWindows:
		return WindowsAssetTarget(asset.Target)
	case contracts.PlatformMacOS:
		return MacOSAssetTarget(asset.Target)
	default:
		return "", fmt.Errorf("unsupported asset target platform: %s", platform)
	}
}

func ValidateAssetTargets(cfg contracts.PackagingConfig, platform contracts.Platform) error {
	for i, asset := range cfg.Assets {
		if _, err := AssetTarget(asset, platform); err != nil {
			return fmt.Errorf("packaging assets[%d].target: %w", i, err)
		}
	}
	return nil
}

func WindowsAssetTarget(target string) (string, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return DefaultWindowsAssetTarget, nil
	}
	if !strings.HasPrefix(target, "/") {
		return "", fmt.Errorf("Windows asset target must start with /: %s", target)
	}
	if strings.Contains(target, `\`) {
		return "", fmt.Errorf("Windows asset target must use / separators: %s", target)
	}
	parts, err := cleanTargetParts(strings.TrimPrefix(target, "/"), target)
	if err != nil {
		return "", err
	}
	if len(parts) == 0 {
		return DefaultWindowsAssetTarget, nil
	}
	return "/" + strings.Join(parts, "/"), nil
}

func MacOSAssetTarget(target string) (string, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return DefaultMacOSAssetTarget, nil
	}
	if strings.HasPrefix(target, "/") || strings.HasPrefix(target, `\`) {
		return "", fmt.Errorf("macOS asset target must be relative to Contents: %s", target)
	}
	if strings.Contains(target, `\`) {
		return "", fmt.Errorf("macOS asset target must use / separators: %s", target)
	}
	parts, err := cleanTargetParts(target, target)
	if err != nil {
		return "", err
	}
	if len(parts) == 0 {
		return DefaultMacOSAssetTarget, nil
	}
	return strings.Join(parts, "/"), nil
}

func WindowsDestDir(target string) string {
	target = strings.Trim(strings.TrimSpace(target), "/")
	if target == "" {
		return "{app}"
	}
	return `{app}\` + strings.ReplaceAll(target, "/", `\`)
}

func JoinWindowsAssetTarget(target, name string) string {
	target = WindowsDestDir(target)
	name = strings.Trim(strings.ReplaceAll(filepath.ToSlash(name), "/", `\`), `\`)
	if name == "" {
		return target
	}
	return target + `\` + name
}

func cleanTargetParts(raw, original string) ([]string, error) {
	if strings.ContainsAny(raw, `:*?"<>|`) {
		return nil, fmt.Errorf("asset target contains illegal characters: %s", original)
	}
	if raw == "" {
		return nil, nil
	}
	parts := strings.Split(raw, "/")
	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			return nil, fmt.Errorf("asset target cannot contain empty, . or .. segments: %s", original)
		}
	}
	return parts, nil
}
