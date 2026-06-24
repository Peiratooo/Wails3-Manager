package config

import (
	"testing"

	"wails3-manager/core/contracts"
)

func TestAssetTargetDefaultsByPlatform(t *testing.T) {
	asset := contracts.PackagingAsset{Src: "README.md", Type: "file"}

	windows, err := AssetTarget(asset, contracts.PlatformWindows)
	if err != nil {
		t.Fatal(err)
	}
	if windows != "/" {
		t.Fatalf("Windows target = %q, want /", windows)
	}

	macos, err := AssetTarget(asset, contracts.PlatformMacOS)
	if err != nil {
		t.Fatal(err)
	}
	if macos != "MacOS" {
		t.Fatalf("macOS target = %q, want MacOS", macos)
	}
}

func TestAssetTargetRejectsInvalidPlatformTargets(t *testing.T) {
	tests := []struct {
		name     string
		platform contracts.Platform
		target   string
	}{
		{name: "windows relative", platform: contracts.PlatformWindows, target: "config"},
		{name: "windows backslash", platform: contracts.PlatformWindows, target: `/data\logs`},
		{name: "windows traversal", platform: contracts.PlatformWindows, target: `/data/../logs`},
		{name: "mac absolute", platform: contracts.PlatformMacOS, target: "/Resources"},
		{name: "mac backslash", platform: contracts.PlatformMacOS, target: `Resources\Images`},
		{name: "mac traversal", platform: contracts.PlatformMacOS, target: "Resources/../MacOS"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := AssetTarget(contracts.PackagingAsset{Target: tt.target}, tt.platform)
			if err == nil {
				t.Fatalf("AssetTarget(%q, %s) expected error", tt.target, tt.platform)
			}
		})
	}
}
