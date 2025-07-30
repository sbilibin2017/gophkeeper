package client

const (
	CommandRegister    = "register"
	CommandLogin       = "login"
	CommandAddBankcard = "add-bankcard"
	CommandAddText     = "add-text"
	CommandAddBinary   = "add-binary"
	CommandAddUser     = "add-user"
	CommandList        = "list"
	CommandSync        = "sync"
	CommandVersion     = "version"
	CommandHelp        = "help"
)

// GetCommand extracts the command from args.
// If no command provided, returns empty string.
func GetCommand(args []string) string {
	if len(args) > 1 {
		return args[1]
	}
	return ""
}
