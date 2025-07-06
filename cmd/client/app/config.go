package app

import (
	"fmt"

	"github.com/sbilibin2017/gophkeeper/internal/configs/file"
	"github.com/spf13/cobra"
)

// newClientConfigGetCommand creates a command to show the current client configuration.
func newClientConfigGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get [<key>]",
		Short: "Display the current client configuration or a specific key",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				val, ok, err := file.GetConfigValue(args[0])
				if err != nil {
					return err
				}
				if !ok {
					return fmt.Errorf("key %q not found", args[0])
				}
				cmd.Printf("%s = %s\n", args[0], val)
				return nil
			}

			cfg, err := file.ListConfig()
			if err != nil {
				return err
			}
			if len(cfg) == 0 {
				cmd.Println("No config values set.")
				return nil
			}
			for k, v := range cfg {
				cmd.Printf("%s = %s\n", k, v)
			}
			return nil
		},
	}
}

// newClientConfigSetCommand creates a command to set a key-value pair in the client configuration.
func newClientConfigSetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a client configuration key to a specified value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, value := args[0], args[1]
			if err := file.SetConfigValue(key, value); err != nil {
				return err
			}
			cmd.Printf("Set %q = %q\n", key, value)
			return nil
		},
	}
}

// newClientConfigUnsetCommand creates a command to remove a key from the client configuration.
func newClientConfigUnsetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "unset <key>",
		Short: "Remove a key from the client configuration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			if err := file.UnsetConfigValue(key); err != nil {
				return err
			}
			cmd.Printf("Unset %q\n", key)
			return nil
		},
	}
}
