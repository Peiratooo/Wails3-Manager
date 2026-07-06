package creator

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type commandRunnerFunc func(context.Context, string, []string) error

func (f commandRunnerFunc) Run(ctx context.Context, workDir string, command []string) error {
	return f(ctx, workDir, command)
}

func TestCreateWailsProjectUsesSelectedDirForWailsAndReturnsProjectDir(t *testing.T) {
	parentDir := t.TempDir()
	projectName := "my-wails-app"
	var commands [][]string
	var workDirs []string

	projectDir, err := CreateWailsProject(
		context.Background(),
		CreateWailsProjectRequest{
			Name:               projectName,
			Dir:                parentDir,
			Template:           "vanilla",
			ProductName:        "My Product",
			ProductDescription: "My Product Description",
			ProductVersion:     "0.1.0",
			ProductCompany:     "My Company",
			ProductCopyright:   "Copyright (c) 2026",
			ProductComments:    "This is a comment",
			ProductIdentifier:  "com.mycompany.myproduct",
		},
		commandRunnerFunc(func(_ context.Context, gotWorkDir string, gotCommand []string) error {
			workDirs = append(workDirs, gotWorkDir)
			commands = append(commands, append([]string(nil), gotCommand...))
			if len(commands) == 1 {
				return writeGeneratedProject(filepath.Join(parentDir, projectName))
			}
			return nil
		}),
	)
	if err != nil {
		t.Fatal(err)
	}

	wantProjectDir := filepath.Clean(filepath.Join(parentDir, projectName))
	if projectDir != wantProjectDir {
		t.Fatalf("projectDir = %q, want %q", projectDir, wantProjectDir)
	}
	if len(commands) != 2 {
		t.Fatalf("commands = %#v, want init and update build assets", commands)
	}
	if workDirs[0] != filepath.Clean(parentDir) {
		t.Fatalf("workDir = %q, want %q", workDirs[0], parentDir)
	}
	if got := flagValue(commands[0], "-d"); got != filepath.Clean(parentDir) {
		t.Fatalf("-d = %q, want %q", got, parentDir)
	}
	if strings.Join(commands[1], " ") != "wails3 task common:update:build-assets" {
		t.Fatalf("update command = %#v", commands[1])
	}

	data, err := os.ReadFile(filepath.Join(projectDir, "build", "config.yml"))
	if err != nil {
		t.Fatal(err)
	}
	config := string(data)
	for _, want := range []string{
		`companyName: "My Company"`,
		`productName: "My Product"`,
		`productIdentifier: "com.mycompany.myproduct"`,
		`description: "My Product Description"`,
		`copyright: "Copyright (c) 2026"`,
		`comments: "This is a comment"`,
		`version: "0.1.0"`,
	} {
		if !strings.Contains(config, want) {
			t.Fatalf("config missing %q\n%s", want, config)
		}
	}
}

func TestTargetProjectDirUsesSelectedDirAsParent(t *testing.T) {
	parentDir := t.TempDir()

	targetDir, err := targetProjectDir(CreateWailsProjectRequest{
		Name: "my-wails-app",
		Dir:  parentDir,
	})
	if err != nil {
		t.Fatal(err)
	}

	want := filepath.Clean(filepath.Join(parentDir, "my-wails-app"))
	if targetDir != want {
		t.Fatalf("targetDir = %q, want %q", targetDir, want)
	}
}

func TestBuildInitCommandUsesTemplateValue(t *testing.T) {
	templateURL := "https://github.com/example/wails-template"
	command := buildInitCommand(CreateWailsProjectRequest{
		Name:        "remote-template-app",
		Template:    templateURL,
		PackageName: "main",
	}, t.TempDir())

	if got := flagValue(command, "-t"); got != templateURL {
		t.Fatalf("-t = %q, want %q", got, templateURL)
	}
}

func flagValue(command []string, flag string) string {
	for i := 0; i < len(command)-1; i++ {
		if command[i] == flag {
			return command[i+1]
		}
	}
	return ""
}

func writeGeneratedProject(projectDir string) error {
	if err := os.MkdirAll(filepath.Join(projectDir, "build"), 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(projectDir, "build", "config.yml"), []byte(`version: '3'
info:
  companyName: "My Company"
  productName: "My Product"
  productIdentifier: "com.mycompany.myproduct"
  description: "A program that does X"
  copyright: "(c) 2025, My Company"
  comments: "Some Product Comments"
  version: "0.0.1"
`), 0644)
}

func TestTargetProjectDirRejectsInvalidProjectNames(t *testing.T) {
	parentDir := t.TempDir()

	for _, name := range []string{
		".",
		"..",
		"bad/name",
		`bad\name`,
		"bad:name",
		"CON",
		"COM1.txt",
		"app.",
	} {
		t.Run(name, func(t *testing.T) {
			_, err := targetProjectDir(CreateWailsProjectRequest{
				Name: name,
				Dir:  parentDir,
			})
			if err == nil {
				t.Fatal("expected invalid project name error")
			}
		})
	}
}
