package hasher

import "golang.org/x/crypto/bcrypt"

type Hasher struct{}

// New creates and returns a new Hasher instance.
func New() *Hasher {
	return &Hasher{}
}

// Hash hashes the given value using bcrypt with default cost.
func (h *Hasher) Hash(value []byte) ([]byte, error) {
	return bcrypt.GenerateFromPassword(value, bcrypt.DefaultCost)
}

// Compare compares a bcrypt hashed value with a plaintext value.
// Returns nil if they match, otherwise an error.
func (h *Hasher) Compare(hashedValue []byte, value []byte) error {
	return bcrypt.CompareHashAndPassword(hashedValue, value)
}
