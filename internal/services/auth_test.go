package services

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/gophkeeper/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestAuthService_Register_SuccessAndExist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserReader := NewMockUserReader(ctrl)
	mockUserWriter := NewMockUserWriter(ctrl)
	mockDeviceReader := NewMockDeviceReader(ctrl)
	mockDeviceWriter := NewMockDeviceWriter(ctrl)
	mockTokenGen := NewMockTokenGenerator(ctrl)

	hashPassword := func(p string) ([]byte, error) { return []byte("hashed_" + p), nil }
	generateRSAKeys := func(bits int) (*rsa.PrivateKey, error) {
		return rsa.GenerateKey(rand.Reader, bits)
	}
	generateRandom := func(size int) ([]byte, error) { return []byte("randomDEK"), nil }
	encryptDEK := func(pubKey *rsa.PublicKey, dek []byte) ([]byte, error) { return []byte("encryptedDEK"), nil }
	encodePrivKey := func(privKey *rsa.PrivateKey) []byte { return []byte("pemKey") }

	auth := NewAuthService(
		mockUserReader,
		mockUserWriter,
		mockDeviceReader,
		mockDeviceWriter,
		mockTokenGen,
		hashPassword,
		generateRSAKeys,
		generateRandom,
		encryptDEK,
		encodePrivKey,
	)

	tests := []struct {
		name           string
		existingUser   bool
		existingDevice bool
		expectedErr    error
	}{
		{
			name:           "successful registration",
			existingUser:   false,
			existingDevice: false,
			expectedErr:    nil,
		},
		{
			name:           "user already exists",
			existingUser:   true,
			existingDevice: false,
			expectedErr:    ErrUserExists,
		},
		{
			name:           "device already exists",
			existingUser:   false,
			existingDevice: true,
			expectedErr:    ErrDeviceExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			username := "user1"
			password := "pass"
			deviceID := "dev1"

			if tt.existingUser {
				mockUserReader.EXPECT().GetByUsername(ctx, username).Return(&models.UserDB{}, nil)
			} else {
				mockUserReader.EXPECT().GetByUsername(ctx, username).Return(nil, nil)
				mockUserWriter.EXPECT().Save(ctx, gomock.Any(), username, gomock.Any()).Return(nil)

				if tt.existingDevice {
					mockDeviceReader.EXPECT().GetByID(ctx, deviceID).Return(&models.DeviceDB{}, nil)
				} else {
					mockDeviceReader.EXPECT().GetByID(ctx, deviceID).Return(nil, nil)
					mockDeviceWriter.EXPECT().Save(ctx, deviceID, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
					mockTokenGen.EXPECT().Generate(gomock.Any()).Return("token123", nil)
				}
			}

			privKey, token, err := auth.Register(ctx, username, password, deviceID)

			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
				assert.Nil(t, privKey)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, []byte("pemKey"), privKey)
				assert.Equal(t, "token123", token)
			}
		})
	}
}

