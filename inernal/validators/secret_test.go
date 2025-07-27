package validators

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateLuhn(t *testing.T) {
	tests := []struct {
		name    string
		number  string
		wantErr bool
	}{
		{"EmptyNumber", "", true},
		{"NonDigitCharacters", "1234a567890", true},
		{"InvalidLuhn", "1234567812345678", true},
		{"ValidVisa", "4111111111111111", false},
		{"ValidMasterCard", "5500000000000004", false},
		{"ValidAmex", "378282246310005", false},
		{"ValidWithSpaces", "4242 4242 4242 4242", true}, // not allowed due to spaces
		{"ValidShortLuhn", "79927398713", false},         // Classic Luhn test
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLuhn(tt.number)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestLuhnValidatorStruct(t *testing.T) {
	v := &LuhnValidator{}

	err := v.Validate("4111111111111111")
	require.NoError(t, err)

	err = v.Validate("1234abc")
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid characters")
}

func TestValidateCVV(t *testing.T) {
	tests := []struct {
		name    string
		cvv     string
		wantErr bool
	}{
		{"EmptyCVV", "", true},
		{"TooShort", "12", true},
		{"TooLong", "1234", true},
		{"NonDigit", "12a", true},
		{"ValidCVV", "123", false},
		{"AllZeroes", "000", false},
		{"Spaces", " 12", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCVV(tt.cvv)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCVVValidatorStruct(t *testing.T) {
	v := &CVVValidator{}

	err := v.Validate("321")
	require.NoError(t, err)

	err = v.Validate("12a")
	require.Error(t, err)
	require.Contains(t, err.Error(), "non-digit")
}
