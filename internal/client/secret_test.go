package client

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/gophkeeper/internal/cryptor"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestSecretReader_Get(t *testing.T) {
	ctx := context.Background()
	secretName := "mySecret"

	// Helper to build encryptedSecret and encryptedData for each test
	makeEncryptedSecret := func(secretType string) (*models.EncryptedSecret, *cryptor.Encrypted) {
		encryptedSecret := &models.EncryptedSecret{
			SecretType: secretType,
			Ciphertext: []byte("ciphertext"),
			AESKeyEnc:  []byte("keyenc"),
		}
		encryptedData := &cryptor.Encrypted{
			Ciphertext: encryptedSecret.Ciphertext,
			AESKeyEnc:  encryptedSecret.AESKeyEnc,
		}
		return encryptedSecret, encryptedData
	}

	t.Run("success bank card secret", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockReader := NewMockReader(ctrl)
		mockDecryptor := NewMockDecryptor(ctrl)
		secretReader := NewSecretReader(mockReader, mockDecryptor)

		payload := models.BankCard{
			Number: "1234567890",
			CVV:    "123",
		}
		plaintext, _ := json.Marshal(payload)
		encryptedSecret, encryptedData := makeEncryptedSecret(models.SecretTypeBankCard)

		mockReader.EXPECT().Get(ctx, secretName).Return(encryptedSecret, nil)
		mockDecryptor.EXPECT().Decrypt(encryptedData).Return(plaintext, nil)

		result, err := secretReader.Get(ctx, secretName)
		assert.NoError(t, err)

		expectedJSON, _ := json.MarshalIndent(payload, "", "  ")
		assert.Equal(t, string(expectedJSON), *result)
	})

	t.Run("success binary secret", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockReader := NewMockReader(ctrl)
		mockDecryptor := NewMockDecryptor(ctrl)
		secretReader := NewSecretReader(mockReader, mockDecryptor)

		payload := models.Binary{
			FileName: "file.txt",
			Data:     []byte{1, 2, 3},
		}
		plaintext, _ := json.Marshal(payload)
		encryptedSecret, encryptedData := makeEncryptedSecret(models.SecretTypeBinary)

		mockReader.EXPECT().Get(ctx, secretName).Return(encryptedSecret, nil)
		mockDecryptor.EXPECT().Decrypt(encryptedData).Return(plaintext, nil)

		result, err := secretReader.Get(ctx, secretName)
		assert.NoError(t, err)

		expectedJSON, _ := json.MarshalIndent(payload, "", "  ")
		assert.Equal(t, string(expectedJSON), *result)
	})

	t.Run("success text secret", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockReader := NewMockReader(ctrl)
		mockDecryptor := NewMockDecryptor(ctrl)
		secretReader := NewSecretReader(mockReader, mockDecryptor)

		payload := models.Text{
			Data: "some text data",
		}
		plaintext, _ := json.Marshal(payload)
		encryptedSecret, encryptedData := makeEncryptedSecret(models.SecretTypeText)

		mockReader.EXPECT().Get(ctx, secretName).Return(encryptedSecret, nil)
		mockDecryptor.EXPECT().Decrypt(encryptedData).Return(plaintext, nil)

		result, err := secretReader.Get(ctx, secretName)
		assert.NoError(t, err)

		expectedJSON, _ := json.MarshalIndent(payload, "", "  ")
		assert.Equal(t, string(expectedJSON), *result)
	})

	t.Run("success user secret", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockReader := NewMockReader(ctrl)
		mockDecryptor := NewMockDecryptor(ctrl)
		secretReader := NewSecretReader(mockReader, mockDecryptor)

		payload := models.User{
			Login:    "user1",
			Password: "pass1",
		}
		plaintext, _ := json.Marshal(payload)
		encryptedSecret, encryptedData := makeEncryptedSecret(models.SecretTypeUser)

		mockReader.EXPECT().Get(ctx, secretName).Return(encryptedSecret, nil)
		mockDecryptor.EXPECT().Decrypt(encryptedData).Return(plaintext, nil)

		result, err := secretReader.Get(ctx, secretName)
		assert.NoError(t, err)

		expectedJSON, _ := json.MarshalIndent(payload, "", "  ")
		assert.Equal(t, string(expectedJSON), *result)
	})

	t.Run("secret not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockReader := NewMockReader(ctrl)
		mockDecryptor := NewMockDecryptor(ctrl)
		secretReader := NewSecretReader(mockReader, mockDecryptor)

		mockReader.EXPECT().Get(ctx, secretName).Return(nil, nil)

		result, err := secretReader.Get(ctx, secretName)
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("reader returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockReader := NewMockReader(ctrl)
		mockDecryptor := NewMockDecryptor(ctrl)
		secretReader := NewSecretReader(mockReader, mockDecryptor)

		mockReader.EXPECT().Get(ctx, secretName).Return(nil, errors.New("db error"))

		result, err := secretReader.Get(ctx, secretName)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("decryptor returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockReader := NewMockReader(ctrl)
		mockDecryptor := NewMockDecryptor(ctrl)
		secretReader := NewSecretReader(mockReader, mockDecryptor)

		encryptedSecret, encryptedData := makeEncryptedSecret(models.SecretTypeBankCard)

		mockReader.EXPECT().Get(ctx, secretName).Return(encryptedSecret, nil)
		mockDecryptor.EXPECT().Decrypt(encryptedData).Return(nil, errors.New("decrypt error"))

		result, err := secretReader.Get(ctx, secretName)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("unmarshal error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockReader := NewMockReader(ctrl)
		mockDecryptor := NewMockDecryptor(ctrl)
		secretReader := NewSecretReader(mockReader, mockDecryptor)

		encryptedSecret, encryptedData := makeEncryptedSecret(models.SecretTypeBankCard)

		mockReader.EXPECT().Get(ctx, secretName).Return(encryptedSecret, nil)
		mockDecryptor.EXPECT().Decrypt(encryptedData).Return([]byte("invalid json"), nil)

		result, err := secretReader.Get(ctx, secretName)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("unknown secret type returns nil", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockReader := NewMockReader(ctrl)
		mockDecryptor := NewMockDecryptor(ctrl)
		secretReader := NewSecretReader(mockReader, mockDecryptor)

		unknownSecret, encryptedData := makeEncryptedSecret("unknown-type")
		mockReader.EXPECT().Get(ctx, secretName).Return(unknownSecret, nil)
		// Expect Decrypt call even for unknown type (returns dummy valid JSON)
		mockDecryptor.EXPECT().Decrypt(encryptedData).Return([]byte("{}"), nil)

		result, err := secretReader.Get(ctx, secretName)
		assert.NoError(t, err)
		assert.Nil(t, result)
	})
}
func TestSecretReader_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReader := NewMockReader(ctrl)
	mockDecryptor := NewMockDecryptor(ctrl)

	secretReader := NewSecretReader(mockReader, mockDecryptor)

	ctx := context.Background()

	// Helper to create encrypted secret and plaintext for a secret type + payload
	makeTestData := func(secretType string, payload interface{}) (*models.EncryptedSecret, *cryptor.Encrypted, []byte) {
		plaintext, _ := json.Marshal(payload)
		encryptedSecret := &models.EncryptedSecret{
			SecretType: secretType,
			SecretName: "secret1",
			Ciphertext: []byte("ciphertext"),
			AESKeyEnc:  []byte("keyenc"),
		}
		encryptedData := &cryptor.Encrypted{
			Ciphertext: encryptedSecret.Ciphertext,
			AESKeyEnc:  encryptedSecret.AESKeyEnc,
		}
		return encryptedSecret, encryptedData, plaintext
	}

	t.Run("success list text secret", func(t *testing.T) {
		payload := models.Text{Data: "Hello world"}
		encryptedSecret, encryptedData, plaintext := makeTestData(models.SecretTypeText, payload)

		mockReader.EXPECT().List(ctx).Return([]*models.EncryptedSecret{encryptedSecret}, nil)
		mockDecryptor.EXPECT().Decrypt(gomock.Eq(encryptedData)).Return(plaintext, nil)

		result, err := secretReader.List(ctx)
		assert.NoError(t, err)

		expectedJSON, _ := json.MarshalIndent(payload, "", "  ")
		assert.Len(t, result, 1)
		assert.Equal(t, string(expectedJSON), result[0])
	})

	t.Run("success list bank card secret", func(t *testing.T) {
		payload := models.BankCard{Number: "1234567890", CVV: "123"}
		encryptedSecret, encryptedData, plaintext := makeTestData(models.SecretTypeBankCard, payload)

		mockReader.EXPECT().List(ctx).Return([]*models.EncryptedSecret{encryptedSecret}, nil)
		mockDecryptor.EXPECT().Decrypt(gomock.Eq(encryptedData)).Return(plaintext, nil)

		result, err := secretReader.List(ctx)
		assert.NoError(t, err)

		expectedJSON, _ := json.MarshalIndent(payload, "", "  ")
		assert.Len(t, result, 1)
		assert.Equal(t, string(expectedJSON), result[0])
	})

	t.Run("success list binary secret", func(t *testing.T) {
		payload := models.Binary{FileName: "file.txt", Data: []byte{1, 2, 3}}
		encryptedSecret, encryptedData, plaintext := makeTestData(models.SecretTypeBinary, payload)

		mockReader.EXPECT().List(ctx).Return([]*models.EncryptedSecret{encryptedSecret}, nil)
		mockDecryptor.EXPECT().Decrypt(gomock.Eq(encryptedData)).Return(plaintext, nil)

		result, err := secretReader.List(ctx)
		assert.NoError(t, err)

		expectedJSON, _ := json.MarshalIndent(payload, "", "  ")
		assert.Len(t, result, 1)
		assert.Equal(t, string(expectedJSON), result[0])
	})

	t.Run("success list user secret", func(t *testing.T) {
		payload := models.User{Login: "user1", Password: "pass1"}
		encryptedSecret, encryptedData, plaintext := makeTestData(models.SecretTypeUser, payload)

		mockReader.EXPECT().List(ctx).Return([]*models.EncryptedSecret{encryptedSecret}, nil)
		mockDecryptor.EXPECT().Decrypt(gomock.Eq(encryptedData)).Return(plaintext, nil)

		result, err := secretReader.List(ctx)
		assert.NoError(t, err)

		expectedJSON, _ := json.MarshalIndent(payload, "", "  ")
		assert.Len(t, result, 1)
		assert.Equal(t, string(expectedJSON), result[0])
	})

	// Existing error & edge cases
	t.Run("reader returns error", func(t *testing.T) {
		mockReader.EXPECT().List(ctx).Return(nil, errors.New("db error"))

		result, err := secretReader.List(ctx)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("decryptor returns error", func(t *testing.T) {
		payload := models.Text{Data: "fail decrypt"}
		encryptedSecret, encryptedData, _ := makeTestData(models.SecretTypeText, payload)

		mockReader.EXPECT().List(ctx).Return([]*models.EncryptedSecret{encryptedSecret}, nil)
		mockDecryptor.EXPECT().Decrypt(gomock.Eq(encryptedData)).Return(nil, errors.New("decrypt error"))

		result, err := secretReader.List(ctx)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("unmarshal error", func(t *testing.T) {
		encryptedSecret := &models.EncryptedSecret{
			SecretType: models.SecretTypeText,
			SecretName: "secret1",
			Ciphertext: []byte("ciphertext"),
			AESKeyEnc:  []byte("keyenc"),
		}
		encryptedData := &cryptor.Encrypted{
			Ciphertext: encryptedSecret.Ciphertext,
			AESKeyEnc:  encryptedSecret.AESKeyEnc,
		}

		mockReader.EXPECT().List(ctx).Return([]*models.EncryptedSecret{encryptedSecret}, nil)
		mockDecryptor.EXPECT().Decrypt(gomock.Eq(encryptedData)).Return([]byte("bad json"), nil)

		result, err := secretReader.List(ctx)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("unknown secret type ignored", func(t *testing.T) {
		unknownSecret := &models.EncryptedSecret{
			SecretType: "unknown-type",
			SecretName: "unknown",
			Ciphertext: []byte("ct"),
			AESKeyEnc:  []byte("key"),
		}

		mockReader.EXPECT().List(ctx).Return([]*models.EncryptedSecret{unknownSecret}, nil)
		// If decryptor called, return nil plaintext and no error (should be ignored anyway)
		mockDecryptor.EXPECT().Decrypt(gomock.Any()).AnyTimes().Return(nil, nil)

		result, err := secretReader.List(ctx)
		assert.NoError(t, err)
		assert.Len(t, result, 0)
	})
}

func TestSecretWriter_AddBinaryTextUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWriter := NewMockWriter(ctrl)
	mockEncryptor := NewMockEncryptor(ctrl)

	secretWriter := NewSecretWriter(mockWriter, mockEncryptor)

	ctx := context.Background()
	secretName := "mySecret"

	t.Run("AddBinary success", func(t *testing.T) {
		payload := models.BinaryPayload{
			FileName: "file.bin",
			Data:     []byte{1, 2, 3},
		}
		plaintext, _ := json.Marshal(payload)
		encrypted := &cryptor.Encrypted{
			Ciphertext: []byte("ciphertext"),
			AESKeyEnc:  []byte("keyenc"),
		}

		mockEncryptor.EXPECT().Encrypt(plaintext).Return(encrypted, nil)

		expectedSecret := &models.EncryptedSecret{
			SecretType: models.SecretTypeBinary,
			SecretName: secretName,
			Ciphertext: encrypted.Ciphertext,
			AESKeyEnc:  encrypted.AESKeyEnc,
		}

		mockWriter.EXPECT().Save(ctx, expectedSecret).Return(nil)

		err := secretWriter.AddBinary(ctx, secretName, payload)
		assert.NoError(t, err)
	})

	t.Run("AddBinary encrypt error", func(t *testing.T) {
		payload := models.BinaryPayload{
			FileName: "file.bin",
			Data:     []byte{4, 5, 6},
		}
		plaintext, _ := json.Marshal(payload)

		mockEncryptor.EXPECT().Encrypt(plaintext).Return(nil, errors.New("encrypt error"))

		err := secretWriter.AddBinary(ctx, secretName, payload)
		assert.Error(t, err)
	})

	t.Run("AddText success", func(t *testing.T) {
		payload := models.TextPayload{
			Data: "some text",
		}
		plaintext, _ := json.Marshal(payload)
		encrypted := &cryptor.Encrypted{
			Ciphertext: []byte("ciphertext-text"),
			AESKeyEnc:  []byte("keyenc-text"),
		}

		mockEncryptor.EXPECT().Encrypt(plaintext).Return(encrypted, nil)

		expectedSecret := &models.EncryptedSecret{
			SecretType: models.SecretTypeText,
			SecretName: secretName,
			Ciphertext: encrypted.Ciphertext,
			AESKeyEnc:  encrypted.AESKeyEnc,
		}

		mockWriter.EXPECT().Save(ctx, expectedSecret).Return(nil)

		err := secretWriter.AddText(ctx, secretName, payload)
		assert.NoError(t, err)
	})

	t.Run("AddText encrypt error", func(t *testing.T) {
		payload := models.TextPayload{
			Data: "fail encrypt",
		}
		plaintext, _ := json.Marshal(payload)

		mockEncryptor.EXPECT().Encrypt(plaintext).Return(nil, errors.New("encrypt error"))

		err := secretWriter.AddText(ctx, secretName, payload)
		assert.Error(t, err)
	})

	t.Run("AddUser success", func(t *testing.T) {
		payload := models.UserPayload{
			Login:    "user1",
			Password: "pass123",
		}
		plaintext, _ := json.Marshal(payload)
		encrypted := &cryptor.Encrypted{
			Ciphertext: []byte("ciphertext-user"),
			AESKeyEnc:  []byte("keyenc-user"),
		}

		mockEncryptor.EXPECT().Encrypt(plaintext).Return(encrypted, nil)

		expectedSecret := &models.EncryptedSecret{
			SecretType: models.SecretTypeUser,
			SecretName: secretName,
			Ciphertext: encrypted.Ciphertext,
			AESKeyEnc:  encrypted.AESKeyEnc,
		}

		mockWriter.EXPECT().Save(ctx, expectedSecret).Return(nil)

		err := secretWriter.AddUser(ctx, secretName, payload)
		assert.NoError(t, err)
	})

	t.Run("AddUser encrypt error", func(t *testing.T) {
		payload := models.UserPayload{
			Login:    "user2",
			Password: "failencrypt",
		}
		plaintext, _ := json.Marshal(payload)

		mockEncryptor.EXPECT().Encrypt(plaintext).Return(nil, errors.New("encrypt error"))

		err := secretWriter.AddUser(ctx, secretName, payload)
		assert.Error(t, err)
	})

	t.Run("AddBankCard success", func(t *testing.T) {
		payload := models.BankCardPayload{
			Number: "1234567890",
			CVV:    "123",
		}
		plaintext, _ := json.Marshal(payload)
		encrypted := &cryptor.Encrypted{
			Ciphertext: []byte("ciphertext-bankcard"),
			AESKeyEnc:  []byte("keyenc-bankcard"),
		}

		mockEncryptor.EXPECT().Encrypt(plaintext).Return(encrypted, nil)

		expectedSecret := &models.EncryptedSecret{
			SecretType: models.SecretTypeBankCard,
			SecretName: secretName,
			Ciphertext: encrypted.Ciphertext,
			AESKeyEnc:  encrypted.AESKeyEnc,
		}

		mockWriter.EXPECT().Save(ctx, expectedSecret).Return(nil)

		err := secretWriter.AddBankCard(ctx, secretName, payload)
		assert.NoError(t, err)
	})
	t.Run("AddBankCard encrypt error", func(t *testing.T) {
		payload := models.BankCardPayload{
			Number: "0000000000",
			CVV:    "999",
		}
		plaintext, _ := json.Marshal(payload)

		mockEncryptor.EXPECT().Encrypt(plaintext).Return(nil, errors.New("encrypt error"))

		err := secretWriter.AddBankCard(ctx, secretName, payload)
		assert.Error(t, err)
	})
}

func TestSecretWriter_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWriter := NewMockWriter(ctrl)
	secretWriter := NewSecretWriter(mockWriter, nil)

	ctx := context.Background()
	secretName := "mySecret"

	t.Run("Delete success", func(t *testing.T) {
		mockWriter.EXPECT().Delete(ctx, secretName).Return(nil)

		err := secretWriter.Delete(ctx, secretName)
		assert.NoError(t, err)
	})

	t.Run("Delete error", func(t *testing.T) {
		mockWriter.EXPECT().Delete(ctx, secretName).Return(errors.New("delete error"))

		err := secretWriter.Delete(ctx, secretName)
		assert.Error(t, err)
	})
}
