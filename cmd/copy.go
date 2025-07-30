package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"gossh/config"
	"gossh/ssh"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var copyCmd = &cobra.Command{
	Use:   "copy [user@host]",
	Short: "Copy local public key to a remote server (similar to ssh-copy-id)",
	Long: `Copy a local public key to a remote server's authorized_keys file.

By default, it operates in a mode similar to ssh-copy-id, taking a user@host argument.
You will be prompted for a password.

Alternatively, you can use a pre-configured connection name with the --name flag.`,
	Run: runCopy,
}

func runCopy(cmd *cobra.Command, args []string) {
	connName, _ := cmd.Flags().GetString("name")
	port, _ := cmd.Flags().GetInt("port")

	var conn *config.Connection

	if connName != "" {
		// --- Connection Name Mode ---
		if len(args) > 0 {
			fmt.Println("Error: Do not provide a user@host argument when using the --name flag.")
			os.Exit(1)
		}
		connections, err := config.LoadConnections()
		if err != nil {
			fmt.Println("Error loading connections:", err)
			os.Exit(1)
		}
		for i, c := range connections {
			if c.Name == connName {
				conn = &connections[i]
				break
			}
		}
		if conn == nil {
			fmt.Printf("Error: connection '%s' not found\n", connName)
			os.Exit(1)
		}
	} else {
		// --- User@Host Mode ---
		if len(args) != 1 {
			fmt.Println("Error: Please provide the target in user@host format or use the --name flag.")
			_ = cmd.Help()
			os.Exit(1)
		}
		parts := strings.Split(args[0], "@")
		if len(parts) != 2 {
			fmt.Println("Error: Invalid format. Please use user@host.")
			os.Exit(1)
		}
		conn = &config.Connection{
			Name: "temp-copy",
			User: parts[0],
			Host: parts[1],
			Port: port,
		}
	}

	pubKeyPath, _ := cmd.Flags().GetString("pubkey")
	if pubKeyPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("Error getting home directory:", err)
			os.Exit(1)
		}
		pubKeyPath = filepath.Join(home, ".ssh", "id_rsa.pub")
	}

	pubKey, err := ioutil.ReadFile(pubKeyPath)
	if err != nil {
		fmt.Printf("Error reading public key from %s: %v\n", pubKeyPath, err)
		os.Exit(1)
	}

	// Command to ensure .ssh directory exists and append the key to authorized_keys
	remoteCmd := fmt.Sprintf("mkdir -p ~/.ssh && chmod 700 ~/.ssh && echo '%s' >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys", strings.TrimSpace(string(pubKey)))

	fmt.Printf("Attempting to copy public key to %s@%s...\n", conn.User, conn.Host)
	// We need to force password authentication for this operation.
	// A temporary connection object is created for this.
	tempConn := *conn
	tempConn.KeyPath = "" // Ensure key auth is not used
	tempConn.CredentialAlias = "" // Ensure stored password is not used either

	err = ssh.ExecuteRemoteCommand(&tempConn, remoteCmd)
	if err != nil {
		fmt.Println("Failed to copy public key.")
		os.Exit(1)
	}

	fmt.Println("Public key copied successfully. You should now be able to connect without a password.")
}

func init() {
	home, err := os.UserHomeDir()
	defaultPubKeyPath := ""
	if err == nil {
		defaultPubKeyPath = filepath.Join(home, ".ssh", "id_rsa.pub")
	}

	copyCmd.Flags().StringP("name", "n", "", "Use a pre-configured connection by its name")
	copyCmd.Flags().IntP("port", "p", 22, "Port number for user@host mode")
	copyCmd.Flags().StringP("pubkey", "i", defaultPubKeyPath, "Path to your public key file")
	rootCmd.AddCommand(copyCmd)
}