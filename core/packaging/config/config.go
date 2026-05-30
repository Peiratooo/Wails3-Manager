package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"wails3-manager/core/contracts"
	"wails3-manager/core/fsx"
)

func LoadPackagingConfig(projectDir string) (contracts.PackagingConfig, error) {
	projectDir, err := fsx.NormalizePath(projectDir)
	if err != nil {
		return contracts.PackagingConfig{}, err
	}
	return LoadPackagingConfigFile(fsx.PackagingConfigPath(projectDir))
}

func LoadPackagingConfigFile(path string) (contracts.PackagingConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return contracts.PackagingConfig{}, err
	}
	var cfg contracts.PackagingConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return contracts.PackagingConfig{}, fmt.Errorf("解析 packaging.json 失败：%w", err)
	}
	NormalizePackagingConfig(&cfg)
	return cfg, nil
}

func SavePackagingConfig(projectDir string, cfg contracts.PackagingConfig) error {
	projectDir, err := fsx.NormalizePath(projectDir)
	if err != nil {
		return err
	}
	return SavePackagingConfigFile(fsx.PackagingConfigPath(projectDir), cfg)
}

func SavePackagingConfigFile(path string, cfg contracts.PackagingConfig) error {
	NormalizePackagingConfig(&cfg)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0644)
}

func NormalizePackagingConfig(cfg *contracts.PackagingConfig) {
	if cfg.SchemaVersion == 0 {
		cfg.SchemaVersion = 1
	}
	if strings.TrimSpace(cfg.Build.Taskfile) == "" {
		cfg.Build.Taskfile = "Taskfile.yml"
	}
	if strings.TrimSpace(cfg.Build.Task) == "" {
		cfg.Build.Task = "builder:release"
	}
	if strings.TrimSpace(cfg.Build.AppName) == "" {
		cfg.Build.AppName = fsx.SafeName(cfg.Project.Name)
	}
	for i := range cfg.Assets {
		cfg.Assets[i].Src = strings.TrimSpace(cfg.Assets[i].Src)
		cfg.Assets[i].Type = strings.ToLower(strings.TrimSpace(cfg.Assets[i].Type))
		if cfg.Assets[i].Type == "" {
			cfg.Assets[i].Type = "file"
		}
	}
	if strings.TrimSpace(cfg.Artifacts.OutputRoot) == "" {
		cfg.Artifacts.OutputRoot = "builder/release"
	}
}
