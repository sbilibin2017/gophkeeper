package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUser(t *testing.T) {
	t.Run("no options", func(t *testing.T) {
		user := NewUser()
		assert.NotNil(t, user)
		assert.Empty(t, user.Username)
		assert.Empty(t, user.Password)
	})

	t.Run("with username only", func(t *testing.T) {
		user := NewUser(WithUsername("john_doe"))
		assert.NotNil(t, user)
		assert.Equal(t, "john_doe", user.Username)
		assert.Empty(t, user.Password)
	})

	t.Run("with password only", func(t *testing.T) {
		user := NewUser(WithPassword("secret123"))
		assert.NotNil(t, user)
		assert.Empty(t, user.Username)
		assert.Equal(t, "secret123", user.Password)
	})

	t.Run("with username and password", func(t *testing.T) {
		user := NewUser(WithUsername("john_doe"), WithPassword("secret123"))
		assert.NotNil(t, user)
		assert.Equal(t, "john_doe", user.Username)
		assert.Equal(t, "secret123", user.Password)
	})
}
