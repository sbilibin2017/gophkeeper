package app

import "github.com/spf13/cobra"

var (
	use   = "gophkeeper"
	short = "GophKeeper is a CLI tool for securely managing personal data"
	long  = `GophKeeper is a CLI tool for securely managing personal data.

Usage:
  gophkeeper [command] [flags]

Available Commands:
  build-info       Show build platform, version, date, and commit
  config           Manage client configuration (get, set, unset)
  register         Register a new user
  login            Authenticate an existing user
  logout           Logout from current session
  add              Add new data/secrets from file or interactively
  get              Retrieve specific data/secret from the server
  list             List stored secrets with filtering and sorting
  sync             Synchronize client with server and resolve conflicts  

Use "gophkeeper [command] --help" for more information about a command.`
)

// NewAppCommand creates the root command "gophkeeper" and adds all top-level child commands to it.
func NewAppCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Long:  long,
	}

	cmd.AddCommand(newBuildInfoCommand())

	cmd.AddCommand(newClientConfigGetCommand())
	cmd.AddCommand(newClientConfigSetCommand())
	cmd.AddCommand(newClientConfigUnsetCommand())

	cmd.AddCommand(newRegisterCommand())
	cmd.AddCommand(newLoginCommand())
	cmd.AddCommand(newLogoutCommand())
	cmd.AddCommand(newAddCommand())
	cmd.AddCommand(newGetCommand())
	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newSyncCommand())

	return cmd
}
