package environment

import (
	"testing"

	"wails3-manager/core/contracts"
)

func TestNormalizeCommandVersion(t *testing.T) {
	tests := []struct {
		name    string
		command string
		output  string
		want    string
	}{
		{
			name:    "go version output",
			command: "go",
			output:  "go version go1.25.6 windows/amd64\n",
			want:    "go1.25.6",
		},
		{
			name:    "git version output",
			command: "git",
			output:  "git version 2.39.2.windows.1\n",
			want:    "2.39.2.windows.1",
		},
		{
			name:    "wails3 plain version output",
			command: "wails3",
			output:  "v3.0.0-alpha.93\n",
			want:    "v3.0.0-alpha.93",
		},
		{
			name:    "wails3 labeled version output",
			command: "wails3",
			output:  "Wails CLI v3.0.0-alpha.93\n",
			want:    "v3.0.0-alpha.93",
		},
		{
			name:    "wails3 hyphenated prerelease output",
			command: "wails3",
			output:  "Wails CLI v3.0.0-alpha-beta.93\n",
			want:    "v3.0.0-alpha-beta.93",
		},
		{
			name:    "strips ansi escapes",
			command: "wails3",
			output:  "\x1b[32mv3.0.0-alpha.93\x1b[0m\n",
			want:    "v3.0.0-alpha.93",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeCommandVersion(tt.command, tt.output); got != tt.want {
				t.Fatalf("normalizeCommandVersion(%q, %q) = %q, want %q", tt.command, tt.output, got, tt.want)
			}
		})
	}
}

func TestVersionAttempts(t *testing.T) {
	tests := []struct {
		command string
		want    [][]string
	}{
		{command: "go", want: [][]string{{"version"}}},
		{command: "wails3", want: [][]string{{"version"}, {"--version"}, {"-v"}}},
		{command: "git", want: [][]string{{"--version"}}},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			attempts := versionAttempts(tt.command)
			if len(attempts) != len(tt.want) {
				t.Fatalf("len(versionAttempts(%q)) = %d, want %d", tt.command, len(attempts), len(tt.want))
			}
			for i, attempt := range attempts {
				if len(attempt.args) != len(tt.want[i]) {
					t.Fatalf("attempt %d arg count = %d, want %d", i, len(attempt.args), len(tt.want[i]))
				}
				for j, arg := range attempt.args {
					if arg != tt.want[i][j] {
						t.Fatalf("attempt %d arg %d = %q, want %q", i, j, arg, tt.want[i][j])
					}
				}
			}
		})
	}
}

func TestOnlyWindowsInstallerNeedsAnExternalTool(t *testing.T) {
	service := &Service{
		Inno: func(string) contracts.ToolRequirement {
			return contracts.ToolRequirement{ID: "inno", Command: "ISCC.exe"}
		},
	}

	windowsChecks := service.platformEnvironmentChecks("windows")
	if len(windowsChecks) != 1 {
		t.Fatalf("Windows checks = %d, want 1", len(windowsChecks))
	}
	if windowsChecks[0].Required {
		t.Fatal("Windows installer check should be optional")
	}
	if darwinChecks := service.platformEnvironmentChecks("darwin"); len(darwinChecks) != 0 {
		t.Fatalf("Darwin checks = %#v, want none", darwinChecks)
	}
}
