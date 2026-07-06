package inno

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"wails3-manager/core/contracts"
	"wails3-manager/core/fsx"
	"wails3-manager/core/packaging/config"
)

func Generate(projectDir string, cfg contracts.PackagingConfig, projectConfig contracts.WailsProjectConfig) (string, error) {
	main := config.WindowsExecutablePath(cfg, projectConfig)
	outputDir := config.WindowsOutputDir(cfg, projectConfig)
	projectName := config.ProjectName(projectConfig)
	if projectName == "" {
		return "", fmt.Errorf("project productName is required for Windows packaging")
	}
	projectVersion := config.ProjectVersion(projectConfig)
	if projectVersion == "" {
		return "", fmt.Errorf("project version is required for Windows packaging")
	}
	outputTemplate := cfg.Windows.OutputBaseName
	if outputTemplate == "" || outputTemplate == "${build.appName}-${project.version}-windows-setup" {
		outputTemplate = "${project.name}-${project.version}-windows-setup"
	}
	outputBase := config.RenderPlaceholders(outputTemplate, cfg, projectConfig)
	appID := config.ProjectBundleID(projectConfig)
	if strings.TrimSpace(appID) == "" {
		return "", fmt.Errorf("project productIdentifier is required for Windows packaging")
	}
	files, err := buildFiles(cfg, projectConfig)
	if err != nil {
		return "", err
	}
	icons := buildIcons(cfg, projectConfig, main)
	registry := buildRegistry(projectConfig)
	changesAssociations := ""
	if registry != "" {
		changesAssociations = "ChangesAssociations=yes"
	}
	run := ""
	appExeName := filepath.Base(main)
	if main != "" {
		run = `Filename: "{app}\{#MyAppExeName}"; Description: "{cm:LaunchProgram,{#StringChange(MyAppName, '&', '&&')}}"; Flags: nowait postinstall skipifsilent`
	}
	content := strings.NewReplacer(
		"{{appName}}", projectName,
		"{{appVersion}}", projectVersion,
		"{{appPublisher}}", config.ProjectPublisher(projectConfig),
		"{{appURL}}", config.RenderPlaceholders(cfg.Windows.AppURL, cfg, projectConfig),
		"{{appExeName}}", appExeName,
		"{{appId}}", appID,
		"{{defaultDirName}}", config.RenderPlaceholders(cfg.Windows.DefaultDirName, cfg, projectConfig),
		"{{privilegesRequired}}", cfg.Windows.PrivilegesRequired,
		"{{changesAssociations}}", changesAssociations,
		"{{sourceDir}}", filepath.ToSlash(projectDir),
		"{{outputDir}}", filepath.ToSlash(fsx.Resolve(projectDir, outputDir)),
		"{{outputBaseFilename}}", outputBase,
		"{{setupIconFile}}", filepath.ToSlash(fsx.Resolve(projectDir, cfg.Windows.SetupIcon)),
		"{{files}}", files,
		"{{registry}}", registry,
		"{{icons}}", icons,
		"{{run}}", run,
	).Replace(defaultTemplate)
	path := fsx.Resolve(projectDir, cfg.Windows.InnoScript)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", err
	}
	return path, os.WriteFile(path, []byte(content), 0644)
}

