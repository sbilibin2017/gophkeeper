package usecases

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/sbilibin2017/gophkeeper/inernal/models"
)

func TestInteractiveSyncUsecase_Sync(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClientLister := NewMockClientLister(ctrl)
	mockServerGetter := NewMockServerGetter(ctrl)
	mockServerSaver := NewMockServerSaver(ctrl)
	mockCryptor := NewMockCryptor(ctrl)

	ctx := context.Background()
	token := "user-token" // token used as secret owner

	secretName := "secret1"
	secretType := "type1"
	updatedAtClient := time.Now()
	updatedAtServer := updatedAtClient.Add(-time.Hour) // server older

	clientSecret := &models.SecretDB{
		SecretName: secretName,
		SecretType: secretType,
		UpdatedAt:  updatedAtClient,
		Ciphertext: []byte("encrypted-client"),
		AESKeyEnc:  []byte("key-client"),
	}

	serverSecret := &models.SecretDB{
		SecretName: secretName,
		SecretType: secretType,
		UpdatedAt:  updatedAtServer,
		Ciphertext: []byte("encrypted-server"),
		AESKeyEnc:  []byte("key-server"),
	}

	clientPlainText := []byte("client plain text")
	serverPlainText := []byte("server plain text")

	tests := []struct {
		name          string
		inputChoice   string
		expectSave    bool
		expectedError error
		serverSecret  *models.SecretDB
	}{
		{
			name:         "Choose client version",
			inputChoice:  "1\n",
			expectSave:   true,
			serverSecret: serverSecret,
		},
		{
			name:         "Choose server version",
			inputChoice:  "2\n",
			expectSave:   false,
			serverSecret: serverSecret,
		},
		{
			name:          "Invalid choice",
			inputChoice:   "3\n",
			expectSave:    false,
			expectedError: errors.New("invalid choice"),
			serverSecret:  serverSecret,
		},
		{
			name:         "No server secret, save client automatically",
			inputChoice:  "",
			expectSave:   true,
			serverSecret: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup expected List call with token
			mockClientLister.EXPECT().
				List(ctx, token).
				Return([]*models.SecretDB{clientSecret}, nil)

			// Setup expected Get call with SecretGetRequest
			mockServerGetter.EXPECT().
				Get(ctx, &models.SecretGetRequest{
					SecretName: secretName,
					SecretType: secretType,
					Token:      token,
				}).
				Return(tt.serverSecret, nil)

			// Decrypt client secret (always)
			mockCryptor.EXPECT().
				Decrypt(&models.Encrypted{
					Ciphertext: clientSecret.Ciphertext,
					AESKeyEnc:  clientSecret.AESKeyEnc,
				}).
				Return(clientPlainText, nil)

			// If server secret exists, decrypt it
			if tt.serverSecret != nil {
				mockCryptor.EXPECT().
					Decrypt(&models.Encrypted{
						Ciphertext: tt.serverSecret.Ciphertext,
						AESKeyEnc:  tt.serverSecret.AESKeyEnc,
					}).
					Return(serverPlainText, nil)
			}

			// Expect Save only if expectSave is true
			if tt.expectSave {
				mockServerSaver.EXPECT().
					Save(ctx, &models.SecretSaveRequest{
						SecretName: clientSecret.SecretName,
						SecretType: clientSecret.SecretType,
						Ciphertext: clientSecret.Ciphertext,
						AESKeyEnc:  clientSecret.AESKeyEnc,
						Token:      token,
					}).
					Return(nil)
			}

			// Prepare input reader with the choice string
			reader := strings.NewReader(tt.inputChoice)

			syncer := NewInteractiveSyncUsecase(mockClientLister, mockServerGetter, mockServerSaver, mockCryptor)
			err := syncer.Sync(ctx, reader, token)

			if tt.expectedError != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestClientSyncUsecase_Sync(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClientLister := NewMockClientLister(ctrl)
	mockServerGetter := NewMockServerGetter(ctrl)
	mockServerSaver := NewMockServerSaver(ctrl)

	ctx := context.Background()
	token := "user-token"

	secretName := "secret1"
	secretType := "typeA"
	now := time.Now()
	older := now.Add(-time.Hour)

	clientSecret := &models.SecretDB{
		SecretName: secretName,
		SecretType: secretType,
		UpdatedAt:  now,
		Ciphertext: []byte("encrypted-client"),
		AESKeyEnc:  []byte("key-client"),
	}

	serverSecretNewer := &models.SecretDB{
		SecretName: secretName,
		SecretType: secretType,
		UpdatedAt:  now.Add(time.Hour),
		Ciphertext: []byte("encrypted-server-newer"),
		AESKeyEnc:  []byte("key-server-newer"),
	}

	serverSecretOlder := &models.SecretDB{
		SecretName: secretName,
		SecretType: secretType,
		UpdatedAt:  older,
		Ciphertext: []byte("encrypted-server-older"),
		AESKeyEnc:  []byte("key-server-older"),
	}

	tests := []struct {
		name          string
		listReturn    []*models.SecretDB
		listErr       error
		getReturn     *models.SecretDB
		getErr        error
		expectSave    bool
		saveErr       error
		expectError   bool
		errorContains string
	}{
		{
			name:          "List returns error",
			listReturn:    nil,
			listErr:       errors.New("list error"),
			expectError:   true,
			errorContains: "failed to list client secrets",
		},
		{
			name:          "ServerGetter returns error",
			listReturn:    []*models.SecretDB{clientSecret},
			getReturn:     nil,
			getErr:        errors.New("get error"),
			expectError:   true,
			errorContains: "get error",
		},
		{
			name:        "Server has newer secret, no save",
			listReturn:  []*models.SecretDB{clientSecret},
			getReturn:   serverSecretNewer,
			expectSave:  false,
			expectError: false,
		},
		{
			name:       "Server secret missing, save client secret",
			listReturn: []*models.SecretDB{clientSecret},
			getReturn:  nil,
			expectSave: true,
		},
		{
			name:       "Server secret older, save client secret",
			listReturn: []*models.SecretDB{clientSecret},
			getReturn:  serverSecretOlder,
			expectSave: true,
		},
		{
			name:          "Save returns error",
			listReturn:    []*models.SecretDB{clientSecret},
			getReturn:     nil,
			expectSave:    true,
			saveErr:       errors.New("save error"),
			expectError:   true,
			errorContains: "failed to save client secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClientLister.EXPECT().List(ctx, token).Return(tt.listReturn, tt.listErr)

			if tt.listErr == nil {
				mockServerGetter.EXPECT().
					Get(ctx, &models.SecretGetRequest{
						SecretName: secretName,
						SecretType: secretType,
						Token:      token,
					}).
					Return(tt.getReturn, tt.getErr)
			}

			if tt.getErr == nil && tt.expectSave {
				mockServerSaver.EXPECT().
					Save(ctx, &models.SecretSaveRequest{
						SecretName: clientSecret.SecretName,
						SecretType: clientSecret.SecretType,
						Ciphertext: clientSecret.Ciphertext,
						AESKeyEnc:  clientSecret.AESKeyEnc,
						Token:      token,
					}).
					Return(tt.saveErr)
			}

			syncer := NewClientSyncUsecase(mockClientLister, mockServerGetter, mockServerSaver)
			err := syncer.Sync(ctx, token)

			if tt.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errorContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestServerSyncUsecase_Sync(t *testing.T) {
	syncer := NewServerSyncUsecase()
	err := syncer.Sync(context.Background())
	require.NoError(t, err)
}
