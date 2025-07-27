package usecases

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/gophkeeper/inernal/models"
	"github.com/stretchr/testify/assert"
)

func TestClientListUsecase_Execute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLister := NewMockServerLister(ctrl) // <-- fix here
	mockDecryptor := NewMockDecryptor(ctrl)

	app := NewClientListUsecase(mockLister, mockDecryptor)
	ctx := context.Background()

	mustMarshal := func(v interface{}) []byte {
		b, err := json.Marshal(v)
		assert.NoError(t, err)
		return b
	}

	bankcard := models.Bankcard{Number: "1234567890123456", Owner: "John Doe", Exp: "12/25", CVV: "123"}
	user := models.User{Username: "johndoe", Password: "secret123"}
	text := models.Text{Data: "This is a secret text"}
	binary := models.Binary{Data: []byte{0x01, 0x02, 0x03, 0x04}}

	tests := []struct {
		name          string
		req           *models.SecretListRequest
		listerResult  []*models.SecretDB
		listerError   error
		decryptErrors map[string]error
		decryptData   map[string][]byte
		expectedBank  []models.Bankcard
		expectedUser  []models.User
		expectedText  []models.Text
		expectedBin   []models.Binary
		expectedErr   string
	}{
		{
			name: "successful listing and decrypt",
			req:  &models.SecretListRequest{Token: "owner1"},
			listerResult: []*models.SecretDB{
				{SecretName: "card1", SecretType: models.SecretTypeBankCard, Ciphertext: []byte("cipher1"), AESKeyEnc: []byte("key1")},
				{SecretName: "user1", SecretType: models.SecretTypeUser, Ciphertext: []byte("cipher2"), AESKeyEnc: []byte("key2")},
				{SecretName: "text1", SecretType: models.SecretTypeText, Ciphertext: []byte("cipher3"), AESKeyEnc: []byte("key3")},
				{SecretName: "bin1", SecretType: models.SecretTypeBinary, Ciphertext: []byte("cipher4"), AESKeyEnc: []byte("key4")},
			},
			decryptData: map[string][]byte{
				"card1": mustMarshal(bankcard),
				"user1": mustMarshal(user),
				"text1": mustMarshal(text),
				"bin1":  mustMarshal(binary),
			},
			expectedBank: []models.Bankcard{bankcard},
			expectedUser: []models.User{user},
			expectedText: []models.Text{text},
			expectedBin:  []models.Binary{binary},
		},
		// ... rest unchanged ...
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.req != nil {
				// Expect List called with the exact request
				mockLister.EXPECT().List(ctx, gomock.Eq(tt.req)).Return(tt.listerResult, tt.listerError)
			}

			if tt.listerResult != nil {
				for _, secret := range tt.listerResult {
					if err, ok := tt.decryptErrors[secret.SecretName]; ok {
						mockDecryptor.EXPECT().Decrypt(gomock.Any()).Return(nil, err)
					} else {
						mockDecryptor.EXPECT().Decrypt(gomock.Any()).Return(tt.decryptData[secret.SecretName], nil)
					}
				}
			}

			bankcards, users, texts, binaries, err := app.Execute(ctx, tt.req)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
				assert.Nil(t, bankcards)
				assert.Nil(t, users)
				assert.Nil(t, texts)
				assert.Nil(t, binaries)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBank, bankcards)
				assert.Equal(t, tt.expectedUser, users)
				assert.Equal(t, tt.expectedText, texts)
				assert.Equal(t, tt.expectedBin, binaries)
			}
		})
	}
}
