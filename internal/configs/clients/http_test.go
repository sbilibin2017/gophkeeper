package clients

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHTTPClient_BaseURL(t *testing.T) {
	baseURL := "https://example.com/api"
	client := NewHTTPClient(baseURL)

	assert.NotNil(t, client)
	assert.Equal(t, baseURL, client.BaseURL)
}
