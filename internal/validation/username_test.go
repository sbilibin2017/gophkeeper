package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateRegisterUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  bool
		errMsg   string
	}{
		{"Valid username min length", "abc", false, ""},
		{"Valid username max length", "a_very_long_username_123456", false, ""},
		{"Too short username", "ab", true, "username must be between 3 and 30 characters"},
		{"Too long username", "a_very_long_username_12345678901234567890", true, "username must be between 3 and 30 characters"},
		{"Invalid characters", "user!name", true, "username can only contain letters, digits, and underscore"},
		{"Valid with underscore", "user_name_123", false, ""},
		{"Empty username", "", true, "username must be between 3 and 30 characters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRegisterUsername(tt.username)
			if tt.wantErr {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
