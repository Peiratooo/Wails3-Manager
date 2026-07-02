package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadManagerWithConfigCompletionWritesMissingInfoKeys(t *testing.T) {
	projectDir := t.TempDir()
	buildDir := filepath.Join(projectDir, "build")
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(buildDir, "config.yml")
	if err := os.WriteFile(path, []byte("version: '3'\ninfo:\n  productName: \"Demo\"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	manager, err := LoadManagerWithConfigCompletion(projectDir)
	if err != nil {
		t.Fatal(err)
	}
	if manager.WailsConfig.Info.ProductName != "Demo" {
		t.Fatalf("productName = %q, want Demo", manager.WailsConfig.Info.ProductName)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, key := range []string{
		"companyName:",
		"productName:",
		"productIdentifier:",
		"description:",
		"copyright:",
		"comments:",
		"version:",
	} {
		if !strings.Contains(text, key) {
			t.Fatalf("completed config is missing %s\n%s", key, text)
		}
	}
	if !strings.Contains(text, `version: "0.0.1"`) {
		t.Fatalf("missing version was not completed with 0.0.1\n%s", text)
	}
}

func TestLoadTaskVarsUsesYamlTaskfile(t *testing.T) {
	projectDir := t.TempDir()
	taskfile := `version: '3'

vars:
  APP_NAME: "yaml-app"
  CGO_ENABLED: "1"
  PRODUCTION: "true"
`
	if err := os.WriteFile(filepath.Join(projectDir, "Taskfile.yaml"), []byte(taskfile), 0644); err != nil {
		t.Fatal(err)
	}

	vars := LoadTaskVars(projectDir)
	if vars.AppName != "yaml-app" || vars.CGOEnabled != "1" || !vars.Production {
		t.Fatalf("LoadTaskVars() = %#v", vars)
	}
}
