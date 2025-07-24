package client

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sbilibin2017/gophkeeper/internal/cryptor"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

func TestSecretReader_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name       string
		secretType string
		plaintext  []byte
		expected   map[string]interface{}
	}{
		{
			name:       "Successful_BankCard",
			secretType: models.SecretTypeBankCard,
			plaintext: mustMarshalJSON(t, models.BankCardPayload{
				Number: "1234",
				Owner:  "John Doe",
				Exp:    "12/25",
				CVV:    "123",
				Meta:   nil,
			}),
			expected: map[string]interface{}{
				"number":       "1234",
				"owner":        "John Doe",
				"exp":          "12/25",
				"cvv":          "123",
				"secret_name":  "",
				"secret_owner": "",
			},
		},
		{
			name:       "Successful_Binary",
			secretType: models.SecretTypeBinary,
			plaintext: mustMarshalJSON(t, models.BinaryPayload{
				FilePath: "/tmp/file.bin",
				Data:     []byte("hello"),
				Meta:     nil,
			}),
			expected: map[string]interface{}{
				"file_path":    "/tmp/file.bin",
				"data":         "aGVsbG8=",
				"secret_name":  "",
				"secret_owner": "",
			},
		},
		{
			name:       "Successful_Text",
			secretType: models.SecretTypeText,
			plaintext: mustMarshalJSON(t, models.TextPayload{
				Data: "some secret text",
				Meta: nil,
			}),
			expected: map[string]interface{}{
				"data":         "some secret text",
				"secret_name":  "",
				"secret_owner": "",
			},
		},
		{
			name:       "Successful_User",
			secretType: models.SecretTypeUser,
			plaintext: mustMarshalJSON(t, models.UserPayload{
				Login:    "user123",
				Password: "pass123",
				Meta:     nil,
			}),
			expected: map[string]interface{}{
				"login":        "user123",
				"password":     "pass123",
				"secret_name":  "",
				"secret_owner": "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGetter := NewMockGetter(ctrl)
			mockDecryptor := NewMockDecryptor(ctrl)

			encryptedSecret := &models.EncryptedSecret{
				Ciphertext: tt.plaintext,
				AESKeyEnc:  []byte("somekey"),
				SecretType: tt.secretType,
			}

			// Setup mocks
			mockGetter.EXPECT().
				Get(gomock.Any(), gomock.Any()).
				Return(encryptedSecret, nil).
				Times(1)

			mockDecryptor.EXPECT().
				Decrypt(gomock.Any()).
				Return(tt.plaintext, nil).
				Times(1)

			sr := &SecretReader{
				getter:    mockGetter,
				decryptor: mockDecryptor,
			}

			secretJSON, err := sr.Get(context.Background(), "any-secret")
			require.NoError(t, err)
			require.NotNil(t, secretJSON)

			var actual map[string]interface{}
			err = json.Unmarshal([]byte(*secretJSON), &actual)
			require.NoError(t, err)

			// Remove "updated_at" if present
			delete(actual, "updated_at")

			assert.Equal(t, tt.expected, actual)
		})
	}
}

func mustMarshalJSON(t *testing.T, v interface{}) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}

