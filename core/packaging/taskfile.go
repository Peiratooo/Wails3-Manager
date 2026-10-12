package packaging

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
	"wails3-manager/core/packaging/config"
)

func syncPackagingTaskfile(projectDir string, cfg contracts.PackagingConfig) error {
	taskfile := strings.TrimSpace(cfg.Build.Taskfile)
	if taskfile == "" {
		return nil
	}
	path := fsx.ResolveProjectFile(projectDir, filepath.ToSlash(taskfile))
	if path == "" {
		return fmt.Errorf("packaging build.taskfile must be a project-relative path: %s", taskfile)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read packaging taskfile %s: %w", taskfile, err)
	}
	next, err := syncPackagingTaskfileText(string(data), cfg)
	if err != nil {
		return fmt.Errorf("failed to sync packaging taskfile %s: %w", taskfile, err)
	}
	if next == string(data) {
		return nil
	}
	if err := writeTaskfileAtomically(path, []byte(next)); err != nil {
		return fmt.Errorf("failed to write packaging taskfile %s: %w", taskfile, err)
	}
	return nil
}

func syncPackagingTaskfileText(text string, cfg contracts.PackagingConfig) (string, error) {
	decoder := yaml.NewDecoder(strings.NewReader(text))
	var doc yaml.Node
	if err := decoder.Decode(&doc); err != nil {
		return "", err
	}
	if len(doc.Content) != 1 || doc.Content[0].Kind != yaml.MappingNode {
		return "", fmt.Errorf("expected a YAML mapping")
	}
	var extra yaml.Node
	if err := decoder.Decode(&extra); err != io.EOF {
		if err != nil {
			return "", err
		}
		return "", fmt.Errorf("expected a single YAML document")
	}
	root := doc.Content[0]
	// Resolve merged keys for presence checks and reject duplicate keys before editing.
	var effective map[string]any
	if err := root.Decode(&effective); err != nil {
		return "", err
	}
	changed := false
	vars := taskfileMappingValue(root, "vars")
	if vars == nil && effective["vars"] == nil {
		vars = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		appendTaskfileValue(root, "vars", vars)
		changed = true
	} else if vars != nil && vars.Kind == yaml.ScalarNode && vars.Tag == "!!null" && vars.Anchor == "" {
		vars.Kind, vars.Tag, vars.Value, vars.Style = yaml.MappingNode, "!!map", "", 0
		changed = true
	}
	// Keep dynamic variables, aliases and anchored mappings intact. Build settings
	// are also passed to Task on its command line, so these need no destructive rewrite.
	if vars != nil && vars.Kind == yaml.MappingNode && vars.Anchor == "" {
		var inherited map[string]any
		if err := vars.Decode(&inherited); err != nil {
			return "", err
		}
		values := []struct{ key, value string }{
			{"APP_NAME", config.AppName(cfg)},
			{"CGO_ENABLED", cgoString(cfg.Build.CGOEnabled)},
			{"PRODUCTION", boolString(cfg.Build.Production)},
		}
		for _, update := range values {
			node := taskfileMappingValue(vars, update.key)
			if node == nil {
				if _, exists := inherited[update.key]; !exists {
					appendTaskfileValue(vars, update.key, taskfileString(update.value))
					changed = true
				}
				continue
			}
			if !isLiteralTaskfileVar(node) || (node.Tag == "!!str" && node.Value == update.value) {
				continue
			}
			node.Tag, node.Value, node.Style = "!!str", update.value, yaml.DoubleQuotedStyle
			changed = true
		}
	}

	tasks := taskfileMappingValue(root, "tasks")
	if tasks == nil && effective["tasks"] == nil {
		tasks = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		appendTaskfileValue(root, "tasks", tasks)
		changed = true
	} else if tasks != nil && tasks.Kind == yaml.ScalarNode && tasks.Tag == "!!null" && tasks.Anchor == "" {
		tasks.Kind, tasks.Tag, tasks.Value, tasks.Style = yaml.MappingNode, "!!map", "", 0
		changed = true
	}
	if tasks != nil && tasks.Kind == yaml.MappingNode && tasks.Anchor == "" {
		var existing map[string]any
		if err := tasks.Decode(&existing); err != nil {
			return "", err
		}
		for _, name := range []string{"build", "release"} {
			if _, exists := existing[name]; exists {
				continue
			}
			appendTaskfileValue(tasks, name, defaultPackagingTask(name))
			changed = true
		}
	}
	if !changed {
		return text, nil
	}
	var out bytes.Buffer
	encoder := yaml.NewEncoder(&out)
	encoder.SetIndent(2)
	if err := encoder.Encode(&doc); err != nil {
		return "", err
	}
	if err := encoder.Close(); err != nil {
		return "", err
	}
	next := out.String()
	if !strings.HasSuffix(text, "\n") {
		next = strings.TrimSuffix(next, "\n")
	}
	if strings.Contains(text, "\r\n") {
		next = strings.ReplaceAll(next, "\n", "\r\n")
	}
	return next, nil
}

func isLiteralTaskfileVar(node *yaml.Node) bool {
	if node.Kind != yaml.ScalarNode || node.Anchor != "" || strings.Contains(node.Value, "{{") {
		return false
	}
	switch node.Tag {
	case "!!str", "!!bool", "!!int", "!!null":
		return node.Style&(yaml.LiteralStyle|yaml.FoldedStyle) == 0
	default:
		return false
	}
}

func taskfileMappingValue(mapping *yaml.Node, key string) *yaml.Node {
	if mapping.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			return mapping.Content[i+1]
		}
	}
	return nil
}

func appendTaskfileValue(mapping *yaml.Node, key string, value *yaml.Node) {
	mapping.Content = append(mapping.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}, value)
}

func taskfileString(value string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value, Style: yaml.DoubleQuotedStyle}
}

func defaultPackagingTask(name string) *yaml.Node {
	summary := "Builds the application"
	if name == "release" {
		summary = "Builds the release application"
	}
	task := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	appendTaskfileValue(task, "summary", &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: summary})
	command := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	appendTaskfileValue(command, "task", taskfileString("{{OS}}:build"))
	vars := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	for _, key := range []string{"APP_NAME", "CGO_ENABLED", "PRODUCTION"} {
		appendTaskfileValue(vars, key, taskfileString("{{."+key+"}}"))
	}
	appendTaskfileValue(command, "vars", vars)
	appendTaskfileValue(task, "cmds", &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq", Content: []*yaml.Node{command}})
	return task
}

func writeTaskfileAtomically(path string, data []byte) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".taskfile-*.tmp")
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
