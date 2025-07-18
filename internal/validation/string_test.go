package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateString(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
		content  string
		wantErr  bool
		errMsg   string
	}{
		{"Non-empty string", "username", "hello", false, ""},
		{"Empty string", "password", "", true, "password must not be empty"},
		{"Whitespace string", "token", "   \t\n", true, "token must not be empty"},
		{"Spaces and text", "name", "  test ", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateString(tt.flagName, tt.content)
			if tt.wantErr {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
