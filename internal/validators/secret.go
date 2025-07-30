package validators

import (
	"errors"
	"unicode"
)

// ValidateLuhn checks if the provided card number passes the Luhn algorithm.
func ValidateLuhn(number string) error {
	if number == "" {
		return errors.New("card number is empty")
	}

	for _, ch := range number {
		if !unicode.IsDigit(ch) {
			return errors.New("card number contains invalid characters")
		}
	}

	sum := 0
	alt := false

	// Iterate over the number from right to left
	for i := len(number) - 1; i >= 0; i-- {
		digit := int(number[i] - '0')
		if alt {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		alt = !alt
	}

	if sum%10 != 0 {
		return errors.New("invalid card number (failed Luhn check)")
	}

	return nil
}

// ValidateCVV checks if the CVV is exactly 3 digits.
func ValidateCVV(cvv string) error {
	if cvv == "" {
		return errors.New("CVV is empty")
	}

	if len(cvv) != 3 {
		return errors.New("CVV must be exactly 3 digits long")
	}

	for _, ch := range cvv {
		if !unicode.IsDigit(ch) {
			return errors.New("CVV contains non-digit characters")
		}
	}

	return nil
}
