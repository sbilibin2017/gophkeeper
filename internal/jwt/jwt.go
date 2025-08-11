package jwt

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/grpc/metadata"
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
func (j *JWT) GetUsername(tokenStr string) (string, error) {
	parsedToken, err := jwt.ParseWithClaims(tokenStr, &claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(j.secret), nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := parsedToken.Claims.(*claims)
	if !ok || !parsedToken.Valid || claims == nil {
		return "", errors.New("invalid token")
	}

	return claims.Username, nil
}

// GetTokenFromHeader extracts the Bearer token string from the Authorization header in the HTTP request.
func GetTokenFromHeader(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header missing")
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", errors.New("invalid authorization header format")
	}
	return parts[1], nil
}

// GetTokenFromContext extracts the Bearer token string from the gRPC context metadata.
// Looks for "authorization" metadata with format "Bearer <token>".
func GetTokenFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("missing metadata in context")
	}

	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return "", errors.New("authorization metadata missing")
	}

	parts := strings.SplitN(authHeaders[0], " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", errors.New("invalid authorization metadata format")
	}

	return parts[1], nil
}
