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
	outputBase := config.RenderPlaceholders(cfg.Windows.OutputBaseName, cfg, projectConfig)
	projectName := config.ProjectName(projectConfig)
	if projectName == "" {
		return "", fmt.Errorf("project productName is required for Windows packaging")
	}
	projectVersion := config.ProjectVersion(projectConfig)
	if projectVersion == "" {
		return "", fmt.Errorf("project version is required for Windows packaging")
	}
	appID := config.ProjectBundleID(projectConfig)
	if strings.TrimSpace(appID) == "" {
		return "", fmt.Errorf("project productIdentifier is required for Windows packaging")
	}
	files := buildFiles(cfg, projectConfig)
	icons := buildIcons(cfg, projectConfig, main)
	run := ""
	appExeName := filepath.Base(main)
	if main != "" {
		run = fmt.Sprintf(`Filename: "{app}\{#MyAppExeName}"; Description: "{cm:LaunchProgram,%s}"; Flags: nowait postinstall skipifsilent`, projectName)
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
		"{{outputDir}}", filepath.ToSlash(fsx.Resolve(projectDir, outputDir)),
		"{{outputBaseFilename}}", outputBase,
		"{{setupIconFile}}", filepath.ToSlash(fsx.Resolve(projectDir, cfg.Windows.SetupIcon)),
		"{{files}}", files,
		"{{icons}}", icons,
		"{{run}}", run,
	).Replace(defaultTemplate)
	path := fsx.Resolve(projectDir, cfg.Windows.InnoScript)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", err
	}
	return path, os.WriteFile(path, []byte(content), 0644)
}

func buildFiles(cfg contracts.PackagingConfig, projectConfig contracts.WailsProjectConfig) string {
	type fileLine struct {
		asset contracts.PackagingAsset
		main  bool
	}
	var items []fileLine
	main := config.WindowsExecutablePath(cfg, projectConfig)
	if main != "" {
		items = append(items, fileLine{asset: contracts.PackagingAsset{Src: main, Type: "file", Required: true}, main: true})
	}
	for _, asset := range cfg.Assets {
		items = append(items, fileLine{asset: asset})
	}
	var lines []string
	for _, line := range items {
		asset := line.asset
		if strings.TrimSpace(asset.Src) == "" {
			continue
		}
		source := asset.Src
		target := "{app}"
		flags := []string{"ignoreversion"}
		if isDirectoryAsset(asset) {
			source = filepath.ToSlash(filepath.Join(asset.Src, "*"))
			target = `{app}\` + strings.ReplaceAll(filepath.Base(strings.TrimRight(asset.Src, `/\`)), "/", `\`)
			flags = append(flags, "recursesubdirs", "createallsubdirs")
		}
		if !asset.Required {
			flags = append(flags, "skipifsourcedoesntexist")
		}
		lines = append(lines, fmt.Sprintf(`Source: "%s"; DestDir: "%s"; Flags: %s`, filepath.ToSlash(source), target, strings.Join(flags, " ")))
	}
	return strings.Join(lines, "\n")
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
PrivilegesRequired={{privilegesRequired}}
OutputDir={{outputDir}}
OutputBaseFilename={{outputBaseFilename}}
SetupIconFile={{setupIconFile}}
Compression=lzma
SolidCompression=yes
WizardStyle=modern

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "desktopicon"; Description: "Create a &desktop shortcut"; GroupDescription: "Additional icons:"; Flags: unchecked

[Files]
{{files}}

[Icons]
{{icons}}

[Run]
{{run}}
`
