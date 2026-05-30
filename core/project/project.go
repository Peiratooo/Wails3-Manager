package project

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"wails3-manager/core/contracts"
	"wails3-manager/core/fsx"
	"wails3-manager/core/project/scanner"
	"wails3-manager/core/runlog"
)

func (s *ProjectService) ScanProject(projectDir string) (contracts.ScanResult, error) {
	return scanner.Scan(projectDir)
}

func (s *ProjectService) ImportProject(projectDir string) (contracts.ProjectRecord, error) {
	projectDir, err := fsx.NormalizePath(projectDir)
	if err != nil {
		return contracts.ProjectRecord{}, err
	}
	scan, err := scanner.Scan(projectDir)
	if err != nil {
		return contracts.ProjectRecord{}, err
	}
	if !scan.IsWailsProject {
		return contracts.ProjectRecord{}, fmt.Errorf("当前目录未通过 Wails3 项目检测，评分 %d，最低需要 %d", scan.Score, scanner.PassingScore)
	}
	// Import only creates project-management files. Packaging config and
	// installer templates are initialized later by PackagingService.
	if err := ensureLayout(projectDir); err != nil {
		return contracts.ProjectRecord{}, err
	}
	if _, err := EnsureInitialSnapshot(projectDir); err != nil {
		return contracts.ProjectRecord{}, fmt.Errorf("创建导入前快照失败：%w", err)
	}
	manager, err := LoadManager(projectDir)
	if err != nil {
		return contracts.ProjectRecord{}, err
	}
	now := contracts.NowUnixTime()
	if existing, ok := LoadProjectRecord(projectDir); ok && !existing.ImportedAt.IsZero() {
		now = existing.ImportedAt
	}
	record := normalizeRecord(contracts.ProjectRecord{ProjectDir: projectDir, Project: manager, ImportedAt: now, LastOpenedAt: contracts.NowUnixTime()}, projectDir)
	if err := SaveProjectRecord(record); err != nil {
		return contracts.ProjectRecord{}, err
	}
	if err := UpsertProjectRecord(record); err != nil {
		return contracts.ProjectRecord{}, err
	}
	return record, nil
}

func (s *ProjectService) SaveProject(record contracts.ProjectRecord) (contracts.ProjectRecord, error) {
	projectDir, err := NormalizeProjectDir(record)
	if err != nil {
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
	if err := (runlog.Runner{Log: s.Log}).Run(context.Background(), projectDir, []string{"wails3", "task", "common:update:build-assets"}); err != nil {
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
	record = normalizeRecord(record, projectDir)
	if err := SaveProjectRecord(record); err != nil {
		return contracts.ProjectRecord{}, err
	}
	if err := UpsertProjectRecord(record); err != nil {
		return contracts.ProjectRecord{}, err
	}
	return record, nil
}

func (s *ProjectService) ReplaceProjectIcon(projectDir string, sourcePath string) (contracts.ProjectRecord, error) {
	projectDir, err := fsx.NormalizePath(projectDir)
	if err != nil {
		return contracts.ProjectRecord{}, err
	}
	record, ok := LoadProjectRecord(projectDir)
	if !ok {
		return contracts.ProjectRecord{}, fmt.Errorf("项目未导入：%s", projectDir)
	}
	source := fsx.Resolve(projectDir, sourcePath)
	if !fsx.FileExists(source) {
		return contracts.ProjectRecord{}, fmt.Errorf("图标源文件不存在：%s", sourcePath)
	}
	// Wails derives platform icons from build/appicon.png, so replacing this
	// single file and running update:build-assets is enough here. The original
	// icon is already protected by the initial import snapshot.
	target := filepath.Join(projectDir, DefaultAppIconRelPath)
	if filepath.Clean(source) != filepath.Clean(target) {
		if err := fsx.CopyFile(source, target, 0644); err != nil {
			return contracts.ProjectRecord{}, fmt.Errorf("写入 build/appicon.png 失败：%w", err)
		}
	}
	if err := (runlog.Runner{Log: s.Log}).Run(context.Background(), projectDir, []string{"wails3", "task", "common:update:build-assets"}); err != nil {
		return contracts.ProjectRecord{}, err
	}
	manager, err := LoadManager(projectDir)
	if err != nil {
		return contracts.ProjectRecord{}, err
	}
	manager.WailsConfig.Icon = filepath.ToSlash(DefaultAppIconRelPath)
	record.ProjectDir = projectDir
	record.Project = manager
	record.LastOpenedAt = contracts.NowUnixTime()
	record = normalizeRecord(record, projectDir)
	if err := SaveProjectRecord(record); err != nil {
		return contracts.ProjectRecord{}, err
	}
	if err := UpsertProjectRecord(record); err != nil {
		return contracts.ProjectRecord{}, err
	}
	return record, nil
}

func ensureLayout(projectDir string) error {
	return os.MkdirAll(BackupRoot(projectDir), 0755)
}

func normalizeRecord(record contracts.ProjectRecord, projectDir string) contracts.ProjectRecord {
	record.ProjectDir = projectDir
	record.Project.ProjectDir = projectDir
	if record.Project.CurrentPlatform == "" {
		record.Project.CurrentPlatform = currentPlatform()
	}
	return record
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
