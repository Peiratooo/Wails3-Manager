// Package project manages Wails project files and exposes the project Wails service.
// It owns build/config.yml, Taskfile.yml, builder/project.json and backups.
package project

import (
	"encoding/json"
	"fmt"
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
	return NormalizeProjectRecord(record, projectDir), true
}

func SaveProjectRecord(record contracts.ProjectRecord) error {
	projectDir, err := NormalizeProjectDir(record)
	if err != nil {
		return err
	}
	record = NormalizeProjectRecord(record, projectDir)
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

func NormalizeProjectRecord(record contracts.ProjectRecord, projectDir string) contracts.ProjectRecord {
	record.ProjectDir = projectDir
	record.Project.ProjectDir = projectDir
	return record
}

func NormalizeProjectDir(record contracts.ProjectRecord) (string, error) {
	projectDir := fsx.FirstNonEmpty(record.ProjectDir, record.Project.ProjectDir)
	if projectDir == "" {
		return "", fmt.Errorf("项目路径不能为空")
	}
	return fsx.NormalizePath(projectDir)
}

func SameProjectPath(a, b string) bool {
	aa, errA := fsx.NormalizePath(a)
	bb, errB := fsx.NormalizePath(b)
	if errA != nil || errB != nil {
		return filepath.Clean(a) == filepath.Clean(b)
	}
	return filepath.Clean(aa) == filepath.Clean(bb)
}
