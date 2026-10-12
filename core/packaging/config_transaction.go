package packaging

import (
	"errors"
	"os"
	"path/filepath"

	"wails3-manager/core/contracts"
	"wails3-manager/core/fsx"
	"wails3-manager/core/packaging/config"
)

type configSnapshot struct {
	path   string
	data   []byte
	mode   os.FileMode
	exists bool
}

func snapshotPackagingFiles(projectDir string, cfg contracts.PackagingConfig, projectConfig contracts.WailsProjectConfig) ([]configSnapshot, error) {
	paths := []string{fsx.PackagingConfigPath(projectDir), fsx.ResolveProjectFile(projectDir, cfg.Build.Taskfile)}
	if cfg.Windows.Enabled && cfg.Windows.InnoScript != "" {
		paths = append(paths, fsx.Resolve(projectDir, cfg.Windows.InnoScript))
	}
	if cfg.MacOS.Enabled && cfg.MacOS.Background != "" {
		paths = append(paths, fsx.Resolve(projectDir, config.RenderPlaceholders(cfg.MacOS.Background, cfg, projectConfig)))
	}
	seen := map[string]bool{}
	var snapshots []configSnapshot
	for _, path := range paths {
		if path == "" || seen[path] {
			continue
		}
		seen[path] = true
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			snapshots = append(snapshots, configSnapshot{path: path})
			continue
		}
		if err != nil {
			return nil, err
		}
		if !info.Mode().IsRegular() {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, configSnapshot{path: path, data: data, mode: info.Mode().Perm(), exists: true})
	}
	return snapshots, nil
}

func restorePackagingFiles(snapshots []configSnapshot) error {
	var failures []error
	for _, snapshot := range snapshots {
		var err error
		if snapshot.exists {
			err = fsx.WriteFileAtomic(snapshot.path, snapshot.data, snapshot.mode)
		} else {
			err = os.Remove(snapshot.path)
			if os.IsNotExist(err) {
				err = nil
			}
		}
		if err != nil {
			failures = append(failures, errors.New("restore "+filepath.Base(snapshot.path)+": "+err.Error()))
		}
	}
	return errors.Join(failures...)
}
