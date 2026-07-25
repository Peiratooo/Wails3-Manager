package project

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"wails3-manager/core/contracts"
	"wails3-manager/core/fsx"
	"wails3-manager/core/project/scanner"
	"wails3-manager/core/runlog"
)

var runProjectCommand = func(log *runlog.Logger, projectDir string, command []string) error {
	return (runlog.Runner{Log: log}).Run(context.Background(), projectDir, command)
}

func (s *Service) ScanProject(projectDir string) (contracts.ScanResult, error) {
	return scanner.Scan(projectDir)
}

func (s *Service) ImportProject(projectDir string) error {
	projectDir, err := fsx.NormalizePath(projectDir)
	if err != nil {
		return err
	}
	scan, err := scanner.Scan(projectDir)
	if err != nil {
		return err
	}
	if !scan.IsWailsProject {
		return fmt.Errorf("directory did not pass Wails3 project detection: score %d, minimum %d", scan.Score, scanner.PassingScore)
	}
	// Import creates project-management files first. Packaging initialization is
	// supplied by the desktop wiring so core/project keeps its module boundary.
	if err := createManagedProjectLayout(projectDir); err != nil {
		return err
	}
	if _, err := EnsureInitialSnapshot(projectDir); err != nil {
		return fmt.Errorf("failed to create the initial import snapshot: %w", err)
	}
	manager, err := LoadManagerWithConfigCompletion(projectDir)
	if err != nil {
		return err
	}
	now := contracts.NowUnixTime()
	if existing, ok := LoadProjectRecord(projectDir); ok && !existing.ImportedAt.IsZero() {
		now = existing.ImportedAt
	}
	record := contracts.ProjectRecord{
		ProjectDir:   projectDir,
		Project:      manager,
		ImportedAt:   now,
		LastOpenedAt: contracts.NowUnixTime(),
	}
	if err := SaveProjectRecord(record); err != nil {
		return err
	}
	if s.InitPackagingFn != nil {
		if err := s.InitPackagingFn(projectDir); err != nil {
			return fmt.Errorf("failed to initialize packaging config: %w", err)
		}
	}
	if err := UpsertProjectRecord(record); err != nil {
		return err
	}
	return nil
}

func (s *Service) SaveProject(record contracts.ProjectRecord) (contracts.ProjectRecord, error) {
	projectDir, err := fsx.NormalizePath(record.ProjectDir)
	if err != nil {
		return contracts.ProjectRecord{}, err
	}
	if err := validateProjectInfoForSave(record.Project.WailsConfig.Info); err != nil {
		return contracts.ProjectRecord{}, err
	}
	// SaveManager is the only place that writes build/config.yml and Taskfile.yml
	// for project metadata. SaveProject intentionally does not create per-edit
	// backups; restore always uses the single initial import snapshot.
	record.Project.ProjectDir = projectDir
	manager, err := SaveManager(projectDir, record.Project)
	if err != nil {
		return contracts.ProjectRecord{}, err
	}
	if err := runProjectCommand(s.Log, projectDir, []string{"wails3", "task", "common:update:build-assets"}); err != nil {
		return contracts.ProjectRecord{}, err
	}
	existing, _ := LoadProjectRecord(projectDir)
	now := contracts.NowUnixTime()
	if record.ImportedAt.IsZero() {
		record.ImportedAt = existing.ImportedAt
	}
	if record.ImportedAt.IsZero() {
		record.ImportedAt = now
	}
	record.ProjectDir = projectDir
	record.Project = manager
	record.LastOpenedAt = now
	if err := SaveProjectRecord(record); err != nil {
		return contracts.ProjectRecord{}, err
	}
	if err := UpsertProjectRecord(record); err != nil {
		return contracts.ProjectRecord{}, err
	}
	return record, nil
}

func validateProjectInfoForSave(info contracts.WailsAppInfo) error {
	required := []struct {
		name  string
		value string
	}{
		{"productName", info.ProductName},
		{"version", info.Version},
		{"companyName", info.CompanyName},
		{"productIdentifier", info.ProductIdentifier},
		{"description", info.Description},
		{"copyright", info.Copyright},
	}
	for _, field := range required {
		if strings.TrimSpace(field.value) == "" {
			return fmt.Errorf("project info %s is required", field.name)
		}
	}
	return nil
}

func (s *Service) ReplaceProjectIcon(projectDir string, pngBase64 string) (contracts.ProjectRecord, error) {
	projectDir, err := fsx.NormalizePath(projectDir)
	if err != nil {
		return contracts.ProjectRecord{}, err
	}
	record, ok := LoadProjectRecord(projectDir)
	if !ok {
		return contracts.ProjectRecord{}, fmt.Errorf("project is not imported: %s", projectDir)
	}
	// Wails derives platform icons from build/appicon.png, so replacing this
	// single file and running update:build-assets is enough here. The original
	// icon is already protected by the initial import snapshot.
	if err := writeProjectIconPNGBase64(projectDir, pngBase64); err != nil {
		return contracts.ProjectRecord{}, fmt.Errorf("failed to write build/appicon.png: %w", err)
	}
	if err := generateProjectIcons(s.Log, projectDir); err != nil {
		return contracts.ProjectRecord{}, err
	}
	if err := runProjectCommand(s.Log, projectDir, []string{"wails3", "task", "common:update:build-assets"}); err != nil {
		return contracts.ProjectRecord{}, err
	}
	manager, err := LoadManagerWithConfigCompletion(projectDir)
	if err != nil {
		return contracts.ProjectRecord{}, err
	}
	manager.WailsConfig.Icon = filepath.ToSlash(DefaultAppIconRelPath)
	record.ProjectDir = projectDir
	record.Project = manager
	record.LastOpenedAt = contracts.NowUnixTime()
	if err := SaveProjectRecord(record); err != nil {
		return contracts.ProjectRecord{}, err
	}
	if err := UpsertProjectRecord(record); err != nil {
		return contracts.ProjectRecord{}, err
	}
	return record, nil
}

func generateProjectIcons(log *runlog.Logger, projectDir string) error {
	for _, dir := range []string{
		filepath.Join(projectDir, "build", "darwin"),
		filepath.Join(projectDir, "build", "windows"),
	} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return runProjectCommand(log, projectDir, []string{
		"wails3", "generate", "icons",
		"-input", "build/appicon.png",
		"-macfilename", "build/darwin/icons.icns",
		"-windowsfilename", "build/windows/icon.ico",
	})
}

func createManagedProjectLayout(projectDir string) error {
	return os.MkdirAll(BackupRoot(projectDir), 0755)
}

func currentPlatform() contracts.Platform {
	switch runtime.GOOS {
	case "windows":
		return contracts.PlatformWindows
	case "darwin":
		return contracts.PlatformMacOS
	default:
		return contracts.Platform(runtime.GOOS)
	}
}
