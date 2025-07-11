package main

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "gophkeeper",
	Short: "GophKeeper CLI — безопасное хранилище паролей и данных",
}

const (
	serverURLEnvKey = "GOPHKEEPER_SERVER_URL"
	tokenEnvKey     = "GOPHKEEPER_TOKEN"
)

func executeCommand() error {
	rootCmd.AddCommand(registerCmd)
	return rootCmd.Execute()
}
