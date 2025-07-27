package validators

import (
	"errors"
	"unicode"
)

// LuhnValidator validates credit card numbers using the Luhn algorithm.
type LuhnValidator struct{}

// NewLuhnValidator creates a new LuhnValidator instance.
func NewLuhnValidator() *LuhnValidator {
	return &LuhnValidator{}
}

// Validate checks if the provided card number passes the Luhn algorithm.
func (v *LuhnValidator) Validate(number string) error {
	return validateLuhn(number)
}

// validateLuhn performs the Luhn algorithm check on the card number.
func validateLuhn(number string) error {
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

// CVVValidator validates credit card CVV codes.
type CVVValidator struct{}

// NewCVVValidator creates a new CVVValidator instance.
func NewCVVValidator() *CVVValidator {
	return &CVVValidator{}
}

// Validate checks if the CVV is exactly 3 digits.
func (v *CVVValidator) Validate(cvv string) error {
	return validateCVV(cvv)
}

// validateCVV validates the CVV code format.
func validateCVV(cvv string) error {
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
