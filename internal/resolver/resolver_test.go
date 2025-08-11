package resolver

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/stretchr/testify/require"
)

func TestClientSyncClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	secretOwner := "user1"

	mockClientLister := NewMockClientLister(ctrl)
	mockServerGetter := NewMockServerGetter(ctrl)
	mockServerSaver := NewMockServerSaver(ctrl)

	clientSecret := &models.SecretDB{
		SecretName: "secret1",
		SecretType: "text",
		Ciphertext: []byte("ciphertext"),
		AESKeyEnc:  []byte("aeskey"),
		UpdatedAt:  time.Now(),
	}

	serverSecret := &models.SecretDB{
		SecretName: "secret1",
		SecretType: "text",
		Ciphertext: []byte("oldcipher"),
		AESKeyEnc:  []byte("oldkey"),
		UpdatedAt:  time.Now().Add(-time.Hour),
	}

	mockClientLister.EXPECT().List(ctx, secretOwner).Return([]*models.SecretDB{clientSecret}, nil)
	mockServerGetter.EXPECT().
		Get(ctx, secretOwner, clientSecret.SecretName, clientSecret.SecretType).
		Return(serverSecret, nil)
	mockServerSaver.EXPECT().
		Save(ctx, secretOwner, clientSecret.SecretName, clientSecret.SecretType, clientSecret.Ciphertext, clientSecret.AESKeyEnc).
		Return(nil)

	err := ClientSyncClient(ctx, mockClientLister, mockServerGetter, mockServerSaver, secretOwner)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClientSyncInteractive_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLister := NewMockClientLister(ctrl)
	mockGetter := NewMockServerGetter(ctrl)
	mockSaver := NewMockServerSaver(ctrl)
	mockDecryptor := NewMockDecryptor(ctrl)

	ctx := context.Background()
	secretOwner := "user1"

	clientSecret := &models.SecretDB{
		SecretName: "text",
		SecretType: "secret1",
		UpdatedAt:  time.Now(),
		Ciphertext: []byte("ciphertext"),
		AESKeyEnc:  []byte("aeskey"),
	}

	serverSecret := &models.SecretDB{
		SecretName: "text",
		SecretType: "secret1",
		UpdatedAt:  clientSecret.UpdatedAt.Add(-time.Hour), // older than client
		Ciphertext: []byte("servercipher"),
		AESKeyEnc:  []byte("serveraes"),
	}

	mockLister.EXPECT().
		List(ctx, secretOwner).
		Return([]*models.SecretDB{clientSecret}, nil)

	mockGetter.EXPECT().
		Get(ctx, secretOwner, clientSecret.SecretName, clientSecret.SecretType).
		Return(serverSecret, nil)

	mockDecryptor.EXPECT().
		Decrypt(gomock.Any()).
		Return([]byte(`{"foo":"bar"}`), nil).
		Times(2)

	mockSaver.EXPECT().
		Save(ctx, secretOwner, clientSecret.SecretName, clientSecret.SecretType,
			clientSecret.Ciphertext, clientSecret.AESKeyEnc).
		Return(nil)

	input := bytes.NewBufferString("1\n")

	err := ClientSyncInteractive(ctx, mockLister, mockGetter, mockSaver, mockDecryptor, secretOwner, input)
	require.NoError(t, err)
}

func TestClientSyncInteractive_InvalidInput(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLister := NewMockClientLister(ctrl)
	mockGetter := NewMockServerGetter(ctrl)
	mockSaver := NewMockServerSaver(ctrl)
	mockDecryptor := NewMockDecryptor(ctrl)

	ctx := context.Background()
	secretOwner := "user1"

	clientSecret := &models.SecretDB{
		SecretName: "text",
		SecretType: "secret1",
		UpdatedAt:  time.Now(),
		Ciphertext: []byte("ciphertext"),
		AESKeyEnc:  []byte("aeskey"),
	}

	serverSecret := &models.SecretDB{
		SecretName: "text",
		SecretType: "secret1",
		UpdatedAt:  clientSecret.UpdatedAt.Add(-time.Hour),
		Ciphertext: []byte("servercipher"),
		AESKeyEnc:  []byte("serveraes"),
	}

	mockLister.EXPECT().
		List(ctx, secretOwner).
		Return([]*models.SecretDB{clientSecret}, nil)

	mockGetter.EXPECT().
		Get(ctx, secretOwner, clientSecret.SecretName, clientSecret.SecretType).
		Return(serverSecret, nil)

	mockDecryptor.EXPECT().
		Decrypt(gomock.Any()).
		Return([]byte(`{"foo":"bar"}`), nil).
		Times(2)

	input := bytes.NewBufferString("invalid\n")

	err := ClientSyncInteractive(ctx, mockLister, mockGetter, mockSaver, mockDecryptor, secretOwner, input)
	require.Error(t, err)
	require.Equal(t, "unsupported input", err.Error())
}
