package validation

import (
	"errors"
	"unicode"
)

// ValidateLoginUsername validates that the username is not empty for login.
//
// Returns an error if username is empty.
func ValidateLoginUsername(username string) error {
	if username == "" {
		return errors.New("username must not be empty")
	}
	return nil
}

// ValidateLoginPassword validates that the password is not empty for login.
//
// Returns an error if password is empty.
func ValidateLoginPassword(password string) error {
	if password == "" {
		return errors.New("password must not be empty")
	}
	return nil
}

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

// ValidateRegisterPassword ensures the password meets strength requirements.
//
// Password must be at least 8 characters long and contain at least one uppercase letter,
// one lowercase letter, and one digit.
func ValidateRegisterPassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	var hasUpper, hasLower, hasDigit bool
	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		}
	}
	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return errors.New("password must contain at least one digit")
	}
	return nil
}
