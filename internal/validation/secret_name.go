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
