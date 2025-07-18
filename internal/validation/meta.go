package validation

import "errors"

// ValidateMeta validates the meta string for printable characters only (optional).
func ValidateMeta(meta string) error {
	if meta == "" {
		return nil
	}
	for _, ch := range meta {
		if (ch < 32 && ch != 9 && ch != 10 && ch != 13) || ch == 127 {
			return errors.New("meta contains invalid control characters")
		}
	}
	return nil
}
