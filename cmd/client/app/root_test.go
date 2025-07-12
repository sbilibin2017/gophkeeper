package app

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExecute_NoArgs(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"app"} // только имя программы

	err := Execute(context.Background())
	require.EqualError(t, err, "не указана команда")
}

func TestExecute_UnknownCommand(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"app", "unknown"}

	err := Execute(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "неизвестная команда: unknown")
}
