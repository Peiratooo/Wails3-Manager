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

func Generate(projectDir string, cfg contracts.PackagingConfig) (string, error) {
	main := cfg.Entry
	outputDir := config.WindowsOutputDir(cfg)
	outputBase := config.RenderPlaceholders(cfg.Windows.OutputBaseName, cfg)
	if outputBase == "" {
		outputBase = config.RenderPlaceholders("${project.name}-${project.version}-windows-setup", cfg)
	}
	files := buildFiles(projectDir, cfg)
	icons := buildIcons(cfg, main)
	run := ""
	if main.ExecutablePath != "" {
		run = fmt.Sprintf(`Filename: "{app}\%s"; Description: "{cm:LaunchProgram,%s}"; Flags: nowait postinstall skipifsilent`, filepath.Base(main.ExecutablePath), cfg.Project.Name)
	}
	content := strings.NewReplacer(
		"{{appName}}", cfg.Project.Name,
		"{{appVersion}}", fsx.StripVersionPrefix(cfg.Project.Version),
		"{{appPublisher}}", cfg.Project.Publisher,
		"{{appURL}}", cfg.Project.Homepage,
		"{{appId}}", fsx.FirstNonEmpty(cfg.Project.BundleID, cfg.Project.Name),
		"{{defaultDirName}}", config.RenderPlaceholders(cfg.Windows.DefaultDirName, cfg),
		"{{privilegesRequired}}", fsx.FirstNonEmpty(cfg.Windows.PrivilegesRequired, "lowest"),
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

func buildFiles(projectDir string, cfg contracts.PackagingConfig) string {
	type fileLine struct {
		asset contracts.PackagingAsset
		main  bool
	}
	var items []fileLine
	main := cfg.Entry
	if main.ExecutablePath != "" {
		items = append(items, fileLine{asset: contracts.PackagingAsset{Src: main.ExecutablePath, Type: "file", Required: true}, main: true})
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
		if isDirectoryAsset(projectDir, asset) {
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

func isDirectoryAsset(projectDir string, asset contracts.PackagingAsset) bool {
	if strings.EqualFold(asset.Type, "directory") {
		return true
	}
	return asset.Type == "" && fsx.DirExists(fsx.Resolve(projectDir, asset.Src))
}

func buildIcons(cfg contracts.PackagingConfig, main contracts.ProgramEntry) string {
	if main.ExecutablePath == "" {
		return ""
	}
	exe := filepath.Base(main.ExecutablePath)
	var lines []string
	lines = append(lines, fmt.Sprintf(`Name: "{autoprograms}\%s"; Filename: "{app}\%s"`, cfg.Project.Name, exe))
	if cfg.Windows.CreateDesktopShortcut {
		lines = append(lines, fmt.Sprintf(`Name: "{autodesktop}\%s"; Filename: "{app}\%s"; Tasks: desktopicon`, cfg.Project.Name, exe))
	}
	return strings.Join(lines, "\n")
}

const defaultTemplate = `#define MyAppName "{{appName}}"
#define MyAppVersion "{{appVersion}}"
#define MyAppPublisher "{{appPublisher}}"
#define MyAppURL "{{appURL}}"

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
