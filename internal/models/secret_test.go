package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecretLoginPasswordDBOptions(t *testing.T) {
	meta := map[string]string{"foo": "bar"}

	lp := NewSecretLoginPasswordDB(
		WithSecretLogin("user123"),
		WithSecretPassword("pass456"),
		WithSecretLoginPasswordMeta(meta),
		WithSecretLoginPasswordSecretID("secret-id-1"),
	)

	assert.Equal(t, "user123", lp.Login)
	assert.Equal(t, "pass456", lp.Password)
	assert.Equal(t, meta, lp.Meta)
	assert.Equal(t, "secret-id-1", lp.SecretID)
}

func TestSecretPayloadTextDBOptions(t *testing.T) {
	meta := map[string]string{"key": "value"}

	pt := NewSecretPayloadTextDB(
		WithSecretTextContent("hello world"),
		WithSecretTextMeta(meta),
		WithSecretTextSecretID("text-secret-1"),
	)

	assert.Equal(t, "hello world", pt.Content)
	assert.Equal(t, meta, pt.Meta)
	assert.Equal(t, "text-secret-1", pt.SecretID)
}

func TestSecretPayloadBinaryDBOptions(t *testing.T) {
	meta := map[string]string{"bin": "data"}
	data := []byte{1, 2, 3, 4}

	pb := NewSecretPayloadBinaryDB(
		WithSecretBinaryData(data),
		WithSecretBinaryMeta(meta),
		WithSecretBinarySecretID("binary-secret-1"),
	)

	assert.Equal(t, data, pb.Data)
	assert.Equal(t, meta, pb.Meta)
	assert.Equal(t, "binary-secret-1", pb.SecretID)
}

func TestSecretPayloadCardDBOptions(t *testing.T) {
	meta := map[string]string{"bank": "Test Bank"}

	pc := NewSecretPayloadCardDB(
		WithSecretCardNumber("4111111111111111"),
		WithSecretCardHolder("John Doe"),
		WithSecretCardExpMonth(12),
		WithSecretCardExpYear(2030),
		WithSecretCardCVV("123"),
		WithSecretCardMeta(meta),
		WithSecretCardSecretID("card-secret-1"),
	)

	assert.Equal(t, "4111111111111111", pc.Number)
	assert.Equal(t, "John Doe", pc.Holder)
	assert.Equal(t, 12, pc.ExpMonth)
	assert.Equal(t, 2030, pc.ExpYear)
	assert.Equal(t, "123", pc.CVV)
	assert.Equal(t, meta, pc.Meta)
	assert.Equal(t, "card-secret-1", pc.SecretID)
}
