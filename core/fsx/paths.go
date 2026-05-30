package fsx

import "path/filepath"

const (
	AppName             = "wails3-manager"
	BuilderDirName      = "builder"
	ProjectConfigName   = "project.json"
	PackagingConfigName = "packaging.json"
	StateFileName       = "state.json"
	SettingsFileName    = "settings.json"
)

func BuilderDir(projectDir string) string { return filepath.Join(projectDir, BuilderDirName) }

func ProjectConfigPath(projectDir string) string {
	return filepath.Join(BuilderDir(projectDir), ProjectConfigName)
}

func PackagingConfigPath(projectDir string) string {
	return filepath.Join(BuilderDir(projectDir), PackagingConfigName)
}

func AppConfigDir() string { return UserConfigDir(AppName) }

func StatePath() string { return filepath.Join(AppConfigDir(), StateFileName) }

func SettingsPath() string { return filepath.Join(AppConfigDir(), SettingsFileName) }
