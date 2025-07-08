package app

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func newSyncCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync [--auto-resolve=client|server] [--interactive] [--server-url <url>]",
		Short: "Synchronize client with server and resolve conflicts",
		RunE: func(cmd *cobra.Command, args []string) error {
			serverURL, _ := cmd.Flags().GetString("server-url")
			autoResolve, _ := cmd.Flags().GetString("auto-resolve")
			interactive, _ := cmd.Flags().GetBool("interactive")

			if serverURL == "" {
				serverURL = "https://default-server-url.example.com"
				cmd.Printf("Using fallback server URL: %s\n", serverURL)
			} else {
				cmd.Printf("Using server URL: %s\n", serverURL)
			}

			// Mock: Detect conflicts - in real app, fetch changes from client & server
			conflicts := []string{"secret1", "secret2"} // example conflicting secret IDs
			if len(conflicts) == 0 {
				cmd.Println("No conflicts detected, syncing completed successfully.")
				return nil
			}

			cmd.Printf("Detected %d conflicts.\n", len(conflicts))

			switch autoResolve {
			case "client":
				cmd.Println("Auto-resolving conflicts in favor of client changes.")
				// mock apply client changes:
				for _, c := range conflicts {
					cmd.Printf("Conflict %s resolved: kept client version.\n", c)
				}
			case "server":
				cmd.Println("Auto-resolving conflicts in favor of server changes.")
				// mock apply server changes:
				for _, c := range conflicts {
					cmd.Printf("Conflict %s resolved: kept server version.\n", c)
				}
			case "":
				if interactive {
					reader := bufio.NewReader(os.Stdin)
					for _, c := range conflicts {
						for {
							cmd.Printf("Conflict detected on '%s'. Choose resolution [client/server/skip]: ", c)
							input, err := reader.ReadString('\n')
							if err != nil {
								return err
							}
							choice := strings.TrimSpace(strings.ToLower(input))
							if choice == "client" {
								cmd.Printf("Conflict %s resolved: kept client version.\n", c)
								break
							} else if choice == "server" {
								cmd.Printf("Conflict %s resolved: kept server version.\n", c)
								break
							} else if choice == "skip" {
								cmd.Printf("Conflict %s skipped.\n", c)
								break
							} else {
								cmd.Println("Invalid choice, please enter 'client', 'server', or 'skip'.")
							}
						}
					}
				} else {
					return errors.New("conflicts detected but no resolution method specified; use --auto-resolve or --interactive")
				}
			default:
				return fmt.Errorf("invalid --auto-resolve option: %q (allowed: client, server)", autoResolve)
			}

			cmd.Println("Synchronization completed successfully.")
			return nil
		},
	}

	cmd.Flags().StringP("server-url", "s", "", "Server URL (optional, fallback to config)")
	cmd.Flags().StringP("auto-resolve", "a", "", "Automatically resolve conflicts using either 'client' or 'server'")
	cmd.Flags().BoolP("interactive", "i", false, "Enable interactive conflict resolution")

	_ = cmd.MarkFlagRequired("server-url")

	return cmd
}
