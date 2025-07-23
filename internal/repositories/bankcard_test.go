package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"

	"github.com/stretchr/testify/require"

	"github.com/sbilibin2017/gophkeeper/internal/models"
)

func prepareTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Connect("sqlite", ":memory:")
	require.NoError(t, err)

	schema := `
	CREATE TABLE bankcards (
		secret_name TEXT PRIMARY KEY,
		secret_owner TEXT NOT NULL,
		number TEXT NOT NULL,
		owner TEXT NOT NULL,
		exp TEXT NOT NULL,
		cvv TEXT NOT NULL,
		meta TEXT,
		updated_at DATETIME NOT NULL
	);
	`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	return db
}

func TestBankCardWriteAndReadRepositories(t *testing.T) {
	ctx := context.Background()
	db := prepareTestDB(t)
	defer db.Close()

	writeRepo := NewBankCardWriteRepository(db)
	readRepo := NewBankCardReadRepository(db)

	meta := "test card"
	card := &models.BankCard{
		SecretName:  "card1",
		SecretOwner: "user1",
		Number:      "4111111111111111",
		Owner:       "John Doe",
		Exp:         "12/25",
		CVV:         "123",
		Meta:        &meta,
		UpdatedAt:   time.Now().UTC(),
	}

	// Add the bank card
	err := writeRepo.Add(ctx, card)
	require.NoError(t, err)

	// List and check
	cards, err := readRepo.List(ctx)
	require.NoError(t, err)
	require.Len(t, cards, 1)

	got := cards[0]
	require.Equal(t, card.SecretName, got.SecretName)
	require.Equal(t, card.SecretOwner, got.SecretOwner)
	require.Equal(t, card.Number, got.Number)
	require.Equal(t, card.Owner, got.Owner)
	require.Equal(t, card.Exp, got.Exp)
	require.Equal(t, card.CVV, got.CVV)
	require.NotNil(t, got.Meta)
	require.Equal(t, *card.Meta, *got.Meta)
	require.WithinDuration(t, card.UpdatedAt, got.UpdatedAt, time.Second)

	// Update the card
	newMeta := "updated card"
	card.Number = "4222222222222"
	card.Meta = &newMeta
	card.UpdatedAt = time.Now().UTC()

	err = writeRepo.Add(ctx, card)
	require.NoError(t, err)

	cards, err = readRepo.List(ctx)
	require.NoError(t, err)
	require.Len(t, cards, 1)

	require.Equal(t, "4222222222222", cards[0].Number)
	require.NotNil(t, cards[0].Meta)
	require.Equal(t, newMeta, *cards[0].Meta)
}

func TestBankCardListEmpty(t *testing.T) {
	ctx := context.Background()
	db := prepareTestDB(t)
	defer db.Close()

	readRepo := NewBankCardReadRepository(db)
	cards, err := readRepo.List(ctx)
	require.NoError(t, err)
	require.Empty(t, cards)
}
