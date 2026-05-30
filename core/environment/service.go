package environment

import (
	"bytes"
	"os/exec"
	"runtime"
	"strings"

	"wails3-manager/core/contracts"
)

type CommandProbe interface {
	LookPath(command string) (string, bool)
	Version(command string) string
}

type EnvironmentService struct {
	Probe     CommandProbe
	Inno      func(string) contracts.ToolRequirement
	CreateDMG func(string) contracts.ToolRequirement
}

func NewService() *EnvironmentService { return &EnvironmentService{} }

func (s *EnvironmentService) CheckEnvironment() contracts.EnvironmentReport {
	probe := s.probe()
	report := contracts.EnvironmentReport{OS: runtime.GOOS, Arch: runtime.GOARCH, OK: true}
	report.Checks = append(report.Checks,
		commandCheck(probe, "go", "Go", "go", true),
		commandCheck(probe, "node", "Node.js", "node", true),
		commandCheck(probe, "npm", "npm", "npm", false),
		commandCheck(probe, "pnpm", "pnpm", "pnpm", false),
		commandCheck(probe, "yarn", "yarn", "yarn", false),
		commandCheck(probe, "wails3", "Wails3 CLI", "wails3", true),
		commandCheck(probe, "git", "Git", "git", false),
	)
	innoCheck := InnoRequirement
	if s.Inno != nil {
		innoCheck = s.Inno
	}
	createDMGCheck := CreateDMGRequirement
	if s.CreateDMG != nil {
		createDMGCheck = s.CreateDMG
	}
	inno := innoCheck("")
	inno.Required = runtime.GOOS == "windows"
	report.Checks = append(report.Checks, toolRequirementCheck(inno))
	dmg := createDMGCheck("")
	dmg.Required = runtime.GOOS == "darwin"
	report.Checks = append(report.Checks, toolRequirementCheck(dmg))
	for _, check := range report.Checks {
		if check.Required && !check.Found {
			report.OK = false
			break
		}
	}
	return report
}

func (s *EnvironmentService) probe() CommandProbe {
	if s.Probe != nil {
		return s.Probe
	}
	return defaultProbe{}
}

func commandCheck(probe CommandProbe, id, name, command string, required bool) contracts.ToolCheck {
	path, found := probe.LookPath(command)
	check := contracts.ToolCheck{ID: id, Name: name, Command: command, Required: required, Found: found, Path: path}
	if check.Found {
		check.Version = probe.Version(command)
		return check
	}
	return check
}

type defaultProbe struct{}

func (defaultProbe) LookPath(command string) (string, bool) {
	path, err := exec.LookPath(command)
	return path, err == nil
}

func (defaultProbe) Version(command string) string { return commandVersion(command) }

func toolRequirementCheck(req contracts.ToolRequirement) contracts.ToolCheck {
	return contracts.ToolCheck{
		ID:       req.ID,
		Name:     req.Name,
		Command:  req.Command,
		Found:    req.Found,
		Version:  req.Version,
		Path:     req.Path,
		Required: req.Required,
		Platform: req.Platform,
	}
}

func commandVersion(command string) string {
	args := []string{"--version"}
	cmd := exec.Command(command, args...)
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
