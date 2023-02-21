package util

import (
	"errors"
	"os"
)

func FileExists(fn string) bool {
	if _, err := os.Stat(fn); err == nil {
		return true

	} else if errors.Is(err, os.ErrNotExist) {
		return false

	} else {
		// Schrodinger: file may or may not exist. See err for details.
		// Therefore, do *NOT* use !os.IsNotExist(err) to test for file existence
		return false
	}
}

func FolderExists(fn string) bool {
	if fInfo, err := os.Stat(fn); err == nil {
		return fInfo.IsDir()
	} else if errors.Is(err, os.ErrNotExist) {
		return false

	} else {
		// Schrodinger: file may or may not exist. See err for details.
		// Therefore, do *NOT* use !os.IsNotExist(err) to test for file existence
		return false
	}
}
