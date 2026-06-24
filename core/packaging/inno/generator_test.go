package inno

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"wails3-manager/core/contracts"
)

func TestGenerateUsesBundleIDAndDefaultExecutable(t *testing.T) {
	projectDir := t.TempDir()
	projectConfig := contracts.WailsProjectConfig{
		Info: contracts.WailsAppInfo{
			CompanyName:       "My Company",
			ProductName:       "B4验机",
			ProductIdentifier: "com.mycompany.myproduct",
			Version:           "2.1.0",
		},
		Icon: "build/appicon.png",
	}
	cfg := contracts.PackagingConfig{
		SchemaVersion: 1,
		Build: contracts.BuildSettings{
			AppName: "demo",
		},
		Windows: contracts.WindowsConfig{
			InnoScript:            "builder/windows/inno.iss",
			DefaultDirName:        `{autopf}\${project.name}`,
			PrivilegesRequired:    "lowest",
			SetupIcon:             "build/windows/icon.ico",
			OutputBaseName:        "${build.appName}-${project.version}-windows-setup",
			AppURL:                "https://b4.cn/",
			CreateDesktopShortcut: true,
		},
		Artifacts: contracts.ArtifactConfig{OutputRoot: "builder/release"},
	}

	path, err := Generate(projectDir, cfg, projectConfig)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	script := string(data)

	for _, want := range []string{
		`#define MyAppURL "https://b4.cn/"`,
		`#define MyAppExeName "demo.exe"`,
		`OutputBaseFilename=B4验机-2.1.0-windows-setup`,
		"AppId=com.mycompany.myproduct",
		`UninstallDisplayIcon={app}\{#MyAppExeName}`,
		"DisableProgramGroupPage=yes",
		"SourceDir=" + filepath.ToSlash(projectDir),
		`Source: "bin/demo.exe"; DestDir: "{app}"; Flags: ignoreversion`,
		`Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked`,
		`Name: "{autoprograms}\B4验机"; Filename: "{app}\{#MyAppExeName}"`,
		`Filename: "{app}\{#MyAppExeName}"; Description: "{cm:LaunchProgram,{#StringChange(MyAppName, '&', '&&')}}"`,
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("generated script missing %q:\n%s", want, script)
		}
	}
	if strings.Contains(script, "[Registry]") || strings.Contains(script, "ChangesAssociations=yes") {
		t.Fatalf("generated script should not include association registry without fileAssociations:\n%s", script)
	}
	if strings.Contains(script, "[Languages]") {
		t.Fatalf("generated script should not include language configuration:\n%s", script)
	}
	assertNoInnoComments(t, script)
}

func TestGenerateUsesFileAssociations(t *testing.T) {
	projectDir := t.TempDir()
	projectConfig := contracts.WailsProjectConfig{
		Info: contracts.WailsAppInfo{
			CompanyName:       "My Company",
			ProductName:       "My Product",
			ProductIdentifier: "com.mycompany.myproduct",
			Version:           "0.0.1",
		},
		FileAssociations: []contracts.WailsFileAssociation{
			{Ext: "myp", Name: `My "Piano" File`},
		},
	}
	cfg := contracts.PackagingConfig{
		SchemaVersion: 1,
		Build: contracts.BuildSettings{
			AppName: "demo",
		},
		Windows: contracts.WindowsConfig{
			InnoScript:            "builder/windows/inno.iss",
			DefaultDirName:        `{autopf}\${project.name}`,
			PrivilegesRequired:    "lowest",
			SetupIcon:             "build/windows/icon.ico",
			OutputBaseName:        "${build.appName}-${project.version}-windows-setup",
			CreateDesktopShortcut: true,
		},
		Artifacts: contracts.ArtifactConfig{OutputRoot: "builder/release"},
	}

	path, err := Generate(projectDir, cfg, projectConfig)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	script := string(data)

	for _, want := range []string{
		"ChangesAssociations=yes",
		"[Registry]",
		`Root: HKA; Subkey: "Software\Classes\.myp\OpenWithProgids"; ValueType: string; ValueName: "com.mycompany.myproduct.myp"; ValueData: ""; Flags: uninsdeletevalue`,
		`Root: HKA; Subkey: "Software\Classes\com.mycompany.myproduct.myp"; ValueType: string; ValueName: ""; ValueData: "My ""Piano"" File"; Flags: uninsdeletekey`,
		`Root: HKA; Subkey: "Software\Classes\com.mycompany.myproduct.myp\DefaultIcon"; ValueType: string; ValueName: ""; ValueData: "{app}\{#MyAppExeName},0"`,
		`Root: HKA; Subkey: "Software\Classes\com.mycompany.myproduct.myp\shell\open\command"; ValueType: string; ValueName: ""; ValueData: """{app}\{#MyAppExeName}"" ""%1"""`,
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("generated script missing %q:\n%s", want, script)
		}
	}
	assertNoInnoComments(t, script)
}

func TestGenerateIncludesBuildOutputAndCustomLaunchProgram(t *testing.T) {
	projectDir := t.TempDir()
	projectConfig := contracts.WailsProjectConfig{
		Info: contracts.WailsAppInfo{
			CompanyName:       "My Company",
			ProductName:       "My Product",
			ProductIdentifier: "com.mycompany.myproduct",
			Version:           "0.0.1",
		},
	}
	cfg := contracts.PackagingConfig{
		SchemaVersion: 1,
		Build: contracts.BuildSettings{
			AppName: "demo",
		},
		Entry: contracts.ProgramEntry{
			ExecutablePath: "launcher/start.exe",
		},
		Windows: contracts.WindowsConfig{
			InnoScript:            "builder/windows/inno.iss",
			DefaultDirName:        `{autopf}\${project.name}`,
			PrivilegesRequired:    "lowest",
			SetupIcon:             "build/windows/icon.ico",
			OutputBaseName:        "${build.appName}-${project.version}-windows-setup",
			CreateDesktopShortcut: true,
		},
		Artifacts: contracts.ArtifactConfig{OutputRoot: "builder/release"},
	}

	path, err := Generate(projectDir, cfg, projectConfig)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	script := string(data)
	for _, want := range []string{
		`#define MyAppExeName "start.exe"`,
		`Source: "bin/demo.exe"; DestDir: "{app}"; Flags: ignoreversion`,
		`Source: "launcher/start.exe"; DestDir: "{app}"; Flags: ignoreversion`,
		`Filename: "{app}\{#MyAppExeName}"; Description: "{cm:LaunchProgram,{#StringChange(MyAppName, '&', '&&')}}"`,
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("generated script missing %q:\n%s", want, script)
		}
	}
	assertNoInnoComments(t, script)
}

func assertNoInnoComments(t *testing.T, script string) {
	t.Helper()
	for _, line := range strings.Split(script, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), ";") {
			t.Fatalf("generated script should not include comment line %q:\n%s", line, script)
		}
	}
}
