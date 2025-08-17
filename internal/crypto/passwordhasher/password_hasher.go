package passwordhasher

import (
	"golang.org/x/crypto/bcrypt"
)

// Hasher оборачивает bcrypt для хэширования и проверки паролей.
type PasswordHasher struct{}

// New создает и возвращает новый экземпляр PasswordHasher.
func New() *PasswordHasher {
	return &PasswordHasher{}
}

// Hash принимает пароль и возвращает его bcrypt-хэш.
func (h *PasswordHasher) Hash(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

// Compare проверяет, соответствует ли пароль данному хэшу.
func (h *PasswordHasher) Compare(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
