package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestText_GetSecretName(t *testing.T) {
	text := &Text{
		SecretName: "my-secret-text",
	}
	assert.Equal(t, "my-secret-text", text.GetSecretName(), "GetSecretName should return SecretName field")
}

func TestText_GetUpdatedAt(t *testing.T) {
	now := time.Now()
	text := &Text{
		UpdatedAt: now,
	}
	assert.Equal(t, now, text.GetUpdatedAt(), "GetUpdatedAt should return UpdatedAt field")
}

func TestTextData_Fields(t *testing.T) {
	meta := "some metadata"
	data := &TextData{
		Data: "this is some secret text",
		Meta: &meta,
	}

	assert.Equal(t, "this is some secret text", data.Data)
	assert.NotNil(t, data.Meta)
	assert.Equal(t, "some metadata", *data.Meta)
}
