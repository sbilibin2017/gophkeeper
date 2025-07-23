package resolver

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/models"
)

type mockListerText struct {
	listFunc func(ctx context.Context) ([]*models.Text, error)
}

func (m *mockListerText) List(ctx context.Context) ([]*models.Text, error) {
	return m.listFunc(ctx)
}

type mockGetterText struct {
	getFunc func(ctx context.Context, secretName string) (*models.Text, error)
}

func (m *mockGetterText) Get(ctx context.Context, secretName string) (*models.Text, error) {
	return m.getFunc(ctx, secretName)
}

type mockSaverText struct {
	saveFunc func(ctx context.Context, secret *models.Text) error
}

func (m *mockSaverText) Save(ctx context.Context, secret *models.Text) error {
	return m.saveFunc(ctx, secret)
}

func TestResolver_Resolve_ClientMode(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	clientSecrets := []*models.Text{
		{SecretName: "secret1", UpdatedAt: now},
		{SecretName: "secret2", UpdatedAt: now.Add(time.Minute)},
	}

	serverSecrets := map[string]*models.Text{
		"secret1": {SecretName: "secret1", UpdatedAt: now.Add(-time.Hour)}, // older on server
		"secret2": {SecretName: "secret2", UpdatedAt: now.Add(time.Hour)},  // newer on server
	}

	var saved []*models.Text

	lister := &mockListerText{
		listFunc: func(ctx context.Context) ([]*models.Text, error) {
			return clientSecrets, nil
		},
	}

	getter := &mockGetterText{
		getFunc: func(ctx context.Context, secretName string) (*models.Text, error) {
			sec, ok := serverSecrets[secretName]
			if !ok {
				return nil, errors.New("not found")
			}
			return sec, nil
		},
	}

	saver := &mockSaverText{
		saveFunc: func(ctx context.Context, secret *models.Text) error {
			saved = append(saved, secret)
			return nil
		},
	}

	resolver := NewResolver[*models.Text](lister, getter, saver, nil)

	err := resolver.Resolve(ctx, "client")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	// Only secret1 should be saved (client newer than server)
	if len(saved) != 1 {
		t.Fatalf("expected 1 saved secret, got %d", len(saved))
	}
	if saved[0].SecretName != "secret1" {
		t.Errorf("expected secret1 to be saved, got %s", saved[0].SecretName)
	}
}

func TestResolver_Resolve_InteractiveMode_KeepClient(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	clientSecret := &models.Text{SecretName: "secret1", UpdatedAt: now}
	serverSecret := &models.Text{SecretName: "secret1", UpdatedAt: now.Add(-time.Minute)}

	lister := &mockListerText{
		listFunc: func(ctx context.Context) ([]*models.Text, error) {
			return []*models.Text{clientSecret}, nil
		},
	}

	getter := &mockGetterText{
		getFunc: func(ctx context.Context, secretName string) (*models.Text, error) {
			return serverSecret, nil
		},
	}

	var saved []*models.Text
	saver := &mockSaverText{
		saveFunc: func(ctx context.Context, secret *models.Text) error {
			saved = append(saved, secret)
			return nil
		},
	}

	// simulate user inputs "1\n" to choose client version
	input := bytes.NewBufferString("1\n")

	resolver := NewResolver[*models.Text](lister, getter, saver, input)

	err := resolver.Resolve(ctx, "interactive")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	if len(saved) != 1 {
		t.Fatalf("expected 1 saved secret, got %d", len(saved))
	}

	if saved[0] != clientSecret {
		t.Errorf("expected saved secret to be clientSecret")
	}
}

func TestResolver_Resolve_InteractiveMode_KeepServer(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	clientSecret := &models.Text{SecretName: "secret1", UpdatedAt: now}
	serverSecret := &models.Text{SecretName: "secret1", UpdatedAt: now.Add(-time.Minute)}

	lister := &mockListerText{
		listFunc: func(ctx context.Context) ([]*models.Text, error) {
			return []*models.Text{clientSecret}, nil
		},
	}

	getter := &mockGetterText{
		getFunc: func(ctx context.Context, secretName string) (*models.Text, error) {
			return serverSecret, nil
		},
	}

	saver := &mockSaverText{
		saveFunc: func(ctx context.Context, secret *models.Text) error {
			t.Fatal("should not save when choosing server version")
			return nil
		},
	}

	// simulate user inputs "2\n" to choose server version (do nothing)
	input := bytes.NewBufferString("2\n")

	resolver := NewResolver[*models.Text](lister, getter, saver, input)

	err := resolver.Resolve(ctx, "interactive")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
}

func TestResolver_Resolve_InteractiveMode_InvalidChoice(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	clientSecret := &models.Text{SecretName: "secret1", UpdatedAt: now}
	serverSecret := &models.Text{SecretName: "secret1", UpdatedAt: now.Add(-time.Minute)}

	lister := &mockListerText{
		listFunc: func(ctx context.Context) ([]*models.Text, error) {
			return []*models.Text{clientSecret}, nil
		},
	}

	getter := &mockGetterText{
		getFunc: func(ctx context.Context, secretName string) (*models.Text, error) {
			return serverSecret, nil
		},
	}

	saver := &mockSaverText{
		saveFunc: func(ctx context.Context, secret *models.Text) error {
			t.Fatal("should not save on invalid input")
			return nil
		},
	}

	// simulate user inputs invalid choice "3\n"
	input := bytes.NewBufferString("3\n")

	resolver := NewResolver[*models.Text](lister, getter, saver, input)

	err := resolver.Resolve(ctx, "interactive")
	if err == nil {
		t.Fatal("expected error for invalid input")
	}
	if err.Error() != "invalid version" {
		t.Fatalf("expected 'invalid version' error, got %v", err)
	}
}

