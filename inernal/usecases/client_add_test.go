package usecases

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/gophkeeper/inernal/models"
	"github.com/stretchr/testify/assert"
)

func TestClientTextAddUsecase_Execute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSaver := NewMockClientSaver(ctrl)
	mockEncryptor := NewMockEncryptor(ctrl)

	app := NewClientTextAddUsecase(mockSaver, mockEncryptor)

	meta := func(s string) *string { return &s }

	tests := []struct {
		name          string
		req           *models.TextAddRequest
		encryptResult *models.Encrypted
		encryptErr    error
		saveErr       error
		wantErr       bool
	}{
		{
			name: "success with meta",
			req: &models.TextAddRequest{
				SecretName: "note1",
				Token:      "user1", // <-- use Token field here, not SecretOwner
				Data:       "Hello World",
				Meta:       meta("text note"),
			},
			encryptResult: &models.Encrypted{
				Ciphertext: []byte("cipher"),
				AESKeyEnc:  []byte("key"),
			},
			wantErr: false,
		},
		{
			name: "success with nil meta",
			req: &models.TextAddRequest{
				SecretName: "note2",
				Token:      "user2",
				Data:       "No Meta",
				Meta:       nil,
			},
			encryptResult: &models.Encrypted{
				Ciphertext: []byte("cipher2"),
				AESKeyEnc:  []byte("key2"),
			},
			wantErr: false,
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
		},
		{
			name: "encryption error",
			req: &models.TextAddRequest{
				SecretName: "note3",
				Token:      "user3",
				Data:       "Fail Data",
				Meta:       meta("fail"),
			},
			encryptErr: errors.New("encryption failed"),
			wantErr:    true,
		},
		{
			name: "save error",
			req: &models.TextAddRequest{
				SecretName: "note4",
				Token:      "user4",
				Data:       "Save Error",
				Meta:       nil,
			},
			encryptResult: &models.Encrypted{
				Ciphertext: []byte("abc"),
				AESKeyEnc:  []byte("xyz"),
			},
			saveErr: errors.New("save failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.req != nil {
				expectedText := models.Text{
					Data: tt.req.Data,
					Meta: tt.req.Meta,
				}

				plaintext, err := json.Marshal(expectedText)
				assert.NoError(t, err)

				mockEncryptor.EXPECT().
					Encrypt(plaintext).
					Return(tt.encryptResult, tt.encryptErr).
					Times(1)

				if tt.encryptErr == nil && tt.encryptResult != nil {
					expectedSecret := &models.SecretDB{
						SecretName:  tt.req.SecretName,
						SecretType:  models.SecretTypeText,
						SecretOwner: tt.req.Token, // Use Token here as per your main code
						Ciphertext:  tt.encryptResult.Ciphertext,
						AESKeyEnc:   tt.encryptResult.AESKeyEnc,
					}
					mockSaver.EXPECT().
						Save(context.Background(), expectedSecret).
						Return(tt.saveErr).
						Times(1)
				}
			}

			err := app.Execute(context.Background(), tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
