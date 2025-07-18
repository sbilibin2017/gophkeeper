package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateTextSecretName(t *testing.T) {
	tests := []struct {
		name       string
		secretName string
		wantErr    bool
	}{
		{"Empty string", "", true},
		{"Valid name with letters", "MySecret", false},
		{"Valid name with digits", "Secret123", false},
		{"Valid name with underscore", "secret_name", false},
		{"Valid name with hyphen and space", "secret-name 1", false},
		{"Invalid name with special char", "secret$name", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTextSecretName(tt.secretName)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateTextContent(t *testing.T) {
	assert.Error(t, ValidateTextContent(""))
	assert.NoError(t, ValidateTextContent("some text content"))
}

func TestValidateTextMeta(t *testing.T) {
	tests := []struct {
		name    string
		meta    string
		wantErr bool
	}{
		{"Empty meta", "", false},
		{"Printable ASCII", "This is meta info", false},
		{"Contains newline", "line1\nline2", false},
		{"Contains carriage return", "line1\rline2", false},
		{"Contains tab", "tab\tseparated", false},
		{"Contains control char", string([]byte{31}), true},
		{"Contains DEL char", string([]byte{127}), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTextMeta(tt.meta)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
