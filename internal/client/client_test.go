package client

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/stretchr/testify/require"
)

func TestClientRegister(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRegisterer := NewMockRegisterer(ctrl)

	tests := []struct {
		name          string
		setupMock     func()
		username      string
		password      string
		expectedToken string
		expectErr     bool
	}{
		{
			name: "success",
			setupMock: func() {
				token := "token123"
				mockRegisterer.EXPECT().
					Register(gomock.Any(), "user", "pass").
					Return(&token, nil)
			},
			username:      "user",
			password:      "pass",
			expectedToken: "token123",
			expectErr:     false,
		},
		{
			name: "error from register",
			setupMock: func() {
				mockRegisterer.EXPECT().
					Register(gomock.Any(), "user", "pass").
					Return(nil, errors.New("register error"))
			},
			username:  "user",
			password:  "pass",
			expectErr: true,
		},
		{
			name: "nil token returned",
			setupMock: func() {
				mockRegisterer.EXPECT().
					Register(gomock.Any(), "user", "pass").
					Return(nil, nil)
			},
			username:  "user",
			password:  "pass",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			token, err := ClientRegister(context.Background(), mockRegisterer, tt.username, tt.password)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedToken, token)
			}
		})
	}
}

func TestClientLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLoginer := NewMockLoginer(ctrl)

	tests := []struct {
		name          string
		setupMock     func()
		username      string
		password      string
		expectedToken string
		expectErr     bool
	}{
		{
			name: "success",
			setupMock: func() {
				token := "token123"
				mockLoginer.EXPECT().
					Login(gomock.Any(), "user", "pass").
					Return(&token, nil)
			},
			username:      "user",
			password:      "pass",
			expectedToken: "token123",
			expectErr:     false,
		},
		{
			name: "error from login",
			setupMock: func() {
				mockLoginer.EXPECT().
					Login(gomock.Any(), "user", "pass").
					Return(nil, errors.New("login error"))
			},
			username:  "user",
			password:  "pass",
			expectErr: true,
		},
		{
			name: "nil token returned",
			setupMock: func() {
				mockLoginer.EXPECT().
					Login(gomock.Any(), "user", "pass").
					Return(nil, nil)
			},
			username:  "user",
			password:  "pass",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			token, err := ClientLogin(context.Background(), mockLoginer, tt.username, tt.password)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedToken, token)
			}
		})
	}
}

func TestClientAddBankcard(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSaver := NewMockClientSaver(ctrl)
	mockEncryptor := NewMockEncryptor(ctrl)

	ctx := context.Background()
	token := "token123"
	secretName := "secretName"
	number := "1234123412341234"
	owner := "Owner Name"
	exp := "12/25"
	cvv := "123"
	meta := "some meta"

	// Expected payload struct (to marshal and verify later)
	expectedPayload := models.BankcardPayload{
		Number: number,
		Owner:  owner,
		Exp:    exp,
		CVV:    cvv,
		Meta:   &meta,
	}
	plaintext, err := json.Marshal(expectedPayload)
	require.NoError(t, err)

	encrypted := models.SecretEncrypted{
		Ciphertext: []byte("encryptedText"),
		AESKeyEnc:  []byte("encryptedKey"),
	}

	mockEncryptor.EXPECT().
		Encrypt(plaintext).
		Return(&encrypted, nil)

	mockSaver.EXPECT().
		Save(ctx, token, secretName, models.SecretTypeBankCard, encrypted.Ciphertext, encrypted.AESKeyEnc).
		Return(nil)

	err = ClientAddBankcard(ctx, mockSaver, mockEncryptor, token, secretName, number, owner, exp, cvv, meta)
	require.NoError(t, err)
}

