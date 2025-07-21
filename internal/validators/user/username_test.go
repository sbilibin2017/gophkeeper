package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		want     bool
	}{
		{"valid_username", "user123", true},
		{"valid_with_underscore", "user_name_1", true},
		{"too_short", "ab", false},
		{"too_long", "a_very_very_long_username_exceeding_thirty_chars", false},
		{"invalid_chars", "user$name", false},
		{"empty_string", "", false},
		{"spaces_not_allowed", "user name", false},
		{"only_letters", "username", true},
		{"only_numbers", "123456", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateUsername(tt.username)
			assert.Equal(t, tt.want, got)
		})
	}
}
