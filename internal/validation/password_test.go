package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateRegisterPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
		errMsg   string
	}{
		{"Valid password", "Abcdef12", false, ""},
		{"Too short", "Abc12", true, "password must be at least 8 characters long"},
		{"No uppercase", "abcdef12", true, "password must contain at least one uppercase letter"},
		{"No lowercase", "ABCDEF12", true, "password must contain at least one lowercase letter"},
		{"No digit", "Abcdefgh", true, "password must contain at least one digit"},
		{"Empty password", "", true, "password must be at least 8 characters long"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRegisterPassword(tt.password)
			if tt.wantErr {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
