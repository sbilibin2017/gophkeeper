package validation

import (
	"errors"
	"unicode"
)

// Rename this function from ValidateTextSecretName to ValidateSecretName
func ValidateSecretName(secretName string) error {
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

// ValidateUser ensures username is not empty and contains only allowed characters.
func ValidateUsername(username string) error {
	if username == "" {
		return errors.New("username must not be empty")
	}
	for _, ch := range username {
		if !(unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_') {
			return errors.New("username can only contain letters, digits, and underscore")
		}
	}
	return nil
}

// ValidatePass ensures password is not empty.
func ValidatePassword(password string) error {
	if password == "" {
		return errors.New("password must not be empty")
	}
	return nil
}

// ValidateMeta validates the meta string for printable characters only (optional).
func ValidateMeta(meta string) error {
	if meta == "" {
		return nil
	}
	for _, ch := range meta {
		if (ch < 32 && ch != 9 && ch != 10 && ch != 13) || ch == 127 {
			return errors.New("meta contains invalid control characters")
		}
	}
	return nil
}
