package settings

import (
	"fmt"
	"os"
	"path/filepath"

	"wails3-manager/core/contracts"
	"wails3-manager/core/fsx"
	"wails3-manager/core/project"
	"wails3-manager/core/runlog"
)

type Service struct {
	Log             *runlog.Logger
	InitPackagingFn func(projectDir string) error
}

func NewService(log *runlog.Logger) *Service {
	return &Service{Log: log}
}

func (s *Service) ListProjects() ([]contracts.ProjectRecord, error) {
	userState := project.LoadUserState()
	project.SortProjects(userState.Projects)
	return userState.Projects, nil
}

func (s *Service) OpenProject(projectDir string) (contracts.ProjectRecord, error) {
	projectDir, err := fsx.NormalizePath(projectDir)
	if err != nil {
		return contracts.ProjectRecord{}, err
	}
	record, ok := project.LoadProjectRecord(projectDir)
	if !ok {
		return contracts.ProjectRecord{}, fmt.Errorf("project is not imported: %s", projectDir)
	}
	manager, err := project.LoadManagerWithConfigCompletion(projectDir)
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
	if s.InitPackagingFn != nil {
		if err := s.InitPackagingFn(projectDir); err != nil {
			return contracts.ProjectRecord{}, fmt.Errorf("failed to initialize packaging config: %w", err)
		}
	}
	return record, nil
}

func (s *Service) RemoveProject(projectDir string, restoreOriginal bool) error {
	projectDir, err := fsx.NormalizePath(projectDir)
	if err != nil {
		return err
	}
	if _, ok := project.LoadProjectRecord(projectDir); !ok {
		removeStaleProjectDirs(projectDir)
		return project.RemoveProjectRecord(projectDir)
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

func (s *Service) GetSettings() (contracts.ManagerSettings, error) {
	return LoadManagerSettings(), nil
}

func (s *Service) SaveSettings(settings contracts.ManagerSettings) (contracts.ManagerSettings, error) {
	saved, err := SaveManagerSettings(settings)
	if err != nil {
		return contracts.ManagerSettings{}, err
	}
	if s.Log != nil {
		s.Log.SetRecordLogs(saved.RecordLogs)
	}
	return saved, nil
}

func (s *Service) ClearLogs() {
	if s.Log != nil {
		s.Log.Clear()
	}
}

func (s *Service) GetABSPath(projectDir, path string) string {
	return fsx.ResolveProjectFile(projectDir, path)
}

func removeStaleProjectDirs(projectDir string) {
	_ = removeBuilderDir(projectDir)
	_ = os.Remove(projectDir)
}

func removeBuilderDir(projectDir string) error {
	builderDir := fsx.BuilderDir(projectDir)
	rel, err := filepath.Rel(projectDir, builderDir)
	if err != nil || rel == "." || !filepath.IsLocal(rel) {
		return fmt.Errorf("refusing to delete an invalid builder directory: %s", builderDir)
	}
	return os.RemoveAll(builderDir)
}
