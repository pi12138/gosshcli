package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"gossh/config"
	"gossh/ssh"
	"os"
	"strings"
)

var cpCmd = &cobra.Command{
	Use:   "cp [source] [destination]",
	Short: "Copy files between local and remote hosts (similar to scp)",
	Long: `Copy files or directories between local and remote hosts using SFTP.

Usage:
  - Upload:   gossh cp <local-path> <connection-name>:<remote-path>
  - Download: gossh cp <connection-name>:<remote-path> <local-path>

Use the -r flag to copy directories recursively.

Examples:
  gossh cp /local/file.txt myserver:/remote/path/
  gossh cp myserver:/remote/file.txt /local/path/
  gossh cp -r /local/directory myserver:/remote/path/
  gossh cp -r myserver:/remote/directory /local/path/`,
	Run: runCp,
}

func init() {
	cpCmd.Flags().BoolP("recursive", "r", false, "Copy directories recursively")
	rootCmd.AddCommand(cpCmd)
}

func runCp(cmd *cobra.Command, args []string) {
	if len(args) != 2 {
		fmt.Println("Error: Please provide both source and destination.")
		_ = cmd.Help()
		os.Exit(1)
	}

	recursive, _ := cmd.Flags().GetBool("recursive")
	source := args[0]
	destination := args[1]

	// Determine if this is an upload or download operation
	// Check if source contains ':'
	sourceHasColon := strings.Contains(source, ":")
	destHasColon := strings.Contains(destination, ":")

	if sourceHasColon && destHasColon {
		fmt.Println("Error: Cannot copy between two remote locations.")
		os.Exit(1)
	}

	if !sourceHasColon && !destHasColon {
		fmt.Println("Error: At least one path must be remote (format: connection-name:path).")
		os.Exit(1)
	}

	connections, err := config.LoadConnections()
	if err != nil {
		fmt.Println("Error loading connections:", err)
		os.Exit(1)
	}

	if sourceHasColon {
		// Download: remote -> local
		parts := strings.SplitN(source, ":", 2)
		if len(parts) != 2 {
			fmt.Println("Error: Invalid remote path format. Use connection-name:path")
			os.Exit(1)
		}
		connName := parts[0]
		remotePath := parts[1]
		localPath := destination

		conn := findConnection(connections, connName)
		if conn == nil {
			fmt.Printf("Error: connection '%s' not found\n", connName)
			os.Exit(1)
		}

		fmt.Printf("Downloading from %s@%s:%s to %s...\n", conn.User, conn.Host, remotePath, localPath)
		err = ssh.DownloadFile(conn, remotePath, localPath, recursive)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Download completed successfully.")
	} else {
		// Upload: local -> remote
		parts := strings.SplitN(destination, ":", 2)
		if len(parts) != 2 {
			fmt.Println("Error: Invalid remote path format. Use connection-name:path")
			os.Exit(1)
		}
		connName := parts[0]
		remotePath := parts[1]
		localPath := source

		conn := findConnection(connections, connName)
		if conn == nil {
			fmt.Printf("Error: connection '%s' not found\n", connName)
			os.Exit(1)
		}

		fmt.Printf("Uploading from %s to %s@%s:%s...\n", localPath, conn.User, conn.Host, remotePath)
		err = ssh.UploadFile(conn, localPath, remotePath, recursive)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Upload completed successfully.")
	}
}

func findConnection(connections []config.Connection, name string) *config.Connection {
	for i, c := range connections {
		if c.Name == name {
			return &connections[i]
		}
	}
	return nil
}
