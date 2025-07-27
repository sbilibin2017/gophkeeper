package usecases

import (
	"context"

	"github.com/sbilibin2017/gophkeeper/inernal/models"
)

// UsernameValidator defines an interface for validating usernames.
type UsernameValidator interface {
	// Validate checks if the provided username meets specific criteria.
	Validate(username string) error
}

// PasswordValidator defines an interface for validating passwords.
type PasswordValidator interface {
	// Validate checks if the provided password meets specific criteria.
	Validate(password string) error
}

// Registerer defines an interface for user registration.
type Registerer interface {
	// Register registers a new user and returns an authentication response.
	Register(ctx context.Context, req *models.AuthRegisterRequest) (*models.AuthResponse, error)
}

// Loginer defines an interface for user login.
type Loginer interface {
	// Login logs in a user and returns an authentication response.
	Login(ctx context.Context, req *models.AuthLoginRequest) (*models.AuthResponse, error)
}

// ClientRegisterHTTPApp handles the user registration process.
type ClientRegisterUsecase struct {
	val1 UsernameValidator
	val2 PasswordValidator
	reg  Registerer
}

// NewClientRegisterUsecase creates a new instance of ClientRegisterHTTPApp.
func NewClientRegisterUsecase(
	val1 UsernameValidator,
	val2 PasswordValidator,
	reg Registerer,
) *ClientRegisterUsecase {
	return &ClientRegisterUsecase{
		val1: val1,
		val2: val2,
		reg:  reg,
	}
}

// Execute validates the input credentials and performs user registration.
//
// It expects an AuthLoginRequest (even though it's for registration), validates
// the username and password using the provided validators, and then calls Registerer.
//
// Returns the AuthResponse if successful or an error if validation or registration fails.
func (r *ClientRegisterUsecase) Execute(
	ctx context.Context,
	req models.AuthRegisterRequest,
) (*models.AuthResponse, error) {
	if err := r.val1.Validate(req.Username); err != nil {
		return nil, err
	}

	if err := r.val2.Validate(req.Password); err != nil {
		return nil, err
	}

	return r.reg.Register(ctx, &req)
}

// ClientLoginApp handles the user login process.
type ClientLoginUsecase struct {
	loginer Loginer
}

// NewClientLoginHTTPApp creates a new instance of ClientLoginHTTPApp.
func NewClientLoginUsecase(loginer Loginer) *ClientLoginUsecase {
	return &ClientLoginUsecase{
		loginer: loginer,
	}
}

// Execute performs the login operation using the Loginer interface.
//
// It sends the login request and returns the authentication response or an error.
func (c *ClientLoginUsecase) Execute(
	ctx context.Context,
	req models.AuthLoginRequest,
) (*models.AuthResponse, error) {
	return c.loginer.Login(ctx, &req)
}
