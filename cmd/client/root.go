package main

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "gophkeeper",
	Short: "Gophkeeper CLI tool",
	Long:  "CLI application for Gophkeeper to login, logout, and manage authentication.",
}

func init() {
	rootCmd.AddCommand(registerCmd)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
}
