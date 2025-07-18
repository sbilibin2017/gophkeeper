package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateBankCardCVV(t *testing.T) {
	tests := []struct {
		name    string
		cvv     string
		wantErr bool
		errMsg  string
	}{
		{"Valid CVV 3 digits", "123", false, ""},
		{"Valid CVV 4 digits", "1234", false, ""},
		{"Too short", "12", true, "cvv must be 3 or 4 digits"},
		{"Too long", "12345", true, "cvv must be 3 or 4 digits"},
		{"Non-digit chars", "12a", true, "cvv must contain only digits"},
		{"Empty string", "", true, "cvv must be 3 or 4 digits"},
		{"All letters", "abc", true, "cvv must contain only digits"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBankCardCVV(tt.cvv)
			if tt.wantErr {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
