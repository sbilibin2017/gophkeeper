package validation

import (
	"errors"
	"strings"
)

// ValidateString ensures the string is not empty or just whitespace.
// flagName is used to provide a descriptive error message.
func ValidateString(flagName, content string) error {
	if strings.TrimSpace(content) == "" {
		return errors.New(flagName + " must not be empty")
	}
	return nil
}
