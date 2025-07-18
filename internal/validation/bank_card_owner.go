package validation

import (
	"errors"
	"unicode"
)

// ValidateBankCardOwner validates the card owner name - non-empty and letters + spaces allowed.
func ValidateBankCardOwner(owner string) error {
	if owner == "" {
		return errors.New("owner name must not be empty")
	}
	for _, ch := range owner {
		if !(unicode.IsLetter(ch) || ch == ' ' || ch == '-') {
			return errors.New("owner name can only contain letters, spaces, and hyphens")
		}
	}
	return nil
}
