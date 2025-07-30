package validators

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{"ValidUsernameLetters", "johnDoe", false},
		{"ValidUsernameWithSpecials", "john_doe!", false},
		{"TooShort", "ab", true},
		{"InvalidCharNonASCII", "юзер", true},
		{"InvalidCharEmoji", "john🙂", true},
		{"InvalidCharSpace", "john doe", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUsername(tt.username)
			if tt.wantErr {
				assert.Error(t, err, "expected an error")
			} else {
				assert.NoError(t, err, "expected no error")
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"ValidPassword", "Passw0rd!", false},
		{"MissingUppercase", "password1!", true},
		{"MissingDigit", "Password!", true},
		{"MissingSpecial", "Password1", true},
		{"TooShort", "P1!", true},
		{"ValidWithManySpecials", "P@ssw0rd!#%", false},
		{"InvalidCharEmoji", "Passw0rd🙂", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if tt.wantErr {
				assert.Error(t, err, "expected an error")
			} else {
				assert.NoError(t, err, "expected no error")
			}
		})
	}
}
