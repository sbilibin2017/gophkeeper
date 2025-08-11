package rsa

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateRSAKeys(t *testing.T) {
	pub, priv, err := GenerateRSAKeys("testuser")
	require.NoError(t, err)
	require.NotNil(t, pub)
	require.NotNil(t, priv)
	assert.Contains(t, string(pub), "BEGIN PUBLIC KEY")
	assert.Contains(t, string(priv), "BEGIN RSA PRIVATE KEY")
}

func TestSaveKeyPair(t *testing.T) {
	username := "testuser_save"
	pub, priv, err := GenerateRSAKeys(username)
	require.NoError(t, err)

	err = SaveKeyPair(username, pub, priv)
	require.NoError(t, err)

	// Check that file exists
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	filePath := filepath.Join(homeDir, ".config", username+".json")
	defer os.Remove(filePath) // clean up after test

	data, err := os.ReadFile(filePath)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	var keyPair RSAKeyPair
	err = json.Unmarshal(data, &keyPair)
	require.NoError(t, err)

	assert.Equal(t, string(pub), keyPair.PublicKey)
	assert.Equal(t, string(priv), keyPair.PrivateKey)
}
