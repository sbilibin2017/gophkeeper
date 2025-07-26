package client

import (
	"context"

	"github.com/sbilibin2017/gophkeeper/inernal/models"
)

type Loginer interface {
	Login(ctx context.Context, req *models.UserLoginRequest) (*models.UserLoginResponse, error)
}

type LoginUsecase struct {
	loginer Loginer
}

func NewLoginerUsecase(loginer Loginer) (*LoginUsecase, error) {
	return &LoginUsecase{loginer: loginer}, nil
}

func (uc *LoginUsecase) Login(
	ctx context.Context,
	req *models.UserLoginRequest,
) (*models.UserLoginResponse, error) {
	return uc.loginer.Login(ctx, req)
}

type Registerer interface {
	Register(ctx context.Context, req *models.UserRegisterRequest) (*models.UserRegisterResponse, error)
}

type RegisterUsecase struct {
	registerer Registerer
}

func NewRegisterUsecase(registerer Registerer) (*RegisterUsecase, error) {
	return &RegisterUsecase{registerer: registerer}, nil
}

func (uc *RegisterUsecase) Register(
	ctx context.Context,
	req *models.UserRegisterRequest,
) (*models.UserRegisterResponse, error) {
	return uc.registerer.Register(ctx, req)
}
