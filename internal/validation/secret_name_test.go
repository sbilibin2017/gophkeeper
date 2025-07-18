package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateSecretName(t *testing.T) {
	tests := []struct {
		name           string
		secretName     string
		wantErr        bool
		expectedErrMsg string
	}{
		{"Valid: simple", "MySecret123", false, ""},
		{"Valid: with spaces", "My Secret Name", false, ""},
		{"Valid: with underscores", "My_Secret_Name", false, ""},
		{"Valid: with hyphens", "My-Secret-Name", false, ""},
		{"Empty string", "", true, "secret name must not be empty"},
		{"Invalid: special char @", "Secret@Name", true, "secret name can only contain letters, digits, underscore, hyphen, and spaces"},
		{"Invalid: newline", "Secret\nName", true, "secret name can only contain letters, digits, underscore, hyphen, and spaces"},
		{"Invalid: tab", "Secret\tName", true, "secret name can only contain letters, digits, underscore, hyphen, and spaces"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSecretName(tt.secretName)
			if tt.wantErr {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
