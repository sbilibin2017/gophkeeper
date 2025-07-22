package user

import "regexp"

// ValidateUsername returns true if the username meets the criteria,
// otherwise false.
func ValidateUsername(username string) bool {
	if len(username) < 3 || len(username) > 30 {
		return false
	}

	// Only allow letters, digits, underscores
	var validUsername = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

	return validUsername.MatchString(username)
}
