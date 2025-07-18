package validation

import (
	"errors"
	"os"
	"unicode"
)

// ValidateBinarySecretName ensures secretName is not empty and valid characters.
func ValidateBinarySecretName(secretName string) error {
	if secretName == "" {
		return errors.New("secret name must not be empty")
	}
	for _, ch := range secretName {
		if !(unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' || ch == '-' || ch == ' ') {
			return errors.New("secret name can only contain letters, digits, underscore, hyphen, and spaces")
		}
	}
	return nil
}

// ValidateBinaryDataPath checks that the dataPath is a valid existing file path.
func ValidateBinaryDataPath(dataPath string) error {
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

// ValidateBinaryMeta validates the meta string for printable characters only (optional).
func ValidateBinaryMeta(meta string) error {
	for _, ch := range meta {
		if ch < 32 || ch == 127 {
			return errors.New("meta contains invalid control characters")
		}
	}
	return nil
}
