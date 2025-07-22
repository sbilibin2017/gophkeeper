package bankcard

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateBankCardNumber(t *testing.T) {
	tests := []struct {
		number   string
		expected bool
		name     string
	}{
		{"", false, "Empty string"},
		{"4111111111111111", true, "Valid Visa card"},
		{"5500000000000004", true, "Valid Mastercard"},
		{"340000000000009", true, "Valid Amex"},
		{"30000000000004", true, "Valid Diners Club"},
		{"1234567812345670", true, "Valid number with correct checksum"},
		{"1234567812345678", false, "Invalid number wrong checksum"},
		{"4111a11111111111", false, "Invalid character in string"},
		{"abcdefg", false, "Non-numeric input"},
		{"0000000000000000", true, "Valid number all zeros"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateNumber(tt.number)
			assert.Equal(t, tt.expected, result)
		})
	}
}