func TestResolve_ListError(t *testing.T) {
	lister := &mockListerText{
		listFunc: func(ctx context.Context) ([]*models.Text, error) {
			return nil, errors.New("list error")
		},
	}
	getter := &mockGetterText{}
	saver := &mockSaverText{}

	r := NewResolver[*models.Text](lister, getter, saver, nil)
	err := r.Resolve(context.Background(), models.SyncModeClient)
	if err == nil || err.Error() != "list error" {
		t.Fatalf("expected 'list error', got %v", err)
	}
}

func TestResolve_GetError(t *testing.T) {
	lister := &mockListerText{
		listFunc: func(ctx context.Context) ([]*models.Text, error) {
			return []*models.Text{{SecretName: "secret1"}}, nil
		},
	}
	getter := &mockGetterText{
		getFunc: func(ctx context.Context, secretName string) (*models.Text, error) {
			return nil, errors.New("get error")
		},
	}
	saver := &mockSaverText{}

	r := NewResolver[*models.Text](lister, getter, saver, nil)
	err := r.Resolve(context.Background(), models.SyncModeClient)
	if err == nil || err.Error() != "get error" {
		t.Fatalf("expected 'get error', got %v", err)
	}
}

func TestResolve_SaveError(t *testing.T) {
	now := time.Now()

	lister := &mockListerText{
		listFunc: func(ctx context.Context) ([]*models.Text, error) {
			return []*models.Text{{SecretName: "secret1", UpdatedAt: now}}, nil
		},
	}
	getter := &mockGetterText{
		getFunc: func(ctx context.Context, secretName string) (*models.Text, error) {
			return &models.Text{SecretName: secretName, UpdatedAt: now.Add(-time.Hour)}, nil
		},
	}
	saver := &mockSaverText{
		saveFunc: func(ctx context.Context, secret *models.Text) error {
			return errors.New("save error")
		},
	}

	r := NewResolver[*models.Text](lister, getter, saver, nil)
	err := r.Resolve(context.Background(), models.SyncModeClient)
	if err == nil || err.Error() != "failed to save client secret: save error" {
		t.Fatalf("expected save error wrapped, got %v", err)
	}
}

func TestResolve_InteractiveReadInputError(t *testing.T) {
	now := time.Now()

	lister := &mockListerText{
		listFunc: func(ctx context.Context) ([]*models.Text, error) {
			return []*models.Text{{SecretName: "secret1", UpdatedAt: now}}, nil
		},
	}
	getter := &mockGetterText{
		getFunc: func(ctx context.Context, secretName string) (*models.Text, error) {
			return &models.Text{SecretName: secretName, UpdatedAt: now.Add(-time.Hour)}, nil
		},
	}
	saver := &mockSaverText{}

	// Empty input to simulate scanner.Scan() == false (EOF)
	r := NewResolver[*models.Text](lister, getter, saver, bytes.NewBuffer(nil))
	err := r.Resolve(context.Background(), models.SyncModeInteractive)
	if err == nil || err.Error() != "failed to read input" {
		t.Fatalf("expected 'failed to read input' error, got %v", err)
	}
}

func TestResolve_InteractiveInvalidChoice(t *testing.T) {
	now := time.Now()

	lister := &mockListerText{
		listFunc: func(ctx context.Context) ([]*models.Text, error) {
			return []*models.Text{{SecretName: "secret1", UpdatedAt: now}}, nil
		},
	}
	getter := &mockGetterText{
		getFunc: func(ctx context.Context, secretName string) (*models.Text, error) {
			return &models.Text{SecretName: secretName, UpdatedAt: now.Add(-time.Hour)}, nil
		},
	}
	saver := &mockSaverText{}

	input := bytes.NewBufferString("invalid\n")

	r := NewResolver[*models.Text](lister, getter, saver, input)
	err := r.Resolve(context.Background(), models.SyncModeInteractive)
	if err == nil || err.Error() != "invalid version" {
		t.Fatalf("expected 'invalid version' error, got %v", err)
	}
}

func TestResolve_ServerMode_ReturnsNil(t *testing.T) {
	r := NewResolver[*models.Text](nil, nil, nil, nil)
	err := r.Resolve(context.Background(), models.SyncModeServer)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestResolve_UnknownMode(t *testing.T) {
	lister := &mockListerText{}
	getter := &mockGetterText{}
	saver := &mockSaverText{}

	r := NewResolver[*models.Text](lister, getter, saver, nil)
	err := r.Resolve(context.Background(), "unknown_mode")
	if err == nil {
		t.Fatal("expected error for unknown sync mode, got nil")
	}

	expected := "unknown sync mode: unknown_mode"
	if err.Error() != expected {
		t.Fatalf("expected error %q, got %q", expected, err.Error())
	}
}
