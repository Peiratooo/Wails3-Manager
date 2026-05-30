package scanner

import (
	"bytes"
	"os"
	"path/filepath"

	"wails3-manager/core/contracts"
	"wails3-manager/core/fsx"
)

const PassingScore = 60

func Scan(projectDir string) (contracts.ScanResult, error) {
	projectDir, err := fsx.NormalizePath(projectDir)
	if err != nil {
		return contracts.ScanResult{}, err
	}
	scan := contracts.ScanResult{ProjectDir: projectDir}
	if !fsx.DirExists(projectDir) {
		return scan, nil
	}

	if fsx.FileExists(filepath.Join(projectDir, "Taskfile.yml")) || fsx.FileExists(filepath.Join(projectDir, "Taskfile.yaml")) {
		scan.Score += 20
	}

	if fsx.FileExists(filepath.Join(projectDir, "build", "config.yml")) {
		scan.Score += 30
	}

	goModPath := filepath.Join(projectDir, "go.mod")
	if fsx.FileExists(goModPath) {
		scan.Score += 15
		data, _ := os.ReadFile(goModPath)
		if bytes.Contains(data, []byte("github.com/wailsapp/wails/v3")) || bytes.Contains(data, []byte("wailsapp/wails/v3")) {
			scan.Score += 25
		}
	}

	for _, dir := range []string{"frontend", "web", "ui"} {
		if fsx.DirExists(filepath.Join(projectDir, dir)) {
			scan.Score += 10
			break
		}
	}

	scan.IsWailsProject = scan.Score >= PassingScore
	return scan, nil
}