func TestClientAddText(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSaver := NewMockClientSaver(ctrl)
	mockEncryptor := NewMockEncryptor(ctrl)

	ctx := context.Background()
	token := "token123"
	secretName := "textSecret"
	data := "some secret text"
	meta := "meta info"

	expectedPayload := models.TextPayload{
		Data: data,
		Meta: &meta,
	}
	plaintext, err := json.Marshal(expectedPayload)
	require.NoError(t, err)

	encrypted := models.SecretEncrypted{
		Ciphertext: []byte("encryptedText"),
		AESKeyEnc:  []byte("encryptedKey"),
	}

	mockEncryptor.EXPECT().
		Encrypt(plaintext).
		Return(&encrypted, nil)

	mockSaver.EXPECT().
		Save(ctx, token, secretName, models.SecretTypeText, encrypted.Ciphertext, encrypted.AESKeyEnc).
		Return(nil)

	err = ClientAddText(ctx, mockSaver, mockEncryptor, token, secretName, data, meta)
	require.NoError(t, err)
}

func TestClientAddBinary(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSaver := NewMockClientSaver(ctrl)
	mockEncryptor := NewMockEncryptor(ctrl)

	ctx := context.Background()
	token := "token123"
	secretName := "binarySecret"
	rawData := []byte{0x1, 0x2, 0x3}
	data := base64.StdEncoding.EncodeToString(rawData)
	meta := "binary meta"

	expectedPayload := models.BinaryPayload{
		Data: rawData,
		Meta: &meta,
	}
	plaintext, err := json.Marshal(expectedPayload)
	require.NoError(t, err)

	encrypted := models.SecretEncrypted{
		Ciphertext: []byte("encryptedBinary"),
		AESKeyEnc:  []byte("encryptedKey"),
	}

	mockEncryptor.EXPECT().
		Encrypt(plaintext).
		Return(&encrypted, nil)

	mockSaver.EXPECT().
		Save(ctx, token, secretName, models.SecretTypeBinary, encrypted.Ciphertext, encrypted.AESKeyEnc).
		Return(nil)

	err = ClientAddBinary(ctx, mockSaver, mockEncryptor, token, secretName, data, meta)
	require.NoError(t, err)
}

func TestClientAddUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSaver := NewMockClientSaver(ctrl)
	mockEncryptor := NewMockEncryptor(ctrl)

	ctx := context.Background()
	token := "token123"
	secretName := "userSecret"
	username := "user1"
	password := "pass1"
	meta := "user meta"

	expectedPayload := models.UserPayload{
		Username: username,
		Password: password,
		Meta:     &meta,
	}
	plaintext, err := json.Marshal(expectedPayload)
	require.NoError(t, err)

	encrypted := models.SecretEncrypted{
		Ciphertext: []byte("encryptedUser"),
		AESKeyEnc:  []byte("encryptedKey"),
	}

	mockEncryptor.EXPECT().
		Encrypt(plaintext).
		Return(&encrypted, nil)

	mockSaver.EXPECT().
		Save(ctx, token, secretName, models.SecretTypeUser, encrypted.Ciphertext, encrypted.AESKeyEnc).
		Return(nil)

	err = ClientAddUser(ctx, mockSaver, mockEncryptor, token, secretName, username, password, meta)
	require.NoError(t, err)
}

