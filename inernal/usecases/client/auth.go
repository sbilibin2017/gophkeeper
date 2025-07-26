package client

import (
	"context"

	"github.com/sbilibin2017/gophkeeper/inernal/models"
)

type UsernameValidator interface {
	Validate(username string) error
}

type PasswordValidator interface {
	Validate(password string) error
}

type Registerer interface {
	Register(ctx context.Context, req *models.UserRegisterRequest) (*models.UserRegisterResponse, error)
}

type RegisterUsecase struct {
	usernameValidator UsernameValidator
	passwordValidator PasswordValidator
	registerer        Registerer
}

func NewRegisterUsecase(
	usernameValidator UsernameValidator,
	passwordValidator PasswordValidator,
	registerer Registerer,
) (*RegisterUsecase, error) {
	return &RegisterUsecase{
		usernameValidator: usernameValidator,
		passwordValidator: passwordValidator,
		registerer:        registerer,
	}, nil
}

func (uc *RegisterUsecase) Execute(
	ctx context.Context,
	req *models.UserRegisterRequest,
) (*models.UserRegisterResponse, error) {
	err := uc.usernameValidator.Validate(req.Username)
	if err != nil {
		return nil, err
	}
	err = uc.passwordValidator.Validate(req.Password)
	if err != nil {
		return nil, err
	}
	return uc.registerer.Register(ctx, req)
}

type Loginer interface {
	Login(ctx context.Context, req *models.UserLoginRequest) (*models.UserLoginResponse, error)
}

type LoginUsecase struct {
	loginer Loginer
}

func NewLoginerUsecase(loginer Loginer) (*LoginUsecase, error) {
	return &LoginUsecase{loginer: loginer}, nil
}

func (uc *LoginUsecase) Execute(
	ctx context.Context,
	req *models.UserLoginRequest,
) (*models.UserLoginResponse, error) {
	return uc.loginer.Login(ctx, req)
}
