// Tool detection lives beside EnvironmentService because both environment
// checks and packaging need the same executable lookup rules.
package environment

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"

	"wails3-manager/core/contracts"
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
		Required:       true,
		ConfiguredPath: strings.TrimSpace(configured),
		Path:           path,
		Found:          path != "",
		DownloadURL:    InnoDownloadURL,
		InstallHint:    "Windows 打包需要 Inno Setup。可以选择 ISCC.exe 或 Inno Setup 安装目录；也可以通过官方下载页或 winget 安装。",
		InstallCommand: []string{"winget", "install", "--id", "JRSoftware.InnoSetup", "-e", "-s", "winget", "-i"},
		CanAutoInstall: true,
		CanChoosePath:  true,
	}
	if req.Found {
		req.Message = "已找到 ISCC。"
		req.Version = executableVersion(path, []string{"/?"})
	} else {
		req.Message = "未找到 ISCC.exe，请选择 Inno Setup 安装目录 / ISCC.exe，或安装 Inno Setup。"
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
		Required:       true,
		ConfiguredPath: strings.TrimSpace(configured),
		Path:           path,
		Found:          path != "",
		DownloadURL:    CreateDMGDownloadURL,
		InstallHint:    "macOS DMG 打包必须使用 create-dmg。推荐通过 Homebrew 安装：brew install create-dmg。",
		InstallCommand: []string{"brew", "install", "create-dmg"},
		CanAutoInstall: true,
		CanChoosePath:  true,
	}
	if req.Found {
		req.Message = "已找到 create-dmg。"
		req.Version = executableVersion(path, []string{"--version"})
	} else {
		req.Message = "未找到 create-dmg，请选择可执行文件，或执行 brew install create-dmg。"
	}
	return req
}

func DetectISCC(configured string) string {
	configured = strings.TrimSpace(strings.Trim(configured, `"`))
	for _, p := range innoCandidates(configured) {
		if resolved := resolveExecutableCandidate(p, "ISCC.exe"); resolved != "" {
			return resolved
		}
	}
	for _, cmd := range []string{"ISCC.exe", "iscc"} {
		if found, err := exec.LookPath(cmd); err == nil {
			return found
		}
	}
	return ""
}

func DetectCreateDMG(configured string) string {
	configured = strings.TrimSpace(strings.Trim(configured, `"`))
	for _, p := range createDMGCandidates(configured) {
		if resolved := resolveExecutableCandidate(p, "create-dmg"); resolved != "" {
			return resolved
		}
	}
	if found, err := exec.LookPath("create-dmg"); err == nil {
		return found
	}
	return ""
}

func innoCandidates(configured string) []string {
	out := []string{}
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
	out := []string{}
	if configured != "" {
		out = append(out, configured)
	}
	return append(out,
		"/opt/homebrew/bin/create-dmg",
		"/usr/local/bin/create-dmg",
		"/usr/bin/create-dmg",
	)
}

func resolveExecutableCandidate(candidate, executableName string) string {
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

func executableVersion(path string, args []string) string {
	if path == "" {
		return ""
	}
	cmd := exec.Command(path, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if cmd.Run() != nil {
		return ""
	}
	line := strings.TrimSpace(out.String())
	if idx := strings.IndexByte(line, '\n'); idx >= 0 {
		line = line[:idx]
	}
	if len(line) > 160 {
		line = line[:160]
	}
	return line
}
