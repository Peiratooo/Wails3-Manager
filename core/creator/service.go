package creator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"wails3-manager/core/contracts"
	"wails3-manager/core/fsx"
	"wails3-manager/core/runlog"
)

type CommandRunner interface {
	Run(ctx context.Context, workDir string, command []string) error
}

type CreatorService struct {
	Log    *runlog.Logger
	Runner CommandRunner
}

func NewService(log *runlog.Logger) *CreatorService {
	return &CreatorService{Log: log}
}

func (s *CreatorService) CreateWailsProject(req contracts.CreateWailsProjectRequest) (string, error) {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return "", fmt.Errorf("project name is required")
	}
	req.Dir = strings.TrimSpace(strings.Trim(req.Dir, `"`))
	req.Template = strings.TrimSpace(req.Template)
	if req.Template == "" {
		req.Template = "vanilla"
	}
	req.PackageName = strings.TrimSpace(req.PackageName)
	if req.PackageName == "" {
		req.PackageName = "main"
	}
	req.GoModule = strings.TrimSpace(req.GoModule)
	req.Git = strings.TrimSpace(req.Git)
	req.ProductName = strings.TrimSpace(req.ProductName)
	req.ProductDescription = strings.TrimSpace(req.ProductDescription)
	req.ProductVersion = strings.TrimSpace(req.ProductVersion)
	req.ProductCompany = strings.TrimSpace(req.ProductCompany)
	req.ProductCopyright = strings.TrimSpace(req.ProductCopyright)
	req.ProductComments = strings.TrimSpace(req.ProductComments)
	req.ProductIdentifier = strings.TrimSpace(req.ProductIdentifier)

	targetDir, err := targetProjectDir(req)
	if err != nil {
		return "", err
	}
	if err := requireEmptyTargetDirectory(targetDir); err != nil {
		return "", err
	}
	parentDir := filepath.Dir(targetDir)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return "", err
	}

	command := buildInitCommand(req, targetDir)
	if err := s.commandRunner().Run(context.Background(), parentDir, command); err != nil {
		return "", err
	}
	if !fsx.DirExists(targetDir) {
		return "", fmt.Errorf("wails3 init completed but project directory was not created: %s", targetDir)
	}
	return targetDir, nil
}

func (s *CreatorService) commandRunner() CommandRunner {
	if s.Runner != nil {
		return s.Runner
	}
	return runlog.Runner{Log: s.Log}
}

func targetProjectDir(req contracts.CreateWailsProjectRequest) (string, error) {
	if req.Dir != "" {
		return fsx.NormalizePath(req.Dir)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return fsx.NormalizePath(filepath.Join(cwd, req.Name))
}

func requireEmptyTargetDirectory(targetDir string) error {
	entries, err := os.ReadDir(targetDir)
	if err == nil {
		if len(entries) > 0 {
			return fmt.Errorf("target directory is not empty: %s", targetDir)
		}
		return nil
	}
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func buildInitCommand(req contracts.CreateWailsProjectRequest, targetDir string) []string {
	command := []string{
		"wails3",
		"init",
		"-n", req.Name,
		"-d", targetDir,
		"-t", req.Template,
		"-p", req.PackageName,
		"-nocolour",
	}
	command = appendFlag(command, "-mod", req.GoModule)
	command = appendFlag(command, "-git", req.Git)
	command = appendFlag(command, "-productname", req.ProductName)
	command = appendFlag(command, "-productdescription", req.ProductDescription)
	command = appendFlag(command, "-productversion", req.ProductVersion)
	command = appendFlag(command, "-productcompany", req.ProductCompany)
	command = appendFlag(command, "-productcopyright", req.ProductCopyright)
	command = appendFlag(command, "-productcomments", req.ProductComments)
	command = appendFlag(command, "-productidentifier", req.ProductIdentifier)
	if req.Quiet {
		command = append(command, "-q")
	}
	if req.SkipRemoteTemplateWarning {
		command = append(command, "-s")
	}
	if req.SkipGoModTidy {
		command = append(command, "-skipgomodtidy")
	}
	return command
}

func appendFlag(command []string, flag string, value string) []string {
	if strings.TrimSpace(value) == "" {
		return command
	}
	return append(command, flag, value)
}
