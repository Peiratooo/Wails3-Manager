package settings

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"wails3-manager/core/contracts"
	"wails3-manager/core/fsx"
	"wails3-manager/core/project"
	"wails3-manager/core/runlog"
)

type SettingsService struct {
	Log *runlog.Logger
}

func NewService(log *runlog.Logger) *SettingsService {
	return &SettingsService{Log: log}
}

func (s *SettingsService) ListProjects() ([]contracts.ProjectRecord, error) {
	userState := project.LoadUserState()
	project.SortProjects(userState.Projects)
	return userState.Projects, nil
}

func (s *SettingsService) OpenProject(projectDir string) (contracts.ProjectRecord, error) {
	projectDir, err := fsx.NormalizePath(projectDir)
	if err != nil {
		return contracts.ProjectRecord{}, err
	}
	record, ok := project.LoadProjectRecord(projectDir)
	if !ok {
		return contracts.ProjectRecord{}, fmt.Errorf("项目未导入：%s", projectDir)
	}
	manager, err := project.LoadManager(projectDir)
	if err != nil {
		return contracts.ProjectRecord{}, err
	}
	record.ProjectDir = projectDir
	record.Project = manager
	record.LastOpenedAt = contracts.NowUnixTime()
	if record.ImportedAt.IsZero() {
		record.ImportedAt = record.LastOpenedAt
	}
	if err := project.SaveProjectRecord(record); err != nil {
		return contracts.ProjectRecord{}, err
	}
	if err := project.UpsertProjectRecord(record); err != nil {
		return contracts.ProjectRecord{}, err
	}
	return record, nil
}

func (s *SettingsService) RemoveProject(projectDir string, restoreOriginal bool) error {
	projectDir, err := fsx.NormalizePath(projectDir)
	if err != nil {
		return err
	}
	if _, ok := project.LoadProjectRecord(projectDir); !ok {
		return fmt.Errorf("项目未导入：%s", projectDir)
	}
	if restoreOriginal {
		if err := project.RestoreInitialSnapshot(projectDir); err != nil {
			return err
		}
		if err := removeBuilderDir(projectDir); err != nil {
			return err
		}
	}
	return project.RemoveProjectRecord(projectDir)
}

func (s *SettingsService) GetSettings() (contracts.ManagerSettings, error) {
	return LoadManagerSettings(), nil
}

func (s *SettingsService) SaveSettings(settings contracts.ManagerSettings) (contracts.ManagerSettings, error) {
	return SaveManagerSettings(settings)
}

func (s *SettingsService) LogsSince(cursor int) contracts.LogSnapshot {
	if s.Log == nil {
		return contracts.LogSnapshot{Cursor: 0, Lines: []string{}}
	}
	next, lines := s.Log.Since(cursor)
	return contracts.LogSnapshot{Cursor: next, Lines: lines}
}

func (s *SettingsService) ClearLogs() {
	if s.Log != nil {
		s.Log.Clear()
	}
}

func (s *SettingsService) GetABSPath(projectDir, path string) string {
	return fsx.ResolveProjectFile(projectDir, path)
}

func removeBuilderDir(projectDir string) error {
	builderDir := fsx.BuilderDir(projectDir)
	rel, err := filepath.Rel(projectDir, builderDir)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return fmt.Errorf("拒绝删除异常 builder 目录：%s", builderDir)
	}
	return os.RemoveAll(builderDir)
}
