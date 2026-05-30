package desktop

import "strings"

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

func isDialogCancel(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())

	return strings.Contains(msg, "cancel") ||
		strings.Contains(msg, "cancelled") ||
		strings.Contains(msg, "canceled")
}
