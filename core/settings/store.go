package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"wails3-manager/core/contracts"
	"wails3-manager/core/fsx"
)

func LoadManagerSettings() contracts.ManagerSettings {
	data, err := os.ReadFile(fsx.SettingsPath())
	if err != nil {
		return defaultSettings()
	}
	var raw managerSettingsJSON
	if json.Unmarshal(data, &raw) != nil {
		return defaultSettings()
	}
	settings := defaultSettings()
	if raw.IsDark != nil {
		settings.IsDark = *raw.IsDark
	} else {
		switch strings.ToLower(strings.TrimSpace(raw.Theme)) {
		case "light":
			settings.IsDark = false
		case "dark":
			settings.IsDark = true
		}
	}
	if raw.Language != "" {
		settings.Language = raw.Language
	}
	if raw.RecordLogs != nil {
		settings.RecordLogs = *raw.RecordLogs
	}
	return applySettingsDefaults(settings)
}

func SaveManagerSettings(settings contracts.ManagerSettings) (contracts.ManagerSettings, error) {
	settings = applySettingsDefaults(settings)
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
	return contracts.ManagerSettings{IsDark: true, Language: "zh-CN", RecordLogs: true}
}

func applySettingsDefaults(settings contracts.ManagerSettings) contracts.ManagerSettings {
	if settings.Language == "" {
		settings.Language = "zh-CN"
	}
	return settings
}

type managerSettingsJSON struct {
	IsDark     *bool  `json:"isDark"`
	Theme      string `json:"theme"`
	Language   string `json:"language"`
	RecordLogs *bool  `json:"recordLogs"`
}
