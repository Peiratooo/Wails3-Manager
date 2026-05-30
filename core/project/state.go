package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

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
	fmt.Println(path)
	return os.WriteFile(path, append(data, '\n'), 0644)
}

func SortProjects(projects []contracts.ProjectRecord) {
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].LastOpenedAt.After(projects[j].LastOpenedAt)
	})
}

func UpsertProjectRecord(record contracts.ProjectRecord) error {
	projectDir, err := NormalizeProjectDir(record)
	if err != nil {
		return err
	}
	record = NormalizeProjectRecord(record, projectDir)
	state := LoadUserState()
	found := false
	for i := range state.Projects {
		if SameProjectPath(state.Projects[i].ProjectDir, projectDir) {
			state.Projects[i] = record
			found = true
			break
		}
	}
	if !found {
		state.Projects = append(state.Projects, record)
	}
	return SaveUserState(state)
}

func RemoveProjectRecord(projectDir string) error {
	state := LoadUserState()
	next := state.Projects[:0]
	for _, record := range state.Projects {
		if SameProjectPath(record.ProjectDir, projectDir) {
			continue
		}
		next = append(next, record)
	}
	state.Projects = next
	return SaveUserState(state)
}
