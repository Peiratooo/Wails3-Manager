package project

import (
	"wails3-manager/core/runlog"
)

type Service struct {
	Log             *runlog.Logger
	InitPackagingFn func(projectDir string) error
}

func NewService(log *runlog.Logger) *Service {
	return &Service{Log: log}
}
