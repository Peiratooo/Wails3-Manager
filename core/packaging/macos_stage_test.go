package packaging

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"howett.net/plist"

	"wails3-manager/core/contracts"
	"wails3-manager/core/fsx"
	"wails3-manager/core/runlog"
)

func TestStageOfficialMacOSBundleRefreshesExecutableAndMetadata(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test task runner uses a shell script")
	}
	projectDir := t.TempDir()
	source := filepath.Join(projectDir, "bin", "renamed.app")
	writeStageTestBundle(t, source)
	mustWriteFile(t, filepath.Join(projectDir, "Taskfile.yml"), "version: '3'\nincludes:\n  darwin: ./darwin.yml\n")
	mustWriteFile(t, filepath.Join(projectDir, "darwin.yml"), "version: '3'\ntasks:\n  create:app:bundle:\n    cmds: []\n")
	tools := filepath.Join(projectDir, "tools")
	mustWriteFile(t, filepath.Join(tools, "wails3"), "#!/bin/sh\nexit 0\n")
	if err := os.Chmod(filepath.Join(tools, "wails3"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", tools+string(os.PathListSeparator)+os.Getenv("PATH"))
	cfg := contracts.PackagingConfig{
		Build: contracts.BuildSettings{Taskfile: "Taskfile.yml", AppName: "renamed"},
		MacOS: contracts.MacOSConfig{AppBundle: "bin/${project.name}.app"},
	}
	projectConfig := contracts.WailsProjectConfig{Info: contracts.WailsAppInfo{
		ProductName: "Demo Product", Version: "2.3.4", ProductIdentifier: "com.example.renamed",
	}}
	bundle, cleanup, err := NewService(nil).stageMacOSBundle(context.Background(), projectDir, cfg, projectConfig, runlog.Transaction{}, true)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	data, err := os.ReadFile(filepath.Join(bundle, "Contents", "Info.plist"))
	if err != nil {
		t.Fatal(err)
	}
	var info map[string]any
	if _, err := plist.Unmarshal(data, &info); err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]string{
		"CFBundleExecutable": "renamed", "CFBundleName": "Demo Product",
		"CFBundleDisplayName": "Demo Product", "CFBundleIdentifier": "com.example.renamed",
		"CFBundleVersion": "2.3.4", "CFBundleShortVersionString": "2.3.4",
		"NSCameraUsageDescription": "Keep the project's privacy declarations",
	} {
		if info[key] != want {
			t.Errorf("%s = %v, want %q", key, info[key], want)
		}
	}
	if !fsx.FileExists(filepath.Join(bundle, "Contents", "MacOS", info["CFBundleExecutable"].(string))) {
		t.Fatal("staged bundle does not contain its declared executable")
	}
	original, err := os.ReadFile(filepath.Join(source, "Contents", "Info.plist"))
	if err != nil || bytes.Equal(original, data) {
		t.Fatalf("expected metadata changes to be confined to staging: %v", err)
	}
}

func TestStageUnmodifiedPrebuiltMacOSBundlePreservesContents(t *testing.T) {
	projectDir := t.TempDir()
	source := filepath.Join(projectDir, "bin", "Selected.app")
	writeStageTestBundle(t, source)
	cfg := contracts.PackagingConfig{
		Build: contracts.BuildSettings{AppName: "renamed"},
		Entry: contracts.ProgramEntry{ExecutablePath: "bin/Selected.app"},
		MacOS: contracts.MacOSConfig{AppBundle: "bin/${project.name}.app"},
	}
	projectConfig := contracts.WailsProjectConfig{Info: contracts.WailsAppInfo{ProductName: "Other Name", Version: "9.0.0"}}
	bundle, cleanup, err := NewService(nil).stageMacOSBundle(context.Background(), projectDir, cfg, projectConfig, runlog.Transaction{}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	for _, relative := range []string{"Info.plist", "MacOS/renamed", "_CodeSignature/CodeResources"} {
		before, err := os.ReadFile(filepath.Join(source, "Contents", relative))
		if err != nil {
			t.Fatal(err)
		}
		after, err := os.ReadFile(filepath.Join(bundle, "Contents", relative))
		if err != nil || !bytes.Equal(before, after) {
			t.Fatalf("prebuilt bundle file %s changed: %v", relative, err)
		}
	}
}

func writeStageTestBundle(t *testing.T, bundle string) {
	t.Helper()
	mustWriteFile(t, filepath.Join(bundle, "Contents", "Info.plist"), `<?xml version="1.0"?>
<plist version="1.0"><dict>
<key>CFBundlePackageType</key><string>APPL</string>
<key>CFBundleExecutable</key><string>old-name</string>
<key>CFBundleIdentifier</key><string>com.example.old</string>
<key>CFBundleVersion</key><string>1.0.0</string>
<key>NSCameraUsageDescription</key><string>Keep the project's privacy declarations</string>
</dict></plist>`)
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	if err := fsx.CopyFile(executable, filepath.Join(bundle, "Contents", "MacOS", "renamed"), 0755); err != nil {
		t.Fatal(err)
	}
	mustWriteFile(t, filepath.Join(bundle, "Contents", "_CodeSignature", "CodeResources"), "original signature resources")
}
