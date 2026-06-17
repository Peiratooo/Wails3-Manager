package packaging

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

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
	next := syncPackagingTaskfileText(string(data), cfg)
	if next == string(data) {
		return nil
	}
	mode := os.FileMode(0644)
	if info, statErr := os.Stat(path); statErr == nil {
		mode = info.Mode()
	}
	if err := os.WriteFile(path, []byte(next), mode); err != nil {
		return fmt.Errorf("failed to write packaging taskfile %s: %w", taskfile, err)
	}
	return nil
}

func syncPackagingTaskfileText(text string, cfg contracts.PackagingConfig) string {
	newline := "\n"
	if strings.Contains(text, "\r\n") {
		newline = "\r\n"
	}
	normalized := strings.ReplaceAll(text, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")
	hasFinalNewline := strings.HasSuffix(normalized, "\n")
	if hasFinalNewline && len(lines) > 0 {
		lines = lines[:len(lines)-1]
	}

	lines = ensurePackagingTaskfileVars(lines, cfg)
	lines = ensurePackagingTaskfileTasks(lines)

	out := strings.Join(lines, "\n")
	if hasFinalNewline {
		out += "\n"
	}
	if newline != "\n" {
		out = strings.ReplaceAll(out, "\n", newline)
	}
	return out
}

func ensurePackagingTaskfileVars(lines []string, cfg contracts.PackagingConfig) []string {
	values := map[string]string{
		"APP_NAME":    config.AppName(cfg),
		"CGO_ENABLED": cgoString(cfg.Build.CGOEnabled),
		"PRODUCTION":  boolString(cfg.Build.Production),
	}
	order := []string{"APP_NAME", "CGO_ENABLED", "PRODUCTION"}

	start, end, ok := findTopLevelBlock(lines, "vars")
	if !ok {
		insertAt := insertionAfterTopLevelKey(lines, "version")
		block := []string{
			"vars:",
			"  APP_NAME: " + quoteTaskfileString(values["APP_NAME"]),
			"  CGO_ENABLED: " + quoteTaskfileString(values["CGO_ENABLED"]),
			"  PRODUCTION: " + quoteTaskfileString(values["PRODUCTION"]),
			"",
		}
		return insertLines(lines, insertAt, block)
	}

	present := map[string]bool{}
	for i := start + 1; i < end; i++ {
		for _, key := range order {
			next, changed := setTaskfileVarLine(lines[i], key, values[key])
			if !changed {
				continue
			}
			lines[i] = next
			present[key] = true
			break
		}
	}

	var missing []string
	for _, key := range order {
		if !present[key] {
			missing = append(missing, "  "+key+": "+quoteTaskfileString(values[key]))
		}
	}
	if len(missing) == 0 {
		return lines
	}
	insertAt := end
	for insertAt > start+1 && strings.TrimSpace(lines[insertAt-1]) == "" {
		insertAt--
	}
	return insertLines(lines, insertAt, missing)
}

func ensurePackagingTaskfileTasks(lines []string) []string {
	start, end, ok := findTopLevelBlock(lines, "tasks")
	if !ok {
		block := append([]string{"tasks:"}, releaseTaskBlock()...)
		return appendWithGap(lines, block)
	}

	buildStart, buildEnd, hasBuild := findTaskBlock(lines, start, end, "build")
	if hasBuild && isManagedBuildTask(lines[buildStart:buildEnd]) {
		lines = replaceLines(lines, buildStart, buildEnd, buildTaskBlock())
		start, end, _ = findTopLevelBlock(lines, "tasks")
	}

	releaseStart, releaseEnd, hasRelease := findTaskBlock(lines, start, end, "release")
	if hasRelease {
		if isManagedReleaseTask(lines[releaseStart:releaseEnd]) {
			lines = replaceLines(lines, releaseStart, releaseEnd, releaseTaskBlock())
		}
	} else {
		insertAt := preferredTaskInsertionPoint(lines, start, end, "build")
		lines = insertLines(lines, insertAt, releaseTaskBlock())
	}

	return lines
}

func buildTaskBlock() []string {
	return []string{
		"  build:",
		"    summary: Builds the application",
		"    cmds:",
		"      - task: \"{{OS}}:build\"",
		"        vars:",
		"          APP_NAME: \"{{.APP_NAME}}\"",
		"          CGO_ENABLED: \"{{.CGO_ENABLED}}\"",
		"          PRODUCTION: \"{{.PRODUCTION}}\"",
		"",
	}
}

func releaseTaskBlock() []string {
	return []string{
		"  release:",
		"    summary: Builds the release application",
		"    cmds:",
		"      - task: \"{{OS}}:build\"",
		"        vars:",
		"          APP_NAME: \"{{.APP_NAME}}\"",
		"          CGO_ENABLED: \"{{.CGO_ENABLED}}\"",
		"          PRODUCTION: \"{{.PRODUCTION}}\"",
		"",
	}
}

func isManagedBuildTask(lines []string) bool {
	return hasSummary(lines, "Builds the application")
}

func isManagedReleaseTask(lines []string) bool {
	return hasSummary(lines, "Builds the release application")
}

func hasSummary(lines []string, summary string) bool {
	want := "summary: " + summary
	for _, line := range lines {
		if strings.TrimSpace(line) == want {
			return true
		}
	}
	return false
}

func findTopLevelBlock(lines []string, key string) (int, int, bool) {
	start := -1
	re := regexp.MustCompile(`^` + regexp.QuoteMeta(key) + `\s*:`)
	for i, line := range lines {
		if re.MatchString(line) {
			start = i
			break
		}
	}
	if start < 0 {
		return 0, 0, false
	}
	end := len(lines)
	for i := start + 1; i < len(lines); i++ {
		if isTopLevelKeyLine(lines[i]) {
			end = i
			break
		}
	}
	return start, end, true
}

func insertionAfterTopLevelKey(lines []string, key string) int {
	re := regexp.MustCompile(`^` + regexp.QuoteMeta(key) + `\s*:`)
	for i, line := range lines {
		if !re.MatchString(line) {
			continue
		}
		insertAt := i + 1
		for insertAt < len(lines) && strings.TrimSpace(lines[insertAt]) == "" {
			insertAt++
		}
		return insertAt
	}
	return 0
}

func findTaskBlock(lines []string, tasksStart, tasksEnd int, taskName string) (int, int, bool) {
	re := regexp.MustCompile(`^\s{2}` + regexp.QuoteMeta(taskName) + `\s*:`)
	start := -1
	for i := tasksStart + 1; i < tasksEnd; i++ {
		if re.MatchString(lines[i]) {
			start = i
			break
		}
	}
	if start < 0 {
		return 0, 0, false
	}
	end := tasksEnd
	for i := start + 1; i < tasksEnd; i++ {
		if isTopLevelKeyLine(lines[i]) || isTaskKeyLine(lines[i]) {
			end = i
			break
		}
	}
	return start, end, true
}

func preferredTaskInsertionPoint(lines []string, tasksStart, tasksEnd int, afterTask string) int {
	if _, afterEnd, ok := findTaskBlock(lines, tasksStart, tasksEnd, afterTask); ok {
		return afterEnd
	}
	insertAt := tasksEnd
	for insertAt > tasksStart+1 && strings.TrimSpace(lines[insertAt-1]) == "" {
		insertAt--
	}
	return insertAt
}

func setTaskfileVarLine(line, key, value string) (string, bool) {
	re := regexp.MustCompile(`^(\s{2}` + regexp.QuoteMeta(key) + `\s*:\s*)([^#]*)(.*)$`)
	parts := re.FindStringSubmatch(line)
	if len(parts) < 4 {
		return line, false
	}
	suffix := strings.TrimRight(parts[3], " \t")
	if strings.TrimSpace(suffix) != "" && !strings.HasPrefix(suffix, " ") {
		suffix = " " + suffix
	}
	return parts[1] + quoteTaskfileString(value) + suffix, true
}

func isTopLevelKeyLine(line string) bool {
	if line == "" || line[0] == ' ' || line[0] == '\t' {
		return false
	}
	trim := strings.TrimSpace(line)
	return trim != "" && !strings.HasPrefix(trim, "#") && strings.Contains(trim, ":")
}

func isTaskKeyLine(line string) bool {
	if !strings.HasPrefix(line, "  ") || strings.HasPrefix(line, "    ") {
		return false
	}
	trim := strings.TrimSpace(line)
	return trim != "" && !strings.HasPrefix(trim, "#") && !strings.HasPrefix(trim, "-") && strings.HasSuffix(trim, ":")
}

func quoteTaskfileString(value string) string {
	return strconv.Quote(strings.TrimSpace(value))
}

func insertLines(lines []string, at int, block []string) []string {
	if at < 0 {
		at = 0
	}
	if at > len(lines) {
		at = len(lines)
	}
	out := make([]string, 0, len(lines)+len(block))
	out = append(out, lines[:at]...)
	out = append(out, block...)
	out = append(out, lines[at:]...)
	return out
}

func replaceLines(lines []string, start, end int, block []string) []string {
	out := make([]string, 0, len(lines)-(end-start)+len(block))
	out = append(out, lines[:start]...)
	out = append(out, block...)
	out = append(out, lines[end:]...)
	return out
}

func appendWithGap(lines []string, block []string) []string {
	if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) != "" {
		lines = append(lines, "")
	}
	return append(lines, block...)
}
