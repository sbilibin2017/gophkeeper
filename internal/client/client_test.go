package client

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

func TestClientRegister(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRegisterer := NewMockRegisterer(ctrl)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		token := "tok"
		mockRegisterer.EXPECT().Register(ctx, "user", "pass").Return(&token, nil)
		got, err := ClientRegister(ctx, mockRegisterer, "user", "pass")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if got != token {
			t.Fatalf("expected token %q, got %q", token, got)
		}
	})

	t.Run("error from Register", func(t *testing.T) {
		mockRegisterer.EXPECT().Register(ctx, "user", "pass").Return(nil, errors.New("fail"))
		_, err := ClientRegister(ctx, mockRegisterer, "user", "pass")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("nil token", func(t *testing.T) {
		mockRegisterer.EXPECT().Register(ctx, "user", "pass").Return(nil, nil)
		_, err := ClientRegister(ctx, mockRegisterer, "user", "pass")
		if err == nil || !strings.Contains(err.Error(), "nil token") {
			t.Fatalf("expected nil token error, got %v", err)
		}
	})
}

func TestClientLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLoginer := NewMockLoginer(ctrl)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		token := "tok"
		mockLoginer.EXPECT().Login(ctx, "user", "pass").Return(&token, nil)
		got, err := ClientLogin(ctx, mockLoginer, "user", "pass")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if got != token {
			t.Fatalf("expected token %q, got %q", token, got)
		}
	})

	t.Run("error from Login", func(t *testing.T) {
		mockLoginer.EXPECT().Login(ctx, "user", "pass").Return(nil, errors.New("fail"))
		_, err := ClientLogin(ctx, mockLoginer, "user", "pass")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("nil token", func(t *testing.T) {
		mockLoginer.EXPECT().Login(ctx, "user", "pass").Return(nil, nil)
		_, err := ClientLogin(ctx, mockLoginer, "user", "pass")
		if err == nil || !strings.Contains(err.Error(), "nil token") {
			t.Fatalf("expected nil token error, got %v", err)
		}
	})
}

func TestClientAddBankcard(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSaver := NewMockClientSaver(ctrl)
	mockEncryptor := NewMockEncryptor(ctrl)
	ctx := context.Background()

	bankcard := models.Bankcard{
		SecretName: "name",
		Number:     "1234",
		Owner:      "owner",
		Exp:        "12/25",
		CVV:        "999",
		Meta:       nil,
	}
	plaintext, _ := json.Marshal(bankcard)

	SecretEncrypted := &models.SecretSecretEncrypted{
		Ciphertext: []byte("cipher"),
		AESKeyEnc:  []byte("key"),
	}

	mockEncryptor.EXPECT().Encrypt(plaintext).Return(SecretEncrypted, nil)
	mockSaver.EXPECT().Save(ctx, "token", "name", models.SecretTypeBankCard, SecretEncrypted.Ciphertext, SecretEncrypted.AESKeyEnc).Return(nil)

	err := ClientAddBankcard(ctx, mockSaver, mockEncryptor, "token", "name", "1234", "owner", "12/25", "999", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	t.Run("encryption failure", func(t *testing.T) {
		mockEncryptor.EXPECT().Encrypt(gomock.Any()).Return(nil, errors.New("enc fail"))
		err := ClientAddBankcard(ctx, mockSaver, mockEncryptor, "token", "name", "1234", "owner", "12/25", "999", "")
		if err == nil || !strings.Contains(err.Error(), "encryption failed") {
			t.Fatalf("expected encryption failed error, got %v", err)
		}
	})
}

func TestClientAddText(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSaver := NewMockClientSaver(ctrl)
	mockEncryptor := NewMockEncryptor(ctrl)
	ctx := context.Background()

	text := models.Text{
		SecretName: "name",
		Data:       "data",
		Meta:       nil,
	}
	plaintext, _ := json.Marshal(text)

	SecretEncrypted := &models.SecretSecretEncrypted{
		Ciphertext: []byte("cipher"),
		AESKeyEnc:  []byte("key"),
	}

	mockEncryptor.EXPECT().Encrypt(plaintext).Return(SecretEncrypted, nil)
	mockSaver.EXPECT().Save(ctx, "token", "name", models.SecretTypeText, SecretEncrypted.Ciphertext, SecretEncrypted.AESKeyEnc).Return(nil)

	err := ClientAddText(ctx, mockSaver, mockEncryptor, "token", "name", "data", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestClientAddBinary(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSaver := NewMockClientSaver(ctrl)
	mockEncryptor := NewMockEncryptor(ctrl)
	ctx := context.Background()

	data := base64.StdEncoding.EncodeToString([]byte("binarydata"))

	binary := models.Binary{
		SecretName: "name",
		Data:       []byte("binarydata"),
		Meta:       nil,
	}
	plaintext, _ := json.Marshal(binary)

	SecretEncrypted := &models.SecretSecretEncrypted{
		Ciphertext: []byte("cipher"),
		AESKeyEnc:  []byte("key"),
	}

	mockEncryptor.EXPECT().Encrypt(plaintext).Return(SecretEncrypted, nil)
	mockSaver.EXPECT().Save(ctx, "token", "name", models.SecretTypeBinary, SecretEncrypted.Ciphertext, SecretEncrypted.AESKeyEnc).Return(nil)

	err := ClientAddBinary(ctx, mockSaver, mockEncryptor, "token", "name", data, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	t.Run("invalid base64", func(t *testing.T) {
		err := ClientAddBinary(ctx, mockSaver, mockEncryptor, "token", "name", "!!!", "")
		if err == nil || !strings.Contains(err.Error(), "failed to decode base64") {
			t.Fatalf("expected base64 decode error, got %v", err)
		}
	})
}

func TestClientAddUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSaver := NewMockClientSaver(ctrl)
	mockEncryptor := NewMockEncryptor(ctrl)
	ctx := context.Background()

	user := models.User{
		SecretName: "name",
		Username:   "u",
		Password:   "p",
		Meta:       nil,
	}
	plaintext, _ := json.Marshal(user)

	SecretEncrypted := &models.SecretSecretEncrypted{
		Ciphertext: []byte("cipher"),
		AESKeyEnc:  []byte("key"),
	}

	mockEncryptor.EXPECT().Encrypt(plaintext).Return(SecretEncrypted, nil)
	mockSaver.EXPECT().Save(ctx, "token", "name", models.SecretTypeUser, SecretEncrypted.Ciphertext, SecretEncrypted.AESKeyEnc).Return(nil)

	err := ClientAddUser(ctx, mockSaver, mockEncryptor, "token", "name", "u", "p", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestClientListSecrets_AllTypes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	token := "user-token"

	mockServerLister := NewMockServerLister(ctrl)
	mockDecryptor := NewMockDecryptor(ctrl)

	// Prepare secrets of all types + one unknown type
	mockSecrets := []*models.Secret{
		{
			SecretName: "bankcard1",
			SecretType: models.SecretTypeBankCard,
			Ciphertext: []byte("SecretEncrypted-bankcard"),
			AESKeyEnc:  []byte("aes-key-enc1"),
		},
		{
			SecretName: "text1",
			SecretType: models.SecretTypeText,
			Ciphertext: []byte("SecretEncrypted-text"),
			AESKeyEnc:  []byte("aes-key-enc2"),
		},
		{
			SecretName: "binary1",
			SecretType: models.SecretTypeBinary,
			Ciphertext: []byte("SecretEncrypted-binary"),
			AESKeyEnc:  []byte("aes-key-enc3"),
		},
		{
			SecretName: "user1",
			SecretType: models.SecretTypeUser,
			Ciphertext: []byte("SecretEncrypted-user"),
			AESKeyEnc:  []byte("aes-key-enc4"),
		},
		{
			SecretName: "unknown1",
			SecretType: "some-unknown-type",
			Ciphertext: []byte("SecretEncrypted-unknown"),
			AESKeyEnc:  []byte("aes-key-enc5"),
		},
	}

	mockServerLister.
		EXPECT().
		List(ctx, token).
		Return(mockSecrets, nil)

	// Prepare decrypted JSON for each secret type
	bankcard := models.Bankcard{
		SecretName: "bankcard1",
		Number:     "1234",
		Owner:      "Alice",
		Exp:        "12/24",
		CVV:        "999",
	}
	bankcardJSON, _ := json.Marshal(bankcard)

	text := models.Text{
		SecretName: "text1",
		Data:       "my secret note",
	}
	textJSON, _ := json.Marshal(text)

	binary := models.Binary{
		SecretName: "binary1",
		Data:       []byte{0x01, 0x02, 0x03},
	}
	binaryJSON, _ := json.Marshal(binary)

	user := models.User{
		SecretName: "user1",
		Username:   "user@example.com",
		Password:   "supersecret",
	}
	userJSON, _ := json.Marshal(user)

	// Setup decryptor expectations in order of calls
	gomock.InOrder(
		mockDecryptor.EXPECT().Decrypt(gomock.Any()).Return(bankcardJSON, nil),
		mockDecryptor.EXPECT().Decrypt(gomock.Any()).Return(textJSON, nil),
		mockDecryptor.EXPECT().Decrypt(gomock.Any()).Return(binaryJSON, nil),
		mockDecryptor.EXPECT().Decrypt(gomock.Any()).Return(userJSON, nil),
		// For unknown type, decrypt still called, but output ignored except message printed
		mockDecryptor.EXPECT().Decrypt(gomock.Any()).Return([]byte{}, nil),
	)

	output, err := ClientListSecrets(ctx, mockServerLister, mockDecryptor, token)
	if err != nil {
		t.Fatalf("ClientListSecrets failed: %v", err)
	}

	// Check that output contains all expected decrypted secret data
	tests := []string{
		`"secret_name": "bankcard1"`, `"number": "1234"`, `"owner": "Alice"`, `"exp": "12/24"`, `"cvv": "999"`,
		`"secret_name": "text1"`, `"data": "my secret note"`,
		`"secret_name": "binary1"`, `"data": "AQID"`, // base64 for 0x01,0x02,0x03
		`"secret_name": "user1"`, `"username": "user@example.com"`, `"password": "supersecret"`,
		"Unknown secret type: some-unknown-type",
	}

	for _, substr := range tests {
		if !strings.Contains(output, substr) {
			t.Errorf("Output missing expected substring: %q\nOutput was:\n%s", substr, output)
		}
	}
}

func TestClientSyncClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	secretOwner := "owner1"

	mockResolver := NewMockClientResolver(ctrl)

	// Success case: Expect Resolve called once with ctx and secretOwner, returns nil error
	mockResolver.EXPECT().
		Resolve(ctx, secretOwner).
		Return(nil).
		Times(1)

	err := ClientSyncClient(ctx, secretOwner, mockResolver)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Error case: Resolve returns error
	mockResolver.EXPECT().
		Resolve(ctx, secretOwner).
		Return(errors.New("resolve failed")).
		Times(1)

	err = ClientSyncClient(ctx, secretOwner, mockResolver)
	if err == nil || err.Error() != "resolve failed" {
		t.Fatalf("expected resolve failed error, got: %v", err)
	}
}

func TestClientSyncInteractive(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	secretOwner := "owner2"
	input := strings.NewReader("user input")

	mockResolver := NewMockInteractiveResolver(ctrl)

	// Success case: Expect Resolve called once with ctx, secretOwner, and input reader, returns nil error
	mockResolver.EXPECT().
		Resolve(ctx, secretOwner, input).
		Return(nil).
		Times(1)

	err := ClientSyncInteractive(ctx, secretOwner, mockResolver, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Error case: Resolve returns error
	mockResolver.EXPECT().
		Resolve(ctx, secretOwner, input).
		Return(errors.New("interactive resolve failed")).
		Times(1)

	err = ClientSyncInteractive(ctx, secretOwner, mockResolver, input)
	if err == nil || err.Error() != "interactive resolve failed" {
		t.Fatalf("expected interactive resolve failed error, got: %v", err)
	}
}
