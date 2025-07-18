package validation

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestValidateBankCardSecretName(t *testing.T) {
	tests := []struct {
		name       string
		secretName string
		wantErr    bool
	}{
		{"Empty", "", true},
		{"Valid simple", "My_Card-1", false},
		{"Valid with space", "Secret Name", false},
		{"Invalid char", "Secret!Name", true},
	}

	for _, tt := range tests {
		err := ValidateBankCardSecretName(tt.secretName)
		if tt.wantErr {
			assert.Error(t, err, tt.name)
		} else {
			assert.NoError(t, err, tt.name)
		}
	}
}

func TestValidateBankCardNumber(t *testing.T) {
	tests := []struct {
		name    string
		number  string
		wantErr bool
	}{
		{"Too short", "12345678901", true},
		{"Too long", "12345678901234567890", true},
		{"Valid length digits", "123456789012", false},
		{"Valid max length", "1234567890123456789", false},
		{"Contains letters", "12345abc6789", true},
	}

	for _, tt := range tests {
		err := ValidateBankCardNumber(tt.number)
		if tt.wantErr {
			assert.Error(t, err, tt.name)
		} else {
			assert.NoError(t, err, tt.name)
		}
	}
}

func TestValidateBankCardOwner(t *testing.T) {
	tests := []struct {
		name    string
		owner   string
		wantErr bool
	}{
		{"Empty owner", "", true},
		{"Valid letters spaces", "John Doe", false},
		{"Valid with hyphen", "Mary-Jane Smith", false},
		{"Invalid chars", "John_Doe", true},
	}

	for _, tt := range tests {
		err := ValidateBankCardOwner(tt.owner)
		if tt.wantErr {
			assert.Error(t, err, tt.name)
		} else {
			assert.NoError(t, err, tt.name)
		}
	}
}

// ...

func TestValidateBankCardExp(t *testing.T) {
	now := time.Now()

	futureMonth := now.Month()
	futureYear := now.Year() + 1

	monthStr := fmt.Sprintf("%02d", futureMonth) // two-digit month string
	yearShort := strconv.Itoa(futureYear - 2000) // last two digits of year as string
	yearLong := strconv.Itoa(futureYear)

	tests := []struct {
		name    string
		exp     string
		wantErr bool
	}{
		{"Empty", "", true},
		{"Bad format", "13/25", true},
		{"Bad format 2", "1/2025", true},
		{"Valid MM/YY future", monthStr + "/" + yearShort, false},
		{"Valid MM/YYYY future", monthStr + "/" + yearLong, false},
		{"Past date MM/YY", "01/20", true},
		{"Past date MM/YYYY", "01/2000", true},
	}

	for _, tt := range tests {
		err := ValidateBankCardExp(tt.exp)
		if tt.wantErr {
			assert.Error(t, err, tt.name)
		} else {
			assert.NoError(t, err, tt.name)
		}
	}
}

func TestValidateBankCardCVV(t *testing.T) {
	tests := []struct {
		name    string
		cvv     string
		wantErr bool
	}{
		{"Too short", "12", true},
		{"Too long", "12345", true},
		{"Valid 3 digits", "123", false},
		{"Valid 4 digits", "1234", false},
		{"Letters included", "12a", true},
	}

	for _, tt := range tests {
		err := ValidateBankCardCVV(tt.cvv)
		if tt.wantErr {
			assert.Error(t, err, tt.name)
		} else {
			assert.NoError(t, err, tt.name)
		}
	}
}

func TestValidateBankCardMeta(t *testing.T) {
	tests := []struct {
		name    string
		meta    string
		wantErr bool
	}{
		{"Empty meta", "", false},
		{"Printable", "Some metadata 123", false},
		{"Contains newline", "line1\nline2", true},
		{"Contains DEL char", string([]byte{127}), true},
		{"Contains tab", "tab\tchar", true},
	}

	for _, tt := range tests {
		err := ValidateBankCardMeta(tt.meta)
		if tt.wantErr {
			assert.Error(t, err, tt.name)
		} else {
			assert.NoError(t, err, tt.name)
		}
	}
}
