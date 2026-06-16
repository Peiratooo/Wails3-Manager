package desktop

import (
	"runtime"
	"strings"
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
