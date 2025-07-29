package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestSecretWriteService_Save_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWriter := NewMockSecretWriter(ctrl)
	mockParser := NewMockJWTParser(ctrl)

	token := "valid.token"
	username := "user123"
	secretName := "note1"
	secretType := "note"
	ciphertext := []byte("encrypted-data")
	aesKeyEnc := []byte("encrypted-key")

	mockParser.EXPECT().
		Parse(token).
		Return(username, nil)

	mockWriter.EXPECT().
		Save(gomock.Any(), username, secretName, secretType, ciphertext, aesKeyEnc).
		Return(nil)

	svc := NewSecretWriteService(mockWriter, mockParser)
	err := svc.Save(context.Background(), token, secretName, secretType, ciphertext, aesKeyEnc)
	assert.NoError(t, err)
}

func TestSecretWriteService_Save_InvalidToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWriter := NewMockSecretWriter(ctrl)
	mockParser := NewMockJWTParser(ctrl)

	mockParser.EXPECT().
		Parse("bad.token").
		Return("", errors.New("parse error"))

	svc := NewSecretWriteService(mockWriter, mockParser)
	err := svc.Save(context.Background(), "bad.token", "x", "x", []byte("c"), []byte("k"))
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestSecretReadService_Get_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReader := NewMockSecretReader(ctrl)
	mockParser := NewMockJWTParser(ctrl)

	token := "valid.token"
	username := "john"
	secretType := "card"
	secretName := "visa"

	now := time.Now()
	expectedSecret := &models.Secret{
		SecretOwner: username,
		SecretType:  secretType,
		SecretName:  secretName,
		Ciphertext:  []byte("secure-ciphertext"),
		AESKeyEnc:   []byte("secure-key"),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	mockParser.EXPECT().
		Parse(token).
		Return(username, nil)

	mockReader.EXPECT().
		Get(gomock.Any(), username, secretType, secretName).
		Return(expectedSecret, nil)

	svc := NewSecretReadService(mockReader, mockParser)
	secret, err := svc.Get(context.Background(), token, secretType, secretName)
	assert.NoError(t, err)
	assert.Equal(t, expectedSecret, secret)
}

func TestSecretReadService_Get_InvalidToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReader := NewMockSecretReader(ctrl)
	mockParser := NewMockJWTParser(ctrl)

	mockParser.EXPECT().
		Parse("bad.token").
		Return("", errors.New("bad jwt"))

	svc := NewSecretReadService(mockReader, mockParser)
	secret, err := svc.Get(context.Background(), "bad.token", "x", "y")
	assert.ErrorIs(t, err, ErrInvalidToken)
	assert.Nil(t, secret)
}

func TestSecretReadService_List_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReader := NewMockSecretReader(ctrl)
	mockParser := NewMockJWTParser(ctrl)

	token := "valid.token"
	username := "alice"

	now := time.Now()
	expectedSecrets := []*models.Secret{
		{
			SecretOwner: username,
			SecretType:  "password",
			SecretName:  "email",
			Ciphertext:  []byte("pass-cipher"),
			AESKeyEnc:   []byte("aes-key-1"),
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			SecretOwner: username,
			SecretType:  "card",
			SecretName:  "bank",
			Ciphertext:  []byte("card-cipher"),
			AESKeyEnc:   []byte("aes-key-2"),
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	mockParser.EXPECT().
		Parse(token).
		Return(username, nil)

	mockReader.EXPECT().
		List(gomock.Any(), username).
		Return(expectedSecrets, nil)

	svc := NewSecretReadService(mockReader, mockParser)
	secrets, err := svc.List(context.Background(), token)
	assert.NoError(t, err)
	assert.Equal(t, expectedSecrets, secrets)
}

func TestSecretReadService_List_InvalidToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReader := NewMockSecretReader(ctrl)
	mockParser := NewMockJWTParser(ctrl)

	mockParser.EXPECT().
		Parse("invalid.jwt").
		Return("", errors.New("token error"))

	svc := NewSecretReadService(mockReader, mockParser)
	secrets, err := svc.List(context.Background(), "invalid.jwt")
	assert.ErrorIs(t, err, ErrInvalidToken)
	assert.Nil(t, secrets)
}
