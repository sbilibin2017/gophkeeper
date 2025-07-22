package client

import (
	"bufio"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

func TestResolveBankCardHTTP(t *testing.T) {
	ctx := context.Background()

	// Common test data
	now := time.Now()
	clientSecrets := []models.BankCardDB{
		{
			SecretName: "card1",
			Number:     "1111222233334444",
			Owner:      "Alice",
			Exp:        "12/25",
			CVV:        "123",
			Meta:       nil,
			UpdatedAt:  now,
		},
	}

	serverSecretNewer := &models.BankCardDB{
		SecretName: "card1",
		Number:     "9999888877776666",
		Owner:      "ServerOwner",
		Exp:        "11/24",
		CVV:        "999",
		Meta:       nil,
		UpdatedAt:  now.Add(1 * time.Hour),
	}
	serverSecretOlder := &models.BankCardDB{
		SecretName: "card1",
		Number:     "5555444433332222",
		Owner:      "ServerOwner",
		Exp:        "10/23",
		CVV:        "555",
		Meta:       nil,
		UpdatedAt:  now.Add(-1 * time.Hour),
	}

	// Dummy DB and client (not used for actual DB or HTTP calls in these tests)
	var dummyDB *sqlx.DB
	var dummyClient *resty.Client

	t.Run("strategy server returns nil", func(t *testing.T) {
		err := ResolveBankCardHTTP(ctx, bufio.NewReader(strings.NewReader("")), "server", nil, nil, nil, dummyDB, dummyClient, "")
		require.NoError(t, err)
	})

	t.Run("strategy client syncs when server secret missing", func(t *testing.T) {
		listCalled := false
		getCalled := false
		addCalled := false

		err := ResolveBankCardHTTP(ctx, nil, "client",
			func(ctx context.Context, db *sqlx.DB) ([]models.BankCardDB, error) {
				listCalled = true
				return clientSecrets, nil
			},
			func(ctx context.Context, client *resty.Client, secretName string) (*models.BankCardDB, error) {
				getCalled = true
				// Server returns nil, simulating missing secret on server
				return nil, nil
			},
			func(ctx context.Context, client *resty.Client, req *models.BankCardAddRequest) error {
				addCalled = true
				// Check that secretName matches expected
				assert.Equal(t, "card1", req.SecretName)
				return nil
			},
			dummyDB, dummyClient, "card1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.True(t, getCalled)
		assert.True(t, addCalled)
	})

	t.Run("strategy client syncs when client secret is newer", func(t *testing.T) {
		listCalled := false
		getCalled := false
		addCalled := false

		err := ResolveBankCardHTTP(ctx, nil, "client",
			func(ctx context.Context, db *sqlx.DB) ([]models.BankCardDB, error) {
				listCalled = true
				return clientSecrets, nil
			},
			func(ctx context.Context, client *resty.Client, secretName string) (*models.BankCardDB, error) {
				getCalled = true
				// Server secret is older than client secret
				return serverSecretOlder, nil
			},
			func(ctx context.Context, client *resty.Client, req *models.BankCardAddRequest) error {
				addCalled = true
				assert.Equal(t, "card1", req.SecretName)
				return nil
			},
			dummyDB, dummyClient, "card1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.True(t, getCalled)
		assert.True(t, addCalled)
	})

	t.Run("strategy client skips when server secret is newer", func(t *testing.T) {
		listCalled := false
		getCalled := false
		addCalled := false

		err := ResolveBankCardHTTP(ctx, nil, "client",
			func(ctx context.Context, db *sqlx.DB) ([]models.BankCardDB, error) {
				listCalled = true
				return clientSecrets, nil
			},
			func(ctx context.Context, client *resty.Client, secretName string) (*models.BankCardDB, error) {
				getCalled = true
				// Server secret is newer, so no add
				return serverSecretNewer, nil
			},
			func(ctx context.Context, client *resty.Client, req *models.BankCardAddRequest) error {
				addCalled = true
				return nil
			},
			dummyDB, dummyClient, "card1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.True(t, getCalled)
		assert.False(t, addCalled)
	})

	t.Run("strategy interactive with server secret nil adds client secret", func(t *testing.T) {
		listCalled := false
		getCalled := false
		addCalled := false

		err := ResolveBankCardHTTP(ctx, bufio.NewReader(strings.NewReader("")), "interactive",
			func(ctx context.Context, db *sqlx.DB) ([]models.BankCardDB, error) {
				listCalled = true
				return clientSecrets, nil
			},
			func(ctx context.Context, client *resty.Client, secretName string) (*models.BankCardDB, error) {
				getCalled = true
				// Server secret missing
				return nil, nil
			},
			func(ctx context.Context, client *resty.Client, req *models.BankCardAddRequest) error {
				addCalled = true
				return nil
			},
			dummyDB, dummyClient, "card1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.True(t, getCalled)
		assert.True(t, addCalled)
	})

	t.Run("strategy interactive skips when client secret not newer", func(t *testing.T) {
		listCalled := false
		getCalled := false
		addCalled := false

		// client secret older than server secret, so no conflict, skip
		clientOlder := []models.BankCardDB{
			{
				SecretName: "card1",
				Number:     "123",
				UpdatedAt:  now.Add(-2 * time.Hour),
			},
		}

		err := ResolveBankCardHTTP(ctx, bufio.NewReader(strings.NewReader("")), "interactive",
			func(ctx context.Context, db *sqlx.DB) ([]models.BankCardDB, error) {
				listCalled = true
				return clientOlder, nil
			},
			func(ctx context.Context, client *resty.Client, secretName string) (*models.BankCardDB, error) {
				getCalled = true
				return serverSecretNewer, nil
			},
			func(ctx context.Context, client *resty.Client, req *models.BankCardAddRequest) error {
				addCalled = true
				return nil
			},
			dummyDB, dummyClient, "card1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.True(t, getCalled)
		assert.False(t, addCalled)
	})

	t.Run("strategy interactive user chooses client version", func(t *testing.T) {
		listCalled := false
		getCalled := false
		addCalled := false

		// client secret newer than server secret, conflict expected, user inputs "client"
		input := "client\n"

		err := ResolveBankCardHTTP(ctx, bufio.NewReader(strings.NewReader(input)), "interactive",
			func(ctx context.Context, db *sqlx.DB) ([]models.BankCardDB, error) {
				listCalled = true
				return clientSecrets, nil
			},
			func(ctx context.Context, client *resty.Client, secretName string) (*models.BankCardDB, error) {
				getCalled = true
				return serverSecretOlder, nil
			},
			func(ctx context.Context, client *resty.Client, req *models.BankCardAddRequest) error {
				addCalled = true
				return nil
			},
			dummyDB, dummyClient, "card1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.True(t, getCalled)
		assert.True(t, addCalled)
	})

	t.Run("strategy interactive user chooses server version", func(t *testing.T) {
		listCalled := false
		getCalled := false
		addCalled := false

		// user inputs "server", so skip adding
		input := "server\n"

		err := ResolveBankCardHTTP(ctx, bufio.NewReader(strings.NewReader(input)), "interactive",
			func(ctx context.Context, db *sqlx.DB) ([]models.BankCardDB, error) {
				listCalled = true
				return clientSecrets, nil
			},
			func(ctx context.Context, client *resty.Client, secretName string) (*models.BankCardDB, error) {
				getCalled = true
				return serverSecretOlder, nil
			},
			func(ctx context.Context, client *resty.Client, req *models.BankCardAddRequest) error {
				addCalled = true
				return nil
			},
			dummyDB, dummyClient, "card1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.True(t, getCalled)
		assert.False(t, addCalled)
	})

	t.Run("strategy interactive invalid user input", func(t *testing.T) {
		// user inputs invalid choice
		input := "invalid\n"

		err := ResolveBankCardHTTP(ctx, bufio.NewReader(strings.NewReader(input)), "interactive",
			func(ctx context.Context, db *sqlx.DB) ([]models.BankCardDB, error) {
				return clientSecrets, nil
			},
			func(ctx context.Context, client *resty.Client, secretName string) (*models.BankCardDB, error) {
				return serverSecretOlder, nil
			},
			func(ctx context.Context, client *resty.Client, req *models.BankCardAddRequest) error {
				return nil
			},
			dummyDB, dummyClient, "card1",
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid choice")
	})

	t.Run("unknown strategy returns error", func(t *testing.T) {
		err := ResolveBankCardHTTP(ctx, nil, "unknown", nil, nil, nil, dummyDB, dummyClient, "card1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown strategy")
	})

	t.Run("listClientFunc error propagates", func(t *testing.T) {
		expectedErr := errors.New("list error")
		err := ResolveBankCardHTTP(ctx, nil, "client",
			func(ctx context.Context, db *sqlx.DB) ([]models.BankCardDB, error) {
				return nil, expectedErr
			},
			nil, nil, dummyDB, dummyClient, "card1")
		require.ErrorIs(t, err, expectedErr)
	})

	t.Run("getServerFunc error propagates", func(t *testing.T) {
		expectedErr := errors.New("get error")
		err := ResolveBankCardHTTP(ctx, nil, "client",
			func(ctx context.Context, db *sqlx.DB) ([]models.BankCardDB, error) {
				return clientSecrets, nil
			},
			func(ctx context.Context, client *resty.Client, secretName string) (*models.BankCardDB, error) {
				return nil, expectedErr
			},
			nil, dummyDB, dummyClient, "card1",
		)
		require.ErrorIs(t, err, expectedErr)
	})

	t.Run("addServerFunc error propagates", func(t *testing.T) {
		expectedErr := errors.New("add error")
		err := ResolveBankCardHTTP(ctx, nil, "client",
			func(ctx context.Context, db *sqlx.DB) ([]models.BankCardDB, error) {
				return clientSecrets, nil
			},
			func(ctx context.Context, client *resty.Client, secretName string) (*models.BankCardDB, error) {
				return nil, nil
			},
			func(ctx context.Context, client *resty.Client, req *models.BankCardAddRequest) error {
				return expectedErr
			},
			dummyDB, dummyClient, "card1",
		)
		require.ErrorIs(t, err, expectedErr)
	})
}

type mockBankCardServiceClient struct{}

func (m *mockBankCardServiceClient) Add(ctx context.Context, in *pb.BankCardAddRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (m *mockBankCardServiceClient) Get(ctx context.Context, in *pb.BankCardFilterRequest, opts ...grpc.CallOption) (*pb.BankCard, error) {
	return nil, nil
}

func (m *mockBankCardServiceClient) List(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (pb.BankCardService_ListClient, error) {
	return nil, nil
}

func TestResolveBankCardGRPC(t *testing.T) {
	ctx := context.Background()

	now := time.Now()
	clientSecrets := []models.BankCardDB{
		{
			SecretName: "card1",
			Number:     "1111222233334444",
			Owner:      "Alice",
			Exp:        "12/25",
			CVV:        "123",
			Meta:       nil,
			UpdatedAt:  now,
		},
	}

	serverSecretNewer := &models.BankCardDB{
		SecretName: "card1",
		Number:     "9999888877776666",
		Owner:      "ServerOwner",
		Exp:        "11/24",
		CVV:        "999",
		Meta:       nil,
		UpdatedAt:  now.Add(1 * time.Hour),
	}
	serverSecretOlder := &models.BankCardDB{
		SecretName: "card1",
		Number:     "5555444433332222",
		Owner:      "ServerOwner",
		Exp:        "10/23",
		CVV:        "555",
		Meta:       nil,
		UpdatedAt:  now.Add(-1 * time.Hour),
	}

	var dummyDB *sqlx.DB
	var dummyClient pb.BankCardServiceClient = &mockBankCardServiceClient{}

	t.Run("strategy server returns nil", func(t *testing.T) {
		err := ResolveBankCardGRPC(ctx, bufio.NewReader(strings.NewReader("")), "server", nil, nil, nil, dummyDB, dummyClient, "")
		require.NoError(t, err)
	})

	t.Run("strategy client syncs when server secret missing", func(t *testing.T) {
		listCalled := false
		getCalled := false
		addCalled := false

		err := ResolveBankCardGRPC(ctx, nil, "client",
			func(ctx context.Context, db *sqlx.DB) ([]models.BankCardDB, error) {
				listCalled = true
				return clientSecrets, nil
			},
			func(ctx context.Context, client pb.BankCardServiceClient, secretName string) (*models.BankCardDB, error) {
				getCalled = true
				return nil, nil
			},
			func(ctx context.Context, client pb.BankCardServiceClient, req *models.BankCardAddRequest) error {
				addCalled = true
				assert.Equal(t, "card1", req.SecretName)
				return nil
			},
			dummyDB, dummyClient, "card1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.True(t, getCalled)
		assert.True(t, addCalled)
	})

	t.Run("strategy client syncs when client secret is newer", func(t *testing.T) {
		listCalled := false
		getCalled := false
		addCalled := false

		err := ResolveBankCardGRPC(ctx, nil, "client",
			func(ctx context.Context, db *sqlx.DB) ([]models.BankCardDB, error) {
				listCalled = true
				return clientSecrets, nil
			},
			func(ctx context.Context, client pb.BankCardServiceClient, secretName string) (*models.BankCardDB, error) {
				getCalled = true
				return serverSecretOlder, nil
			},
			func(ctx context.Context, client pb.BankCardServiceClient, req *models.BankCardAddRequest) error {
				addCalled = true
				assert.Equal(t, "card1", req.SecretName)
				return nil
			},
			dummyDB, dummyClient, "card1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.True(t, getCalled)
		assert.True(t, addCalled)
	})

	t.Run("strategy client skips when server secret is newer", func(t *testing.T) {
		listCalled := false
		getCalled := false
		addCalled := false

		err := ResolveBankCardGRPC(ctx, nil, "client",
			func(ctx context.Context, db *sqlx.DB) ([]models.BankCardDB, error) {
				listCalled = true
				return clientSecrets, nil
			},
			func(ctx context.Context, client pb.BankCardServiceClient, secretName string) (*models.BankCardDB, error) {
				getCalled = true
				return serverSecretNewer, nil
			},
			func(ctx context.Context, client pb.BankCardServiceClient, req *models.BankCardAddRequest) error {
				addCalled = true
				return nil
			},
			dummyDB, dummyClient, "card1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.True(t, getCalled)
		assert.False(t, addCalled)
	})

	t.Run("strategy interactive with server secret nil adds client secret", func(t *testing.T) {
		listCalled := false
		getCalled := false
		addCalled := false

		err := ResolveBankCardGRPC(ctx, bufio.NewReader(strings.NewReader("")), "interactive",
			func(ctx context.Context, db *sqlx.DB) ([]models.BankCardDB, error) {
				listCalled = true
				return clientSecrets, nil
			},
			func(ctx context.Context, client pb.BankCardServiceClient, secretName string) (*models.BankCardDB, error) {
				getCalled = true
				return nil, nil
			},
			func(ctx context.Context, client pb.BankCardServiceClient, req *models.BankCardAddRequest) error {
				addCalled = true
				return nil
			},
			dummyDB, dummyClient, "card1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.True(t, getCalled)
		assert.True(t, addCalled)
	})

	t.Run("strategy interactive skips when client secret not newer", func(t *testing.T) {
		listCalled := false
		getCalled := false
		addCalled := false

		clientOlder := []models.BankCardDB{
			{
				SecretName: "card1",
				Number:     "123",
				UpdatedAt:  now.Add(-2 * time.Hour),
			},
		}

		err := ResolveBankCardGRPC(ctx, bufio.NewReader(strings.NewReader("")), "interactive",
			func(ctx context.Context, db *sqlx.DB) ([]models.BankCardDB, error) {
				listCalled = true
				return clientOlder, nil
			},
			func(ctx context.Context, client pb.BankCardServiceClient, secretName string) (*models.BankCardDB, error) {
				getCalled = true
				return serverSecretNewer, nil
			},
			func(ctx context.Context, client pb.BankCardServiceClient, req *models.BankCardAddRequest) error {
				addCalled = true
				return nil
			},
			dummyDB, dummyClient, "card1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.True(t, getCalled)
		assert.False(t, addCalled)
	})

	t.Run("strategy interactive user chooses client version", func(t *testing.T) {
		listCalled := false
		getCalled := false
		addCalled := false

		input := "client\n"

		err := ResolveBankCardGRPC(ctx, bufio.NewReader(strings.NewReader(input)), "interactive",
			func(ctx context.Context, db *sqlx.DB) ([]models.BankCardDB, error) {
				listCalled = true
				return clientSecrets, nil
			},
			func(ctx context.Context, client pb.BankCardServiceClient, secretName string) (*models.BankCardDB, error) {
				getCalled = true
				return serverSecretOlder, nil
			},
			func(ctx context.Context, client pb.BankCardServiceClient, req *models.BankCardAddRequest) error {
				addCalled = true
				return nil
			},
			dummyDB, dummyClient, "card1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.True(t, getCalled)
		assert.True(t, addCalled)
	})

	t.Run("strategy interactive user chooses server version", func(t *testing.T) {
		listCalled := false
		getCalled := false
		addCalled := false

		input := "server\n"

		err := ResolveBankCardGRPC(ctx, bufio.NewReader(strings.NewReader(input)), "interactive",
			func(ctx context.Context, db *sqlx.DB) ([]models.BankCardDB, error) {
				listCalled = true
				return clientSecrets, nil
			},
			func(ctx context.Context, client pb.BankCardServiceClient, secretName string) (*models.BankCardDB, error) {
				getCalled = true
				return serverSecretOlder, nil
			},
			func(ctx context.Context, client pb.BankCardServiceClient, req *models.BankCardAddRequest) error {
				addCalled = true
				return nil
			},
			dummyDB, dummyClient, "card1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.True(t, getCalled)
		assert.False(t, addCalled)
	})

	t.Run("strategy interactive invalid user input", func(t *testing.T) {
		input := "invalid\n"

		err := ResolveBankCardGRPC(ctx, bufio.NewReader(strings.NewReader(input)), "interactive",
			func(ctx context.Context, db *sqlx.DB) ([]models.BankCardDB, error) {
				return clientSecrets, nil
			},
			func(ctx context.Context, client pb.BankCardServiceClient, secretName string) (*models.BankCardDB, error) {
				return serverSecretOlder, nil
			},
			func(ctx context.Context, client pb.BankCardServiceClient, req *models.BankCardAddRequest) error {
				return nil
			},
			dummyDB, dummyClient, "card1",
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid choice")
	})

	t.Run("unknown strategy returns error", func(t *testing.T) {
		err := ResolveBankCardGRPC(ctx, nil, "unknown", nil, nil, nil, dummyDB, dummyClient, "card1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown strategy")
	})

	t.Run("listClientFunc error propagates", func(t *testing.T) {
		expectedErr := errors.New("list error")
		err := ResolveBankCardGRPC(ctx, nil, "client",
			func(ctx context.Context, db *sqlx.DB) ([]models.BankCardDB, error) {
				return nil, expectedErr
			},
			nil, nil, dummyDB, dummyClient, "card1")
		require.ErrorIs(t, err, expectedErr)
	})

	t.Run("getServerFunc error propagates", func(t *testing.T) {
		expectedErr := errors.New("get error")
		err := ResolveBankCardGRPC(ctx, nil, "client",
			func(ctx context.Context, db *sqlx.DB) ([]models.BankCardDB, error) {
				return clientSecrets, nil
			},
			func(ctx context.Context, client pb.BankCardServiceClient, secretName string) (*models.BankCardDB, error) {
				return nil, expectedErr
			},
			nil, dummyDB, dummyClient, "card1",
		)
		require.ErrorIs(t, err, expectedErr)
	})

	t.Run("addServerFunc error propagates", func(t *testing.T) {
		expectedErr := errors.New("add error")
		err := ResolveBankCardGRPC(ctx, nil, "client",
			func(ctx context.Context, db *sqlx.DB) ([]models.BankCardDB, error) {
				return clientSecrets, nil
			},
			func(ctx context.Context, client pb.BankCardServiceClient, secretName string) (*models.BankCardDB, error) {
				return nil, nil
			},
			func(ctx context.Context, client pb.BankCardServiceClient, req *models.BankCardAddRequest) error {
				return expectedErr
			},
			dummyDB, dummyClient, "card1",
		)
		require.ErrorIs(t, err, expectedErr)
	})
}

func TestResolveTextHTTP(t *testing.T) {
	ctx := context.Background()

	now := time.Now()
	clientSecrets := []models.TextDB{
		{
			SecretName: "text1",
			Content:    "Hello World",
			Meta:       nil,
			UpdatedAt:  now,
		},
	}

	serverSecretNewer := &models.TextDB{
		SecretName: "text1",
		Content:    "Server New Content",
		Meta:       nil,
		UpdatedAt:  now.Add(1 * time.Hour),
	}
	serverSecretOlder := &models.TextDB{
		SecretName: "text1",
		Content:    "Server Old Content",
		Meta:       nil,
		UpdatedAt:  now.Add(-1 * time.Hour),
	}

	var dummyDB *sqlx.DB
	var dummyClient *resty.Client

	t.Run("strategy server returns nil", func(t *testing.T) {
		err := ResolveTextHTTP(ctx, bufio.NewReader(strings.NewReader("")), "server", nil, nil, nil, dummyDB, dummyClient, "")
		require.NoError(t, err)
	})

	t.Run("strategy client syncs when server secret missing", func(t *testing.T) {
		listCalled := false
		getCalled := false
		addCalled := false

		err := ResolveTextHTTP(ctx, nil, "client",
			func(ctx context.Context, db *sqlx.DB) ([]models.TextDB, error) {
				listCalled = true
				return clientSecrets, nil
			},
			func(ctx context.Context, client *resty.Client, secretName string) (*models.TextDB, error) {
				getCalled = true
				return nil, nil // server secret missing
			},
			func(ctx context.Context, client *resty.Client, req *models.TextAddRequest) error {
				addCalled = true
				assert.Equal(t, "text1", req.SecretName)
				return nil
			},
			dummyDB, dummyClient, "text1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.True(t, getCalled)
		assert.True(t, addCalled)
	})

	t.Run("strategy client syncs when client secret is newer", func(t *testing.T) {
		listCalled := false
		getCalled := false
		addCalled := false

		err := ResolveTextHTTP(ctx, nil, "client",
			func(ctx context.Context, db *sqlx.DB) ([]models.TextDB, error) {
				listCalled = true
				return clientSecrets, nil
			},
			func(ctx context.Context, client *resty.Client, secretName string) (*models.TextDB, error) {
				getCalled = true
				return serverSecretOlder, nil
			},
			func(ctx context.Context, client *resty.Client, req *models.TextAddRequest) error {
				addCalled = true
				assert.Equal(t, "text1", req.SecretName)
				return nil
			},
			dummyDB, dummyClient, "text1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.True(t, getCalled)
		assert.True(t, addCalled)
	})

	t.Run("strategy client skips when server secret is newer", func(t *testing.T) {
		listCalled := false
		getCalled := false
		addCalled := false

		err := ResolveTextHTTP(ctx, nil, "client",
			func(ctx context.Context, db *sqlx.DB) ([]models.TextDB, error) {
				listCalled = true
				return clientSecrets, nil
			},
			func(ctx context.Context, client *resty.Client, secretName string) (*models.TextDB, error) {
				getCalled = true
				return serverSecretNewer, nil
			},
			func(ctx context.Context, client *resty.Client, req *models.TextAddRequest) error {
				addCalled = true
				return nil
			},
			dummyDB, dummyClient, "text1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.True(t, getCalled)
		assert.False(t, addCalled)
	})

	t.Run("strategy interactive with server secret nil adds client secret", func(t *testing.T) {
		listCalled := false
		getCalled := false
		addCalled := false

		err := ResolveTextHTTP(ctx, bufio.NewReader(strings.NewReader("")), "interactive",
			func(ctx context.Context, db *sqlx.DB) ([]models.TextDB, error) {
				listCalled = true
				return clientSecrets, nil
			},
			func(ctx context.Context, client *resty.Client, secretName string) (*models.TextDB, error) {
				getCalled = true
				return nil, nil
			},
			func(ctx context.Context, client *resty.Client, req *models.TextAddRequest) error {
				addCalled = true
				return nil
			},
			dummyDB, dummyClient, "text1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.True(t, getCalled)
		assert.True(t, addCalled)
	})

	t.Run("strategy interactive skips when client secret not newer", func(t *testing.T) {
		listCalled := false
		getCalled := false
		addCalled := false

		clientOlder := []models.TextDB{
			{
				SecretName: "text1",
				Content:    "Old content",
				UpdatedAt:  now.Add(-2 * time.Hour),
			},
		}

		err := ResolveTextHTTP(ctx, bufio.NewReader(strings.NewReader("")), "interactive",
			func(ctx context.Context, db *sqlx.DB) ([]models.TextDB, error) {
				listCalled = true
				return clientOlder, nil
			},
			func(ctx context.Context, client *resty.Client, secretName string) (*models.TextDB, error) {
				getCalled = true
				return serverSecretNewer, nil
			},
			func(ctx context.Context, client *resty.Client, req *models.TextAddRequest) error {
				addCalled = true
				return nil
			},
			dummyDB, dummyClient, "text1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.True(t, getCalled)
		assert.False(t, addCalled)
	})

	t.Run("strategy interactive user chooses client version", func(t *testing.T) {
		listCalled := false
		getCalled := false
		addCalled := false

		input := "client\n"

		err := ResolveTextHTTP(ctx, bufio.NewReader(strings.NewReader(input)), "interactive",
			func(ctx context.Context, db *sqlx.DB) ([]models.TextDB, error) {
				listCalled = true
				return clientSecrets, nil
			},
			func(ctx context.Context, client *resty.Client, secretName string) (*models.TextDB, error) {
				getCalled = true
				return serverSecretOlder, nil
			},
			func(ctx context.Context, client *resty.Client, req *models.TextAddRequest) error {
				addCalled = true
				return nil
			},
			dummyDB, dummyClient, "text1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.True(t, getCalled)
		assert.True(t, addCalled)
	})

	t.Run("strategy interactive user chooses server version", func(t *testing.T) {
		listCalled := false
		getCalled := false
		addCalled := false

		input := "server\n"

		err := ResolveTextHTTP(ctx, bufio.NewReader(strings.NewReader(input)), "interactive",
			func(ctx context.Context, db *sqlx.DB) ([]models.TextDB, error) {
				listCalled = true
				return clientSecrets, nil
			},
			func(ctx context.Context, client *resty.Client, secretName string) (*models.TextDB, error) {
				getCalled = true
				return serverSecretOlder, nil
			},
			func(ctx context.Context, client *resty.Client, req *models.TextAddRequest) error {
				addCalled = true
				return nil
			},
			dummyDB, dummyClient, "text1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.True(t, getCalled)
		assert.False(t, addCalled)
	})

	t.Run("strategy interactive invalid user input", func(t *testing.T) {
		input := "invalid\n"

		err := ResolveTextHTTP(ctx, bufio.NewReader(strings.NewReader(input)), "interactive",
			func(ctx context.Context, db *sqlx.DB) ([]models.TextDB, error) {
				return clientSecrets, nil
			},
			func(ctx context.Context, client *resty.Client, secretName string) (*models.TextDB, error) {
				return serverSecretOlder, nil
			},
			func(ctx context.Context, client *resty.Client, req *models.TextAddRequest) error {
				return nil
			},
			dummyDB, dummyClient, "text1",
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid choice")
	})

	t.Run("unknown strategy returns error", func(t *testing.T) {
		err := ResolveTextHTTP(ctx, nil, "unknown", nil, nil, nil, dummyDB, dummyClient, "text1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown strategy")
	})
}

func TestResolveBinaryHTTP(t *testing.T) {
	ctx := context.Background()

	now := time.Now()
	clientSecrets := []models.BinaryDB{
		{
			SecretName: "binary1",
			Data:       []byte{1, 2, 3},
			Meta:       nil,
			UpdatedAt:  now,
		},
	}

	serverSecretNewer := &models.BinaryDB{
		SecretName: "binary1",
		Data:       []byte{4, 5, 6},
		Meta:       nil,
		UpdatedAt:  now.Add(1 * time.Hour),
	}
	serverSecretOlder := &models.BinaryDB{
		SecretName: "binary1",
		Data:       []byte{7, 8, 9},
		Meta:       nil,
		UpdatedAt:  now.Add(-1 * time.Hour),
	}

	var dummyDB *sqlx.DB
	var dummyClient *resty.Client

	t.Run("strategy server returns nil", func(t *testing.T) {
		err := ResolveBinaryHTTP(ctx, bufio.NewReader(strings.NewReader("")), "server", nil, nil, nil, dummyDB, dummyClient, "")
		require.NoError(t, err)
	})

	t.Run("strategy client syncs all client secrets to server", func(t *testing.T) {
		listCalled := false
		addCalls := 0

		err := ResolveBinaryHTTP(ctx, nil, "client",
			func(ctx context.Context, db *sqlx.DB) ([]models.BinaryDB, error) {
				listCalled = true
				return clientSecrets, nil
			},
			nil,
			func(ctx context.Context, client *resty.Client, req *models.BinaryAddRequest) error {
				addCalls++
				assert.Equal(t, "binary1", req.SecretName)
				assert.Equal(t, []byte{1, 2, 3}, req.Data)
				return nil
			},
			dummyDB, dummyClient, "binary1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.Equal(t, 1, addCalls)
	})

	t.Run("strategy interactive adds client secret if server secret missing", func(t *testing.T) {
		listCalled := false
		getCalled := false
		addCalled := false

		err := ResolveBinaryHTTP(ctx, bufio.NewReader(strings.NewReader("")), "interactive",
			func(ctx context.Context, db *sqlx.DB) ([]models.BinaryDB, error) {
				listCalled = true
				return clientSecrets, nil
			},
			func(ctx context.Context, client *resty.Client, secretName string) (*models.BinaryDB, error) {
				getCalled = true
				return nil, nil
			},
			func(ctx context.Context, client *resty.Client, req *models.BinaryAddRequest) error {
				addCalled = true
				return nil
			},
			dummyDB, dummyClient, "binary1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.True(t, getCalled)
		assert.True(t, addCalled)
	})

	t.Run("strategy interactive skips add when client secret not newer", func(t *testing.T) {
		listCalled := false
		getCalled := false
		addCalled := false

		clientOlder := []models.BinaryDB{
			{
				SecretName: "binary1",
				Data:       []byte{0},
				UpdatedAt:  now.Add(-2 * time.Hour),
			},
		}

		err := ResolveBinaryHTTP(ctx, bufio.NewReader(strings.NewReader("")), "interactive",
			func(ctx context.Context, db *sqlx.DB) ([]models.BinaryDB, error) {
				listCalled = true
				return clientOlder, nil
			},
			func(ctx context.Context, client *resty.Client, secretName string) (*models.BinaryDB, error) {
				getCalled = true
				return serverSecretNewer, nil
			},
			func(ctx context.Context, client *resty.Client, req *models.BinaryAddRequest) error {
				addCalled = true
				return nil
			},
			dummyDB, dummyClient, "binary1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.True(t, getCalled)
		assert.False(t, addCalled)
	})

	t.Run("strategy interactive user chooses client version", func(t *testing.T) {
		listCalled := false
		getCalled := false
		addCalled := false

		input := "client\n"

		err := ResolveBinaryHTTP(ctx, bufio.NewReader(strings.NewReader(input)), "interactive",
			func(ctx context.Context, db *sqlx.DB) ([]models.BinaryDB, error) {
				listCalled = true
				return clientSecrets, nil
			},
			func(ctx context.Context, client *resty.Client, secretName string) (*models.BinaryDB, error) {
				getCalled = true
				return serverSecretOlder, nil
			},
			func(ctx context.Context, client *resty.Client, req *models.BinaryAddRequest) error {
				addCalled = true
				return nil
			},
			dummyDB, dummyClient, "binary1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.True(t, getCalled)
		assert.True(t, addCalled)
	})

	t.Run("strategy interactive user chooses server version", func(t *testing.T) {
		listCalled := false
		getCalled := false
		addCalled := false

		input := "server\n"

		err := ResolveBinaryHTTP(ctx, bufio.NewReader(strings.NewReader(input)), "interactive",
			func(ctx context.Context, db *sqlx.DB) ([]models.BinaryDB, error) {
				listCalled = true
				return clientSecrets, nil
			},
			func(ctx context.Context, client *resty.Client, secretName string) (*models.BinaryDB, error) {
				getCalled = true
				return serverSecretOlder, nil
			},
			func(ctx context.Context, client *resty.Client, req *models.BinaryAddRequest) error {
				addCalled = true
				return nil
			},
			dummyDB, dummyClient, "binary1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.True(t, getCalled)
		assert.False(t, addCalled)
	})

	t.Run("strategy interactive invalid user input returns error", func(t *testing.T) {
		input := "invalid\n"

		err := ResolveBinaryHTTP(ctx, bufio.NewReader(strings.NewReader(input)), "interactive",
			func(ctx context.Context, db *sqlx.DB) ([]models.BinaryDB, error) {
				return clientSecrets, nil
			},
			func(ctx context.Context, client *resty.Client, secretName string) (*models.BinaryDB, error) {
				return serverSecretOlder, nil
			},
			func(ctx context.Context, client *resty.Client, req *models.BinaryAddRequest) error {
				return nil
			},
			dummyDB, dummyClient, "binary1",
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid choice")
	})

	t.Run("unknown strategy returns error", func(t *testing.T) {
		err := ResolveBinaryHTTP(ctx, nil, "unknown", nil, nil, nil, dummyDB, dummyClient, "binary1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown strategy")
	})
}

func TestResolveUserHTTP(t *testing.T) {
	ctx := context.Background()

	now := time.Now()
	clientSecrets := []models.UserDB{
		{
			SecretName:  "user1",
			SecretOwner: "owner1",
			Username:    "alice",
			Password:    "password123",
			Meta:        nil,
			UpdatedAt:   now,
		},
	}

	serverSecretNewer := &models.UserDB{
		SecretName:  "user1",
		SecretOwner: "owner1",
		Username:    "alice_server",
		Password:    "serverpass",
		Meta:        nil,
		UpdatedAt:   now.Add(1 * time.Hour),
	}
	serverSecretOlder := &models.UserDB{
		SecretName:  "user1",
		SecretOwner: "owner1",
		Username:    "alice_old",
		Password:    "oldpass",
		Meta:        nil,
		UpdatedAt:   now.Add(-1 * time.Hour),
	}

	var dummyDB *sqlx.DB
	var dummyClient *resty.Client

	t.Run("strategy server returns nil", func(t *testing.T) {
		err := ResolveUserHTTP(ctx, bufio.NewReader(strings.NewReader("")), "server", nil, nil, nil, dummyDB, dummyClient, "")
		require.NoError(t, err)
	})

	t.Run("strategy client syncs all client secrets to server", func(t *testing.T) {
		listCalled := false
		addCalls := 0

		err := ResolveUserHTTP(ctx, nil, "client",
			func(ctx context.Context, db *sqlx.DB) ([]models.UserDB, error) {
				listCalled = true
				return clientSecrets, nil
			},
			nil,
			func(ctx context.Context, client *resty.Client, req *models.UserAddRequest) error {
				addCalls++
				assert.Equal(t, "user1", req.SecretName)
				assert.Equal(t, "alice", req.Username)
				assert.Equal(t, "password123", req.Password)
				return nil
			},
			dummyDB, dummyClient, "user1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.Equal(t, 1, addCalls)
	})

	t.Run("strategy interactive adds client secret if server secret missing", func(t *testing.T) {
		listCalled := false
		getCalled := false
		addCalled := false

		err := ResolveUserHTTP(ctx, bufio.NewReader(strings.NewReader("")), "interactive",
			func(ctx context.Context, db *sqlx.DB) ([]models.UserDB, error) {
				listCalled = true
				return clientSecrets, nil
			},
			func(ctx context.Context, client *resty.Client, secretName string) (*models.UserDB, error) {
				getCalled = true
				return nil, nil
			},
			func(ctx context.Context, client *resty.Client, req *models.UserAddRequest) error {
				addCalled = true
				return nil
			},
			dummyDB, dummyClient, "user1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.True(t, getCalled)
		assert.True(t, addCalled)
	})

	t.Run("strategy interactive skips add when client secret not newer", func(t *testing.T) {
		listCalled := false
		getCalled := false
		addCalled := false

		clientOlder := []models.UserDB{
			{
				SecretName:  "user1",
				SecretOwner: "owner1",
				Username:    "oldclient",
				Password:    "oldpass",
				UpdatedAt:   now.Add(-2 * time.Hour),
			},
		}

		err := ResolveUserHTTP(ctx, bufio.NewReader(strings.NewReader("")), "interactive",
			func(ctx context.Context, db *sqlx.DB) ([]models.UserDB, error) {
				listCalled = true
				return clientOlder, nil
			},
			func(ctx context.Context, client *resty.Client, secretName string) (*models.UserDB, error) {
				getCalled = true
				return serverSecretNewer, nil
			},
			func(ctx context.Context, client *resty.Client, req *models.UserAddRequest) error {
				addCalled = true
				return nil
			},
			dummyDB, dummyClient, "user1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.True(t, getCalled)
		assert.False(t, addCalled)
	})

	t.Run("strategy interactive user chooses client version", func(t *testing.T) {
		listCalled := false
		getCalled := false
		addCalled := false

		input := "client\n"

		err := ResolveUserHTTP(ctx, bufio.NewReader(strings.NewReader(input)), "interactive",
			func(ctx context.Context, db *sqlx.DB) ([]models.UserDB, error) {
				listCalled = true
				return clientSecrets, nil
			},
			func(ctx context.Context, client *resty.Client, secretName string) (*models.UserDB, error) {
				getCalled = true
				return serverSecretOlder, nil
			},
			func(ctx context.Context, client *resty.Client, req *models.UserAddRequest) error {
				addCalled = true
				return nil
			},
			dummyDB, dummyClient, "user1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.True(t, getCalled)
		assert.True(t, addCalled)
	})

	t.Run("strategy interactive user chooses server version", func(t *testing.T) {
		listCalled := false
		getCalled := false
		addCalled := false

		input := "server\n"

		err := ResolveUserHTTP(ctx, bufio.NewReader(strings.NewReader(input)), "interactive",
			func(ctx context.Context, db *sqlx.DB) ([]models.UserDB, error) {
				listCalled = true
				return clientSecrets, nil
			},
			func(ctx context.Context, client *resty.Client, secretName string) (*models.UserDB, error) {
				getCalled = true
				return serverSecretOlder, nil
			},
			func(ctx context.Context, client *resty.Client, req *models.UserAddRequest) error {
				addCalled = true
				return nil
			},
			dummyDB, dummyClient, "user1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.True(t, getCalled)
		assert.False(t, addCalled)
	})

	t.Run("strategy interactive invalid user input returns error", func(t *testing.T) {
		input := "invalid\n"

		err := ResolveUserHTTP(ctx, bufio.NewReader(strings.NewReader(input)), "interactive",
			func(ctx context.Context, db *sqlx.DB) ([]models.UserDB, error) {
				return clientSecrets, nil
			},
			func(ctx context.Context, client *resty.Client, secretName string) (*models.UserDB, error) {
				return serverSecretOlder, nil
			},
			func(ctx context.Context, client *resty.Client, req *models.UserAddRequest) error {
				return nil
			},
			dummyDB, dummyClient, "user1",
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid choice")
	})

	t.Run("unknown strategy returns error", func(t *testing.T) {
		err := ResolveUserHTTP(ctx, nil, "unknown", nil, nil, nil, dummyDB, dummyClient, "user1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown strategy")
	})
}

type mockTextServiceClient struct{}

func (m *mockTextServiceClient) Add(ctx context.Context, in *pb.TextAddRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (m *mockTextServiceClient) Get(ctx context.Context, in *pb.TextFilterRequest, opts ...grpc.CallOption) (*pb.TextDB, error) {
	return nil, nil
}

func (m *mockTextServiceClient) List(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (pb.TextService_ListClient, error) {
	return nil, nil
}

func TestResolveTextGRPC(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	clientSecrets := []models.TextDB{
		{
			SecretName: "text1",
			Content:    "Hello World",
			Meta:       nil,
			UpdatedAt:  now,
		},
	}

	serverSecretNewer := &models.TextDB{
		SecretName: "text1",
		Content:    "Hello from server",
		Meta:       nil,
		UpdatedAt:  now.Add(1 * time.Hour),
	}

	serverSecretOlder := &models.TextDB{
		SecretName: "text1",
		Content:    "Old server text",
		Meta:       nil,
		UpdatedAt:  now.Add(-1 * time.Hour),
	}

	var dummyDB *sqlx.DB
	var dummyClient pb.TextServiceClient = &mockTextServiceClient{}

	t.Run("strategy client syncs when server secret missing", func(t *testing.T) {
		listCalled, addCalled := false, false

		err := ResolveTextGRPC(ctx, nil, "client",
			func(ctx context.Context, db *sqlx.DB) ([]models.TextDB, error) {
				listCalled = true
				return clientSecrets, nil
			},
			func(ctx context.Context, client pb.TextServiceClient, secretName string) (*models.TextDB, error) {
				return nil, nil // server secret missing
			},
			func(ctx context.Context, client pb.TextServiceClient, req *models.TextAddRequest) error {
				addCalled = true
				assert.Equal(t, "text1", req.SecretName)
				assert.Equal(t, "Hello World", req.Content)
				return nil
			},
			dummyDB, dummyClient, "text1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.True(t, addCalled)
	})

	t.Run("strategy client syncs when client secret is newer", func(t *testing.T) {
		listCalled, addCalled := false, false

		err := ResolveTextGRPC(ctx, nil, "client",
			func(ctx context.Context, db *sqlx.DB) ([]models.TextDB, error) {
				listCalled = true
				return clientSecrets, nil
			},
			func(ctx context.Context, client pb.TextServiceClient, secretName string) (*models.TextDB, error) {
				return serverSecretOlder, nil
			},
			func(ctx context.Context, client pb.TextServiceClient, req *models.TextAddRequest) error {
				addCalled = true
				return nil
			},
			dummyDB, dummyClient, "text1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.True(t, addCalled)
	})

	t.Run("strategy client skips when server secret is newer", func(t *testing.T) {
		listCalled, addCalled := false, false

		err := ResolveTextGRPC(ctx, nil, "client",
			func(ctx context.Context, db *sqlx.DB) ([]models.TextDB, error) {
				listCalled = true
				return clientSecrets, nil
			},
			func(ctx context.Context, client pb.TextServiceClient, secretName string) (*models.TextDB, error) {
				return serverSecretNewer, nil
			},
			func(ctx context.Context, client pb.TextServiceClient, req *models.TextAddRequest) error {
				addCalled = true
				return nil
			},
			dummyDB, dummyClient, "text1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.False(t, addCalled)
	})

	t.Run("strategy interactive user chooses client version", func(t *testing.T) {
		listCalled, addCalled := false, false

		input := "client\n"

		err := ResolveTextGRPC(ctx, bufio.NewReader(strings.NewReader(input)), "interactive",
			func(ctx context.Context, db *sqlx.DB) ([]models.TextDB, error) {
				listCalled = true
				return clientSecrets, nil
			},
			func(ctx context.Context, client pb.TextServiceClient, secretName string) (*models.TextDB, error) {
				return serverSecretOlder, nil
			},
			func(ctx context.Context, client pb.TextServiceClient, req *models.TextAddRequest) error {
				addCalled = true
				return nil
			},
			dummyDB, dummyClient, "text1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.True(t, addCalled)
	})

	t.Run("strategy interactive user chooses server version", func(t *testing.T) {
		listCalled, addCalled := false, false

		input := "server\n"

		err := ResolveTextGRPC(ctx, bufio.NewReader(strings.NewReader(input)), "interactive",
			func(ctx context.Context, db *sqlx.DB) ([]models.TextDB, error) {
				listCalled = true
				return clientSecrets, nil
			},
			func(ctx context.Context, client pb.TextServiceClient, secretName string) (*models.TextDB, error) {
				return serverSecretOlder, nil
			},
			func(ctx context.Context, client pb.TextServiceClient, req *models.TextAddRequest) error {
				addCalled = true
				return nil
			},
			dummyDB, dummyClient, "text1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.False(t, addCalled)
	})

	t.Run("strategy interactive invalid user input", func(t *testing.T) {
		input := "invalid\n"

		err := ResolveTextGRPC(ctx, bufio.NewReader(strings.NewReader(input)), "interactive",
			func(ctx context.Context, db *sqlx.DB) ([]models.TextDB, error) {
				return clientSecrets, nil
			},
			func(ctx context.Context, client pb.TextServiceClient, secretName string) (*models.TextDB, error) {
				return serverSecretOlder, nil
			},
			func(ctx context.Context, client pb.TextServiceClient, req *models.TextAddRequest) error {
				return nil
			},
			dummyDB, dummyClient, "text1",
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid choice")
	})

	t.Run("unknown strategy returns error", func(t *testing.T) {
		err := ResolveTextGRPC(ctx, nil, "unknown", nil, nil, nil, dummyDB, dummyClient, "text1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown strategy")
	})

	t.Run("listClientFunc error propagates", func(t *testing.T) {
		expectedErr := errors.New("list error")
		err := ResolveTextGRPC(ctx, nil, "client",
			func(ctx context.Context, db *sqlx.DB) ([]models.TextDB, error) {
				return nil, expectedErr
			},
			nil, nil, dummyDB, dummyClient, "text1")
		require.ErrorIs(t, err, expectedErr)
	})

}

type mockBinaryServiceClient struct{}

func (m *mockBinaryServiceClient) Add(ctx context.Context, in *pb.BinaryAddRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (m *mockBinaryServiceClient) Get(ctx context.Context, in *pb.BinaryFilterRequest, opts ...grpc.CallOption) (*pb.BinaryDB, error) {
	return nil, nil
}

func (m *mockBinaryServiceClient) List(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (pb.BinaryService_ListClient, error) {
	return nil, nil
}

func TestResolveBinaryGRPC(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	clientSecrets := []models.BinaryDB{
		{
			SecretName: "binary1",
			Data:       []byte{0x01, 0x02},
			Meta:       nil,
			UpdatedAt:  now,
		},
	}

	serverSecretNewer := &models.BinaryDB{
		SecretName: "binary1",
		Data:       []byte{0x03, 0x04},
		Meta:       nil,
		UpdatedAt:  now.Add(1 * time.Hour),
	}

	serverSecretOlder := &models.BinaryDB{
		SecretName: "binary1",
		Data:       []byte{0x05, 0x06},
		Meta:       nil,
		UpdatedAt:  now.Add(-1 * time.Hour),
	}

	var dummyDB *sqlx.DB
	var dummyClient pb.BinaryServiceClient = &mockBinaryServiceClient{}

	t.Run("strategy client syncs when server secret missing", func(t *testing.T) {
		listCalled, addCalled := false, false

		err := ResolveBinaryGRPC(ctx, nil, "client",
			func(ctx context.Context, db *sqlx.DB) ([]models.BinaryDB, error) {
				listCalled = true
				return clientSecrets, nil
			},
			func(ctx context.Context, client pb.BinaryServiceClient, secretName string) (*models.BinaryDB, error) {
				return nil, nil // server secret missing
			},
			func(ctx context.Context, client pb.BinaryServiceClient, req *models.BinaryAddRequest) error {
				addCalled = true
				assert.Equal(t, "binary1", req.SecretName)
				assert.Equal(t, []byte{0x01, 0x02}, req.Data)
				return nil
			},
			dummyDB, dummyClient, "binary1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.True(t, addCalled)
	})

	t.Run("strategy client syncs when client secret is newer", func(t *testing.T) {
		listCalled, addCalled := false, false

		err := ResolveBinaryGRPC(ctx, nil, "client",
			func(ctx context.Context, db *sqlx.DB) ([]models.BinaryDB, error) {
				listCalled = true
				return clientSecrets, nil
			},
			func(ctx context.Context, client pb.BinaryServiceClient, secretName string) (*models.BinaryDB, error) {
				return serverSecretOlder, nil
			},
			func(ctx context.Context, client pb.BinaryServiceClient, req *models.BinaryAddRequest) error {
				addCalled = true
				return nil
			},
			dummyDB, dummyClient, "binary1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.True(t, addCalled)
	})

	t.Run("strategy client skips when server secret is newer", func(t *testing.T) {
		listCalled, addCalled := false, false

		err := ResolveBinaryGRPC(ctx, nil, "client",
			func(ctx context.Context, db *sqlx.DB) ([]models.BinaryDB, error) {
				listCalled = true
				return clientSecrets, nil
			},
			func(ctx context.Context, client pb.BinaryServiceClient, secretName string) (*models.BinaryDB, error) {
				return serverSecretNewer, nil
			},
			func(ctx context.Context, client pb.BinaryServiceClient, req *models.BinaryAddRequest) error {
				addCalled = true
				return nil
			},
			dummyDB, dummyClient, "binary1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.False(t, addCalled)
	})

	t.Run("strategy interactive user chooses client version", func(t *testing.T) {
		listCalled, addCalled := false, false

		input := "client\n"

		err := ResolveBinaryGRPC(ctx, bufio.NewReader(strings.NewReader(input)), "interactive",
			func(ctx context.Context, db *sqlx.DB) ([]models.BinaryDB, error) {
				listCalled = true
				return clientSecrets, nil
			},
			func(ctx context.Context, client pb.BinaryServiceClient, secretName string) (*models.BinaryDB, error) {
				return serverSecretOlder, nil
			},
			func(ctx context.Context, client pb.BinaryServiceClient, req *models.BinaryAddRequest) error {
				addCalled = true
				return nil
			},
			dummyDB, dummyClient, "binary1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.True(t, addCalled)
	})

	t.Run("strategy interactive user chooses server version", func(t *testing.T) {
		listCalled, addCalled := false, false

		input := "server\n"

		err := ResolveBinaryGRPC(ctx, bufio.NewReader(strings.NewReader(input)), "interactive",
			func(ctx context.Context, db *sqlx.DB) ([]models.BinaryDB, error) {
				listCalled = true
				return clientSecrets, nil
			},
			func(ctx context.Context, client pb.BinaryServiceClient, secretName string) (*models.BinaryDB, error) {
				return serverSecretOlder, nil
			},
			func(ctx context.Context, client pb.BinaryServiceClient, req *models.BinaryAddRequest) error {
				addCalled = true
				return nil
			},
			dummyDB, dummyClient, "binary1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.False(t, addCalled)
	})

	t.Run("strategy interactive invalid user input", func(t *testing.T) {
		input := "invalid\n"

		err := ResolveBinaryGRPC(ctx, bufio.NewReader(strings.NewReader(input)), "interactive",
			func(ctx context.Context, db *sqlx.DB) ([]models.BinaryDB, error) {
				return clientSecrets, nil
			},
			func(ctx context.Context, client pb.BinaryServiceClient, secretName string) (*models.BinaryDB, error) {
				return serverSecretOlder, nil
			},
			func(ctx context.Context, client pb.BinaryServiceClient, req *models.BinaryAddRequest) error {
				return nil
			},
			dummyDB, dummyClient, "binary1",
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid choice")
	})

	t.Run("unknown strategy returns error", func(t *testing.T) {
		err := ResolveBinaryGRPC(ctx, nil, "unknown", nil, nil, nil, dummyDB, dummyClient, "binary1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown strategy")
	})

	t.Run("listClientFunc error propagates", func(t *testing.T) {
		expectedErr := errors.New("list error")
		err := ResolveBinaryGRPC(ctx, nil, "client",
			func(ctx context.Context, db *sqlx.DB) ([]models.BinaryDB, error) {
				return nil, expectedErr
			},
			nil, nil, dummyDB, dummyClient, "binary1")
		require.ErrorIs(t, err, expectedErr)
	})
}

type mockUserServiceClient struct{}

func (m *mockUserServiceClient) Add(ctx context.Context, in *pb.UserAddRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (m *mockUserServiceClient) Get(ctx context.Context, in *pb.UserFilterRequest, opts ...grpc.CallOption) (*pb.UserDB, error) {
	return nil, nil
}

func (m *mockUserServiceClient) List(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (pb.UserService_ListClient, error) {
	return nil, nil
}

func TestResolveUserGRPC(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	clientUsers := []models.UserDB{
		{
			SecretName: "user1",
			Username:   "alice",
			Password:   "pass123",
			Meta:       nil,
			UpdatedAt:  now,
		},
	}

	serverUserNewer := &models.UserDB{
		SecretName: "user1",
		Username:   "alice",
		Password:   "newpass456",
		Meta:       nil,
		UpdatedAt:  now.Add(1 * time.Hour),
	}

	serverUserOlder := &models.UserDB{
		SecretName: "user1",
		Username:   "alice",
		Password:   "oldpass789",
		Meta:       nil,
		UpdatedAt:  now.Add(-1 * time.Hour),
	}

	var dummyDB *sqlx.DB
	var dummyClient pb.UserServiceClient = &mockUserServiceClient{}

	t.Run("strategy client syncs when server user missing", func(t *testing.T) {
		listCalled, addCalled := false, false

		err := ResolveUserGRPC(ctx, nil, "client",
			func(ctx context.Context, db *sqlx.DB) ([]models.UserDB, error) {
				listCalled = true
				return clientUsers, nil
			},
			func(ctx context.Context, client pb.UserServiceClient, secretName string) (*models.UserDB, error) {
				return nil, nil // server user missing
			},
			func(ctx context.Context, client pb.UserServiceClient, req *models.UserAddRequest) error {
				addCalled = true
				assert.Equal(t, "user1", req.SecretName)
				assert.Equal(t, "alice", req.Username)
				assert.Equal(t, "pass123", req.Password)
				return nil
			},
			dummyDB, dummyClient, "user1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.True(t, addCalled)
	})

	t.Run("strategy client syncs when client user is newer", func(t *testing.T) {
		listCalled, addCalled := false, false

		err := ResolveUserGRPC(ctx, nil, "client",
			func(ctx context.Context, db *sqlx.DB) ([]models.UserDB, error) {
				listCalled = true
				return clientUsers, nil
			},
			func(ctx context.Context, client pb.UserServiceClient, secretName string) (*models.UserDB, error) {
				return serverUserOlder, nil
			},
			func(ctx context.Context, client pb.UserServiceClient, req *models.UserAddRequest) error {
				addCalled = true
				return nil
			},
			dummyDB, dummyClient, "user1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.True(t, addCalled)
	})

	t.Run("strategy client skips when server user is newer", func(t *testing.T) {
		listCalled, addCalled := false, false

		err := ResolveUserGRPC(ctx, nil, "client",
			func(ctx context.Context, db *sqlx.DB) ([]models.UserDB, error) {
				listCalled = true
				return clientUsers, nil
			},
			func(ctx context.Context, client pb.UserServiceClient, secretName string) (*models.UserDB, error) {
				return serverUserNewer, nil
			},
			func(ctx context.Context, client pb.UserServiceClient, req *models.UserAddRequest) error {
				addCalled = true
				return nil
			},
			dummyDB, dummyClient, "user1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.False(t, addCalled)
	})

	t.Run("strategy interactive user chooses client version", func(t *testing.T) {
		listCalled, addCalled := false, false

		input := "client\n"

		err := ResolveUserGRPC(ctx, bufio.NewReader(strings.NewReader(input)), "interactive",
			func(ctx context.Context, db *sqlx.DB) ([]models.UserDB, error) {
				listCalled = true
				return clientUsers, nil
			},
			func(ctx context.Context, client pb.UserServiceClient, secretName string) (*models.UserDB, error) {
				return serverUserOlder, nil
			},
			func(ctx context.Context, client pb.UserServiceClient, req *models.UserAddRequest) error {
				addCalled = true
				return nil
			},
			dummyDB, dummyClient, "user1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.True(t, addCalled)
	})

	t.Run("strategy interactive user chooses server version", func(t *testing.T) {
		listCalled, addCalled := false, false

		input := "server\n"

		err := ResolveUserGRPC(ctx, bufio.NewReader(strings.NewReader(input)), "interactive",
			func(ctx context.Context, db *sqlx.DB) ([]models.UserDB, error) {
				listCalled = true
				return clientUsers, nil
			},
			func(ctx context.Context, client pb.UserServiceClient, secretName string) (*models.UserDB, error) {
				return serverUserOlder, nil
			},
			func(ctx context.Context, client pb.UserServiceClient, req *models.UserAddRequest) error {
				addCalled = true
				return nil
			},
			dummyDB, dummyClient, "user1",
		)
		require.NoError(t, err)
		assert.True(t, listCalled)
		assert.False(t, addCalled)
	})

	t.Run("strategy interactive invalid user input", func(t *testing.T) {
		input := "invalid\n"

		err := ResolveUserGRPC(ctx, bufio.NewReader(strings.NewReader(input)), "interactive",
			func(ctx context.Context, db *sqlx.DB) ([]models.UserDB, error) {
				return clientUsers, nil
			},
			func(ctx context.Context, client pb.UserServiceClient, secretName string) (*models.UserDB, error) {
				return serverUserOlder, nil
			},
			func(ctx context.Context, client pb.UserServiceClient, req *models.UserAddRequest) error {
				return nil
			},
			dummyDB, dummyClient, "user1",
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid choice")
	})

	t.Run("unknown strategy returns error", func(t *testing.T) {
		err := ResolveUserGRPC(ctx, nil, "unknown", nil, nil, nil, dummyDB, dummyClient, "user1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown strategy")
	})

	t.Run("listClientFunc error propagates", func(t *testing.T) {
		expectedErr := errors.New("list error")
		err := ResolveUserGRPC(ctx, nil, "client",
			func(ctx context.Context, db *sqlx.DB) ([]models.UserDB, error) {
				return nil, expectedErr
			},
			nil, nil, dummyDB, dummyClient, "user1")
		require.ErrorIs(t, err, expectedErr)
	})
}
