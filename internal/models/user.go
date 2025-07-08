package models

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// UserOpt defines a functional option for User
type UserOpt func(*User)

// NewUser creates a new User instance with given functional options
func NewUser(opts ...UserOpt) *User {
	u := &User{}
	for _, opt := range opts {
		opt(u)
	}
	return u
}

// WithUsername sets the Username field
func WithUsername(username string) UserOpt {
	return func(u *User) {
		u.Username = username
	}
}

// WithPassword sets the Password field
func WithPassword(password string) UserOpt {
	return func(u *User) {
		u.Password = password
	}
}
