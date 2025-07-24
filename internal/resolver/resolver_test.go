package resolver

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/gophkeeper/internal/cryptor"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/stretchr/testify/require"
)

func TestResolver_ResolveServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClientReader := NewMockClientSecretReader(ctrl)
	mockServerReader := NewMockServerSecretReader(ctrl)
	mockServerWriter := NewMockServerSecretWriter(ctrl)
	mockCryptor := NewMockCryptor(ctrl)

	r := NewResolver(mockClientReader, mockServerReader, mockServerWriter, mockCryptor)
	ctx := context.Background()

	err := r.ResolveServer(ctx)
	require.NoError(t, err) // Just returns nil currently
}

func TestResolver_ResolveClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClientReader := NewMockClientSecretReader(ctrl)
	mockServerReader := NewMockServerSecretReader(ctrl)
	mockServerWriter := NewMockServerSecretWriter(ctrl)
	mockCryptor := NewMockCryptor(ctrl) // Not used in ResolveClient, but needed for NewResolver

	r := NewResolver(mockClientReader, mockServerReader, mockServerWriter, mockCryptor)
	ctx := context.Background()

	clientSecretNewer := &models.EncryptedSecret{
		SecretName: "secret1",
		Timestamp:  5,
	}
	clientSecretOlder := &models.EncryptedSecret{
		SecretName: "secret2",
		Timestamp:  3,
	}

	serverSecretOlder := &models.EncryptedSecret{
		SecretName: "secret1",
		Timestamp:  2,
	}
	serverSecretNewer := &models.EncryptedSecret{
		SecretName: "secret2",
		Timestamp:  4,
	}

	mockClientReader.EXPECT().List(ctx).Return([]*models.EncryptedSecret{clientSecretNewer, clientSecretOlder}, nil)

	// For secret1, server timestamp < client, expect save
	mockServerReader.EXPECT().Get(ctx, "secret1").Return(serverSecretOlder, nil)
	mockServerWriter.EXPECT().Save(ctx, clientSecretNewer).Return(nil)

	// For secret2, server timestamp > client, expect no save
	mockServerReader.EXPECT().Get(ctx, "secret2").Return(serverSecretNewer, nil)

	err := r.ResolveClient(ctx)
	require.NoError(t, err)
}

func TestResolver_ResolveClient_ServerSecretNil(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClientReader := NewMockClientSecretReader(ctrl)
	mockServerReader := NewMockServerSecretReader(ctrl)
	mockServerWriter := NewMockServerSecretWriter(ctrl)
	mockCryptor := NewMockCryptor(ctrl)

	r := NewResolver(mockClientReader, mockServerReader, mockServerWriter, mockCryptor)
	ctx := context.Background()

	clientSecret := &models.EncryptedSecret{
		SecretName: "secret1",
		Timestamp:  5,
	}

	mockClientReader.EXPECT().List(ctx).Return([]*models.EncryptedSecret{clientSecret}, nil)
	mockServerReader.EXPECT().Get(ctx, clientSecret.SecretName).Return(nil, nil)
	mockServerWriter.EXPECT().Save(ctx, clientSecret).Return(nil)

	err := r.ResolveClient(ctx)
	require.NoError(t, err)
}

