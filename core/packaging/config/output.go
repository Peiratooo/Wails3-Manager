package config

import (
	"path/filepath"
	"strings"

	"wails3-manager/core/contracts"
	"wails3-manager/core/fsx"
)

// ExportRoot is the single user-facing release directory.
func ExportRoot(cfg contracts.PackagingConfig) string {
	return RenderPlaceholders(fsx.FirstNonEmpty(cfg.Artifacts.OutputRoot, "builder/release"), cfg)
}

func WindowsOutputDir(cfg contracts.PackagingConfig) string {
	root := strings.TrimSpace(ExportRoot(cfg))
	if root == "" {
		root = "builder/release"
	}
	return filepath.ToSlash(filepath.Join(root, "windows"))
}

func MacOSOutputDir(cfg contracts.PackagingConfig) string {
	root := strings.TrimSpace(ExportRoot(cfg))
	if root == "" {
		root = "builder/release"
	}
	return filepath.ToSlash(filepath.Join(root, "macos"))
}
