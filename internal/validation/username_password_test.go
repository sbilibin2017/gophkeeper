package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateLoginUsername(t *testing.T) {
	tests := []struct {
		username string
		wantErr  bool
	}{
		{"user", false},
		{"", true},
	}

	for _, tt := range tests {
		err := ValidateLoginUsername(tt.username)
		if tt.wantErr {
			assert.Error(t, err, "username: %q", tt.username)
		} else {
			assert.NoError(t, err, "username: %q", tt.username)
		}
	}
}

func TestValidateLoginPassword(t *testing.T) {
	tests := []struct {
		password string
		wantErr  bool
	}{
		{"pass", false},
		{"", true},
	}

	for _, tt := range tests {
		err := ValidateLoginPassword(tt.password)
		if tt.wantErr {
			assert.Error(t, err, "password: %q", tt.password)
		} else {
			assert.NoError(t, err, "password: %q", tt.password)
		}
	}
}

func TestRegisterValidateRegisterUsername(t *testing.T) {
	tests := []struct {
		username string
		wantErr  bool
	}{
		{"abc", false},
		{"a1_b2", false},
		{"ab", true},
		{"a_very_long_username_more_than_30_chars", true},
		{"invalid-char!", true},
	}

	for _, tt := range tests {
		err := ValidateRegisterUsername(tt.username)
		if tt.wantErr {
			assert.Error(t, err, "username: %s", tt.username)
		} else {
			assert.NoError(t, err, "username: %s", tt.username)
		}
	}
}

func TestRegisterValidateRegisterPassword(t *testing.T) {
	tests := []struct {
		password string
		wantErr  bool
	}{
		{"Strong1A", false},
		{"weak", true},
		{"nouppercase1", true},
		{"NOLOWERCASE1", true},
		{"NoDigits", true},
	}

	for _, tt := range tests {
		err := ValidateRegisterPassword(tt.password)
		if tt.wantErr {
			assert.Error(t, err, "password: %s", tt.password)
		} else {
			assert.NoError(t, err, "password: %s", tt.password)
		}
	}
}
