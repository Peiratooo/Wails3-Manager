package environment

import (
	"bytes"
	"os/exec"
	"regexp"
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

func NewService() *EnvironmentService {
	return &EnvironmentService{}
}

func (s *EnvironmentService) CheckEnvironment() contracts.EnvironmentReport {
	probe := s.probe()

	report := contracts.EnvironmentReport{
		OS:     runtime.GOOS,
		Arch:   runtime.GOARCH,
		OK:     true,
		Checks: make([]contracts.ToolCheck, 0, 8),
	}

	report.Checks = append(report.Checks, commonEnvironmentChecks(probe)...)
	report.Checks = append(report.Checks, s.platformEnvironmentChecks(runtime.GOOS)...)

	report.OK = environmentOK(report.Checks)

	return report
}

func (s *EnvironmentService) probe() CommandProbe {
	if s.Probe != nil {
		return s.Probe
	}
	return defaultProbe{}
}

func commonEnvironmentChecks(probe CommandProbe) []contracts.ToolCheck {
	return []contracts.ToolCheck{
		commandCheck(probe, "go", "Go", "go", true),
		commandCheck(probe, "node", "Node.js", "node", true),
		commandCheck(probe, "npm", "npm", "npm", false),
		commandCheck(probe, "wails3", "Wails3 CLI", "wails3", true),
		commandCheck(probe, "git", "Git", "git", false),
	}
}

func (s *EnvironmentService) platformEnvironmentChecks(goos string) []contracts.ToolCheck {
	switch goos {
	case "windows":
		req := s.innoRequirement("")
		req.Required = true
		return []contracts.ToolCheck{toolRequirementCheck(req)}

	case "darwin":
		req := s.createDMGRequirement("")
		req.Required = true
		return []contracts.ToolCheck{toolRequirementCheck(req)}

	default:
		return nil
	}
}

func (s *EnvironmentService) innoRequirement(configured string) contracts.ToolRequirement {
	if s.Inno != nil {
		return s.Inno(configured)
	}
	return InnoRequirement(configured)
}

func (s *EnvironmentService) createDMGRequirement(configured string) contracts.ToolRequirement {
	if s.CreateDMG != nil {
		return s.CreateDMG(configured)
	}
	return CreateDMGRequirement(configured)
}

func environmentOK(checks []contracts.ToolCheck) bool {
	for _, check := range checks {
		if check.Required && !check.Found {
			return false
		}
	}
	return true
}

func commandCheck(probe CommandProbe, id string, name string, command string, required bool) contracts.ToolCheck {
	path, found := probe.LookPath(command)

	check := contracts.ToolCheck{
		ID:       id,
		Name:     name,
		Command:  command,
		Required: required,
		Found:    found,
		Path:     path,
	}

	if check.Found {
		check.Version = probe.Version(command)
	}

	return check
}

type defaultProbe struct{}

func (defaultProbe) LookPath(command string) (string, bool) {
	path, err := exec.LookPath(command)
	return path, err == nil
}

func (defaultProbe) Version(command string) string {
	return commandVersion(command)
}

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

type versionAttempt struct {
	args []string
}

var (
	ansiEscapePattern = regexp.MustCompile(`\x1b\[[0-9;]*[A-Za-z]`)
	versionPattern    = regexp.MustCompile(`v?\d+\.\d+\.\d+(?:[-+][0-9A-Za-z.-]+)*`)
)

func commandVersion(command string) string {
	for _, attempt := range versionAttempts(command) {
		output, ok := commandOutput(command, attempt.args)
		if !ok {
			continue
		}

		if version := normalizeCommandVersion(command, output); version != "" {
			return version
		}

		if line := firstOutputLine(output); line != "" {
			return line
		}
	}

	return ""
}

func versionAttempts(command string) []versionAttempt {
	switch command {
	case "go":
		return []versionAttempt{
			{args: []string{"version"}},
		}

	case "wails3":
		return []versionAttempt{
			{args: []string{"version"}},
			{args: []string{"--version"}},
			{args: []string{"-v"}},
		}

	default:
		return []versionAttempt{
			{args: []string{"--version"}},
		}
	}
}

func commandOutput(command string, args []string) (string, bool) {
	cmd := exec.Command(command, args...)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	if err != nil && out.Len() == 0 {
		return "", false
	}

	return out.String(), true
}

func normalizeCommandVersion(command string, output string) string {
	line := firstOutputLine(output)
	if line == "" {
		return ""
	}

	switch command {
	case "go":
		for _, field := range strings.Fields(line) {
			if strings.HasPrefix(field, "go") && hasDigit(field) {
				return field
			}
		}

	case "git":
		line = strings.TrimSpace(strings.TrimPrefix(line, "git version"))
		if fields := strings.Fields(line); len(fields) > 0 {
			return fields[0]
		}

	case "wails3":
		if match := versionPattern.FindString(line); match != "" {
			return match
		}
	}

	return line
}

func firstOutputLine(output string) string {
	line := ansiEscapePattern.ReplaceAllString(output, "")
	line = strings.TrimSpace(line)

	if idx := strings.IndexByte(line, '\n'); idx >= 0 {
		line = line[:idx]
	}

	line = strings.TrimSpace(line)

	if len(line) > 160 {
		line = line[:160]
	}

	return line
}

func hasDigit(value string) bool {
	for _, r := range value {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	return false
}
