package project

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"wails3-manager/core/contracts"
	"wails3-manager/core/fsx"
)

const RootTaskfileRelPath = "Taskfile.yml"

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
	path := filepath.Join(projectDir, RootTaskfileRelPath)
	data, err := os.ReadFile(path)
	if err != nil {
		return contracts.WailsTaskVars{}
	}
	text := string(data)
	return contracts.WailsTaskVars{
		AppName:    cleanYAMLScalar(getVarValue(text, "APP_NAME")),
		Production: ResolveBoolean(cleanYAMLScalar(getVarValue(text, "PRODUCTION"))),
		CGOEnabled: cleanYAMLScalar(getVarValue(text, "CGO_ENABLED")),
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

func parseFileAssociations(text string) []contracts.WailsFileAssociation {
	block := sectionBlock(text, "fileAssociations", 0)
	if strings.TrimSpace(block) == "" {
		return nil
	}
	items := regexp.MustCompile(`(?m)^\s*-\s+ext\s*:`).FindAllStringIndex(block, -1)
	if len(items) == 0 {
		return nil
	}
	var out []contracts.WailsFileAssociation
	for i, idx := range items {
		start := idx[0]
		end := len(block)
		if i+1 < len(items) {
			end = items[i+1][0]
		}
		item := block[start:end]
		out = append(out, contracts.WailsFileAssociation{
			Ext:         getInlineYAMLValue(item, "ext"),
			Name:        getInlineYAMLValue(item, "name"),
			Description: getInlineYAMLValue(item, "description"),
			IconName:    getInlineYAMLValue(item, "iconName"),
			Role:        getInlineYAMLValue(item, "role"),
			MimeType:    getInlineYAMLValue(item, "mimeType"),
		})
	}
	return out
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
	text := string(data)
	text = ensureInfoSection(text)
	updates := map[string]string{
		"companyName":       cfg.Info.CompanyName,
		"productName":       cfg.Info.ProductName,
		"productIdentifier": cfg.Info.ProductIdentifier,
		"description":       cfg.Info.Description,
		"copyright":         cfg.Info.Copyright,
		"comments":          cfg.Info.Comments,
		"version":           ensureVersionPrefix(cfg.Info.Version),
	}
	for key, value := range updates {
		text = setInfoValue(text, key, value)
	}
	if err := os.WriteFile(path, []byte(text), 0644); err != nil {
		return contracts.WailsProjectConfig{}, fmt.Errorf("failed to write build/config.yml: %w", err)
	}
	return LoadWailsConfig(projectDir)
}

func getVarValue(text, key string) string {
	re := regexp.MustCompile(`(?m)^\s{2}` + regexp.QuoteMeta(key) + `\s*:\s*(.+)$`)
	m := re.FindStringSubmatch(text)
	if len(m) < 2 {
		return ""
	}
	return strings.TrimSpace(m[1])
}

func sectionBlock(text, key string, indent int) string {
	prefix := strings.Repeat(" ", indent)
	startRe := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(prefix+key) + `\s*:\s*(?:#.*)?$`)
	loc := startRe.FindStringIndex(text)
	if loc == nil {
		return ""
	}
	start := loc[0]
	lines := strings.SplitAfter(text[start:], "\n")
	var b strings.Builder
	for i, line := range lines {
		trim := strings.TrimSpace(line)
		if i > 0 && trim != "" && !strings.HasPrefix(line, prefix+" ") && !strings.HasPrefix(trim, "#") {
			break
		}
		b.WriteString(line)
	}
	return b.String()
}

func getInlineYAMLValue(text, key string) string {
	re := regexp.MustCompile(`(?m)` + regexp.QuoteMeta(key) + `\s*:\s*([^#\n]+)`)
	m := re.FindStringSubmatch(text)
	if len(m) < 2 {
		return ""
	}
	return cleanYAMLScalar(m[1])
}
