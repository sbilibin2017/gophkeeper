package app_test

import (
	"testing"

	"github.com/sbilibin2017/gophkeeper/cmd/client/app"
	"github.com/stretchr/testify/assert"
)

func TestNewCommand(t *testing.T) {
	cmd := app.NewCommand()

	// Проверяем, что команда создана и имеет правильное имя
	assert.NotNil(t, cmd)
	assert.Equal(t, "gophkeeper", cmd.Use)
	assert.Contains(t, cmd.Short, "GophKeeper")

	// Проверяем, что в списке подкоманд есть команда "auth"
	foundAuth := false
	for _, c := range cmd.Commands() {
		if c.Name() == "auth" {
			foundAuth = true
			break
		}
	}

	assert.True(t, foundAuth, "expected to find 'auth' subcommand")
}
