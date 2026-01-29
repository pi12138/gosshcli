package cmd

import (
	"fmt"
	"gossh/config"
	"gossh/ssh"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var scpCmd = &cobra.Command{
	Use:   "scp [source] [destination]",
	Short: "Copy files between local and remote hosts (secure copy protocol)",
	Long: `Copy files or directories between local and remote hosts using SFTP.

Usage:
  - Upload:   gossh scp <local-path> <connection-name>:<remote-path>
  - Download: gossh scp <connection-name>:<remote-path> <local-path>

Use the -r flag to copy directories recursively.

Examples:
  gossh scp /local/file.txt myserver:/remote/path/
  gossh scp myserver:/remote/file.txt /local/path/
  gossh scp -r /local/directory myserver:/remote/path/
  gossh scp -r myserver:/remote/directory /local/path/`,
	Run: runScp,
}

func init() {
	scpCmd.Flags().BoolP("recursive", "r", false, "Copy directories recursively")
	scpCmd.Flags().BoolP("force", "f", false, "Force overwrite of existing files")
	rootCmd.AddCommand(scpCmd)
}

func runScp(cmd *cobra.Command, args []string) {
	if len(args) != 2 {
		fmt.Println("Error: Please provide both source and destination.")
		_ = cmd.Help()
		os.Exit(1)
	}

	recursive, _ := cmd.Flags().GetBool("recursive")
	force, _ := cmd.Flags().GetBool("force")
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
		err = ssh.DownloadFileWithOpts(conn, remotePath, localPath, recursive, force)
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
		err = ssh.UploadFileWithOpts(conn, localPath, remotePath, recursive, force)
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
