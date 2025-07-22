package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		password string
		valid    bool
		name     string
	}{
		{"", false, "Empty password"},
		{"short7!", false, "Too short, 7 chars"},
		{"alllowercase1!", false, "No uppercase letters"},
		{"ALLUPPERCASE1!", false, "No lowercase letters"},
		{"NoDigits!!", false, "No digits"},
		{"NoSpecialChar1", false, "No special characters"},
		{"Valid1!", false, "Too short, 7 chars with all types"}, // Fixed here: 7 chars now
		{"ValidPass1!", true, "Valid password"},
		{"A1!aaaaaa", true, "Valid with min length 8"},
		{"ThisIsAValidPassword123!", true, "Long valid password"},
		{"NoSpecialCharButVeryLong123456789012345678901234567890123456789012345678901234567890", false, "Long but no special chars"},
		{"Special!ButNoDigit", false, "Has special chars but no digit"},
		{"UPPERlower123!", true, "Mixed case with digit and special char"},
		{"Symbols#%&123aA", true, "Valid with multiple special chars"},
		{"TooLong" + string(make([]byte, 130)), false, "Too long password"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidatePassword(tt.password)
			assert.Equal(t, tt.valid, result)
		})
	}
}
