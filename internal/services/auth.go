package services

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"

	"github.com/google/uuid"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrUserExists возвращается, если пользователь с указанным именем уже существует.
	ErrUserExists = errors.New("user already exists")

	// ErrDeviceExists возвращается, если устройство с указанным ID уже зарегистрировано.
	ErrDeviceExists = errors.New("device already exists")
)

// UserReader описывает интерфейс для чтения данных о пользователе.
type UserReader interface {
	GetByUsername(ctx context.Context, username string) (*models.UserDB, error)
}

// UserWriter описывает интерфейс для записи данных о пользователе.
type UserWriter interface {
	Save(ctx context.Context, userID string, username string, password string) error
}

// DeviceReader описывает интерфейс для чтения данных об устройстве пользователя.
type DeviceReader interface {
	GetByID(ctx context.Context, deviceID string) (*models.DeviceDB, error)
}

// DeviceWriter описывает интерфейс для записи данных об устройстве пользователя.
type DeviceWriter interface {
	Save(ctx context.Context, deviceID string, userID string, publicKey string, encryptedDEK string) error
}

// TokenGenerator описывает интерфейс генерации токенов авторизации.
type TokenGenerator interface {
	Generate(userID string) (string, error)
}

// AuthService предоставляет методы для регистрации пользователей и устройств,
// используя внедрение зависимостей криптографических функций.
type AuthService struct {
	userReader     UserReader
	userWriter     UserWriter
	deviceReader   DeviceReader
	deviceWriter   DeviceWriter
	tokenGenerator TokenGenerator
}

// NewAuthService создаёт новый экземпляр AuthService.
func NewAuthService(
	userReader UserReader,
	userWriter UserWriter,
	deviceReader DeviceReader,
	deviceWriter DeviceWriter,
	tokenGenerator TokenGenerator,
) *AuthService {
	return &AuthService{
		userReader:     userReader,
		userWriter:     userWriter,
		deviceReader:   deviceReader,
		deviceWriter:   deviceWriter,
		tokenGenerator: tokenGenerator,
	}
}

// Register регистрирует нового пользователя и его устройство.
// Возвращает приватный ключ пользователя в PEM формате и токен авторизации.
//
// Параметры:
// - ctx: контекст выполнения запроса
// - username: имя пользователя
// - password: пароль пользователя
// - deviceID: уникальный идентификатор устройства
//
// Ошибки:
// - ErrUserExists: если пользователь с таким именем уже существует
// - ErrDeviceExists: если устройство с таким ID уже зарегистрировано
func (s *AuthService) Register(ctx context.Context, username, password, deviceID string) ([]byte, string, error) {
	// Проверка пользователя
	user, err := s.userReader.GetByUsername(ctx, username)
	if err != nil {
		return nil, "", err
	}
	if user != nil {
		return nil, "", ErrUserExists
	}

	// Хеш пароля
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	// Сохранение пользователя
	userID := uuid.New().String()
	if err := s.userWriter.Save(ctx, userID, username, string(hash)); err != nil {
		return nil, "", err
	}

	// Проверка устройства
	device, err := s.deviceReader.GetByID(ctx, deviceID)
	if err != nil {
		return nil, "", err
	}
	if device != nil {
		return nil, "", ErrDeviceExists
	}

	// Генерация ключей
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, "", err
	}
	pubKey := &privKey.PublicKey

	// Генерация DEK
	dek := make([]byte, 32)
	if _, err := rand.Read(dek); err != nil {
		return nil, "", err
	}

	// Шифрование DEK
	encryptedDEK, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pubKey, dek, nil)
	if err != nil {
		return nil, "", err
	}

	// Сохранение устройства
	if err := s.deviceWriter.Save(
		ctx,
		deviceID,
		userID,
		base64.StdEncoding.EncodeToString(x509.MarshalPKCS1PublicKey(pubKey)),
		base64.StdEncoding.EncodeToString(encryptedDEK),
	); err != nil {
		return nil, "", err
	}

	// Приватный ключ в PEM
	privBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privKey),
	})

	// Генерация токена
	token, err := s.tokenGenerator.Generate(userID)
	if err != nil {
		return nil, "", err
	}

	return privBytes, token, nil
}
