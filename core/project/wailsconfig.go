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

const DefaultConfigRelPath = "build/config.yml"
const DefaultAppIconRelPath = "build/appicon.png"

func ConfigPath(projectDir string) string {
	return filepath.Join(projectDir, DefaultConfigRelPath)
}

func LoadWailsConfig(projectDir string) (contracts.WailsProjectConfig, error) {
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
	cfg := contracts.WailsProjectConfig{
		Info:             parseInfo(text),
		Icon:             filepath.ToSlash(DefaultAppIconRelPath),
		FileAssociations: parseFileAssociations(text),
	}
	return cfg, nil
}

func WailsConfigNeedsCompletion(projectDir string) (bool, error) {
	projectDir, err := fsx.NormalizePath(projectDir)
	if err != nil {
		return false, err
	}
	data, err := os.ReadFile(ConfigPath(projectDir))
	if err != nil {
		return false, fmt.Errorf("failed to read build/config.yml: %w", err)
	}
	text := string(data)
	info := sectionBlock(text, "info", 0)
	if strings.TrimSpace(info) == "" {
		return true, nil
	}
	for _, key := range []string{
		"companyName",
		"productName",
		"productIdentifier",
		"description",
		"copyright",
		"comments",
		"version",
	} {
		if !regexp.MustCompile(`(?m)^\s{2}` + regexp.QuoteMeta(key) + `\s*:`).MatchString(info) {
			return true, nil
		}
	}
	return false, nil
}

func parseInfo(text string) contracts.WailsAppInfo {
	return contracts.WailsAppInfo{
		CompanyName:       getInfoValue(text, "companyName"),
		ProductName:       getInfoValue(text, "productName"),
		ProductIdentifier: getInfoValue(text, "productIdentifier"),
		Description:       getInfoValue(text, "description"),
		Copyright:         getInfoValue(text, "copyright"),
		Comments:          getInfoValue(text, "comments"),
		Version:           fsx.StripVersionPrefix(getInfoValue(text, "version")),
	}
}

func getInfoValue(text, key string) string {
	re := regexp.MustCompile(`(?m)^\s{2}` + regexp.QuoteMeta(key) + `\s*:\s*([^#\n]+)`)
	m := re.FindStringSubmatch(text)
	if len(m) < 2 {
		return ""
	}
	return cleanYAMLScalar(m[1])
}

func setInfoValue(text, key, value string) string {
	value = quoteYAML(value)
	re := regexp.MustCompile(`(?m)^(\s{2}` + regexp.QuoteMeta(key) + `\s*:\s*)([^#\n]*)(.*)$`)
	if !re.MatchString(text) {
		return insertInfoLine(text, key, value)
	}
	return re.ReplaceAllStringFunc(text, func(line string) string {
		parts := re.FindStringSubmatch(line)
		if len(parts) < 4 {
			return line
		}
		suffix := strings.TrimRight(parts[3], " \t")
		if strings.TrimSpace(suffix) != "" && !strings.HasPrefix(suffix, " ") {
			suffix = " " + suffix
		}
		return parts[1] + value + suffix
	})
}

func ensureInfoSection(text string) string {
	if regexp.MustCompile(`(?m)^info\s*:`).MatchString(text) {
		return text
	}
	prefix := "info:\n"
	if strings.HasPrefix(text, "version:") {
		lines := strings.SplitAfter(text, "\n")
		if len(lines) > 0 {
			return lines[0] + "\n" + prefix + strings.Join(lines[1:], "")
		}
	}
	return prefix + text
}

func insertInfoLine(text, key, value string) string {
	lines := strings.SplitAfter(text, "\n")
	insertAt := -1
	inInfo := false
	for i, line := range lines {
		trim := strings.TrimSpace(line)
		if trim == "info:" {
			inInfo = true
			insertAt = i + 1
			continue
		}
		if inInfo {
			if trim != "" && !strings.HasPrefix(line, "  ") && !strings.HasPrefix(trim, "#") {
				break
			}
			insertAt = i + 1
		}
	}
	if insertAt < 0 {
		return text + "\ninfo:\n  " + key + ": " + value + "\n"
	}
	line := "  " + key + ": " + value + "\n"
	lines = append(lines[:insertAt], append([]string{line}, lines[insertAt:]...)...)
	return strings.Join(lines, "")
}

func cleanYAMLScalar(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, `"'`)
	return value
}

func quoteYAML(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	return `"` + value + `"`
}

func normalizedVersion(version string) string {
	version = strings.TrimSpace(version)
	if version == "" {
		return "0.0.1"
	}
	return version
}
