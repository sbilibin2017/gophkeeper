package app_test

import (
	"testing"

	"github.com/sbilibin2017/gophkeeper/cmd/client/app"
	"github.com/stretchr/testify/assert"
)

func TestNewCobraCommand_CommandsRegistered(t *testing.T) {
	root := app.NewCobraCommand()

	expectedCommands := []string{
		"register",
		"login",
		"add-secret-bank-card",
		"add-secret-binary",
		"add-secret-text",
		"add-secret-username-password",
		"list-secrets",
	}

	actualCommands := []string{}
	for _, cmd := range root.Commands() {
		actualCommands = append(actualCommands, cmd.Name())
	}

	for _, expected := range expectedCommands {
		assert.Contains(t, actualCommands, expected, "Команда %q должна быть зарегистрирована", expected)
	}
}
