package execenv

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

const loginShellTimeout = 2 * time.Second

var (
	loginShellPathOnce  sync.Once
	loginShellPathValue string
)

// LookPath searches the GUI process environment plus common developer-tool
// locations that are often missing when a desktop app is launched outside a
// terminal.
func LookPath(command string) (string, error) {
	command = strings.TrimSpace(command)
	if command == "" {
		return "", exec.ErrNotFound
	}

	if path, err := exec.LookPath(command); err == nil {
		return path, nil
	}

	if hasPathSeparator(command) {
		if isExecutable(command) {
			return command, nil
		}
		return "", exec.ErrNotFound
	}

	for _, dir := range SearchDirs() {
		if path := findExecutable(dir, command); path != "" {
			return path, nil
		}
	}

	return "", exec.ErrNotFound
}

func Environ(overrides map[string]string) []string {
	env := envMap(os.Environ())
	env["PATH"] = EnrichedPath()

	for key, value := range overrides {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		key = canonicalEnvKey(key)
		if strings.EqualFold(key, "PATH") {
			env[key] = mergePath(value, env["PATH"])
			continue
		}
		env[key] = value
	}

	keys := make([]string, 0, len(env))
	for key := range env {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	out := make([]string, 0, len(keys))
	for _, key := range keys {
		out = append(out, key+"="+env[key])
	}
	return out
}

func EnrichedPath() string {
	return strings.Join(SearchDirs(), string(os.PathListSeparator))
}

func SearchDirs() []string {
	var dirs []string
	addPathList(&dirs, os.Getenv("PATH"))
	addPathList(&dirs, loginShellPath())
	addCommonToolDirs(&dirs)
	return uniqueDirs(dirs)
}

func addCommonToolDirs(dirs *[]string) {
	home, _ := os.UserHomeDir()

	switch runtime.GOOS {
	case "windows":
		addWindowsToolDirs(dirs, home)
	default:
		addUnixToolDirs(dirs, home)
	}
}

func addUnixToolDirs(dirs *[]string, home string) {
	addDirs(dirs,
		"/opt/homebrew/bin",
		"/opt/homebrew/sbin",
		"/usr/local/bin",
		"/usr/local/sbin",
		"/usr/local/go/bin",
		"/usr/bin",
		"/bin",
		"/usr/sbin",
		"/sbin",
	)

	if home == "" {
		return
	}

	addDirs(dirs,
		filepath.Join(home, "go", "bin"),
		filepath.Join(home, "bin"),
		filepath.Join(home, ".local", "bin"),
		filepath.Join(home, ".cargo", "bin"),
		filepath.Join(home, ".volta", "bin"),
		filepath.Join(home, ".bun", "bin"),
		filepath.Join(home, ".deno", "bin"),
		filepath.Join(home, ".asdf", "shims"),
		filepath.Join(home, ".local", "share", "mise", "shims"),
		filepath.Join(home, ".config", "mise", "shims"),
	)
	addGlobDirs(dirs, filepath.Join(home, ".nvm", "versions", "node", "*", "bin"))
	addGlobDirs(dirs, filepath.Join(home, ".local", "share", "fnm", "node-versions", "*", "installation", "bin"))
}

func addWindowsToolDirs(dirs *[]string, home string) {
	programFiles := os.Getenv("ProgramFiles")
	programFilesX86 := os.Getenv("ProgramFiles(x86)")
	localAppData := os.Getenv("LocalAppData")
	appData := os.Getenv("AppData")

	addUnder := func(base string, parts ...string) {
		if strings.TrimSpace(base) == "" {
			return
		}
		addDirs(dirs, filepath.Join(append([]string{base}, parts...)...))
	}

	addUnder(programFiles, "Go", "bin")
	addUnder(programFiles, "nodejs")
	addUnder(programFiles, "Git", "cmd")
	addUnder(programFiles, "Git", "bin")
	addUnder(programFiles, "Inno Setup 7")
	addUnder(programFiles, "Inno Setup 6")
	addUnder(programFilesX86, "Inno Setup 7")
	addUnder(programFilesX86, "Inno Setup 6")
	addUnder(localAppData, "Programs", "Git", "cmd")
	addUnder(localAppData, "Programs", "Git", "bin")
	addUnder(appData, "npm")

	if home != "" {
		addDirs(dirs,
			filepath.Join(home, "go", "bin"),
			filepath.Join(home, ".cargo", "bin"),
			filepath.Join(home, "AppData", "Roaming", "npm"),
		)
	}
}

func loginShellPath() string {
	if runtime.GOOS == "windows" {
		return ""
	}

	loginShellPathOnce.Do(func() {
		shell := strings.TrimSpace(os.Getenv("SHELL"))
		if shell == "" {
			for _, candidate := range []string{"/bin/zsh", "/bin/bash", "/bin/sh"} {
				if isExecutable(candidate) {
					shell = candidate
					break
				}
			}
		}
		if shell == "" || !isExecutable(shell) {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), loginShellTimeout)
		defer cancel()

		cmd := exec.CommandContext(ctx, shell, "-l", "-c", `printf '%s\n' "$PATH"`)
		cmd.Env = os.Environ()

		out, err := cmd.Output()
		if err != nil {
			return
		}

		loginShellPathValue = lastPathLikeLine(string(out))
	})

	return loginShellPathValue
}

