package settings

import (
	"os"
	"path/filepath"
	"testing"

	"wails3-manager/core/contracts"
	packagingConfig "wails3-manager/core/packaging/config"
	"wails3-manager/core/project"
)

func TestRemoveProjectDeletesStaleRecordWhenProjectDirectoryIsGone(t *testing.T) {
	setTestConfigRoot(t)

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
	setTestConfigRoot(t)

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

func TestRemoveProjectRestoreCleansManagedPackageOutputs(t *testing.T) {
	setTestConfigRoot(t)

	projectDir := filepath.Join(t.TempDir(), "managed-project")
	buildDir := filepath.Join(projectDir, "build")
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(buildDir, "config.yml"), []byte("productName: Demo\n"), 0644); err != nil {
		t.Fatal(err)
	}
	taskfile := filepath.Join(projectDir, "Taskfile.yml")
	if err := os.WriteFile(taskfile, []byte("version: '3'\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := project.EnsureInitialSnapshot(projectDir); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(buildDir, "config.yml"), []byte("productName: Changed\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(taskfile, []byte("version: '3'\ntasks: {}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	wailsConfig := contracts.WailsProjectConfig{
		Info: contracts.WailsAppInfo{
			ProductName: "Demo",
			Version:     "1.0.0",
		},
	}
	record := contracts.ProjectRecord{
		ProjectDir: projectDir,
		Project: contracts.WailsProjectManager{
			ProjectDir:  projectDir,
			WailsConfig: wailsConfig,
		},
	}
	if err := project.SaveProjectRecord(record); err != nil {
		t.Fatal(err)
	}
	if err := project.SaveUserState(contracts.UserState{Projects: []contracts.ProjectRecord{record}}); err != nil {
		t.Fatal(err)
	}

	cfg := contracts.PackagingConfig{
		SchemaVersion: 1,
		Build: contracts.BuildSettings{
			Taskfile: "Taskfile.yml",
			Task:     "release",
			AppName:  "demo_binary",
		},
		Artifacts: contracts.ArtifactConfig{OutputRoot: "builder/release"},
		MacOS: contracts.MacOSConfig{
			Enabled:       true,
			AppBundle:     "bin/${project.name}.app",
			DMGScript:     "builder/macos/dmg.sh",
			OutputName:    packagingConfig.InstallerOutputNameTemplate,
			CreateDMGPath: "create-dmg",
			WindowWidth:   640,
			WindowHeight:  420,
			IconSize:      96,
		},
	}
	if err := packagingConfig.SavePackagingConfig(projectDir, cfg); err != nil {
		t.Fatal(err)
	}

	managedBinary := filepath.Join(projectDir, "bin", "demo_binary")
	managedBundle := filepath.Join(projectDir, "bin", "Demo.app")
	userBinFile := filepath.Join(projectDir, "bin", "keep.txt")
	if err := os.MkdirAll(managedBundle, 0755); err != nil {
		t.Fatal(err)
	}
	for path, data := range map[string][]byte{
		managedBinary: []byte("binary"),
		userBinFile:   []byte("keep"),
	} {
		if err := os.WriteFile(path, data, 0755); err != nil {
			t.Fatal(err)
		}
	}

	if err := NewService(nil).RemoveProject(projectDir, true); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(managedBinary); !os.IsNotExist(err) {
		t.Fatalf("managed binary still exists, stat err=%v", err)
	}
	if _, err := os.Stat(managedBundle); !os.IsNotExist(err) {
		t.Fatalf("managed app bundle still exists, stat err=%v", err)
	}
	if _, err := os.Stat(userBinFile); err != nil {
		t.Fatalf("user bin file should remain: %v", err)
	}
	if _, err := os.Stat(filepath.Join(projectDir, "builder")); !os.IsNotExist(err) {
		t.Fatalf("builder dir still exists, stat err=%v", err)
	}
	configData, err := os.ReadFile(filepath.Join(buildDir, "config.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if string(configData) != "productName: Demo\n" {
		t.Fatalf("build/config.yml was not restored: %q", configData)
	}
	state := project.LoadUserState()
	if len(state.Projects) != 0 {
		t.Fatalf("project record was not removed: %#v", state.Projects)
	}
}

func setTestConfigRoot(t *testing.T) {
	t.Helper()
	configRoot := t.TempDir()
	t.Setenv("HOME", configRoot)
	t.Setenv("APPDATA", configRoot)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(configRoot, ".config"))
}
