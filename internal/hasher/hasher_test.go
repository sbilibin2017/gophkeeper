package hasher

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasher_Hash(t *testing.T) {
	h := New()

	password := []byte("mySecret123")

	hashed, err := h.Hash(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hashed)
}

func TestHasher_Compare(t *testing.T) {
	h := New()

	password := []byte("mySecret123")
	hashed, err := h.Hash(password)
	assert.NoError(t, err)

	// Compare correct password
	err = h.Compare(hashed, password)
	assert.NoError(t, err)

	// Compare incorrect password
	err = h.Compare(hashed, []byte("wrongPassword"))
	assert.Error(t, err)
}
