package app

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoginCommand_EmptyUsernameFlag(t *testing.T) {
	cmd := newLoginCommand()

	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)
	cmd.SetArgs([]string{
		"--server-url", "http://localhost:8080",
		"--username", "",
		"--password", "pass123",
	})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "username cannot be empty")
	assert.NotContains(t, errBuf.String(), "Usage:")
}

func TestLoginCommand_EmptyPasswordFlag(t *testing.T) {
	cmd := newLoginCommand()

	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)
	cmd.SetArgs([]string{
		"--server-url", "http://localhost:8080",
		"--username", "user1",
		"--password", "",
	})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "password cannot be empty")
	assert.NotContains(t, errBuf.String(), "Usage:")
}

func TestLoginCommand_MissingServerURLFlag(t *testing.T) {
	cmd := newLoginCommand()

	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	cmd.SetOut(outBuf) // usage prints here
	cmd.SetErr(errBuf)
	cmd.SetArgs([]string{
		"--username", "user1",
		"--password", "pass123",
	})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required flag(s) \"server-url\" not set")
	// Usage message is printed to stdout, not stderr
	assert.Contains(t, outBuf.String(), "Usage:")
}