func TestClientListSecrets(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	token := "token123"

	tests := []struct {
		name        string
		mockSetup   func(*MockServerLister, *MockDecryptor)
		expectedOut string
		expectedErr string
	}{
		{
			name: "Bankcard secret",
			mockSetup: func(l *MockServerLister, d *MockDecryptor) {
				secrets := []*models.Secret{
					{
						SecretName: "card1",
						SecretType: models.SecretTypeBankCard,
						Ciphertext: []byte("cipher1"),
						AESKeyEnc:  []byte("key1"),
					},
				}
				l.EXPECT().List(ctx, token).Return(secrets, nil)
				d.EXPECT().Decrypt(gomock.Any()).DoAndReturn(func(secret *models.SecretEncrypted) ([]byte, error) {
					meta := "meta1"
					bankcard := models.BankcardPayload{
						Number: "1234567890",
						Owner:  "Alice",
						Exp:    "12/25",
						CVV:    "123",
						Meta:   &meta,
					}
					return json.Marshal(bankcard)
				}).AnyTimes()
			},
			expectedOut: `{
  "number": "1234567890",
  "owner": "Alice",
  "exp": "12/25",
  "cvv": "123",
  "meta": "meta1"
}`,
		},
		{
			name: "Text secret",
			mockSetup: func(l *MockServerLister, d *MockDecryptor) {
				secrets := []*models.Secret{
					{
						SecretName: "text1",
						SecretType: models.SecretTypeText,
						Ciphertext: []byte("cipher2"),
						AESKeyEnc:  []byte("key2"),
					},
				}
				l.EXPECT().List(ctx, token).Return(secrets, nil)
				d.EXPECT().Decrypt(gomock.Any()).DoAndReturn(func(secret *models.SecretEncrypted) ([]byte, error) {
					meta := "meta2"
					text := models.TextPayload{
						Data: "Hello, World!",
						Meta: &meta,
					}
					return json.Marshal(text)
				}).AnyTimes()
			},
			expectedOut: `{
  "data": "Hello, World!",
  "meta": "meta2"
}`,
		},
		{
			name: "Binary secret",
			mockSetup: func(l *MockServerLister, d *MockDecryptor) {
				secrets := []*models.Secret{
					{
						SecretName: "bin1",
						SecretType: models.SecretTypeBinary,
						Ciphertext: []byte("cipher3"),
						AESKeyEnc:  []byte("key3"),
					},
				}
				l.EXPECT().List(ctx, token).Return(secrets, nil)
				d.EXPECT().Decrypt(gomock.Any()).DoAndReturn(func(secret *models.SecretEncrypted) ([]byte, error) {
					meta := "meta3"
					binary := models.BinaryPayload{
						Data: []byte{0x01, 0x02, 0x03},
						Meta: &meta,
					}
					return json.Marshal(binary)
				}).AnyTimes()
			},
			expectedOut: `{
  "data": "AQID",
  "meta": "meta3"
}`,
		},
		{
			name: "User secret",
			mockSetup: func(l *MockServerLister, d *MockDecryptor) {
				secrets := []*models.Secret{
					{
						SecretName: "user1",
						SecretType: models.SecretTypeUser,
						Ciphertext: []byte("cipher4"),
						AESKeyEnc:  []byte("key4"),
					},
				}
				l.EXPECT().List(ctx, token).Return(secrets, nil)
				d.EXPECT().Decrypt(gomock.Any()).DoAndReturn(func(secret *models.SecretEncrypted) ([]byte, error) {
					meta := "meta4"
					user := models.UserPayload{
						Username: "bob",
						Password: "secret",
						Meta:     &meta,
					}
					return json.Marshal(user)
				}).AnyTimes()
			},
			expectedOut: `{
  "username": "bob",
  "password": "secret",
  "meta": "meta4"
}`,
		},
		{
			name: "Unknown secret type",
			mockSetup: func(l *MockServerLister, d *MockDecryptor) {
				secrets := []*models.Secret{
					{
						SecretName: "unknown1",
						SecretType: "unknownType",
						Ciphertext: []byte("cipher5"),
						AESKeyEnc:  []byte("key5"),
					},
				}
				l.EXPECT().List(ctx, token).Return(secrets, nil)
				d.EXPECT().Decrypt(gomock.Any()).Return([]byte{}, nil).AnyTimes()
			},
			expectedOut: "Unknown secret type: unknownType",
		},
		{
			name: "Decrypt error",
			mockSetup: func(l *MockServerLister, d *MockDecryptor) {
				secrets := []*models.Secret{
					{
						SecretName: "faildecrypt",
						SecretType: models.SecretTypeText,
						Ciphertext: []byte("cipher6"),
						AESKeyEnc:  []byte("key6"),
					},
				}
				l.EXPECT().List(ctx, token).Return(secrets, nil)
				d.EXPECT().Decrypt(gomock.Any()).Return(nil, errors.New("decryption error"))
			},
			expectedErr: "failed to decrypt secret faildecrypt",
		},
		{
			name: "List error",
			mockSetup: func(l *MockServerLister, d *MockDecryptor) {
				l.EXPECT().List(ctx, token).Return(nil, errors.New("list error"))
			},
			expectedErr: "list error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLister := NewMockServerLister(ctrl)
			mockDecryptor := NewMockDecryptor(ctrl)

			tt.mockSetup(mockLister, mockDecryptor)

			out, err := ClientListSecrets(ctx, mockLister, mockDecryptor, token)

			if tt.expectedErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedErr)
				return
			}
			require.NoError(t, err)

			out = strings.ReplaceAll(out, "\r\n", "\n") // Normalize newlines for Windows

			require.Equal(t, strings.TrimSpace(tt.expectedOut), strings.TrimSpace(out))
		})
	}
}

