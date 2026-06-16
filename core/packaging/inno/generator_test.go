package inno

import (
	"os"
	"strings"
	"testing"

	"wails3-manager/core/contracts"
)

func TestGenerateUsesBundleIDAndDefaultExecutable(t *testing.T) {
	projectDir := t.TempDir()
	projectConfig := contracts.WailsProjectConfig{
		Info: contracts.WailsAppInfo{
			CompanyName:       "My Company",
			ProductName:       "My Product",
			ProductIdentifier: "com.mycompany.myproduct",
			Version:           "0.0.1",
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
		"AppId=com.mycompany.myproduct",
		`Source: "bin/demo.exe"; DestDir: "{app}"; Flags: ignoreversion`,
		`Name: "{autoprograms}\My Product"; Filename: "{app}\{#MyAppExeName}"`,
		`Filename: "{app}\{#MyAppExeName}"; Description: "{cm:LaunchProgram,My Product}"`,
		"OutputBaseFilename=demo-0.0.1-windows-setup",
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("generated script missing %q:\n%s", want, script)
		}
	}
}
