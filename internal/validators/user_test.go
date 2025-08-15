package validators

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{"too short", "ab", true, "username must be at least 3 characters long"},
		{"valid letters", "abc", false, ""},
		{"valid letters and digits", "user123", false, ""},
		{"valid with specials", "user_!@#", false, ""},
		{"invalid chars", "user😊", true, "username contains invalid characters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUsername(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{"too short", "Ab1!", true, "password must be at least 6 characters long"},
		{"missing uppercase", "abc123!", true, "password must contain at least one uppercase letter"},
		{"missing digit", "Abcdef!", true, "password must contain at least one digit"},
		{"missing special", "Abc1234", true, "password must contain at least one special character"},
		{"valid password", "Abc123!", false, ""},
		{"invalid char", "Abc123😊", true, "password contains invalid characters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
