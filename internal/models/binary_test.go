package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBinary_GetSecretName(t *testing.T) {
	b := &Binary{
		SecretName: "my-secret",
	}

	assert.Equal(t, "my-secret", b.GetSecretName(), "GetSecretName should return the SecretName field")
}

func TestBinary_GetUpdatedAt(t *testing.T) {
	now := time.Now()
	b := &Binary{
		UpdatedAt: now,
	}

	assert.Equal(t, now, b.GetUpdatedAt(), "GetUpdatedAt should return the UpdatedAt field")
}

func TestBinaryData_Fields(t *testing.T) {
	meta := "some meta"
	data := []byte{0x01, 0x02, 0x03}
	bd := &BinaryPayload{
		FilePath: "/tmp/file.bin",
		Data:     data,
		Meta:     &meta,
	}

	assert.Equal(t, "/tmp/file.bin", bd.FilePath)
	assert.Equal(t, data, bd.Data)
	assert.NotNil(t, bd.Meta)
	assert.Equal(t, "some meta", *bd.Meta)
}
