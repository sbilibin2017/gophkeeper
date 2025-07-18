package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateBankCardOwner(t *testing.T) {
	tests := []struct {
		name    string
		owner   string
		wantErr bool
		errMsg  string
	}{
		{"Valid name", "John Doe", false, ""},
		{"Valid with hyphen", "Anne-Marie Smith", false, ""},
		{"Valid single letter", "A", false, ""},
		{"Empty string", "", true, "owner name must not be empty"},
		{"Contains digit", "John Doe2", true, "owner name can only contain letters, spaces, and hyphens"},
		{"Contains special char", "John_Doe", true, "owner name can only contain letters, spaces, and hyphens"},
		{"Contains punctuation", "John, Doe", true, "owner name can only contain letters, spaces, and hyphens"},
		{"Contains newline", "John\nDoe", true, "owner name can only contain letters, spaces, and hyphens"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBankCardOwner(tt.owner)
			if tt.wantErr {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
