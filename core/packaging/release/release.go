package release

import (
	"os"
	"path/filepath"

	"wails3-manager/core/contracts"
	"wails3-manager/core/fsx"
)

func ReleaseArtifacts(projectDir string) []contracts.Artifact {
	return listArtifacts(filepath.Join(fsx.BuilderDir(projectDir), "release"))
}

func listArtifacts(root string) []contracts.Artifact {
	items := []contracts.Artifact{}
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		info, _ := d.Info()
		created := contracts.NowUnixTime()
		var size int64
		if info != nil {
			created = contracts.UnixTimeFrom(info.ModTime())
			size = info.Size()
		}
		items = append(items, contracts.Artifact{Kind: classify(path), Name: d.Name(), Path: path, Size: size, CreatedAt: created})
		return nil
	})
	return items
}

func classify(path string) string {
	switch filepath.Ext(path) {
	case ".exe":
		return "windows-installer"
	case ".dmg":
		return "macos-dmg"
	}
	if fsx.DirExists(path) {
		return "directory"
	}
	return "file"
}
