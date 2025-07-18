package validation

import (
	"errors"
	"unicode"
)

// ValidateBankCardNumber validates bank card number using basic length and digit checks.
func ValidateBankCardNumber(number string) error {
	if len(number) < 12 || len(number) > 19 {
		return errors.New("card number must be between 12 and 19 digits")
	}
	for _, ch := range number {
		if !unicode.IsDigit(ch) {
			return errors.New("card number must contain only digits")
		}
	}
	return nil
}
