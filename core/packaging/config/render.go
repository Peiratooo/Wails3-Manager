package config

import (
	"path/filepath"
	"strings"

	"wails3-manager/core/contracts"
	"wails3-manager/core/fsx"
)

// RenderPlaceholders resolves the small set of placeholders supported by
// packaging.json. Keeping this helper central avoids each packager inventing its
// own replacement rules.
func RenderPlaceholders(s string, cfg contracts.PackagingConfig, project contracts.WailsProjectConfig) string {
	repl := map[string]string{
		"${project.name}":        ProjectName(project),
		"${project.version}":     ProjectVersion(project),
		"${project.bundleId}":    ProjectBundleID(project),
		"${project.publisher}":   ProjectPublisher(project),
		"${project.description}": ProjectDescription(project),
		"${project.copyright}":   ProjectCopyright(project),
		"${build.appName}":       AppName(cfg),
	}
	for k, v := range repl {
		s = strings.ReplaceAll(s, k, v)
	}
	return s
}

func ProjectName(project contracts.WailsProjectConfig) string {
	return strings.TrimSpace(project.Info.ProductName)
}

func ProjectVersion(project contracts.WailsProjectConfig) string {
	return fsx.StripVersionPrefix(project.Info.Version)
}

func ProjectBundleID(project contracts.WailsProjectConfig) string {
	return project.Info.ProductIdentifier
}

func ProjectPublisher(project contracts.WailsProjectConfig) string {
	return project.Info.CompanyName
}

func ProjectDescription(project contracts.WailsProjectConfig) string {
	return project.Info.Description
}

func ProjectCopyright(project contracts.WailsProjectConfig) string {
	return project.Info.Copyright
}

func AppName(cfg contracts.PackagingConfig) string {
	return strings.TrimSpace(cfg.Build.AppName)
}

func DefaultExecutablePath(cfg contracts.PackagingConfig, platform contracts.Platform) string {
	switch platform {
	case contracts.PlatformMacOS:
		return filepath.ToSlash(filepath.Join("bin", AppName(cfg)+".app"))
	default:
		return filepath.ToSlash(filepath.Join("bin", AppName(cfg)+".exe"))
	}
}

func DefaultMacOSBinaryPath(cfg contracts.PackagingConfig) string {
	return filepath.ToSlash(filepath.Join("bin", AppName(cfg)))
}

func DefaultMacOSAppBundlePath(cfg contracts.PackagingConfig, project contracts.WailsProjectConfig) string {
	return filepath.ToSlash(filepath.Join("bin", MacOSAppBundleName(project)+".app"))
}

func MacOSAppBundleName(project contracts.WailsProjectConfig) string {
	name := ProjectName(project)
	if name == "" {
		return "app"
	}
	name = strings.NewReplacer("/", "-", `\`, "-", ":", "-").Replace(name)
	name = strings.Trim(name, ". ")
	if name == "" {
		return "app"
	}
	return name
}

func ResolveExecutablePath(cfg contracts.PackagingConfig, project contracts.WailsProjectConfig, platform contracts.Platform) contracts.PackagingRuntimeInfo {
	defaultPath := DefaultExecutablePath(cfg, platform)
	configuredPath := strings.TrimSpace(cfg.Entry.ExecutablePath)
	if configuredPath == "" {
		return contracts.PackagingRuntimeInfo{
			DefaultExecutablePath:   defaultPath,
			EffectiveExecutablePath: defaultPath,
			UsingDefaultExecutable:  true,
		}
	}
	return contracts.PackagingRuntimeInfo{
		DefaultExecutablePath:   defaultPath,
		EffectiveExecutablePath: RenderPlaceholders(configuredPath, cfg, project),
		UsingDefaultExecutable:  false,
	}
}

func WindowsExecutablePath(cfg contracts.PackagingConfig, project contracts.WailsProjectConfig) string {
	return ResolveExecutablePath(cfg, project, contracts.PlatformWindows).EffectiveExecutablePath
}

func MacOSAppBundlePath(cfg contracts.PackagingConfig, project contracts.WailsProjectConfig) string {
	return ResolveMacOSAppBundlePath(cfg, project).EffectiveExecutablePath
}

func ResolveMacOSAppBundlePath(cfg contracts.PackagingConfig, project contracts.WailsProjectConfig) contracts.PackagingRuntimeInfo {
	defaultPath := DefaultMacOSAppBundlePath(cfg, project)
	configuredPath := strings.TrimSpace(cfg.MacOS.AppBundle)
	if configuredPath == "" || configuredPath == "bin/${build.appName}.app" {
		return contracts.PackagingRuntimeInfo{
			DefaultExecutablePath:   defaultPath,
			EffectiveExecutablePath: defaultPath,
			UsingDefaultExecutable:  true,
		}
	}

	effectivePath := RenderPlaceholders(configuredPath, cfg, project)
	return contracts.PackagingRuntimeInfo{
		DefaultExecutablePath:   defaultPath,
		EffectiveExecutablePath: effectivePath,
		UsingDefaultExecutable:  effectivePath == defaultPath,
	}
}
