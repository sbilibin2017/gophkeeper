package validation

import (
	"errors"
	"regexp"
	"strconv"
	"time"
)

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