func TestResolver_ResolveClient_ErrorCases(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClientReader := NewMockClientSecretReader(ctrl)
	mockServerReader := NewMockServerSecretReader(ctrl)
	mockServerWriter := NewMockServerSecretWriter(ctrl)
	mockCryptor := NewMockCryptor(ctrl)

	r := NewResolver(mockClientReader, mockServerReader, mockServerWriter, mockCryptor)
	ctx := context.Background()

	// Error listing client secrets
	mockClientReader.EXPECT().List(ctx).Return(nil, errors.New("list error"))
	err := r.ResolveClient(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to list client secrets")

	// Successful list but error fetching server secret
	clientSecret := &models.EncryptedSecret{SecretName: "secret1", Timestamp: 1}
	mockClientReader.EXPECT().List(ctx).Return([]*models.EncryptedSecret{clientSecret}, nil)
	mockServerReader.EXPECT().Get(ctx, "secret1").Return(nil, errors.New("get error"))
	err = r.ResolveClient(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "get error")

	// Successful get but error saving to server
	mockClientReader.EXPECT().List(ctx).Return([]*models.EncryptedSecret{clientSecret}, nil)
	mockServerReader.EXPECT().Get(ctx, "secret1").Return(nil, nil)
	mockServerWriter.EXPECT().Save(ctx, clientSecret).Return(errors.New("save error"))
	err = r.ResolveClient(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to save client secret")
}

func TestResolver_ResolveInteractive(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClientReader := NewMockClientSecretReader(ctrl)
	mockServerReader := NewMockServerSecretReader(ctrl)
	mockServerWriter := NewMockServerSecretWriter(ctrl)
	mockCryptor := NewMockCryptor(ctrl)

	r := NewResolver(mockClientReader, mockServerReader, mockServerWriter, mockCryptor)
	ctx := context.Background()

	clientSecret := &models.EncryptedSecret{
		SecretName: "secret1",
		Ciphertext: []byte("clientcipher"),
		AESKeyEnc:  []byte("clientkey"),
		Timestamp:  2,
	}

	serverSecret := &models.EncryptedSecret{
		SecretName: "secret1",
		Ciphertext: []byte("servercipher"),
		AESKeyEnc:  []byte("serverkey"),
		Timestamp:  1,
	}

	mockClientReader.EXPECT().List(ctx).Return([]*models.EncryptedSecret{clientSecret}, nil)
	mockServerReader.EXPECT().Get(ctx, clientSecret.SecretName).Return(serverSecret, nil)

	// Decrypt called twice: client and server
	mockCryptor.EXPECT().Decrypt(gomock.Any()).DoAndReturn(func(enc *cryptor.Encrypted) ([]byte, error) {
		if bytes.Equal(enc.Ciphertext, clientSecret.Ciphertext) {
			return []byte("clientplaintext"), nil
		}
		if bytes.Equal(enc.Ciphertext, serverSecret.Ciphertext) {
			return []byte("serverplaintext"), nil
		}
		return nil, errors.New("unknown ciphertext")
	}).Times(2)

	mockServerWriter.EXPECT().Save(ctx, clientSecret).Return(nil)

	input := strings.NewReader("1\n") // User chooses client version

	err := r.ResolveInteractive(ctx, input)
	require.NoError(t, err)
}

func TestResolver_ResolveInteractive_InvalidChoice(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClientReader := NewMockClientSecretReader(ctrl)
	mockServerReader := NewMockServerSecretReader(ctrl)
	mockServerWriter := NewMockServerSecretWriter(ctrl)
	mockCryptor := NewMockCryptor(ctrl)

	r := NewResolver(mockClientReader, mockServerReader, mockServerWriter, mockCryptor)
	ctx := context.Background()

	clientSecret := &models.EncryptedSecret{
		SecretName: "secret1",
		Ciphertext: []byte("clientcipher"),
		AESKeyEnc:  []byte("clientkey"),
		Timestamp:  2,
	}

	serverSecret := &models.EncryptedSecret{
		SecretName: "secret1",
		Ciphertext: []byte("servercipher"),
		AESKeyEnc:  []byte("serverkey"),
		Timestamp:  1,
	}

	mockClientReader.EXPECT().List(ctx).Return([]*models.EncryptedSecret{clientSecret}, nil)
	mockServerReader.EXPECT().Get(ctx, clientSecret.SecretName).Return(serverSecret, nil)

	// Decrypt called twice
	mockCryptor.EXPECT().Decrypt(gomock.Any()).Return([]byte("plaintext"), nil).Times(2)

	input := strings.NewReader("3\n") // Invalid choice

	err := r.ResolveInteractive(ctx, input)
	require.Error(t, err)
	require.Equal(t, "invalid choice", err.Error())
}

func TestResolver_ResolveInteractive_ServerSecretNil_AutoSaveClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClientReader := NewMockClientSecretReader(ctrl)
	mockServerReader := NewMockServerSecretReader(ctrl)
	mockServerWriter := NewMockServerSecretWriter(ctrl)
	mockCryptor := NewMockCryptor(ctrl)

	r := NewResolver(mockClientReader, mockServerReader, mockServerWriter, mockCryptor)
	ctx := context.Background()

	clientSecret := &models.EncryptedSecret{
		SecretName: "secret1",
		Ciphertext: []byte("clientcipher"),
		AESKeyEnc:  []byte("clientkey"),
		Timestamp:  2,
	}

	mockClientReader.EXPECT().List(ctx).Return([]*models.EncryptedSecret{clientSecret}, nil)
	mockServerReader.EXPECT().Get(ctx, clientSecret.SecretName).Return(nil, nil)

	mockCryptor.EXPECT().Decrypt(gomock.Any()).Return([]byte("plaintext"), nil).Times(1)

	mockServerWriter.EXPECT().Save(ctx, clientSecret).Return(nil)

	// No user input needed since server secret is nil and client saved automatically
	err := r.ResolveInteractive(ctx, strings.NewReader(""))
	require.NoError(t, err)
}
