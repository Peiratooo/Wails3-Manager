package settings

import (
	"encoding/json"
	"os"
	"path/filepath"

	"wails3-manager/core/contracts"
	"wails3-manager/core/fsx"
)

func LoadManagerSettings() contracts.ManagerSettings {
	data, err := os.ReadFile(fsx.SettingsPath())
	if err != nil {
		return defaultSettings()
	}
	var settings contracts.ManagerSettings
	if json.Unmarshal(data, &settings) != nil {
		return defaultSettings()
	}
	return normalizeSettings(settings)
}

func SaveManagerSettings(settings contracts.ManagerSettings) (contracts.ManagerSettings, error) {
	settings = normalizeSettings(settings)
	path := fsx.SettingsPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return contracts.ManagerSettings{}, err
	}
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return contracts.ManagerSettings{}, err
	}
	if err := os.WriteFile(path, append(data, '\n'), 0644); err != nil {
		return contracts.ManagerSettings{}, err
	}
	return settings, nil
}

func defaultSettings() contracts.ManagerSettings {
	return contracts.ManagerSettings{Theme: "system", Language: "zh-CN", RecordLogs: true}
}

func normalizeSettings(settings contracts.ManagerSettings) contracts.ManagerSettings {
	if settings.Theme == "" {
		settings.Theme = "system"
	}
	if settings.Language == "" {
		settings.Language = "zh-CN"
	}
	return settings
}
