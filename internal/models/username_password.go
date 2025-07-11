package models

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// UsernamePassword хранит пару логин и пароль.
type UsernamePassword struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// NewUsernamePasswordFromInteractive запрашивает у пользователя логин и пароль через консоль (stdin).
// Возвращает структуру UsernamePassword или ошибку при чтении данных.
func NewUsernamePasswordFromInteractive(r io.Reader) (*UsernamePassword, error) {
	reader := bufio.NewReader(r)

	fmt.Print("Enter username: ")
	inputLogin, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	login := strings.TrimSpace(inputLogin)

	fmt.Print("Enter password: ")
	inputPassword, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	password := strings.TrimSpace(inputPassword)

	return &UsernamePassword{
		Username: login,
		Password: password,
	}, nil
}

// NewUsernamePasswordFromArgs получает логин и пароль из аргументов командной строки.
// Возвращает ошибку, если передано не ровно два аргумента.
func NewUsernamePasswordFromArgs(args []string) (*UsernamePassword, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("expected exactly 2 arguments: username and password")
	}
	return &UsernamePassword{
		Username: args[0],
		Password: args[1],
	}, nil
}

// ValidateUsernamePassword проверяет, что логин и пароль не пустые.
// Возвращает ошибку с описанием, если логин или пароль пусты.
func ValidateUsernamePassword(secret *UsernamePassword) error {
	if secret.Username == "" {
		return fmt.Errorf("username cannot be empty")
	}
	if secret.Password == "" {
		return fmt.Errorf("password cannot be empty")
	}
	return nil
}
