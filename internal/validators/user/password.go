package user

import (
	"unicode"
)

// ValidatePassword checks if the password meets the requirements:
// at least one lowercase, one uppercase, one digit, one special character, and length 8-128.
func ValidatePassword(pwd string) bool {
	var (
		hasLower   bool
		hasUpper   bool
		hasDigit   bool
		hasSpecial bool
	)

	if len(pwd) < 8 || len(pwd) > 128 {
		return false
	}

	for _, ch := range pwd {
		switch {
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsDigit(ch):
			hasDigit = true
		case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
			hasSpecial = true
		}
	}

	return hasLower && hasUpper && hasDigit && hasSpecial
}
