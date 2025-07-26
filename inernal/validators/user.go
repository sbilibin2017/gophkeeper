package validators

import (
	"errors"
	"unicode"
)

type UsernameValidator struct{}

func NewUsernameValidator() *UsernameValidator {
	return &UsernameValidator{}
}

func (u *UsernameValidator) Validate(username string) error {
	if len(username) < 3 {
		return errors.New("username must be at least 3 characters long")
	}

	for _, ch := range username {
		switch {
		case unicode.IsLetter(ch) && ch <= unicode.MaxASCII:
		case unicode.IsDigit(ch):
		case isAllowedSpecial(ch):
		default:
			return errors.New("username contains invalid characters")
		}
	}

	return nil
}

func isAllowedSpecial(ch rune) bool {
	specials := "!@#$%^&*()_+-={}[]:\";'<>?,./~|\\"
	for _, s := range specials {
		if ch == s {
			return true
		}
	}
	return false
}

type PasswordValidator struct{}

func NewPasswordValidator() *PasswordValidator {
	return &PasswordValidator{}
}

func (p *PasswordValidator) Validate(password string) error {
	if len(password) < 6 {
		return errors.New("password must be at least 6 characters long")
	}

	var hasUpper, hasDigit, hasSpecial bool

	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsDigit(ch):
			hasDigit = true
		case isSpecial(ch):
			hasSpecial = true
		case unicode.IsLower(ch):
			// lowercase allowed
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

func isSpecial(ch rune) bool {
	specials := "!@#$%^&*()_+-={}[]:\";'<>?,./~|\\"
	for _, s := range specials {
		if ch == s {
			return true
		}
	}
	return false
}
