package creator

import (
	"context"

	"wails3-manager/core/runlog"
)

type CommandRunner interface {
	Run(ctx context.Context, workDir string, command []string) error
}

type Service struct {
	Log    *runlog.Logger
	Runner CommandRunner
}

func NewService(log *runlog.Logger) *Service {
	return &Service{Log: log}
}

func (s *Service) CreateWailsProject(req CreateWailsProjectRequest) (string, error) {
	return CreateWailsProject(context.Background(), req, s.commandRunner())
}

func (s *Service) ListTemplates() ([]WailsTemplate, error) {
	return ListWailsTemplates()
}

func (s *Service) commandRunner() CommandRunner {
	if s.Runner != nil {
		return s.Runner
	}
	return runlog.Runner{Log: s.Log}
}
