package desktop

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenPathRejectsEmptyPath(t *testing.T) {
	err := (&AppService{}).OpenPath("")
	if err == nil || !strings.Contains(err.Error(), "path is required") {
		t.Fatalf("OpenPath() error = %v, want path is required", err)
	}
}

func TestOpenPathRejectsMissingPath(t *testing.T) {
	err := (&AppService{}).OpenPath(filepath.Join(t.TempDir(), "missing"))
	if err == nil {
		t.Fatal("expected missing path to be rejected")
	}
}
