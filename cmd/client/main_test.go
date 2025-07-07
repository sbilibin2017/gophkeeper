package main_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type AppSuite struct {
	suite.Suite
	binPath string
	server  *httptest.Server
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

	// Запускаем мок-сервер, который отвечает на POST /register
	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/register" && r.Method == http.MethodPost {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok","message":"user registered"}`))
			return
		}
		http.NotFound(w, r)
	}))
}

func (s *AppSuite) TearDownSuite() {
	if s.server != nil {
		s.server.Close()
	}
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

func (s *AppSuite) TestRegisterCommandSuccess() {
	// Передаем базовый URL, клиент должен добавить /register сам
	out, err := s.runCommand(
		"register",
		"--server-url", s.server.URL,
		"--username", "testuser",
		"--password", "testpass",
	)

	s.T().Logf("Output:\n%s", out)
	require.NoError(s.T(), err)
}

func TestAppSuite(t *testing.T) {
	suite.Run(t, new(AppSuite))
}
