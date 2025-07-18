package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateMeta(t *testing.T) {
	tests := []struct {
		name    string
		meta    string
		wantErr bool
		errMsg  string
	}{
		{"Empty string", "", false, ""},
		{"Valid printable ASCII", "Hello, world!", false, ""},
		{"Valid with tabs and newlines", "Line1\nLine2\tEnd", false, ""},
		{"Invalid with control char 0x01", "Hello\x01World", true, "meta contains invalid control characters"},
		{"Invalid with DEL char", "Invalid\x7FChar", true, "meta contains invalid control characters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMeta(tt.meta)
			if tt.wantErr {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
