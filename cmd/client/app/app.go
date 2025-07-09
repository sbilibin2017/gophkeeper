package app

import "github.com/spf13/cobra"

var (
	use   = "gophkeeper"
	short = "GophKeeper — CLI tool for secure management of personal data"
	long  = `GophKeeper — CLI tool for secure management of personal data.

Usage:
  gophkeeper [command] [flags]

Available commands:
  build-info       Show build information: platform, version, date, and commit  
  register         Register a new user
  login            Authenticate an existing user  
  add              Add new data/secrets from a file or interactively
  get              Retrieve specific data/secret from the server
  list             List saved secrets with filtering and sorting
  sync             Synchronize client with server and resolve conflicts  

Use "gophkeeper [command] --help" for more information about a command.`
)

func NewAppCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Long:  long,
	}

	cmd.AddCommand(newBuildInfoCommand())
	cmd.AddCommand(newRegisterCommand())
	cmd.AddCommand(newConfigureCommand())
	cmd.AddCommand(newLoginCommand())
	cmd.AddCommand(newAddLoginPasswordCommand())
	cmd.AddCommand(newAddTextSecretCommand())
	cmd.AddCommand(newAddBinarySecretCommand())
	cmd.AddCommand(newAddCardSecretCommand())
	cmd.AddCommand(newGetLoginPasswordCommand())
	cmd.AddCommand(newGetTextCommand())
	cmd.AddCommand(newGetBinaryCommand())
	cmd.AddCommand(newGetCardCommand())
	cmd.AddCommand(newClientListCommand())
	cmd.AddCommand(newSyncCommand())

	return cmd
}
