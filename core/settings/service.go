package settings

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"wails3-manager/core/contracts"
	"wails3-manager/core/fsx"
	packagingconfig "wails3-manager/core/packaging/config"
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
	record, ok := project.LoadProjectRecord(projectDir)
	if !ok {
		removeStaleProjectDirs(projectDir)
		return project.RemoveProjectRecord(projectDir)
	}
	if restoreOriginal {
		cfg, cfgErr := packagingconfig.LoadPackagingConfig(projectDir)
		if err := project.RestoreInitialSnapshot(projectDir); err != nil {
			return err
		}
		if cfgErr == nil {
			if err := removeManagedPackageOutputs(projectDir, cfg, record.Project.WailsConfig); err != nil {
				return err
			}
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

func removeManagedPackageOutputs(projectDir string, cfg contracts.PackagingConfig, projectConfig contracts.WailsProjectConfig) error {
	paths := []string{
		packagingconfig.DefaultExecutablePath(cfg, contracts.PlatformWindows),
		packagingconfig.DefaultMacOSBinaryPath(cfg),
		packagingconfig.DefaultMacOSAppBundlePath(cfg, projectConfig),
		packagingconfig.MacOSAppBundlePath(cfg, projectConfig),
	}
	seen := map[string]bool{}
	for _, path := range paths {
		abs := fsx.Resolve(projectDir, path)
		if abs == "" || !isManagedBinPath(projectDir, abs) || seen[abs] {
			continue
		}
		seen[abs] = true
		if err := os.RemoveAll(abs); err != nil {
			return err
		}
	}
	removeEmptyBinDir(projectDir)
	return nil
}

func isManagedBinPath(projectDir, absPath string) bool {
	rel, err := filepath.Rel(projectDir, absPath)
	if err != nil || rel == "." || !filepath.IsLocal(rel) {
		return false
	}
	return strings.HasPrefix(filepath.ToSlash(rel), "bin/")
}

func removeEmptyBinDir(projectDir string) {
	binDir := filepath.Join(projectDir, "bin")
	entries, err := os.ReadDir(binDir)
	if err != nil || len(entries) != 0 {
		return
	}
	_ = os.Remove(binDir)
}
