package execenv

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestLookPathFindsUserGoBinWhenPathIsMinimal(t *testing.T) {
	home := t.TempDir()
	setHome(t, home)
	t.Setenv("PATH", "")
	t.Setenv("SHELL", "")

	name := "wails3-manager-test-tool"
	filename := name
	if runtime.GOOS == "windows" {
		filename += ".exe"
	}

	expected := filepath.Join(home, "go", "bin", filename)
	if err := os.MkdirAll(filepath.Dir(expected), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(expected, []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatal(err)
	}

	got, err := LookPath(name)
	if err != nil {
		t.Fatal(err)
	}
	if got != expected {
		t.Fatalf("LookPath(%q) = %q, want %q", name, got, expected)
	}
}

func TestEnvironMergesPathOverride(t *testing.T) {
	base := t.TempDir()
	override := t.TempDir()
	t.Setenv("PATH", base)
	t.Setenv("SHELL", "")

	env := envMap(Environ(map[string]string{"PATH": override}))
	parts := filepath.SplitList(env["PATH"])
	if len(parts) < 2 {
		t.Fatalf("PATH = %q, want at least override and base", env["PATH"])
	}
	if parts[0] != override {
		t.Fatalf("PATH first entry = %q, want override %q", parts[0], override)
	}
	if parts[1] != base {
		t.Fatalf("PATH second entry = %q, want base %q", parts[1], base)
	}
}

func TestLastPathLikeLineIgnoresShellNoise(t *testing.T) {
	output := strings.Join([]string{
		"hello from profile",
		"/usr/local/bin:/usr/bin:/bin",
	}, "\n")

	if got, want := lastPathLikeLine(output), "/usr/local/bin:/usr/bin:/bin"; got != want {
		t.Fatalf("lastPathLikeLine() = %q, want %q", got, want)
	}
}

func setHome(t *testing.T, home string) {
	t.Helper()

	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", home)
		t.Setenv("HOMEDRIVE", "")
		t.Setenv("HOMEPATH", "")
		return
	}

	t.Setenv("HOME", home)
}
