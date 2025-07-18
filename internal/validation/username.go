package validation

import (
	"errors"
	"unicode"
)

// ValidateRegisterUsername ensures the username is valid.
//
// Username must be between 3 and 30 characters and contain only letters, digits, and underscores.
func ValidateRegisterUsername(username string) error {
	if len(username) < 3 || len(username) > 30 {
		return errors.New("username must be between 3 and 30 characters")
	}
	for _, ch := range username {
		if !(unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_') {
			return errors.New("username can only contain letters, digits, and underscore")
		}
	}
	return nil
}