func buildFiles(cfg contracts.PackagingConfig, projectConfig contracts.WailsProjectConfig) (string, error) {
	var lines []string
	for i, asset := range config.EffectiveAssets(cfg, projectConfig, contracts.PlatformWindows) {
		if strings.TrimSpace(asset.Src) == "" {
			continue
		}
		target, err := config.AssetTarget(asset, contracts.PlatformWindows)
		if err != nil {
			return "", fmt.Errorf("Windows asset %d: %w", i, err)
		}
		source := asset.Src
		destDir := config.WindowsDestDir(target)
		flags := []string{"ignoreversion"}
		if isDirectoryAsset(asset) {
			source = filepath.ToSlash(filepath.Join(asset.Src, "*"))
			destDir = config.JoinWindowsAssetTarget(target, filepath.Base(strings.TrimRight(asset.Src, `/\`)))
			flags = append(flags, "recursesubdirs", "createallsubdirs")
		}
		if !asset.Required {
			flags = append(flags, "skipifsourcedoesntexist")
		}
		lines = append(lines, fmt.Sprintf(`Source: "%s"; DestDir: "%s"; Flags: %s`, filepath.ToSlash(source), destDir, strings.Join(flags, " ")))
	}
	return strings.Join(lines, "\n"), nil
}

func isDirectoryAsset(asset contracts.PackagingAsset) bool {
	return asset.Type == "directory"
}

func buildIcons(cfg contracts.PackagingConfig, projectConfig contracts.WailsProjectConfig, main string) string {
	if main == "" {
		return ""
	}
	projectName := config.ProjectName(projectConfig)
	var lines []string
	lines = append(lines, fmt.Sprintf(`Name: "{autoprograms}\%s"; Filename: "{app}\{#MyAppExeName}"`, projectName))
	if cfg.Windows.CreateDesktopShortcut {
		lines = append(lines, fmt.Sprintf(`Name: "{autodesktop}\%s"; Filename: "{app}\{#MyAppExeName}"; Tasks: desktopicon`, projectName))
	}
	return strings.Join(lines, "\n")
}

func buildRegistry(projectConfig contracts.WailsProjectConfig) string {
	projectName := config.ProjectName(projectConfig)
	var lines []string
	for _, assoc := range projectConfig.FileAssociations {
		ext := normalizeAssocExt(assoc.Ext)
		if ext == "" {
			continue
		}
		name := strings.TrimSpace(assoc.Name)
		if name == "" {
			name = strings.TrimSpace(assoc.Description)
		}
		if name == "" {
			name = projectName + " " + strings.TrimPrefix(ext, ".") + " File"
		}
		key := assocRegistryKey(projectConfig, ext)
		lines = append(lines,
			fmt.Sprintf(`Root: HKA; Subkey: "Software\Classes\%s\OpenWithProgids"; ValueType: string; ValueName: "%s"; ValueData: ""; Flags: uninsdeletevalue`, escapeInnoString(ext), escapeInnoString(key)),
			fmt.Sprintf(`Root: HKA; Subkey: "Software\Classes\%s"; ValueType: string; ValueName: ""; ValueData: "%s"; Flags: uninsdeletekey`, escapeInnoString(key), escapeInnoString(name)),
			fmt.Sprintf(`Root: HKA; Subkey: "Software\Classes\%s\DefaultIcon"; ValueType: string; ValueName: ""; ValueData: "{app}\{#MyAppExeName},0"`, escapeInnoString(key)),
			fmt.Sprintf(`Root: HKA; Subkey: "Software\Classes\%s\shell\open\command"; ValueType: string; ValueName: ""; ValueData: """{app}\{#MyAppExeName}"" ""%%1"""`, escapeInnoString(key)),
		)
	}
	if len(lines) == 0 {
		return ""
	}
	return "\n[Registry]\n" + strings.Join(lines, "\n")
}

func normalizeAssocExt(ext string) string {
	ext = strings.TrimSpace(ext)
	if ext == "" {
		return ""
	}
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	return ext
}

func assocRegistryKey(projectConfig contracts.WailsProjectConfig, ext string) string {
	base := strings.TrimSpace(config.ProjectBundleID(projectConfig))
	if base == "" {
		base = strings.TrimSpace(config.ProjectName(projectConfig))
	}
	return cleanRegistryKey(base + ext)
}

func cleanRegistryKey(value string) string {
	var b strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9',
			r == '.', r == '_', r == '-':
			b.WriteRune(r)
		}
	}
	out := strings.Trim(b.String(), ".")
	if out == "" {
		return "WailsApp.File"
	}
	return out
}

func escapeInnoString(value string) string {
	return strings.ReplaceAll(value, `"`, `""`)
}

const defaultTemplate = `#define MyAppName "{{appName}}"
#define MyAppVersion "{{appVersion}}"
#define MyAppPublisher "{{appPublisher}}"
#define MyAppURL "{{appURL}}"
#define MyAppExeName "{{appExeName}}"

[Setup]
AppId={{appId}}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
AppSupportURL={#MyAppURL}
AppUpdatesURL={#MyAppURL}
DefaultDirName={{defaultDirName}}
DefaultGroupName={#MyAppName}
AllowNoIcons=yes
UninstallDisplayIcon={app}\{#MyAppExeName}
PrivilegesRequired={{privilegesRequired}}
DisableProgramGroupPage=yes
{{changesAssociations}}
SourceDir={{sourceDir}}
OutputDir={{outputDir}}
OutputBaseFilename={{outputBaseFilename}}
SetupIconFile={{setupIconFile}}
Compression=lzma
SolidCompression=yes
WizardStyle=modern

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked

[Files]
{{files}}
{{registry}}

[Icons]
{{icons}}

[Run]
{{run}}
`
