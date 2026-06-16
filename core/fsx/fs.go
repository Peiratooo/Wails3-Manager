// Package fsx keeps low-level filesystem helpers in one place.
// Business packages use these helpers for path normalization and copy behavior
// so safety checks stay consistent across project/settings/packaging.
package fsx

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func FileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func DirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func NormalizePath(path string) (string, error) {
	path = strings.TrimSpace(strings.Trim(path, `"`))
	if path == "" {
		return "", fmt.Errorf("path is required")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.Clean(abs), nil
}

func CopyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func CopyDir(src, dst string, recursive bool) error {
	if !recursive {
		return fmt.Errorf("directory copy requires recursive=true: %s", src)
	}
	return filepath.WalkDir(src, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		to := filepath.Join(dst, rel)
		info, err := d.Info()
		if err != nil {
			return err
		}
		if d.IsDir() {
			return os.MkdirAll(to, info.Mode())
		}
		return CopyFile(path, to, info.Mode())
	})
}

func WriteIfMissing(path string, data []byte, mode os.FileMode) (bool, error) {
	if FileExists(path) {
		return false, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return false, err
	}
	return true, os.WriteFile(path, data, mode)
}

func SafeName(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "app"
	}
	var b strings.Builder
	lastDash := false
	for _, r := range s {
		invalid := r < 32 || strings.ContainsRune(`<>:"/\\|?*`, r)
		space := r == ' ' || r == '\t' || r == '\n' || r == '\r'
		if invalid || space {
			if !lastDash {
				b.WriteRune('-')
				lastDash = true
			}
			continue
		}
		b.WriteRune(r)
		lastDash = false
	}
	out := strings.Trim(b.String(), "-_. ")
	if out == "" {
		return "app"
	}
	return out
}

func StripVersionPrefix(v string) string {
	v = strings.TrimSpace(v)
	return strings.TrimPrefix(strings.TrimPrefix(v, "v"), "V")
}

func Resolve(projectDir, path string) string {
	path = strings.TrimSpace(strings.Trim(path, `"`))
	if path == "" {
		return ""
	}
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Clean(filepath.Join(projectDir, path))
}

func UserConfigDir(appName string) string {
	if dir, err := os.UserConfigDir(); err == nil && dir != "" {
		return filepath.Join(dir, appName)
	}
	if runtime.GOOS == "windows" {
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, appName)
		}
	}
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, "."+appName)
	}
	return "." + appName
}

func Timestamp() string { return time.Now().Format("20060102-150405") }

func ResolveProjectFile(projectDir, relPath string) string {
	projectDir = strings.TrimSpace(strings.Trim(projectDir, `"`))
	relPath = strings.TrimSpace(strings.Trim(relPath, `"`))

	if projectDir == "" || relPath == "" {
		return ""
	}

	projectAbs, err := NormalizePath(projectDir)
	if err != nil {
		return ""
	}

	if !filepath.IsAbs(projectAbs) {
		return ""
	}

	// Project file paths must be relative to the project root.
	if filepath.IsAbs(relPath) {
		return ""
	}

	cleanRel := filepath.Clean(filepath.FromSlash(relPath))

	fullPath := filepath.Join(projectAbs, cleanRel)

	fullAbs, err := filepath.Abs(fullPath)
	if err != nil {
		return ""
	}

	fullAbs = filepath.Clean(fullAbs)

	// Prevent ../ from escaping the project root.
	relToProject, err := filepath.Rel(projectAbs, fullAbs)
	if err != nil {
		return ""
	}

	if relToProject == ".." ||
		strings.HasPrefix(relToProject, ".."+string(filepath.Separator)) ||
		filepath.IsAbs(relToProject) {
		return ""
	}

	return fullAbs
}