func lastPathLikeLine(output string) string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		if strings.Contains(line, string(os.PathListSeparator)) {
			return line
		}
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.TrimSpace(lines[len(lines)-1])
}

func addPathList(dirs *[]string, value string) {
	for _, dir := range filepath.SplitList(value) {
		addDirs(dirs, dir)
	}
}

func addDirs(dirs *[]string, values ...string) {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		*dirs = append(*dirs, filepath.Clean(value))
	}
}

func addGlobDirs(dirs *[]string, pattern string) {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return
	}
	for _, match := range matches {
		info, err := os.Stat(match)
		if err == nil && info.IsDir() {
			addDirs(dirs, match)
		}
	}
}

func uniqueDirs(dirs []string) []string {
	seen := make(map[string]struct{}, len(dirs))
	out := make([]string, 0, len(dirs))
	for _, dir := range dirs {
		key := filepath.Clean(strings.TrimSpace(dir))
		if key == "." || key == "" {
			continue
		}
		if runtime.GOOS == "windows" {
			key = strings.ToLower(key)
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, filepath.Clean(dir))
	}
	return out
}

func findExecutable(dir string, command string) string {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return ""
	}

	for _, name := range executableNames(command) {
		path := filepath.Join(dir, name)
		if isExecutable(path) {
			return path
		}
	}
	return ""
}

func executableNames(command string) []string {
	if runtime.GOOS != "windows" || filepath.Ext(command) != "" {
		return []string{command}
	}

	exts := filepath.SplitList(os.Getenv("PATHEXT"))
	if len(exts) == 0 {
		exts = strings.Split(".COM;.EXE;.BAT;.CMD", ";")
	}

	names := []string{command}
	seen := map[string]struct{}{strings.ToLower(command): {}}
	for _, ext := range exts {
		ext = strings.TrimSpace(ext)
		if ext == "" {
			continue
		}
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		name := command + ext
		key := strings.ToLower(name)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		names = append(names, name)
	}
	return names
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	if runtime.GOOS == "windows" {
		return true
	}
	return info.Mode()&0111 != 0
}

func hasPathSeparator(value string) bool {
	return strings.ContainsAny(value, `/\`)
}

func envMap(pairs []string) map[string]string {
	env := make(map[string]string, len(pairs))
	for _, pair := range pairs {
		key, value, ok := strings.Cut(pair, "=")
		if !ok || strings.TrimSpace(key) == "" {
			continue
		}
		key = canonicalEnvKey(key)
		env[key] = value
	}
	return env
}

func canonicalEnvKey(key string) string {
	if runtime.GOOS == "windows" && strings.EqualFold(key, "PATH") {
		return "PATH"
	}
	return key
}

func mergePath(first string, second string) string {
	var dirs []string
	addPathList(&dirs, first)
	addPathList(&dirs, second)
	return strings.Join(uniqueDirs(dirs), string(os.PathListSeparator))
}
