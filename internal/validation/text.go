package validation

import (
	"errors"
	"unicode"
)

// ValidateTextSecretName ensures secretName is not empty and contains valid characters.
func ValidateTextSecretName(secretName string) error {
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

// ValidateTextContent ensures content is not empty.
func ValidateTextContent(content string) error {
	if content == "" {
		return errors.New("content must not be empty")
	}
	return nil
}

func ValidateTextMeta(meta string) error {
	for _, ch := range meta {
		if (ch < 32 && ch != 9 && ch != 10 && ch != 13) || ch == 127 {
			return errors.New("meta contains invalid control characters")
		}
	}
	return nil
}
