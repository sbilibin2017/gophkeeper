package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// JWT holds config for signing and verifying tokens.
type JWT struct {
	secret   string
	lifetime time.Duration
}

// Opt defines a functional option for JWT configuration.
type Opt func(*JWT)

// WithSecret sets the signing secret.
func WithSecret(secret string) Opt {
	return func(j *JWT) {
		j.secret = secret
	}
}

// WithLifetime sets the token lifetime.
func WithLifetime(duration time.Duration) Opt {
	return func(j *JWT) {
		j.lifetime = duration
	}
}

// New constructs a JWT instance with given options.
func New(opts ...Opt) *JWT {
	j := &JWT{}
	for _, opt := range opts {
		opt(j)
	}
	return j
}

// claims defines the JWT claims with Username and standard fields.
type claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// Generate creates a signed JWT token string including username.
func (j *JWT) Generate(username string) (string, error) {
	now := time.Now()

	claims := claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.lifetime)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secret))
}

// GetUsername extracts the username from a JWT token string.
func (j *JWT) Parse(tokenStr string) (string, error) {
	parsedToken, err := jwt.ParseWithClaims(tokenStr, &claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(j.secret), nil
	})
	if err != nil {
		return "", err
	}

	if claims, ok := parsedToken.Claims.(*claims); ok && parsedToken.Valid {
		return claims.Username, nil
	}

	return "", errors.New("invalid token")
}
