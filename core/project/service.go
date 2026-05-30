package project

import (
	"wails3-manager/core/runlog"
)

type ProjectService struct {
	Log *runlog.Logger
}

func NewService(log *runlog.Logger) *ProjectService {
	return &ProjectService{Log: log}
}
