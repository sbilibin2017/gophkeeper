package clients

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewHTTPClient(t *testing.T) {
	baseURL := "http://example.com"
	client := NewHTTPClient(baseURL)

	assert.Equal(t, baseURL, client.BaseURL, "BaseURL должен быть установлен правильно")
	assert.Equal(t, 3, client.RetryCount, "RetryCount должен быть 3")
	assert.Equal(t, 500*time.Millisecond, client.RetryWaitTime, "RetryWaitTime должен быть 500ms")
	assert.Equal(t, 2*time.Second, client.RetryMaxWaitTime, "RetryMaxWaitTime должен быть 2s")
}
