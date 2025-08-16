package validators

import (
	"errors"
	"unicode"
)

// ValidateUsername checks username length and allowed characters.
// It ensures the username is at least 3 characters long and contains
// only ASCII letters, digits, or allowed special characters.
func ValidateUsername(username string) error {
	if len(username) < 3 {
		return errors.New("username must be at least 3 characters long")
	}

	const specials = "!@#$%^&*()_+-={}[]:\";'<>?,./~|\\"

	for _, ch := range username {
		switch {
		case unicode.IsLetter(ch) && ch <= unicode.MaxASCII:
		case unicode.IsDigit(ch):
		case func(r rune) bool {
			for _, c := range specials {
				if c == r {
					return true
				}
			}
			return false
		}(ch):
		default:
			return errors.New("username contains invalid characters")
		}
	}

	return nil
}

// ValidatePassword checks password complexity requirements.
// Password must be at least 6 characters long, contain at least one uppercase
// letter, one digit, and one special character from the allowed set.
func ValidatePassword(password string) error {
	if len(password) < 6 {
		return errors.New("password must be at least 6 characters long")
	}

	const specials = "!@#$%^&*()_+-={}[]:\";'<>?,./~|\\"

	var hasUpper, hasDigit, hasSpecial bool

	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsDigit(ch):
			hasDigit = true
		case func(r rune) bool {
			for _, c := range specials {
				if c == r {
					return true
				}
			}
			return false
		}(ch):
			hasSpecial = true
		case unicode.IsLower(ch):
			// lowercase allowed, no flag needed
		default:
			return errors.New("password contains invalid characters")
		}
	}

	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !hasDigit {
		return errors.New("password must contain at least one digit")
	}
	if !hasSpecial {
		return errors.New("password must contain at least one special character")
	}

	return nil
}
