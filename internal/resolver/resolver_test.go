package resolver_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/resolver"
	"github.com/stretchr/testify/assert"
)

func makeSecret(name string, ts int64) *models.EncryptedSecret {
	return &models.EncryptedSecret{
		SecretName: name,
		Timestamp:  ts,
	}
}

func TestResolveClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLister := resolver.NewMockLister(ctrl)
	mockGetter := resolver.NewMockGetter(ctrl)
	mockSaver := resolver.NewMockSaver(ctrl)

	ctx := context.Background()

	clientSecrets := []*models.EncryptedSecret{
		makeSecret("secret1", 10),
		makeSecret("secret2", 20),
	}

	mockLister.EXPECT().List(ctx).Return(clientSecrets, nil)

	// For secret1: server secret is older => save called
	mockGetter.EXPECT().Get(ctx, "secret1").Return(makeSecret("secret1", 5), nil)
	mockSaver.EXPECT().Save(ctx, clientSecrets[0]).Return(nil)

	// For secret2: server secret is newer => no save
	mockGetter.EXPECT().Get(ctx, "secret2").Return(makeSecret("secret2", 25), nil)

	res := resolver.NewResolver(mockLister, mockGetter, mockSaver)

	err := res.ResolveClient(ctx)
	assert.NoError(t, err)
}

func TestResolveClient_GetError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLister := resolver.NewMockLister(ctrl)
	mockGetter := resolver.NewMockGetter(ctrl)
	mockSaver := resolver.NewMockSaver(ctrl)

	ctx := context.Background()

	clientSecrets := []*models.EncryptedSecret{
		makeSecret("secret1", 10),
	}

	mockLister.EXPECT().List(ctx).Return(clientSecrets, nil)
	mockGetter.EXPECT().Get(ctx, "secret1").Return(nil, errors.New("get error"))

	res := resolver.NewResolver(mockLister, mockGetter, mockSaver)

	err := res.ResolveClient(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get error")
}

func TestResolveServer(t *testing.T) {
	res := resolver.NewResolver(nil, nil, nil)
	err := res.ResolveServer(context.Background())
	assert.NoError(t, err)
}

func TestResolveInteractive(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLister := resolver.NewMockLister(ctrl)
	mockGetter := resolver.NewMockGetter(ctrl)
	mockSaver := resolver.NewMockSaver(ctrl)

	ctx := context.Background()

	clientSecrets := []*models.EncryptedSecret{
		makeSecret("secret1", 20),
		makeSecret("secret2", 30),
	}

	mockLister.EXPECT().List(ctx).Return(clientSecrets, nil)

	// secret1 server secret older => conflict
	mockGetter.EXPECT().Get(ctx, "secret1").Return(makeSecret("secret1", 10), nil)
	// secret2 server secret newer => no conflict (no get call for saving)

	mockGetter.EXPECT().Get(ctx, "secret2").Return(makeSecret("secret2", 40), nil)

	// Simulate user input: choose "1" for secret1 to save client version
	mockSaver.EXPECT().Save(ctx, clientSecrets[0]).Return(nil)

	// Prepare input for scanner
	input := bytes.NewBufferString("1\n")

	res := resolver.NewResolver(mockLister, mockGetter, mockSaver)
	err := res.ResolveInteractive(ctx, input)
	assert.NoError(t, err)
}

func TestResolveInteractive_InvalidInput(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLister := resolver.NewMockLister(ctrl)
	mockGetter := resolver.NewMockGetter(ctrl)
	mockSaver := resolver.NewMockSaver(ctrl)

	ctx := context.Background()

	clientSecrets := []*models.EncryptedSecret{
		makeSecret("secret1", 20),
	}

	mockLister.EXPECT().List(ctx).Return(clientSecrets, nil)
	mockGetter.EXPECT().Get(ctx, "secret1").Return(makeSecret("secret1", 10), nil)

	// Provide invalid input "3"
	input := bytes.NewBufferString("3\n")

	res := resolver.NewResolver(mockLister, mockGetter, mockSaver)
	err := res.ResolveInteractive(ctx, input)
	assert.Error(t, err)
	assert.Equal(t, "invalid version", err.Error())
}

func TestResolveInteractive_ScanFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLister := resolver.NewMockLister(ctrl)
	mockGetter := resolver.NewMockGetter(ctrl)
	mockSaver := resolver.NewMockSaver(ctrl)

	ctx := context.Background()

	clientSecrets := []*models.EncryptedSecret{
		makeSecret("secret1", 20),
	}

	mockLister.EXPECT().List(ctx).Return(clientSecrets, nil)
	mockGetter.EXPECT().Get(ctx, "secret1").Return(makeSecret("secret1", 10), nil)

	// Provide empty input, scanner.Scan() will return false
	input := bytes.NewBufferString("")

	res := resolver.NewResolver(mockLister, mockGetter, mockSaver)
	err := res.ResolveInteractive(ctx, input)
	assert.Error(t, err)
	assert.Equal(t, "failed to read input", err.Error())
}
