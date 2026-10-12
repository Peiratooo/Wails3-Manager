package project

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

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
	data, err := os.ReadFile(ConfigPath(projectDir))
	if err != nil {
		return contracts.WailsProjectConfig{}, fmt.Errorf("failed to read build/config.yml: %w", err)
	}
	doc, err := parseYAMLDocument(data)
	if err != nil {
		return contracts.WailsProjectConfig{}, fmt.Errorf("invalid build/config.yml: %w", err)
	}
	return wailsConfigFromYAML(doc.Content[0])
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
	doc, err := parseYAMLDocument(data)
	if err != nil {
		return false, fmt.Errorf("invalid build/config.yml: %w", err)
	}
	var fields struct {
		Info map[string]any `yaml:"info"`
	}
	if err := doc.Content[0].Decode(&fields); err != nil {
		return false, fmt.Errorf("invalid build/config.yml info: %w", err)
	}
	for _, key := range infoKeys {
		if _, exists := fields.Info[key]; !exists {
			return true, nil
		}
	}
	return false, nil
}

var infoKeys = []string{"companyName", "productName", "productIdentifier", "description", "copyright", "comments", "version"}

func wailsConfigFromYAML(root *yaml.Node) (contracts.WailsProjectConfig, error) {
	var fields struct {
		Info struct {
			CompanyName       string `yaml:"companyName"`
			ProductName       string `yaml:"productName"`
			ProductIdentifier string `yaml:"productIdentifier"`
			Description       string `yaml:"description"`
			Copyright         string `yaml:"copyright"`
			Comments          string `yaml:"comments"`
			Version           string `yaml:"version"`
		} `yaml:"info"`
		FileAssociations []struct {
			Ext         string `yaml:"ext"`
			Name        string `yaml:"name"`
			Description string `yaml:"description"`
			IconName    string `yaml:"iconName"`
			Role        string `yaml:"role"`
			MimeType    string `yaml:"mimeType"`
		} `yaml:"fileAssociations"`
	}
	if err := root.Decode(&fields); err != nil {
		return contracts.WailsProjectConfig{}, fmt.Errorf("invalid build/config.yml: %w", err)
	}
	cfg := contracts.WailsProjectConfig{
		Info: contracts.WailsAppInfo{
			CompanyName: fields.Info.CompanyName, ProductName: fields.Info.ProductName,
			ProductIdentifier: fields.Info.ProductIdentifier, Description: fields.Info.Description,
			Copyright: fields.Info.Copyright, Comments: fields.Info.Comments,
			Version: fsx.StripVersionPrefix(fields.Info.Version),
		},
		Icon: filepath.ToSlash(DefaultAppIconRelPath),
	}
	for _, item := range fields.FileAssociations {
		cfg.FileAssociations = append(cfg.FileAssociations, contracts.WailsFileAssociation{
			Ext: item.Ext, Name: item.Name, Description: item.Description,
			IconName: item.IconName, Role: item.Role, MimeType: item.MimeType,
		})
	}
	return cfg, nil
}

func parseYAMLDocument(data []byte) (*yaml.Node, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	var doc yaml.Node
	if err := decoder.Decode(&doc); err != nil {
		return nil, err
	}
	if len(doc.Content) != 1 || doc.Content[0].Kind != yaml.MappingNode {
		return nil, fmt.Errorf("expected a YAML mapping")
	}
	var extra yaml.Node
	if err := decoder.Decode(&extra); err != io.EOF {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("expected a single YAML document")
	}
	// Decode once to reject duplicate keys before making any edits.
	var values map[string]any
	if err := doc.Content[0].Decode(&values); err != nil {
		return nil, err
	}
	return &doc, nil
}

func yamlMappingValue(mapping *yaml.Node, key string) *yaml.Node {
	if mapping == nil || mapping.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			return mapping.Content[i+1]
		}
	}
	return nil
}

func yamlString(value string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value, Style: yaml.DoubleQuotedStyle}
}

func appendYAMLValue(mapping *yaml.Node, key string, value *yaml.Node) {
	mapping.Content = append(mapping.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}, value)
}

func encodeYAMLDocument(doc *yaml.Node, original []byte) ([]byte, error) {
	var out bytes.Buffer
	encoder := yaml.NewEncoder(&out)
	encoder.SetIndent(2)
	if err := encoder.Encode(doc); err != nil {
		return nil, err
	}
	if err := encoder.Close(); err != nil {
		return nil, err
	}
	data := out.Bytes()
	if bytes.Contains(original, []byte("\r\n")) {
		data = bytes.ReplaceAll(data, []byte("\n"), []byte("\r\n"))
	}
	return data, nil
}

func writeConfigAtomically(path string, data []byte) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".config-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if err := file.Chmod(info.Mode().Perm()); err != nil {
		file.Close()
		return err
	}
	if _, err := file.Write(data); err != nil {
		file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}

func normalizedVersion(version string) string {
	version = strings.TrimSpace(version)
	if version == "" {
		return "0.0.1"
	}
	return version
}
