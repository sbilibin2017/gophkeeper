package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoginPassword(t *testing.T) {
	now := time.Now()
	meta := map[string]string{"env": "prod"}

	lp := NewLoginPassword(
		WithLoginPasswordSecretID("id123"),
		WithLoginPasswordLogin("user123"),
		WithLoginPasswordPassword("pass123"),
		WithLoginPasswordMeta(meta),
		WithLoginPasswordUpdatedAt(now),
	)

	assert.Equal(t, "id123", lp.SecretID)
	assert.Equal(t, "user123", lp.Login)
	assert.Equal(t, "pass123", lp.Password)
	assert.Equal(t, meta, lp.Meta)
	assert.Equal(t, now, lp.UpdatedAt)
}

func TestText(t *testing.T) {
	now := time.Now()
	meta := map[string]string{"note": "important"}

	text := NewText(
		WithTextSecretID("textid"),
		WithTextContent("This is secret text"),
		WithTextMeta(meta),
		WithTextUpdatedAt(now),
	)

	assert.Equal(t, "textid", text.SecretID)
	assert.Equal(t, "This is secret text", text.Content)
	assert.Equal(t, meta, text.Meta)
	assert.Equal(t, now, text.UpdatedAt)
}

func TestBinary(t *testing.T) {
	now := time.Now()
	meta := map[string]string{"filetype": "png"}
	data := []byte{0x01, 0x02, 0x03}

	bin := NewBinary(
		WithBinarySecretID("binid"),
		WithBinaryData(data),
		WithBinaryMeta(meta),
		WithBinaryUpdatedAt(now),
	)

	assert.Equal(t, "binid", bin.SecretID)
	assert.Equal(t, data, bin.Data)
	assert.Equal(t, meta, bin.Meta)
	assert.Equal(t, now, bin.UpdatedAt)
}

func TestCard(t *testing.T) {
	now := time.Now()
	meta := map[string]string{"issuer": "bank"}

	card := NewCard(
		WithCardSecretID("cardid"),
		WithCardNumber("4111111111111111"),
		WithCardHolder("John Doe"),
		WithCardExpMonth(12),
		WithCardExpYear(2030),
		WithCardCVV("123"),
		WithCardMeta(meta),
		WithCardUpdatedAt(now),
	)

	assert.Equal(t, "cardid", card.SecretID)
	assert.Equal(t, "4111111111111111", card.Number)
	assert.Equal(t, "John Doe", card.Holder)
	assert.Equal(t, 12, card.ExpMonth)
	assert.Equal(t, 2030, card.ExpYear)
	assert.Equal(t, "123", card.CVV)
	assert.Equal(t, meta, card.Meta)
	assert.Equal(t, now, card.UpdatedAt)
}
