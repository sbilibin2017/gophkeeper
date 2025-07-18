package validation

import (
	"errors"
	"os"
)

// ValidateDataPath checks that the dataPath is a valid existing file path.
func ValidateDataPath(dataPath string) error {
	if dataPath == "" {
		return errors.New("data path must not be empty")
	}
	info, err := os.Stat(dataPath)
	if err != nil {
		return errors.New("data path does not exist or cannot be accessed")
	}
	if info.IsDir() {
		return errors.New("data path must be a file, not a directory")
	}
	return nil
}
