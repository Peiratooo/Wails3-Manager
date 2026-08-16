package desktop

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"wails3-manager/core/execenv"
)

type AppService struct{}

func (a *AppService) ChooseFolder() (string, error) {
	path, err := App.Dialog.
		OpenFile().
		CanChooseDirectories(true).
		CanChooseFiles(false).
		PromptForSingleSelection()

	if err != nil {
		if isDialogCancel(err) {
			return "", nil
		}
		return "", err
	}

	if path == "" {
		return "", nil
	}

	return path, nil
}

func (a *AppService) OpenPath(path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return errors.New("path is required")
	}
	absolutePath, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return err
	}
	info, err := os.Stat(absolutePath)
	if err != nil {
		return err
	}

	cmd := openPathCommand(runtime.GOOS, absolutePath, info.IsDir())
	execenv.HideWindow(cmd)
	return cmd.Start()
}

func openPathCommand(goos string, path string, isDir bool) *exec.Cmd {
	switch goos {
	case "windows":
		if isDir {
			return exec.Command("explorer.exe", path)
		}
		return exec.Command("explorer.exe", "/select,"+path)
	case "darwin":
		return exec.Command("open", path)
	default:
		return exec.Command("xdg-open", path)
	}
}

func (a *AppService) ChooseIcon() (string, error) {
	path, err := App.Dialog.
		OpenFile().
		CanChooseDirectories(false).
		CanChooseFiles(true).
		AddFilter("PNG File", "*.png").
		PromptForSingleSelection()

	if err != nil {
		if isDialogCancel(err) {
			return "", nil
		}
		return "", err
	}

	if path == "" {
		return "", nil
	}

	return path, nil
}

func (a *AppService) ChooseSetupIcon() (string, error) {
	dialog := App.Dialog.
		OpenFile().
		CanChooseDirectories(false).
		CanChooseFiles(true)

	switch runtime.GOOS {
	case "windows":
		dialog = dialog.
			AddFilter("ICO File", "*.ico")

	case "darwin":
		dialog = dialog.
			AddFilter("ICNS File", "*.icns")

	default:
		dialog = dialog.
			AddFilter("ICO File", "*.ico").
			AddFilter("ICNS File", "*.icns")
	}

	path, err := dialog.PromptForSingleSelection()
	if err != nil {
		if isDialogCancel(err) {
			return "", nil
		}
		return "", err
	}

	if path == "" {
		return "", nil
	}

	return path, nil
}
func (a *AppService) ChooseFile() (string, error) {
	path, err := App.Dialog.
		OpenFile().
		CanChooseDirectories(false).
		CanChooseFiles(true).
		PromptForSingleSelection()

	if err != nil {
		if isDialogCancel(err) {
			return "", nil
		}
		return "", err
	}

	if path == "" {
		return "", nil
	}

	return path, nil
}

func isDialogCancel(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())

	return strings.Contains(msg, "cancel") ||
		strings.Contains(msg, "cancelled") ||
		strings.Contains(msg, "canceled")
}
