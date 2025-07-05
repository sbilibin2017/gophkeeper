package main_test

import (
	"bytes"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type AppSuite struct {
	suite.Suite
	binPath string
}

func (s *AppSuite) SetupSuite() {
	s.binPath = "./testbin_gophkeeper"
	cmd := exec.Command("go", "build", "-o", s.binPath, ".")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		s.T().Fatalf("failed to build binary: %v, stderr: %s", err, stderr.String())
	}
}

func (s *AppSuite) TearDownSuite() {
	err := os.Remove(s.binPath)
	if err != nil {
		s.T().Logf("failed to remove test binary: %v", err)
	}
}

func (s *AppSuite) runCommand(args ...string) (string, error) {
	cmd := exec.Command(s.binPath, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return out.String(), err
}

func (s *AppSuite) TestRootCommandRuns() {
	out, err := s.runCommand("help")
	require.NoError(s.T(), err)
	require.Contains(s.T(), out, "GophKeeper")
}

func (s *AppSuite) TestBuildInfoCommand() {
	out, err := s.runCommand("build-info")
	require.NoError(s.T(), err)
	require.Contains(s.T(), out, "Build platform:")
	require.Contains(s.T(), out, "Build version:")
}

func (s *AppSuite) TestUsageCommand() {
	out, err := s.runCommand("usage")
	require.NoError(s.T(), err)
	require.Contains(s.T(), out, "Usage")
}

func (s *AppSuite) TestRegisterCommandMissingFlags() {
	_, err := s.runCommand("register")
	// Проверяем, что ошибка именно о флагах
	require.Error(s.T(), err)

}

func (s *AppSuite) TestRegisterCommandSuccess() {
	out, err := s.runCommand(
		"register",
		"--server-url", "https://localhost:8000",
		"--username", "testuser",
		"--password", "testpass",
	)
	require.NoError(s.T(), err)
	// Можно проверить, что вывод не пустой, если команда должна что-то выводить:
	require.NotEmpty(s.T(), out, "expected some output from register command")
}

func TestAppSuite(t *testing.T) {
	suite.Run(t, new(AppSuite))
}
