package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"wails3-manager/core/contracts"
	"wails3-manager/core/fsx"
)

const RootTaskfileRelPath = "Taskfile.yml"

func RootTaskfileRelPathForProject(projectDir string) string {
	for _, name := range []string{RootTaskfileRelPath, "Taskfile.yaml"} {
		if fsx.FileExists(filepath.Join(projectDir, name)) {
			return name
		}
	}
	return RootTaskfileRelPath
}

func LoadManager(projectDir string) (contracts.WailsProjectManager, error) {
	projectDir, err := fsx.NormalizePath(projectDir)
	if err != nil {
		return contracts.WailsProjectManager{}, err
	}
	wc, err := LoadWailsConfig(projectDir)
	if err != nil {
		return contracts.WailsProjectManager{}, err
	}
	platform := currentPlatform()
	manager := contracts.WailsProjectManager{
		ProjectDir:      projectDir,
		CurrentPlatform: platform,
		WailsConfig:     wc,
	}
	return manager, nil
}

func LoadManagerWithConfigCompletion(projectDir string) (contracts.WailsProjectManager, error) {
	manager, err := LoadManager(projectDir)
	if err != nil {
		return contracts.WailsProjectManager{}, err
	}
	needsWrite, err := WailsConfigNeedsCompletion(projectDir)
	if err != nil {
		return contracts.WailsProjectManager{}, err
	}
	if !needsWrite {
		return manager, nil
	}
	return SaveManager(projectDir, manager)
}

func SaveManager(projectDir string, manager contracts.WailsProjectManager) (contracts.WailsProjectManager, error) {
	projectDir, err := fsx.NormalizePath(projectDir)
	if err != nil {
		return contracts.WailsProjectManager{}, err
	}
	manager.ProjectDir = projectDir
	if _, err := SaveWailsConfig(projectDir, manager.WailsConfig); err != nil {
		return contracts.WailsProjectManager{}, err
	}
	return LoadManager(projectDir)
}

func LoadTaskVars(projectDir string) contracts.WailsTaskVars {
	path := filepath.Join(projectDir, RootTaskfileRelPathForProject(projectDir))
	data, err := os.ReadFile(path)
	if err != nil {
		return contracts.WailsTaskVars{}
	}
	doc, err := parseYAMLDocument(data)
	if err != nil {
		return contracts.WailsTaskVars{}
	}
	var fields struct {
		Vars map[string]any `yaml:"vars"`
	}
	if err := doc.Content[0].Decode(&fields); err != nil {
		return contracts.WailsTaskVars{}
	}
	production := true // Wails beta templates build production unless DEV is true.
	if value, exists := fields.Vars["PRODUCTION"]; exists {
		production = ResolveBoolean(taskScalarValue(value))
	} else if value, exists := fields.Vars["DEV"]; exists {
		production = !ResolveBoolean(taskScalarValue(value))
	}
	return contracts.WailsTaskVars{
		AppName:    taskScalarValue(fields.Vars["APP_NAME"]),
		Production: production,
		CGOEnabled: taskScalarValue(fields.Vars["CGO_ENABLED"]),
	}
}

func taskScalarValue(value any) string {
	switch value := value.(type) {
	case string:
		return strings.TrimSpace(value)
	case bool, int, int64, uint64, float64:
		return fmt.Sprint(value)
	default:
		return ""
	}
}

func ResolveBoolean(value string) bool {
	value = strings.TrimSpace(strings.ToLower(value))
	switch value {
	case "true", "1", "yes", "y", "on", "enable", "enabled":
		return true
	default:
		return false
	}
}

func SaveWailsConfig(projectDir string, cfg contracts.WailsProjectConfig) (contracts.WailsProjectConfig, error) {
	projectDir, err := fsx.NormalizePath(projectDir)
	if err != nil {
		return contracts.WailsProjectConfig{}, err
	}
	path := ConfigPath(projectDir)
	data, err := os.ReadFile(path)
	if err != nil {
		return contracts.WailsProjectConfig{}, fmt.Errorf("failed to read build/config.yml: %w", err)
	}
	doc, err := parseYAMLDocument(data)
	if err != nil {
		return contracts.WailsProjectConfig{}, fmt.Errorf("invalid build/config.yml: %w", err)
	}
	root := doc.Content[0]
	if _, err := wailsConfigFromYAML(root); err != nil {
		return contracts.WailsProjectConfig{}, err
	}
	info := yamlMappingValue(root, "info")
	if info == nil {
		var inherited struct {
			Info *yaml.Node `yaml:"info"`
		}
		if err := root.Decode(&inherited); err != nil {
			return contracts.WailsProjectConfig{}, fmt.Errorf("invalid build/config.yml info: %w", err)
		}
		if inherited.Info != nil {
			return contracts.WailsProjectConfig{}, fmt.Errorf("build/config.yml info is inherited and cannot be edited safely")
		}
		info = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		appendYAMLValue(root, "info", info)
	} else if info.Kind == yaml.ScalarNode && info.Tag == "!!null" && info.Anchor == "" {
		info.Kind, info.Tag, info.Value, info.Style = yaml.MappingNode, "!!map", "", 0
	}
	// An anchored or aliased info mapping may also be used by another section.
	// Refuse that write rather than changing another section through the alias.
	if info.Kind != yaml.MappingNode || info.Anchor != "" {
		return contracts.WailsProjectConfig{}, fmt.Errorf("build/config.yml info must be an unanchored mapping to edit safely")
	}
	values := []string{
		cfg.Info.CompanyName, cfg.Info.ProductName, cfg.Info.ProductIdentifier,
		cfg.Info.Description, cfg.Info.Copyright, cfg.Info.Comments,
		normalizedVersion(cfg.Info.Version),
	}
	for i, key := range infoKeys {
		value := yamlMappingValue(info, key)
		if value == nil {
			appendYAMLValue(info, key, yamlString(values[i]))
			continue
		}
		if value.Anchor != "" {
			return contracts.WailsProjectConfig{}, fmt.Errorf("build/config.yml info.%s is anchored and cannot be edited safely", key)
		}
		head, line, foot := value.HeadComment, value.LineComment, value.FootComment
		*value = *yamlString(values[i])
		value.HeadComment, value.LineComment, value.FootComment = head, line, foot
	}
	// Validate the complete result before replacing the original file.
	if _, err := wailsConfigFromYAML(root); err != nil {
		return contracts.WailsProjectConfig{}, err
	}
	next, err := encodeYAMLDocument(doc, data)
	if err != nil {
		return contracts.WailsProjectConfig{}, fmt.Errorf("failed to encode build/config.yml: %w", err)
	}
	if err := writeConfigAtomically(path, next); err != nil {
		return contracts.WailsProjectConfig{}, fmt.Errorf("failed to write build/config.yml: %w", err)
	}
	return LoadWailsConfig(projectDir)
}
