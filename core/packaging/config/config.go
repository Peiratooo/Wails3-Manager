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
	cfg, err := ReadPackagingConfigFile(path)
	if err != nil {
		return contracts.PackagingConfig{}, err
	}
	if err := validatePackagingConfig(cfg); err != nil {
		return contracts.PackagingConfig{}, err
	}
	return cfg, nil
}

func ReadPackagingConfigFile(path string) (contracts.PackagingConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return contracts.PackagingConfig{}, err
	}
	var cfg contracts.PackagingConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return contracts.PackagingConfig{}, fmt.Errorf("failed to parse packaging.json: %w", err)
	}
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
	if err := validatePackagingConfig(cfg); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0644)
}

func validatePackagingConfig(cfg contracts.PackagingConfig) error {
	if cfg.SchemaVersion != 1 {
		return fmt.Errorf("packaging schemaVersion must be 1")
	}
	if strings.TrimSpace(cfg.Build.Taskfile) == "" {
		return fmt.Errorf("packaging build.taskfile is required")
	}
	if strings.TrimSpace(cfg.Build.Task) == "" && len(cfg.Build.Command) == 0 {
		return fmt.Errorf("packaging build.task is required when build.command is empty")
	}
	if strings.TrimSpace(cfg.Build.AppName) == "" {
		return fmt.Errorf("packaging build.appName is required")
	}
	if strings.TrimSpace(cfg.Artifacts.OutputRoot) == "" {
		return fmt.Errorf("packaging artifacts.outputRoot is required")
	}
	for i, asset := range cfg.Assets {
		if strings.TrimSpace(asset.Src) == "" {
			return fmt.Errorf("packaging assets[%d].src is required", i)
		}
		if asset.Type != "file" && asset.Type != "directory" {
			return fmt.Errorf("packaging assets[%d].type must be file or directory", i)
		}
	}
	if cfg.Windows.Enabled {
		if err := validateWindowsConfig(cfg.Windows); err != nil {
			return err
		}
	}
	if cfg.MacOS.Enabled {
		if err := validateMacOSConfig(cfg.MacOS); err != nil {
			return err
		}
	}
	return nil
}

func validateWindowsConfig(cfg contracts.WindowsConfig) error {
	if strings.TrimSpace(cfg.InnoScript) == "" {
		return fmt.Errorf("packaging windows.innoScript is required")
	}
	if strings.TrimSpace(cfg.DefaultDirName) == "" {
		return fmt.Errorf("packaging windows.defaultDirName is required")
	}
	if strings.TrimSpace(cfg.PrivilegesRequired) == "" {
		return fmt.Errorf("packaging windows.privilegesRequired is required")
	}
	if strings.TrimSpace(cfg.SetupIcon) == "" {
		return fmt.Errorf("packaging windows.setupIcon is required")
	}
	return nil
}

func validateMacOSConfig(cfg contracts.MacOSConfig) error {
	if strings.TrimSpace(cfg.AppBundle) == "" {
		return fmt.Errorf("packaging macos.appBundle is required")
	}
	if strings.TrimSpace(cfg.DMGScript) == "" {
		return fmt.Errorf("packaging macos.dmgScript is required")
	}
	if strings.TrimSpace(cfg.CreateDMGPath) == "" {
		return fmt.Errorf("packaging macos.createDmgPath is required")
	}
	if cfg.WindowWidth <= 0 || cfg.WindowHeight <= 0 || cfg.IconSize <= 0 {
		return fmt.Errorf("packaging macos window size and icon size must be greater than zero")
	}
	return nil
}
