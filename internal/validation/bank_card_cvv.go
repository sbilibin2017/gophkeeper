package validation

import (
	"errors"
	"unicode"
)

// ValidateBankCardCVV validates the CVV is 3 or 4 digits.
func ValidateBankCardCVV(cvv string) error {
	if len(cvv) != 3 && len(cvv) != 4 {
		return errors.New("cvv must be 3 or 4 digits")
	}
	for _, ch := range cvv {
		if !unicode.IsDigit(ch) {
			return errors.New("cvv must contain only digits")
		}
	}
	return nil
}
