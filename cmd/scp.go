package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"gossh/config"
	"gossh/scp"
	"os"
	"path/filepath"
	"strings"
)

var scpCmd = &cobra.Command{
	Use:   "scp [flags] source destination",
	Short: "Securely copy files between local and remote systems",
	Long: `Copy files between local and remote systems using SCP protocol.

Supports standard SCP syntax:
- Local to remote: gossh scp localfile.txt user@host:/path/to/destination
- Remote to local: gossh scp user@host:/path/to/file.txt localfile.txt
- Using connection names: gossh scp --name myserver localfile.txt /path/to/destination

Examples:
  gossh scp file.txt user@server:/home/user/
  gossh scp user@server:/etc/config.conf ./config.conf
  gossh scp --name myserver localdir/ /remote/path/ -r
  gossh scp -r localdir/ user@server:/remote/path/`,
	Run: runScp,
}

var (
	recursive bool
	preserveTimes bool
	compression bool
)

func init() {
	scpCmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Copy directories recursively")
	scpCmd.Flags().BoolVarP(&preserveTimes, "preserve", "p", false, "Preserve modification times")
	scpCmd.Flags().BoolVarP(&compression, "compress", "C", false, "Enable compression")
	scpCmd.Flags().StringP("name", "n", "", "Use a pre-configured connection by its name")
	scpCmd.Flags().IntP("port", "P", 22, "Port number for user@host mode")
	rootCmd.AddCommand(scpCmd)
}

func runScp(cmd *cobra.Command, args []string) {
	if len(args) != 2 {
		fmt.Println("Error: SCP requires exactly 2 arguments: source and destination")
		cmd.Help()
		os.Exit(1)
	}

	source := args[0]
	destination := args[1]

	connName, _ := cmd.Flags().GetString("name")
	port, _ := cmd.Flags().GetInt("port")

	// Parse source and destination to determine direction and connection info
	sourceIsRemote, sourceConn, sourcePath := parsePath(source, connName, port)
	destIsRemote, destConn, destPath := parsePath(destination, connName, port)

	// Validate that we have exactly one remote and one local path
	if sourceIsRemote == destIsRemote {
		fmt.Println("Error: One of source or destination must be remote, the other must be local")
		os.Exit(1)
	}

	var conn *config.Connection
	var remotePath, localPath string
	var isUpload bool

	if sourceIsRemote {
		conn = sourceConn
		remotePath = sourcePath
		localPath = destPath
		isUpload = false
	} else {
		conn = destConn
		remotePath = destPath
		localPath = sourcePath
		isUpload = true
	}

	// Create SCP session
	scpSession, err := scp.NewSCPSession(conn)
	if err != nil {
		fmt.Printf("Error creating SCP session: %v\n", err)
		os.Exit(1)
	}
	defer scpSession.Close()

	// Check if we're dealing with directories
	sourceIsDir := isDirectory(localPath)
	
	if isUpload {
		// Upload: local to remote
		if sourceIsDir {
			if !recursive {
				fmt.Println("Error: Source is a directory, use -r flag for recursive copy")
				os.Exit(1)
			}
			fmt.Printf("Uploading directory %s to %s:%s\n", localPath, conn.Host, remotePath)
			err = scpSession.UploadDirectory(localPath, remotePath, preserveTimes)
		} else {
			fmt.Printf("Uploading file %s to %s:%s\n", localPath, conn.Host, remotePath)
			err = scpSession.UploadFile(localPath, remotePath, preserveTimes)
		}
	} else {
		// Download: remote to local
		// First check if remote path is a directory
		remoteIsDir := isRemoteDirectory(scpSession, remotePath)
		
		if remoteIsDir {
			if !recursive {
				fmt.Println("Error: Remote source is a directory, use -r flag for recursive copy")
				os.Exit(1)
			}
			fmt.Printf("Downloading directory %s:%s to %s\n", conn.Host, remotePath, localPath)
			err = scpSession.DownloadDirectory(remotePath, localPath, preserveTimes)
		} else {
			fmt.Printf("Downloading file %s:%s to %s\n", conn.Host, remotePath, localPath)
			err = scpSession.DownloadFile(remotePath, localPath, preserveTimes)
		}
	}

	if err != nil {
		fmt.Printf("Error during file transfer: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("File transfer completed successfully!")
}

// parsePath determines if a path is local or remote and extracts connection info
func parsePath(pathStr, connName string, port int) (isRemote bool, conn *config.Connection, filePath string) {
	if strings.Contains(pathStr, ":") {
		// Remote path format: [user@]host:path
		parts := strings.SplitN(pathStr, ":", 2)
		hostPart := parts[0]
		filePath = parts[1]
		
		isRemote = true
		
		if connName != "" {
			// Use pre-configured connection
			connections, err := config.LoadConnections()
			if err != nil {
				fmt.Printf("Error loading connections: %v\n", err)
				os.Exit(1)
			}
			
			for _, c := range connections {
				if c.Name == connName {
					conn = &c
					break
				}
			}
			
			if conn == nil {
				fmt.Printf("Error: connection '%s' not found\n", connName)
				os.Exit(1)
			}
		} else {
			// Parse user@host format
			var user, host string
			if strings.Contains(hostPart, "@") {
				userHostParts := strings.SplitN(hostPart, "@", 2)
				user = userHostParts[0]
				host = userHostParts[1]
			} else {
				fmt.Println("Error: Remote path must be in format user@host:path or use --name flag")
				os.Exit(1)
			}
			
			conn = &config.Connection{
				Name: "temp-scp",
				User: user,
				Host: host,
				Port: port,
			}
		}
	} else {
		// Local path
		isRemote = false
		filePath = pathStr
		
		// Expand ~ to home directory
		if strings.HasPrefix(filePath, "~/") {
			home, err := os.UserHomeDir()
			if err == nil {
				filePath = filepath.Join(home, filePath[2:])
			}
		}
	}
	
	return
}

// isDirectory checks if a local path is a directory
func isDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// isRemoteDirectory checks if a remote path is a directory
func isRemoteDirectory(session *scp.SCPSession, path string) bool {
	// Simple check: try to list the directory
	// This is a basic implementation - could be improved
	output, err := session.ExecuteCommand(fmt.Sprintf("test -d %s && echo 'DIR' || echo 'FILE'", path))
	if err != nil {
		return false
	}
	return strings.TrimSpace(output) == "DIR"
}