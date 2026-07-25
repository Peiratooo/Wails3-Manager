// Tool detection lives beside the environment service because both environment
// checks and packaging need the same executable lookup rules.
package environment

import (
	"os/exec"
	"path/filepath"
	"strings"

	"wails3-manager/core/contracts"
	"wails3-manager/core/execenv"
	"wails3-manager/core/fsx"
)

const (
	InnoDownloadURL      = "https://jrsoftware.org/isdl.php/Inno-Setup-Downloads"
	CreateDMGDownloadURL = "https://github.com/create-dmg/create-dmg"
)

func InnoRequirement(configured string) contracts.ToolRequirement {
	path := DetectISCC(configured)

	req := contracts.ToolRequirement{
		ID:             "inno",
		Name:           "Inno Setup / ISCC",
		Platform:       contracts.PlatformWindows,
		Command:        "ISCC.exe",
		Required:       false,
		ConfiguredPath: cleanConfiguredPath(configured),
		Path:           path,
		Found:          path != "",
		DownloadURL:    InnoDownloadURL,
		InstallHint:    "Windows packaging requires Inno Setup. Choose ISCC.exe or the Inno Setup install directory, or install it from the official download page or winget.",
		InstallCommand: []string{"winget", "install", "--id", "JRSoftware.InnoSetup", "-e", "-s", "winget", "-i"},
		CanAutoInstall: true,
		CanChoosePath:  true,
	}

	if req.Found {
		req.Message = "ISCC was found."
		req.Version = executableVersion(path, []string{"/?"})
	} else {
		req.Message = "ISCC.exe was not found. Choose an Inno Setup install directory / ISCC.exe, or install Inno Setup."
	}

	return req
}

func CreateDMGRequirement(configured string) contracts.ToolRequirement {
	path := DetectCreateDMG(configured)

	req := contracts.ToolRequirement{
		ID:             "create-dmg",
		Name:           "create-dmg",
		Platform:       contracts.PlatformMacOS,
		Command:        "create-dmg",
		Required:       false,
		ConfiguredPath: cleanConfiguredPath(configured),
		Path:           path,
		Found:          path != "",
		DownloadURL:    CreateDMGDownloadURL,
		InstallHint:    "macOS DMG packaging requires create-dmg. Install it with Homebrew: brew install create-dmg.",
		InstallCommand: []string{"brew", "install", "create-dmg"},
		CanAutoInstall: true,
		CanChoosePath:  true,
	}

	if req.Found {
		req.Message = "create-dmg was found."
		req.Version = executableVersion(path, []string{"--version"})
	} else {
		req.Message = "create-dmg was not found. Choose the executable, or run brew install create-dmg."
	}

	return req
}

func DetectISCC(configured string) string {
	configured = cleanConfiguredPath(configured)

	for _, p := range innoCandidates(configured) {
		if resolved := resolveExecutableCandidate(p, "ISCC.exe"); resolved != "" {
			return resolved
		}
	}

	for _, command := range []string{"ISCC.exe", "iscc"} {
		if found, err := execenv.LookPath(command); err == nil {
			return found
		}
	}

	return ""
}

func DetectCreateDMG(configured string) string {
	configured = cleanConfiguredPath(configured)

	for _, p := range createDMGCandidates(configured) {
		if resolved := resolveExecutableCandidate(p, "create-dmg"); resolved != "" {
			return resolved
		}
	}

	if found, err := execenv.LookPath("create-dmg"); err == nil {
		return found
	}

	return ""
}

func innoCandidates(configured string) []string {
	out := make([]string, 0, 9)

	if configured != "" {
		out = append(out, configured)
	}

	return append(out,
		`C:\Program Files\Inno Setup 7\ISCC.exe`,
		`C:\Program Files (x86)\Inno Setup 7\ISCC.exe`,
		`C:\Program Files\Inno Setup 6\ISCC.exe`,
		`C:\Program Files (x86)\Inno Setup 6\ISCC.exe`,
		`C:\Program Files\Inno Setup 7`,
		`C:\Program Files (x86)\Inno Setup 7`,
		`C:\Program Files\Inno Setup 6`,
		`C:\Program Files (x86)\Inno Setup 6`,
	)
}

func createDMGCandidates(configured string) []string {
	out := make([]string, 0, 4)

	if configured != "" {
		out = append(out, configured)
	}

	return append(out,
		"/opt/homebrew/bin/create-dmg",
		"/usr/local/bin/create-dmg",
		"/usr/bin/create-dmg",
	)
}

func resolveExecutableCandidate(candidate string, executableName string) string {
	candidate = cleanConfiguredPath(candidate)
	if candidate == "" {
		return ""
	}

	if fsx.FileExists(candidate) {
		return candidate
	}

	if fsx.DirExists(candidate) {
		joined := filepath.Join(candidate, executableName)
		if fsx.FileExists(joined) {
			return joined
		}
	}

	return ""
}

func cleanConfiguredPath(value string) string {
	return strings.TrimSpace(strings.Trim(value, `"`))
}

func executableVersion(path string, args []string) string {
	path = cleanConfiguredPath(path)
	if path == "" {
		return ""
	}

	cmd := exec.Command(path, args...)
	execenv.HideWindow(cmd)
	cmd.Env = execenv.Environ(nil)

	out, err := cmd.CombinedOutput()
	if err != nil && len(out) == 0 {
		return ""
	}
	return firstOutputLine(string(out))
}
