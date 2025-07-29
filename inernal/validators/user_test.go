package validators

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{"TooShort", "ab", true},
		{"ValidASCIIOnly", "abc123", false},
		{"ValidWithSpecials", "user_name!", false},
		{"InvalidWithEmoji", "user🙂", true},
		{"InvalidWithSpace", "user name", true},
		{"ValidWithSymbols", "user@123", false},
		{"NonASCIILetters", "юзер", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUsername(tt.username)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUsernameValidatorStruct(t *testing.T) {
	v := &UsernameValidator{}

	err := v.Validate("good_user123!")
	require.NoError(t, err)

	err = v.Validate("😎")
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid characters")
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"TooShort", "Ab1!", true},
		{"NoUppercase", "abc123!", true},
		{"NoDigit", "Abcdef!", true},
		{"NoSpecial", "Abcdef1", true},
		{"ValidStrongPassword", "Abc123!", false},
		{"WithUnicode", "Abc123🙂", true},
		{"AllLowercase", "abcdef!", true},
		{"OnlySpecials", "!@#$%^", true},
		{"UpperDigitSpecial", "A1@aaa", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePassword(tt.password)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPasswordValidatorStruct(t *testing.T) {
	v := &PasswordValidator{}

	err := v.Validate("StrongPass1@")
	require.NoError(t, err)

	err = v.Validate("weak")
	require.Error(t, err)
	require.Contains(t, err.Error(), "at least 6 characters")
}
