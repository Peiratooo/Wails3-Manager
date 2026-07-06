package project

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"wails3-manager/core/runlog"
)

const onePixelPNGBase64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAAEUlEQVR4nGJiYGBgAAQAAP//AA8AA/6P688AAAAASUVORK5CYII="

func TestWriteProjectIconPNGBase64SupportsDataURL(t *testing.T) {
	projectDir := t.TempDir()
	input := "data:image/png;base64," + onePixelPNGBase64

	if err := writeProjectIconPNGBase64(projectDir, input); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(projectDir, DefaultAppIconRelPath))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(data, pngSignature) {
		t.Fatalf("written icon is not a PNG")
	}
}

func TestDecodePNGBase64RejectsNonPNG(t *testing.T) {
	if _, err := decodePNGBase64("bm90IGEgcG5n"); err == nil {
		t.Fatal("expected non-PNG base64 to be rejected")
	}
}

func TestGenerateProjectIconsRunsWailsIconCommand(t *testing.T) {
	projectDir := t.TempDir()
	var gotWorkDir string
	var gotCommand []string
	oldRun := runProjectCommand
	runProjectCommand = func(_ *runlog.Logger, workDir string, command []string) error {
		gotWorkDir = workDir
		gotCommand = append([]string(nil), command...)
		return nil
	}
	t.Cleanup(func() {
		runProjectCommand = oldRun
	})

	if err := generateProjectIcons(nil, projectDir); err != nil {
		t.Fatal(err)
	}

	want := []string{
		"wails3", "generate", "icons",
		"-input", "build/appicon.png",
		"-macfilename", "build/darwin/icons.icns",
		"-windowsfilename", "build/windows/icon.ico",
	}
	if gotWorkDir != projectDir {
		t.Fatalf("workDir = %q, want %q", gotWorkDir, projectDir)
	}
	if !reflect.DeepEqual(gotCommand, want) {
		t.Fatalf("command = %#v, want %#v", gotCommand, want)
	}
}
