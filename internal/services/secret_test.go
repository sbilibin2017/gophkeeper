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

func TestSecretWriteService_Save(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWriter := NewMockSecretWriter(ctrl)
	service := NewSecretWriteService(mockWriter)

	ctx := context.Background()
	username := "alice"
	secretName := "mysecret"
	secretType := "password"
	ciphertext := []byte("cipherdata")
	aesKeyEnc := []byte("keydata")

	tests := []struct {
		name      string
		saveErr   error
		expectErr error
	}{
		{"success", nil, nil},
		{"save fails", errors.New("save error"), errors.New("save error")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWriter.EXPECT().
				Save(ctx, username, secretName, secretType, ciphertext, aesKeyEnc).
				Return(tt.saveErr)

			err := service.Save(ctx, username, secretName, secretType, ciphertext, aesKeyEnc)
			if tt.expectErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSecretReadService_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReader := NewMockSecretReader(ctrl)
	service := NewSecretReadService(mockReader)

	ctx := context.Background()
	username := "alice"
	secretType := "password"
	secretName := "mysecret"

	now := time.Now()
	expectedSecret := &models.Secret{
		SecretName:  secretName,
		SecretType:  secretType,
		SecretOwner: username,
		Ciphertext:  []byte("secretdata"),
		AESKeyEnc:   []byte("keydata"),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	tests := []struct {
		name        string
		getSecret   *models.Secret
		getErr      error
		expectErr   error
		expectValue *models.Secret
	}{
		{"success", expectedSecret, nil, nil, expectedSecret},
		{"get fails", nil, errors.New("not found"), errors.New("not found"), nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReader.EXPECT().
				Get(ctx, username, secretType, secretName).
				Return(tt.getSecret, tt.getErr)

			secret, err := service.Get(ctx, username, secretType, secretName)
			if tt.expectErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectErr.Error())
				assert.Nil(t, secret)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectValue, secret)
			}
		})
	}
}

func TestSecretReadService_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReader := NewMockSecretReader(ctrl)
	service := NewSecretReadService(mockReader)

	ctx := context.Background()
	username := "alice"

	now := time.Now()
	secrets := []*models.Secret{
		{
			SecretName:  "secret1",
			SecretType:  "password",
			SecretOwner: username,
			Ciphertext:  []byte("data1"),
			AESKeyEnc:   []byte("key1"),
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			SecretName:  "secret2",
			SecretType:  "note",
			SecretOwner: username,
			Ciphertext:  []byte("data2"),
			AESKeyEnc:   []byte("key2"),
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	tests := []struct {
		name        string
		listSecrets []*models.Secret
		listErr     error
		expectErr   error
		expectValue []*models.Secret
	}{
		{"success", secrets, nil, nil, secrets},
		{"list fails", nil, errors.New("list error"), errors.New("list error"), nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReader.EXPECT().
				List(ctx, username).
				Return(tt.listSecrets, tt.listErr)

			result, err := service.List(ctx, username)
			if tt.expectErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectErr.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectValue, result)
			}
		})
	}
}
