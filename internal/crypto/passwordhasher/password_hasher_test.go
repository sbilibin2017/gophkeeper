package passwordhasher

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswordHasher_HashAndCompare(t *testing.T) {
	h := New()

	password := "mySecretPassword123!"

	// Тестируем хэширование пароля
	hash, err := h.Hash(password)
	require.NoError(t, err, "ошибка при хэшировании пароля")
	require.NotEmpty(t, hash, "хэш не должен быть пустым")

	// Тестируем сравнение правильного пароля
	err = h.Compare(string(hash), password)
	assert.NoError(t, err, "пароль должен совпадать с хэшем")

	// Тестируем сравнение неправильного пароля
	err = h.Compare(string(hash), "wrongPassword")
	assert.Error(t, err, "неправильный пароль должен вернуть ошибку")
}

func TestPasswordHasher_HashUniqueness(t *testing.T) {
	h := New()

	password := "samePassword"

	hash1, err := h.Hash(password)
	require.NoError(t, err)

	hash2, err := h.Hash(password)
	require.NoError(t, err)

	// Два хэша одного и того же пароля должны быть разными из-за соли
	assert.NotEqual(t, hash1, hash2, "хэши одного пароля должны быть разными")
}

func TestPasswordHasher_EmptyPassword(t *testing.T) {
	h := New()

	// Проверяем пустой пароль
	hash, err := h.Hash("")
	require.NoError(t, err)
	require.NotEmpty(t, hash)

	// Проверка сравнения пустого пароля
	err = h.Compare(string(hash), "")
	assert.NoError(t, err)
}