// helper to create a sample secret
func makeSecret(name, secretType string, updatedAt time.Time) *models.Secret {
	return &models.Secret{
		SecretName: name,
		SecretType: secretType,
		UpdatedAt:  updatedAt,
		Ciphertext: []byte(`{"foo":"bar"}`),
		AESKeyEnc:  []byte(`key`),
	}
}

func TestClientSyncClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	owner := "owner1"

	cl := NewMockClientLister(ctrl)
	sg := NewMockServerGetter(ctrl)
	ss := NewMockServerSaver(ctrl)

	clientSecret := makeSecret("secretA", "typeA", time.Now())
	serverSecret := makeSecret("secretA", "typeA", time.Now().Add(-time.Hour))

	// Client secret is newer, so Save should be called
	cl.EXPECT().List(ctx, owner).Return([]*models.Secret{clientSecret}, nil)
	sg.EXPECT().Get(ctx, owner, clientSecret.SecretType, clientSecret.SecretName).Return(serverSecret, nil)
	ss.EXPECT().Save(ctx, owner, clientSecret.SecretName, clientSecret.SecretType, clientSecret.Ciphertext, clientSecret.AESKeyEnc).Return(nil)

	err := ClientSyncClient(ctx, cl, sg, ss, owner)
	require.NoError(t, err)
}

func TestClientSyncInteractive(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	owner := "owner1"

	cl := NewMockClientLister(ctrl)
	sg := NewMockServerGetter(ctrl)
	ss := NewMockServerSaver(ctrl)
	d := NewMockDecryptor(ctrl)

	now := time.Now()

	// Secret missing on server, must Save
	clientSecretMissingOnServer := makeSecret("secretX", "typeX", now)
	// Secret exists with conflict
	clientSecretConflict := makeSecret("secretY", "typeY", now)
	serverSecretConflict := makeSecret("secretY", "typeY", now.Add(-time.Hour))

	cl.EXPECT().List(ctx, owner).Return([]*models.Secret{
		clientSecretMissingOnServer,
		clientSecretConflict,
	}, nil)

	// Get calls for both secrets
	sg.EXPECT().Get(ctx, owner, clientSecretMissingOnServer.SecretType, clientSecretMissingOnServer.SecretName).Return(nil, nil)
	sg.EXPECT().Get(ctx, owner, clientSecretConflict.SecretType, clientSecretConflict.SecretName).Return(serverSecretConflict, nil)

	// Decrypt calls for conflict secret
	gomock.InOrder(
		// Save for missing secret first
		ss.EXPECT().Save(
			ctx, owner,
			clientSecretMissingOnServer.SecretName,
			clientSecretMissingOnServer.SecretType,
			clientSecretMissingOnServer.Ciphertext,
			clientSecretMissingOnServer.AESKeyEnc,
		).Return(nil),

		// Decrypt client conflict secret
		d.EXPECT().Decrypt(gomock.AssignableToTypeOf(&models.SecretEncrypted{})).Return(clientSecretConflict.Ciphertext, nil),
		// Decrypt server conflict secret
		d.EXPECT().Decrypt(gomock.AssignableToTypeOf(&models.SecretEncrypted{})).Return(serverSecretConflict.Ciphertext, nil),

		// Save for conflict secret when client chooses version "1"
		ss.EXPECT().Save(
			ctx, owner,
			clientSecretConflict.SecretName,
			clientSecretConflict.SecretType,
			clientSecretConflict.Ciphertext,
			clientSecretConflict.AESKeyEnc,
		).Return(nil),
	)

	// Input simulates choosing client version "1"
	input := "1\n"
	reader := strings.NewReader(input)

	err := ClientSyncInteractive(ctx, cl, sg, ss, d, owner, reader)
	require.NoError(t, err)

	// Test invalid input returns error (optional: separate test)
}
