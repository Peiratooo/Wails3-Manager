package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"wails3-manager/core/contracts"
	"wails3-manager/core/fsx"
)

func BackupRoot(projectDir string) string { return filepath.Join(fsx.BuilderDir(projectDir), "backup") }

// EnsureInitialSnapshot captures the files we may later restore when a project
// is removed from the manager. It is idempotent: re-importing an already
// managed project must not overwrite the original snapshot. This is the only
// backup this app creates for project-management files.
func EnsureInitialSnapshot(projectDir string) (string, error) {
	dir := filepath.Join(BackupRoot(projectDir), "initial")
	manifestPath := filepath.Join(dir, "manifest.json")
	if fsx.FileExists(manifestPath) {
		return dir, nil
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	manifest := SnapshotManifest{ID: "initial", Reason: "initial-import", CreatedAt: contracts.NowUnixTime(), Files: []SnapshotFile{}}
	buildSrc := filepath.Join(projectDir, "build")
	if fsx.DirExists(buildSrc) {
		buildDst := filepath.Join(dir, "build")
		if err := fsx.CopyDir(buildSrc, buildDst, true); err != nil {
			return "", err
		}
		manifest.Files = append(manifest.Files, SnapshotFile{Source: "build", Backup: filepath.ToSlash(filepath.Join("initial", "build"))})
	}
	for _, name := range []string{"Taskfile.yml", "Taskfile.yaml"} {
		src := filepath.Join(projectDir, name)
		if !fsx.FileExists(src) {
			continue
		}
		dst := filepath.Join(dir, name)
		if err := fsx.CopyFile(src, dst, 0644); err != nil {
			return "", err
		}
		manifest.Files = append(manifest.Files, SnapshotFile{Source: name, Backup: filepath.ToSlash(filepath.Join("initial", name))})
		break
	}
	if len(manifest.Files) == 0 {
		return "", fmt.Errorf("no initial project files are available for backup")
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return "", err
	}
	return dir, os.WriteFile(manifestPath, append(data, '\n'), 0644)
}

func RestoreInitialSnapshot(projectDir string) error {
	dir := filepath.Join(BackupRoot(projectDir), "initial")
	if !fsx.FileExists(filepath.Join(dir, "manifest.json")) {
		return fmt.Errorf("initial import snapshot was not found")
	}
	buildSrc := filepath.Join(dir, "build")
	if fsx.DirExists(buildSrc) {
		buildDst := filepath.Join(projectDir, "build")
		if err := os.RemoveAll(buildDst); err != nil {
			return err
		}
		if err := fsx.CopyDir(buildSrc, buildDst, true); err != nil {
			return err
		}
	}
	restoredTaskfile := false
	for _, name := range []string{"Taskfile.yml", "Taskfile.yaml"} {
		src := filepath.Join(dir, name)
		if !fsx.FileExists(src) {
			continue
		}
		if err := fsx.CopyFile(src, filepath.Join(projectDir, name), 0644); err != nil {
			return err
		}
		restoredTaskfile = true
		break
	}
	if !restoredTaskfile {
		return fmt.Errorf("initial import snapshot is missing the Taskfile")
	}
	return nil
}

type SnapshotFile struct {
	Source string `json:"source"`
	Backup string `json:"backup"`
}

type SnapshotManifest struct {
	ID        string             `json:"id"`
	Reason    string             `json:"reason"`
	CreatedAt contracts.UnixTime `json:"createdAt"`
	Files     []SnapshotFile     `json:"files"`
}
