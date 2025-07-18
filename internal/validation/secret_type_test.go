package validation

import (
	"testing"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestValidateSecretType(t *testing.T) {
	tests := []struct {
		name    string
		typ     string
		wantErr bool
		errMsg  string
	}{
		{"Valid BankCard", models.SecretTypeBankCard, false, ""},
		{"Valid Binary", models.SecretTypeBinary, false, ""},
		{"Valid Text", models.SecretTypeText, false, ""},
		{"Valid UsernamePassword", models.SecretTypeUsernamePassword, false, ""},
		{"Invalid Type", "invalid_type", true, `unknown secret type: "invalid_type"`},
		{"Empty Type", "", true, `unknown secret type: ""`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSecretType(tt.typ)
			if tt.wantErr {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
