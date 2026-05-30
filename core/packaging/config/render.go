package config

import (
	"strings"

	"wails3-manager/core/contracts"
	"wails3-manager/core/fsx"
)

// RenderPlaceholders resolves the small set of placeholders supported by
// packaging.json. Keeping this helper central avoids each packager inventing its
// own replacement rules.
func RenderPlaceholders(s string, cfg contracts.PackagingConfig) string {
	repl := map[string]string{
		"${project.name}":      cfg.Project.Name,
		"${project.version}":   fsx.StripVersionPrefix(cfg.Project.Version),
		"${project.bundleId}":  cfg.Project.BundleID,
		"${project.publisher}": cfg.Project.Publisher,
		"${build.appName}":     fsx.FirstNonEmpty(cfg.Build.AppName, fsx.SafeName(cfg.Project.Name)),
	}
	for k, v := range repl {
		s = strings.ReplaceAll(s, k, v)
	}
	return s
}
