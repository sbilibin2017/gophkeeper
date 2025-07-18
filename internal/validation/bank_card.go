package validation

import (
	"errors"
	"regexp"
	"strconv"
	"time"
	"unicode"
)

// ValidateBankCardSecretName ensures secretName is not empty and has valid characters.
func ValidateBankCardSecretName(secretName string) error {
	if secretName == "" {
		return errors.New("secret name must not be empty")
	}
	for _, ch := range secretName {
		if !(unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' || ch == '-' || ch == ' ') {
			return errors.New("secret name can only contain letters, digits, underscore, hyphen, and spaces")
		}
	}
	return nil
}

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

// ValidateBankCardExp validates expiration date in MM/YY or MM/YYYY format and ensures it's a future date.
func ValidateBankCardExp(exp string) error {
	if exp == "" {
		return errors.New("expiration date must not be empty")
	}

	var month, year int
	var err error

	reShort := regexp.MustCompile(`^(0[1-9]|1[0-2])/(\d{2})$`)
	reLong := regexp.MustCompile(`^(0[1-9]|1[0-2])/(\d{4})$`)

	switch {
	case reShort.MatchString(exp):
		parts := reShort.FindStringSubmatch(exp)
		month, err = strconv.Atoi(parts[1])
		if err != nil {
			return errors.New("invalid month in expiration date")
		}
		year, err = strconv.Atoi(parts[2])
		if err != nil {
			return errors.New("invalid year in expiration date")
		}
		year += 2000
	case reLong.MatchString(exp):
		parts := reLong.FindStringSubmatch(exp)
		month, err = strconv.Atoi(parts[1])
		if err != nil {
			return errors.New("invalid month in expiration date")
		}
		year, err = strconv.Atoi(parts[2])
		if err != nil {
			return errors.New("invalid year in expiration date")
		}
	default:
		return errors.New("expiration date must be in MM/YY or MM/YYYY format")
	}

	now := time.Now()
	expDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	expDate = expDate.AddDate(0, 1, -1)

	if expDate.Before(now) {
		return errors.New("expiration date must be in the future")
	}

	return nil
}

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

// ValidateBankCardMeta validates the meta string, optional - allows empty or printable characters only.
func ValidateBankCardMeta(meta string) error {
	if meta == "" {
		return nil
	}
	for _, ch := range meta {
		if ch < 32 || ch == 127 {
			return errors.New("meta contains invalid control characters")
		}
	}
	return nil
}
