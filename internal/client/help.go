package client

// GetHelp returns a string containing the full usage guide and available commands
// for the gophkeeper CLI client. This includes instructions for registering,
// logging in, adding secrets (bankcard, text, binary, user credentials), listing
// and syncing secrets, and viewing version information.
//
// Each section includes the required flags and an example of usage.
func GetHelp() string {
	return `Usage:
  gophkeeper <command> [options]

Commands:
  register    Register a new user
  login       Login and get authentication token
  add-bankcard Add a new bankcard secret
  add-text    Add a new text secret
  add-binary  Add a new binary secret
  add-user    Add a new user secret
  list        List all secrets (requires private key for decryption)
  sync        Synchronize secrets between client and server (requires private key)
  version     Show version information

Options:

Register:
  --username      Username for registration (required)
  --password      Password for registration (required)
  --server-url    Server URL (required)

Example:
  gophkeeper register --username alice --password secret123 --server-url http://localhost:8080

Login:
  --username      Username for login (required)
  --password      Password for login (required)
  --server-url    Server URL (required)

Example:
  gophkeeper login --username alice --password secret123 --server-url http://localhost:8080

Add Bankcard:
  --token         Authentication token (required)
  --secret-name   Name for the bankcard (required)
  --number        Bankcard number (required)
  --owner         Bankcard owner (required)
  --exp           Expiry date (required)
  --cvv           CVV code (required)
  --meta          Optional metadata
  --pubkey        Public key PEM for encryption (required)

Example:
  gophkeeper add-bankcard --token <token> --secret-name "MyCard" --number 1234567890123456 --owner "Alice" --exp "12/24" --cvv 123 --meta "personal" --pubkey "<public_key_pem>"

Add Text:
  --token         Authentication token (required)
  --secret-name   Name for the text secret (required)
  --data          Text data (required)
  --meta          Optional metadata
  --pubkey        Public key PEM for encryption (required)

Example:
  gophkeeper add-text --token <token> --secret-name "Note" --data "My secret note" --meta "work" --pubkey "<public_key_pem>"

Add Binary:
  --token         Authentication token (required)
  --secret-name   Name for the binary secret (required)
  --data          Binary data (base64 encoded) (required)
  --meta          Optional metadata
  --pubkey        Public key PEM for encryption (required)

Example:
  gophkeeper add-binary --token <token> --secret-name "File" --data "<base64_data>" --meta "backup" --pubkey "<public_key_pem>"

Add User:
  --token         Authentication token (required)
  --secret-name   Name for the user secret (required)
  --username      Username (required)
  --password      Password (required)
  --meta          Optional metadata
  --pubkey        Public key PEM for encryption (required)

Example:
  gophkeeper add-user --token <token> --secret-name "EmailAccount" --username "user@example.com" --password "passw0rd" --meta "personal" --pubkey "<public_key_pem>"

List:
  --token         Authentication token (required)
  --privkey       Private key PEM for decryption (required)

Example:
  gophkeeper list --token <token> --privkey "<private_key_pem>"

Sync:
  --token         Authentication token (required)
  --sync-mode     Sync mode: server, client, or interactive (required)
  --privkey       Private key PEM (required)
  --server-url    Server URL (required)

Example:
  gophkeeper sync --token <token> --sync-mode client --privkey "<private_key_pem>" --server-url http://localhost:8080

Version:
  Show version and build date

Example:
  gophkeeper version`
}
