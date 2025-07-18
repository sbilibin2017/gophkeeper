package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateBankCardNumber(t *testing.T) {
	tests := []struct {
		name    string
		number  string
		wantErr bool
		errMsg  string
	}{
		{"Valid 12 digits", "123456789012", false, ""},
		{"Valid 19 digits", "1234567890123456789", false, ""},
		{"Too short", "12345678901", true, "card number must be between 12 and 19 digits"},
		{"Too long", "12345678901234567890", true, "card number must be between 12 and 19 digits"},
		{"Contains letters", "12345abc6789", true, "card number must contain only digits"},
		{"Contains spaces", "1234 56789012", true, "card number must contain only digits"},
		{"Empty string", "", true, "card number must be between 12 and 19 digits"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBankCardNumber(tt.number)
			if tt.wantErr {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
