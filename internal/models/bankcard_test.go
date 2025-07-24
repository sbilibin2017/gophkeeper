package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBankCard_GetSecretName(t *testing.T) {
	card := &BankCard{
		SecretName: "card-secret",
	}
	assert.Equal(t, "card-secret", card.GetSecretName(), "GetSecretName should return the SecretName field")
}

func TestBankCard_GetUpdatedAt(t *testing.T) {
	now := time.Now()
	card := &BankCard{
		UpdatedAt: now,
	}
	assert.Equal(t, now, card.GetUpdatedAt(), "GetUpdatedAt should return the UpdatedAt field")
}

func TestBankCardData_Fields(t *testing.T) {
	meta := "metadata info"
	data := &BankCardPayload{
		Number: "1234-5678-9012-3456",
		Owner:  "John Doe",
		Exp:    "12/25",
		CVV:    "123",
		Meta:   &meta,
	}

	assert.Equal(t, "1234-5678-9012-3456", data.Number)
	assert.Equal(t, "John Doe", data.Owner)
	assert.Equal(t, "12/25", data.Exp)
	assert.Equal(t, "123", data.CVV)
	assert.NotNil(t, data.Meta)
	assert.Equal(t, "metadata info", *data.Meta)
}
