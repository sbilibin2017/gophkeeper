package validators

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateLuhn(t *testing.T) {
	tests := []struct {
		name    string
		number  string
		wantErr bool
	}{
		{"ValidCard_Visa", "4539578763621486", false},
		{"ValidCard_MasterCard", "5500005555555559", false},
		{"InvalidCard_LuhnFail", "1234567812345678", true}, // fixed
		{"EmptyInput", "", true},
		{"NonDigitChars", "4539-5787-6362-1486", true},
		{"AlphabetInNumber", "4539A78763621486", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLuhn(tt.number)
			if tt.wantErr {
				assert.Error(t, err, "expected error")
			} else {
				assert.NoError(t, err, "expected no error")
			}
		})
	}
}

func TestValidateCVV(t *testing.T) {
	tests := []struct {
		name    string
		cvv     string
		wantErr bool
	}{
		{"ValidCVV", "123", false},
		{"TooShort", "12", true},
		{"TooLong", "1234", true},
		{"EmptyCVV", "", true},
		{"NonDigitCVV", "12a", true},
		{"SpecialCharCVV", "1@3", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCVV(tt.cvv)
			if tt.wantErr {
				assert.Error(t, err, "expected error")
			} else {
				assert.NoError(t, err, "expected no error")
			}
		})
	}
}
