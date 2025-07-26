package validators

import (
	"errors"
	"unicode"
)

type LuhnValidator struct{}

func NewLuhnValidator() *LuhnValidator {
	return &LuhnValidator{}
}

func (cn *LuhnValidator) Validate(number string) error {
	if number == "" {
		return errors.New("card number is empty")
	}

	for _, ch := range number {
		if !unicode.IsDigit(ch) {
			return errors.New("card number contains invalid characters")
		}
	}

	var sum int
	alt := false
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

type CVVValidator struct{}

func NewCVVValidator() *CVVValidator {
	return &CVVValidator{}
}

func (cc *CVVValidator) Validate(cvv string) error {
	if cvv == "" {
		return errors.New("CVV is empty")
	}

	if len(cvv) != 3 && len(cvv) != 4 {
		return errors.New("CVV must be 3 or 4 digits long")
	}

	for _, ch := range cvv {
		if !unicode.IsDigit(ch) {
			return errors.New("CVV contains non-digit characters")
		}
	}

	return nil
}
