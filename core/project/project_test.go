package project

import (
	"strings"
	"testing"

	"wails3-manager/core/contracts"
)

func TestSaveProjectRejectsMissingRequiredInfo(t *testing.T) {
	record := contracts.ProjectRecord{
		ProjectDir: t.TempDir(),
		Project: contracts.WailsProjectManager{
			WailsConfig: contracts.WailsProjectConfig{
				Info: contracts.WailsAppInfo{
					ProductName:       "Demo",
					CompanyName:       "Acme",
					ProductIdentifier: "com.example.demo",
					Description:       "Demo app",
					Copyright:         "Copyright (c) 2026",
				},
			},
		},
	}

	_, err := (&Service{}).SaveProject(record)
	if err == nil || !strings.Contains(err.Error(), "version") {
		t.Fatalf("SaveProject error = %v, want missing version", err)
	}
}
