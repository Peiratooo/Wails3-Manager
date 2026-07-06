package config

import (
	"path/filepath"
	"strings"

	"wails3-manager/core/contracts"
)

// ExportRoot is the single user-facing release directory.
func ExportRoot(cfg contracts.PackagingConfig, project contracts.WailsProjectConfig) string {
	return RenderPlaceholders(cfg.Artifacts.OutputRoot, cfg, project)
}

func WindowsOutputDir(cfg contracts.PackagingConfig, project contracts.WailsProjectConfig) string {
	root := strings.TrimSpace(ExportRoot(cfg, project))
	return filepath.ToSlash(filepath.Join(root, "windows"))
}

func MacOSOutputDir(cfg contracts.PackagingConfig, project contracts.WailsProjectConfig) string {
	root := strings.TrimSpace(ExportRoot(cfg, project))
	return filepath.ToSlash(filepath.Join(root, "macos"))
}
