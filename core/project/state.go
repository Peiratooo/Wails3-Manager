package project

import (
	"cmp"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"

	"wails3-manager/core/contracts"
	"wails3-manager/core/fsx"
)

// state.json is the manager's imported-project index. Project import/save writes
// it directly so the project service only needs a logger dependency.
func LoadUserState() contracts.UserState {
	data, err := os.ReadFile(fsx.StatePath())
	if err != nil {
		return contracts.UserState{}
	}
	var state contracts.UserState
	if json.Unmarshal(data, &state) != nil {
		return contracts.UserState{}
	}
	SortProjects(state.Projects)
	return state
}

func SaveUserState(state contracts.UserState) error {
	SortProjects(state.Projects)
	path := fsx.StatePath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0644)
}

func SortProjects(projects []contracts.ProjectRecord) {
	slices.SortFunc(projects, func(a, b contracts.ProjectRecord) int {
		return cmp.Compare(b.LastOpenedAt, a.LastOpenedAt)
	})
}

func UpsertProjectRecord(record contracts.ProjectRecord) error {
	if record.ProjectDir == "" {
		return errors.New("project path is required")
	}
	record.Project.ProjectDir = record.ProjectDir
	state := LoadUserState()
	index := slices.IndexFunc(state.Projects, func(existing contracts.ProjectRecord) bool {
		return SameProjectPath(existing.ProjectDir, record.ProjectDir)
	})
	if index >= 0 {
		state.Projects[index] = record
	} else {
		state.Projects = append(state.Projects, record)
	}
	return SaveUserState(state)
}

func RemoveProjectRecord(projectDir string) error {
	state := LoadUserState()
	state.Projects = slices.DeleteFunc(state.Projects, func(record contracts.ProjectRecord) bool {
		return SameProjectPath(record.ProjectDir, projectDir)
	})
	return SaveUserState(state)
}
