package project

import (
	"testing"

	"wails3-manager/core/contracts"
)

func TestSortProjectsNewestFirst(t *testing.T) {
	projects := []contracts.ProjectRecord{
		{LastOpenedAt: 1},
		{LastOpenedAt: 3},
		{LastOpenedAt: 2},
	}
	SortProjects(projects)
	for i, want := range []contracts.UnixTime{3, 2, 1} {
		if projects[i].LastOpenedAt != want {
			t.Fatalf("projects[%d].LastOpenedAt = %d, want %d", i, projects[i].LastOpenedAt, want)
		}
	}
}
