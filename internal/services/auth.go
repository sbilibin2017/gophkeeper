package services

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"errors"

	"github.com/google/uuid"
	"github.com/sbilibin2017/gophkeeper/internal/models"
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

	hashPassword    func(string) ([]byte, error)
	generateRSAKeys func(bits int) (*rsa.PrivateKey, error)
	generateRandom  func(size int) ([]byte, error)
	encryptDEK      func(pubKey *rsa.PublicKey, dek []byte) ([]byte, error)
	encodePrivKey   func(privKey *rsa.PrivateKey) []byte
}

// NewAuthService создаёт новый экземпляр AuthService.
func NewAuthService(
	userReader UserReader,
	userWriter UserWriter,
	deviceReader DeviceReader,
	deviceWriter DeviceWriter,
	tokenGenerator TokenGenerator,
	hashPassword func(string) ([]byte, error),
	generateRSAKeys func(bits int) (*rsa.PrivateKey, error),
	generateRandom func(size int) ([]byte, error),
	encryptDEK func(pubKey *rsa.PublicKey, dek []byte) ([]byte, error),
	encodePrivKey func(privKey *rsa.PrivateKey) []byte,
) *AuthService {
	return &AuthService{
		userReader:      userReader,
		userWriter:      userWriter,
		deviceReader:    deviceReader,
		deviceWriter:    deviceWriter,
		tokenGenerator:  tokenGenerator,
		hashPassword:    hashPassword,
		generateRSAKeys: generateRSAKeys,
		generateRandom:  generateRandom,
		encryptDEK:      encryptDEK,
		encodePrivKey:   encodePrivKey,
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
	// Проверка существующего пользователя
	user, err := s.userReader.GetByUsername(ctx, username)
	if err != nil {
		return nil, "", err
	}
	if user != nil {
		return nil, "", ErrUserExists
	}

	// Хеширование пароля
	hash, err := s.hashPassword(password)
	if err != nil {
		return nil, "", err
	}

	// Сохранение пользователя
	userID := uuid.New().String()
	if err := s.userWriter.Save(ctx, userID, username, string(hash)); err != nil {
		return nil, "", err
	}

	// Проверка существующего устройства
	device, err := s.deviceReader.GetByID(ctx, deviceID)
	if err != nil {
		return nil, "", err
	}
	if device != nil {
		return nil, "", ErrDeviceExists
	}

	// Генерация RSA ключей
	privKey, err := s.generateRSAKeys(2048)
	if err != nil {
		return nil, "", err
	}
	pubKey := &privKey.PublicKey

	// Генерация DEK
	dek, err := s.generateRandom(32)
	if err != nil {
		return nil, "", err
	}

	// Шифрование DEK
	encryptedDEK, err := s.encryptDEK(pubKey, dek)
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
	privBytes := s.encodePrivKey(privKey)

	// Генерация токена
	token, err := s.tokenGenerator.Generate(userID)
	if err != nil {
		return nil, "", err
	}

	return privBytes, token, nil
}
