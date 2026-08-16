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

func TestOpenPathCommandWindows(t *testing.T) {
	const path = `C:\Users\Peirato\Desktop\dist\Wails3 Manager Setup.exe`

	directoryCmd := openPathCommand("windows", path, true)
	if got, want := directoryCmd.Args, []string{"explorer.exe", path}; !equalStrings(got, want) {
		t.Fatalf("directory command args = %q, want %q", got, want)
	}

	fileCmd := openPathCommand("windows", path, false)
	if got, want := fileCmd.Args, []string{"explorer.exe", "/select," + path}; !equalStrings(got, want) {
		t.Fatalf("file command args = %q, want %q", got, want)
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
