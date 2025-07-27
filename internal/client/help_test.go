package client

import (
	"strings"
	"testing"
)

func TestGetHelp(t *testing.T) {
	help := GetHelp()

	if !strings.Contains(help, "Usage:") {
		t.Error("GetHelp output missing 'Usage:'")
	}

	if !strings.Contains(help, "register") {
		t.Error("GetHelp output missing 'register' command")
	}

	if !strings.Contains(help, "login") {
		t.Error("GetHelp output missing 'login' command")
	}

	if !strings.Contains(help, "add-bankcard") {
		t.Error("GetHelp output missing 'add-bankcard' command")
	}

	if !strings.Contains(help, "version") {
		t.Error("GetHelp output missing 'version' command")
	}

	// Optionally check the output length to ensure it is not empty or truncated
	if len(help) < 100 {
		t.Error("GetHelp output unexpectedly short")
	}
}
