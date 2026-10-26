package packaging

import (
	"os"
	"path/filepath"
	"testing"

	"howett.net/plist"

	"wails3-manager/core/contracts"
)

const catalogIconPlist = `<?xml version="1.0"?><plist version="1.0"><dict>
<key>CFBundleIconName</key><string>appicon</string>
<key>CFBundleIconFile</key><string>custom</string>
<key>NSCameraUsageDescription</key><string>Preserve privacy metadata</string>
</dict></plist>`

func TestRenderedMacOSBundleUsesICNSIcon(t *testing.T) {
	projectDir := t.TempDir()
	writeIconTestPlist(t, filepath.Join(projectDir, "build", "darwin", "Info.plist"))
	text, err := renderMacOSInfoPlist(projectDir, contracts.PackagingConfig{
		Build: contracts.BuildSettings{AppName: "demo"},
	}, contracts.WailsProjectConfig{})
	if err != nil {
		t.Fatal(err)
	}
	assertIconTestPlist(t, []byte(text), true)
}

func TestStagedMacOSIconSelection(t *testing.T) {
	for _, useProjectIcon := range []bool{true, false} {
		name := "preserve existing bundle icon"
		if useProjectIcon {
			name = "use generated project icon"
		}
		t.Run(name, func(t *testing.T) {
			bundle := t.TempDir()
			path := filepath.Join(bundle, "Contents", "Info.plist")
			writeIconTestPlist(t, path)
			if err := updateStagedBundlePlist(bundle, contracts.PackagingConfig{}, contracts.WailsProjectConfig{}, false, useProjectIcon); err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			assertIconTestPlist(t, data, useProjectIcon)
		})
	}
}

func writeIconTestPlist(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(catalogIconPlist), 0644); err != nil {
		t.Fatal(err)
	}
}

func assertIconTestPlist(t *testing.T, data []byte, useProjectIcon bool) {
	t.Helper()
	var info map[string]any
	if _, err := plist.Unmarshal(data, &info); err != nil {
		t.Fatal(err)
	}
	if info["NSCameraUsageDescription"] != "Preserve privacy metadata" {
		t.Fatal("unrelated bundle metadata changed")
	}
	if useProjectIcon {
		if _, exists := info["CFBundleIconName"]; exists || info["CFBundleIconFile"] != "icons" {
			t.Fatalf("generated ICNS icon is overridden: %#v", info)
		}
	} else if info["CFBundleIconName"] != "appicon" || info["CFBundleIconFile"] != "custom" {
		t.Fatalf("existing bundle icon changed: %#v", info)
	}
}
