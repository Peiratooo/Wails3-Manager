package fsx

import (
	"path/filepath"
	"testing"
)

func TestResolveProjectFileRejectsEscapes(t *testing.T) {
	root := t.TempDir()
	if got := ResolveProjectFile(root, "build/appicon.png"); got != filepath.Join(root, "build", "appicon.png") {
		t.Fatalf("ResolveProjectFile() = %q", got)
	}
	for _, path := range []string{"../outside", "build/../../outside", filepath.Join(root, "outside")} {
		if got := ResolveProjectFile(root, path); got != "" {
			t.Fatalf("ResolveProjectFile(%q) = %q, want empty", path, got)
		}
	}
}
