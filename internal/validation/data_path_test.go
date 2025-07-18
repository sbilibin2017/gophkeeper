package validation

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateDataPath(t *testing.T) {
	// Create a temporary file for testing valid file path
	tmpFile, err := ioutil.TempFile("", "testfile")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Create a temporary directory for testing directory case
	tmpDir, err := ioutil.TempDir("", "testdir")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name    string
		path    string
		wantErr bool
		errMsg  string
	}{
		{"Empty path", "", true, "data path must not be empty"},
		{"Non-existent path", "/path/does/not/exist", true, "data path does not exist or cannot be accessed"},
		{"Directory path", tmpDir, true, "data path must be a file, not a directory"},
		{"Valid file path", tmpFile.Name(), false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDataPath(tt.path)
			if tt.wantErr {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