func TestAuthService_Register_ErrorCases(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserReader := NewMockUserReader(ctrl)
	mockUserWriter := NewMockUserWriter(ctrl)
	mockDeviceReader := NewMockDeviceReader(ctrl)
	mockDeviceWriter := NewMockDeviceWriter(ctrl)
	mockTokenGen := NewMockTokenGenerator(ctrl)

	hashPassword := func(p string) ([]byte, error) { return []byte("hashed_" + p), nil }
	generateRSAKeys := func(bits int) (*rsa.PrivateKey, error) {
		return rsa.GenerateKey(rand.Reader, bits)
	}
	generateRandom := func(size int) ([]byte, error) { return []byte("randomDEK"), nil }
	encryptDEK := func(pubKey *rsa.PublicKey, dek []byte) ([]byte, error) { return []byte("encryptedDEK"), nil }
	encodePrivKey := func(privKey *rsa.PrivateKey) []byte { return []byte("pemKey") }

	auth := NewAuthService(
		mockUserReader,
		mockUserWriter,
		mockDeviceReader,
		mockDeviceWriter,
		mockTokenGen,
		hashPassword,
		generateRSAKeys,
		generateRandom,
		encryptDEK,
		encodePrivKey,
	)

	ctx := context.Background()
	username := "user1"
	password := "pass"
	deviceID := "dev1"

	errorTests := []struct {
		name          string
		mockSetup     func()
		expectedError string
	}{
		{
			name: "userReader error",
			mockSetup: func() {
				mockUserReader.EXPECT().GetByUsername(ctx, username).Return(nil, errors.New("db error"))
			},
			expectedError: "db error",
		},
		{
			name: "hashPassword error",
			mockSetup: func() {
				mockUserReader.EXPECT().GetByUsername(ctx, username).Return(nil, nil)
				auth.hashPassword = func(p string) ([]byte, error) { return nil, errors.New("hash failed") }
			},
			expectedError: "hash failed",
		},
		{
			name: "userWriter.Save error",
			mockSetup: func() {
				mockUserReader.EXPECT().GetByUsername(ctx, username).Return(nil, nil)
				mockUserWriter.EXPECT().Save(ctx, gomock.Any(), username, gomock.Any()).Return(errors.New("save user failed"))
			},
			expectedError: "save user failed",
		},
		{
			name: "deviceReader.GetByID error",
			mockSetup: func() {
				mockUserReader.EXPECT().GetByUsername(ctx, username).Return(nil, nil)
				mockUserWriter.EXPECT().Save(ctx, gomock.Any(), username, gomock.Any()).Return(nil)
				mockDeviceReader.EXPECT().GetByID(ctx, deviceID).Return(nil, errors.New("device read error"))
			},
			expectedError: "device read error",
		},
		{
			name: "generateRSAKeys error",
			mockSetup: func() {
				mockUserReader.EXPECT().GetByUsername(ctx, username).Return(nil, nil)
				mockUserWriter.EXPECT().Save(ctx, gomock.Any(), username, gomock.Any()).Return(nil)
				mockDeviceReader.EXPECT().GetByID(ctx, deviceID).Return(nil, nil)
				auth.generateRSAKeys = func(bits int) (*rsa.PrivateKey, error) { return nil, errors.New("rsa error") }
			},
			expectedError: "rsa error",
		},
		{
			name: "generateRandom error",
			mockSetup: func() {
				mockUserReader.EXPECT().GetByUsername(ctx, username).Return(nil, nil)
				mockUserWriter.EXPECT().Save(ctx, gomock.Any(), username, gomock.Any()).Return(nil)
				mockDeviceReader.EXPECT().GetByID(ctx, deviceID).Return(nil, nil)
				auth.generateRandom = func(size int) ([]byte, error) { return nil, errors.New("random error") }
			},
			expectedError: "random error",
		},
		{
			name: "encryptDEK error",
			mockSetup: func() {
				mockUserReader.EXPECT().GetByUsername(ctx, username).Return(nil, nil)
				mockUserWriter.EXPECT().Save(ctx, gomock.Any(), username, gomock.Any()).Return(nil)
				mockDeviceReader.EXPECT().GetByID(ctx, deviceID).Return(nil, nil)
				auth.encryptDEK = func(pubKey *rsa.PublicKey, dek []byte) ([]byte, error) { return nil, errors.New("encrypt error") }
			},
			expectedError: "encrypt error",
		},
		{
			name: "deviceWriter.Save error",
			mockSetup: func() {
				mockUserReader.EXPECT().GetByUsername(ctx, username).Return(nil, nil)
				mockUserWriter.EXPECT().Save(ctx, gomock.Any(), username, gomock.Any()).Return(nil)
				mockDeviceReader.EXPECT().GetByID(ctx, deviceID).Return(nil, nil)
				mockDeviceWriter.EXPECT().Save(ctx, deviceID, gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("save device error"))
			},
			expectedError: "save device error",
		},
		{
			name: "tokenGenerator.Generate error",
			mockSetup: func() {
				mockUserReader.EXPECT().GetByUsername(ctx, username).Return(nil, nil)
				mockUserWriter.EXPECT().Save(ctx, gomock.Any(), username, gomock.Any()).Return(nil)
				mockDeviceReader.EXPECT().GetByID(ctx, deviceID).Return(nil, nil)
				mockDeviceWriter.EXPECT().Save(ctx, deviceID, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				mockTokenGen.EXPECT().Generate(gomock.Any()).Return("", errors.New("token error"))
			},
			expectedError: "token error",
		},
	}

	for _, tt := range errorTests {
		t.Run(tt.name, func(t *testing.T) {
			auth.hashPassword = hashPassword
			auth.generateRSAKeys = generateRSAKeys
			auth.generateRandom = generateRandom
			auth.encryptDEK = encryptDEK
			auth.encodePrivKey = encodePrivKey

			tt.mockSetup()

			priv, token, err := auth.Register(ctx, username, password, deviceID)
			assert.Nil(t, priv)
			assert.Empty(t, token)
			assert.EqualError(t, err, tt.expectedError)
		})
	}
}
