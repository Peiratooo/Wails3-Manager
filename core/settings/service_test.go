package settings

import (
	"os"
	"path/filepath"
	"testing"

	"wails3-manager/core/contracts"
	"wails3-manager/core/project"
)

func TestRemoveProjectDeletesStaleRecordWhenProjectDirectoryIsGone(t *testing.T) {
	configRoot := t.TempDir()
	t.Setenv("APPDATA", configRoot)
	t.Setenv("XDG_CONFIG_HOME", configRoot)

	projectDir := filepath.Join(t.TempDir(), "deleted-project")
	if err := project.SaveUserState(contracts.UserState{
		Projects: []contracts.ProjectRecord{{ProjectDir: projectDir}},
	}); err != nil {
		t.Fatal(err)
	}

	if err := NewService(nil).RemoveProject(projectDir, true); err != nil {
		t.Fatal(err)
	}

	state := project.LoadUserState()
	if len(state.Projects) != 0 {
		t.Fatalf("stale project record was not removed: %#v", state.Projects)
	}
}

func TestRemoveProjectCleansStaleBuilderButKeepsNonEmptyProjectDir(t *testing.T) {
	configRoot := t.TempDir()
	t.Setenv("APPDATA", configRoot)
	t.Setenv("XDG_CONFIG_HOME", configRoot)

	projectDir := filepath.Join(t.TempDir(), "partially-deleted-project")
	builderDir := filepath.Join(projectDir, "builder")
	if err := os.MkdirAll(builderDir, 0755); err != nil {
		t.Fatal(err)
	}
	sourceFile := filepath.Join(projectDir, "main.go")
	if err := os.WriteFile(sourceFile, []byte("package main\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(builderDir, "stale.txt"), []byte("stale"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := project.SaveUserState(contracts.UserState{
		Projects: []contracts.ProjectRecord{{ProjectDir: projectDir}},
	}); err != nil {
		t.Fatal(err)
	}

	if err := NewService(nil).RemoveProject(projectDir, true); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(builderDir); !os.IsNotExist(err) {
		t.Fatalf("stale builder dir still exists, stat err=%v", err)
	}
	if _, err := os.Stat(sourceFile); err != nil {
		t.Fatalf("non-empty project dir content should remain: %v", err)
	}
}
