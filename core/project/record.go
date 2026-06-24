// Package project manages Wails project files and exposes the project Wails service.
// It owns build/config.yml, Taskfile.yml, builder/project.json and backups.
package project

import (
	"encoding/json"
	"os"
	"path/filepath"

	"wails3-manager/core/contracts"
	"wails3-manager/core/fsx"
)

// LoadProjectRecord reads builder/project.json from the target project. The
// boolean return keeps callers simple: false means "not imported yet".
func LoadProjectRecord(projectDir string) (contracts.ProjectRecord, bool) {
	projectDir, err := fsx.NormalizePath(projectDir)
	if err != nil {
		return contracts.ProjectRecord{}, false
	}
	data, err := os.ReadFile(fsx.ProjectConfigPath(projectDir))
	if err != nil {
		return contracts.ProjectRecord{}, false
	}
	var record contracts.ProjectRecord
	if json.Unmarshal(data, &record) != nil {
		return contracts.ProjectRecord{}, false
	}
	record.ProjectDir = projectDir
	record.Project.ProjectDir = projectDir
	return record, true
}

func SaveProjectRecord(record contracts.ProjectRecord) error {
	projectDir, err := fsx.NormalizePath(record.ProjectDir)
	if err != nil {
		return err
	}
	record.ProjectDir = projectDir
	record.Project.ProjectDir = projectDir
	path := fsx.ProjectConfigPath(projectDir)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, append(data, '\n'), 0644)
}

func SameProjectPath(a, b string) bool {
	aa, errA := fsx.NormalizePath(a)
	bb, errB := fsx.NormalizePath(b)
	if errA != nil || errB != nil {
		return false
	}
	return aa == bb
}