func TestSecretReader_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLister := NewMockLister(ctrl)
	mockDecryptor := NewMockDecryptor(ctrl)

	reader := SecretReader{
		lister:    mockLister,
		decryptor: mockDecryptor,
	}

	tests := []struct {
		name             string
		listerResp       []*models.EncryptedSecret
		listerErr        error
		decryptResponses map[string][]byte // map SecretType to decrypted JSON bytes
		decryptErr       error
		expectedCount    int
		expectedErr      bool
	}{
		{
			name: "Successful list including Binary and User",
			listerResp: []*models.EncryptedSecret{
				{
					SecretType: models.SecretTypeBankCard,
					Ciphertext: []byte("c1"),
					AESKeyEnc:  []byte("k1"),
				},
				{
					SecretType: models.SecretTypeBinary,
					Ciphertext: []byte("c2"),
					AESKeyEnc:  []byte("k2"),
				},
				{
					SecretType: models.SecretTypeUser,
					Ciphertext: []byte("c3"),
					AESKeyEnc:  []byte("k3"),
				},
				{
					SecretType: models.SecretTypeText,
					Ciphertext: []byte("c4"),
					AESKeyEnc:  []byte("k4"),
				},
				{
					SecretType: "unknown",
					Ciphertext: []byte("c5"),
					AESKeyEnc:  []byte("k5"),
				},
			},
			decryptResponses: map[string][]byte{
				models.SecretTypeBankCard: []byte(`{"number":"1234","owner":"John","exp":"12/25","cvv":"123"}`),
				models.SecretTypeBinary:   []byte(`{"file_path":"/tmp/file.bin","data":"aGVsbG8="}`),
				models.SecretTypeUser:     []byte(`{"login":"user123","password":"pass123"}`),
				models.SecretTypeText:     []byte(`{"data":"some secret text"}`),
			},
			expectedCount: 4,
		},
		{
			name:        "Lister returns error",
			listerErr:   errors.New("list failed"),
			expectedErr: true,
		},
		{
			name: "Decrypt returns error",
			listerResp: []*models.EncryptedSecret{
				{
					SecretType: models.SecretTypeText,
					Ciphertext: []byte("c"),
					AESKeyEnc:  []byte("k"),
				},
			},
			decryptErr:  errors.New("decrypt failed"),
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			mockLister.EXPECT().
				List(ctx).
				Return(tt.listerResp, tt.listerErr).
				Times(1)

			if tt.listerErr == nil {
				for _, es := range tt.listerResp {
					if tt.decryptErr != nil {
						mockDecryptor.EXPECT().
							Decrypt(gomock.Any()).
							Return(nil, tt.decryptErr).
							Times(1)
						break
					}
					resp, ok := tt.decryptResponses[es.SecretType]
					if !ok {
						// For unknown types or if no response provided, return empty JSON object
						resp = []byte(`{}`)
					}
					mockDecryptor.EXPECT().
						Decrypt(gomock.Any()).
						Return(resp, nil).
						Times(1)
				}
			}

			result, err := reader.List(ctx)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				// Only count known secret types (exclude "unknown")
				expectedCount := 0
				for _, s := range tt.listerResp {
					switch s.SecretType {
					case models.SecretTypeBankCard, models.SecretTypeBinary, models.SecretTypeUser, models.SecretTypeText:
						expectedCount++
					}
				}
				assert.Len(t, result, expectedCount)
			}
		})
	}
}
func TestSecretWriter_AddBankCard(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSaver := NewMockSaver(ctrl)
	mockEncryptor := NewMockEncryptor(ctrl)

	writer := SecretWriter{
		saver:     mockSaver,
		encryptor: mockEncryptor,
	}

	payload := models.BankCardPayload{
		Number: "1234",
		Owner:  "John Doe",
	}

	tests := []struct {
		name        string
		encryptResp *cryptor.Encrypted
		encryptErr  error
		saveErr     error
		expectedErr bool
	}{
		{
			name: "Success",
			encryptResp: &cryptor.Encrypted{
				Ciphertext: []byte("cipher"),
				AESKeyEnc:  []byte("key"),
			},
		},
		{
			name:        "Encrypt error",
			encryptErr:  errors.New("encrypt failed"),
			expectedErr: true,
		},
		{
			name:        "Save error",
			encryptResp: &cryptor.Encrypted{Ciphertext: []byte("cipher"), AESKeyEnc: []byte("key")},
			saveErr:     errors.New("save failed"),
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			mockEncryptor.EXPECT().
				Encrypt(gomock.Any()).
				Return(tt.encryptResp, tt.encryptErr).
				Times(1)

			if tt.encryptErr == nil {
				mockSaver.EXPECT().
					Save(ctx, gomock.AssignableToTypeOf(&models.EncryptedSecret{})).
					Return(tt.saveErr).
					Times(1)
			}

			err := writer.AddBankCard(ctx, "secret1", payload)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSecretWriter_AddBinary(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSaver := NewMockSaver(ctrl)
	mockEncryptor := NewMockEncryptor(ctrl)

	writer := SecretWriter{
		saver:     mockSaver,
		encryptor: mockEncryptor,
	}

	payload := models.BinaryPayload{
		Data: []byte{0x01, 0x02},
	}

	tests := []struct {
		name        string
		encryptResp *cryptor.Encrypted
		encryptErr  error
		saveErr     error
		expectedErr bool
	}{
		{
			name: "Success",
			encryptResp: &cryptor.Encrypted{
				Ciphertext: []byte("cipher"),
				AESKeyEnc:  []byte("key"),
			},
		},
		{
			name:        "Encrypt error",
			encryptErr:  errors.New("encrypt failed"),
			expectedErr: true,
		},
		{
			name:        "Save error",
			encryptResp: &cryptor.Encrypted{Ciphertext: []byte("cipher"), AESKeyEnc: []byte("key")},
			saveErr:     errors.New("save failed"),
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			mockEncryptor.EXPECT().
				Encrypt(gomock.Any()).
				Return(tt.encryptResp, tt.encryptErr).
				Times(1)

			if tt.encryptErr == nil {
				mockSaver.EXPECT().
					Save(ctx, gomock.AssignableToTypeOf(&models.EncryptedSecret{})).
					Return(tt.saveErr).
					Times(1)
			}

			err := writer.AddBinary(ctx, "secret1", payload)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSecretWriter_AddText(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSaver := NewMockSaver(ctrl)
	mockEncryptor := NewMockEncryptor(ctrl)

	writer := SecretWriter{
		saver:     mockSaver,
		encryptor: mockEncryptor,
	}

	payload := models.TextPayload{
		Data: "some text",
	}

	tests := []struct {
		name        string
		encryptResp *cryptor.Encrypted
		encryptErr  error
		saveErr     error
		expectedErr bool
	}{
		{
			name: "Success",
			encryptResp: &cryptor.Encrypted{
				Ciphertext: []byte("cipher"),
				AESKeyEnc:  []byte("key"),
			},
		},
		{
			name:        "Encrypt error",
			encryptErr:  errors.New("encrypt failed"),
			expectedErr: true,
		},
		{
			name:        "Save error",
			encryptResp: &cryptor.Encrypted{Ciphertext: []byte("cipher"), AESKeyEnc: []byte("key")},
			saveErr:     errors.New("save failed"),
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			mockEncryptor.EXPECT().
				Encrypt(gomock.Any()).
				Return(tt.encryptResp, tt.encryptErr).
				Times(1)

			if tt.encryptErr == nil {
				mockSaver.EXPECT().
					Save(ctx, gomock.AssignableToTypeOf(&models.EncryptedSecret{})).
					Return(tt.saveErr).
					Times(1)
			}

			err := writer.AddText(ctx, "secret1", payload)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSecretWriter_AddUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSaver := NewMockSaver(ctrl)
	mockEncryptor := NewMockEncryptor(ctrl)

	writer := SecretWriter{
		saver:     mockSaver,
		encryptor: mockEncryptor,
	}

	payload := models.UserPayload{
		Login:    "user1",
		Password: "pass1",
	}

	tests := []struct {
		name        string
		encryptResp *cryptor.Encrypted
		encryptErr  error
		saveErr     error
		expectedErr bool
	}{
		{
			name: "Success",
			encryptResp: &cryptor.Encrypted{
				Ciphertext: []byte("cipher"),
				AESKeyEnc:  []byte("key"),
			},
		},
		{
			name:        "Encrypt error",
			encryptErr:  errors.New("encrypt failed"),
			expectedErr: true,
		},
		{
			name:        "Save error",
			encryptResp: &cryptor.Encrypted{Ciphertext: []byte("cipher"), AESKeyEnc: []byte("key")},
			saveErr:     errors.New("save failed"),
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			mockEncryptor.EXPECT().
				Encrypt(gomock.Any()).
				Return(tt.encryptResp, tt.encryptErr).
				Times(1)

			if tt.encryptErr == nil {
				mockSaver.EXPECT().
					Save(ctx, gomock.AssignableToTypeOf(&models.EncryptedSecret{})).
					Return(tt.saveErr).
					Times(1)
			}

			err := writer.AddUser(ctx, "secret1", payload)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSecretWriter_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDeleter := NewMockDeleter(ctrl)

	writer := SecretWriter{
		deleter: mockDeleter,
	}

	tests := []struct {
		name        string
		deleteErr   error
		expectedErr bool
	}{
		{
			name:      "Success",
			deleteErr: nil,
		},
		{
			name:        "Delete error",
			deleteErr:   errors.New("delete failed"),
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockDeleter.EXPECT().
				Delete(ctx, "secret1").
				Return(tt.deleteErr).
				Times(1)

			err := writer.Delete(ctx, "secret1")

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
