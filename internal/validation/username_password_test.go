package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateSecretName(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
		errMsg  string
	}{
		{"Valid_Name-123", false, ""},
		{"", true, "secret name must not be empty"},
		{"Invalid!", true, "secret name can only contain letters, digits, underscore, hyphen, and spaces"},
		{"Name With Spaces", false, ""},
	}

	for _, tt := range tests {
		err := ValidateSecretName(tt.input)
		if tt.wantErr {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		} else {
			assert.NoError(t, err)
		}
	}
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
		errMsg  string
	}{
		{"valid_user123", false, ""},
		{"", true, "username must not be empty"},
		{"invalid-user!", true, "username can only contain letters, digits, and underscore"},
	}

	for _, tt := range tests {
		err := ValidateUsername(tt.input)
		if tt.wantErr {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		} else {
			assert.NoError(t, err)
		}
	}
}

func TestValidatePassword(t *testing.T) {
	assert.NoError(t, ValidatePassword("anything"))
	assert.ErrorContains(t, ValidatePassword(""), "password must not be empty")
}

func TestValidateMeta(t *testing.T) {
	validMeta := "This is some valid meta info.\nWith newlines and tabs\tallowed."
	invalidMeta := string([]byte{0x01, 0x02, 0x03}) // control chars

	assert.NoError(t, ValidateMeta(""))        // empty meta allowed
	assert.NoError(t, ValidateMeta(validMeta)) // valid meta passes
	err := ValidateMeta(invalidMeta)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "meta contains invalid control characters")
}
