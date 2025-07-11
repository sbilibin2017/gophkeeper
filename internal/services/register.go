package services

import (
	"context"

	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// Registerer определяет интерфейс, который должен реализовать любой фасад регистрации (HTTP/gRPC).
type Registerer interface {
	Register(ctx context.Context, secret *models.UsernamePassword) (string, error)
}

// RegisterService реализует сервис регистрации, делегируя вызов конкретному фасаду.
type RegisterService struct {
	facade Registerer
}

// NewRegisterService создает новый экземпляр RegisterService с заданным фасадом.
func NewRegisterService(facade Registerer) *RegisterService {
	return &RegisterService{facade: facade}
}

// Register вызывает метод фасада регистрации и возвращает токен или ошибку.
func (s *RegisterService) Register(
	ctx context.Context, secret *models.UsernamePassword,
) (string, error) {
	return s.facade.Register(ctx, secret)
}
