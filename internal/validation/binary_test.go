package validation

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateBinarySecretName(t *testing.T) {
	tests := []struct {
		name       string
		secretName string
		wantErr    bool
	}{
		{"Empty secretName", "", true},
		{"Valid letters digits underscore", "abc_123", false},
		{"Valid with spaces and hyphen", "secret-name 1", false},
		{"Invalid with special char", "secret$name", true},
		{"Invalid with punctuation", "secret.name", true},
	}

	for _, tt := range tests {
		err := ValidateBinarySecretName(tt.secretName)
		if tt.wantErr {
			assert.Error(t, err, tt.name)
		} else {
			assert.NoError(t, err, tt.name)
		}
	}
}

func TestValidateBinaryDataPath(t *testing.T) {
	// Create a temp file for testing
	tmpFile, err := os.CreateTemp("", "testfile")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Create a temp directory for testing
	tmpDir, err := os.MkdirTemp("", "testdir")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name     string
		dataPath string
		wantErr  bool
	}{
		{"Empty dataPath", "", true},
		{"Non-existent file", "nonexistent.file", true},
		{"Existing file", tmpFile.Name(), false},
		{"Directory instead of file", tmpDir, true},
	}

	for _, tt := range tests {
		err := ValidateBinaryDataPath(tt.dataPath)
		if tt.wantErr {
			assert.Error(t, err, tt.name)
		} else {
			assert.NoError(t, err, tt.name)
		}
	}
}

func TestValidateBinaryMeta(t *testing.T) {
	tests := []struct {
		name    string
		meta    string
		wantErr bool
	}{
		{"Empty meta", "", false},
		{"Printable ASCII", "This is a meta string.", false},
		{"Contains newline", "line1\nline2", true},
		{"Contains tab", "tab\tchar", true},
		{"Contains DEL char", string([]byte{127}), true},
		{"Contains valid extended ASCII", "ñáéíóú", false}, // Unicode printable chars
	}

	for _, tt := range tests {
		err := ValidateBinaryMeta(tt.meta)
		if tt.wantErr {
			assert.Error(t, err, tt.name)
		} else {
			assert.NoError(t, err, tt.name)
		}
	}
}
