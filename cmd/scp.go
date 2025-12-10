package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"gossh/config"
	"gossh/ssh"
	"os"
	"path/filepath"
	"strings"
)

var scpCmd = &cobra.Command{
	Use:   "scp [source] [target]",
	Short: "Copy files between local and remote servers (SCP compatible)",
	Long: `Copy files between local and remote servers using SCP protocol.

Usage:
  scp [options] local_file remote_name:remote_path
  scp [options] remote_name:remote_path local_file
  scp [options] local_file1 local_file2 remote_name:remote_path
  scp [options] -r local_dir remote_name:remote_path
  scp [options] -r remote_name:remote_path local_dir

Options:
  -r	Recursively copy directories`,
	Run: runScp,
}

var recursive bool

func init() {
	scpCmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Recursively copy directories")
	rootCmd.AddCommand(scpCmd)
}

func runScp(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		fmt.Println("Error: Please provide at least two arguments: source and target")
		_ = cmd.Help()
		os.Exit(1)
	}

	// Check if source or target is remote
	isSourceRemote := strings.Contains(args[0], ":")
	isTargetRemote := strings.Contains(args[len(args)-1], ":")

	if isSourceRemote && isTargetRemote {
		fmt.Println("Error: Both source and target cannot be remote. Use local intermediary.")
		os.Exit(1)
	} else if !isSourceRemote && !isTargetRemote {
		fmt.Println("Error: At least one of source or target must be remote.")
		os.Exit(1)
	}

	if isSourceRemote {
		// Download from remote to local
		downloadFiles(cmd, args)
	} else {
		// Upload from local to remote
		uploadFiles(cmd, args)
	}
}

func uploadFiles(cmd *cobra.Command, args []string) {
	// Last argument is remote target
	remoteTarget := args[len(args)-1]
	localSources := args[:len(args)-1]

	// Parse remote target: remote_name:remote_path
	parts := strings.Split(remoteTarget, ":")
	if len(parts) != 2 {
		fmt.Println("Error: Invalid remote target format. Use 'remote_name:remote_path'")
		os.Exit(1)
	}

	remoteName := parts[0]
	remotePath := parts[1]

	// Load connection configuration
	connections, err := config.LoadConnections()
	if err != nil {
		fmt.Println("Error loading connections:", err)
		os.Exit(1)
	}

	var conn *config.Connection
	for i, c := range connections {
		if c.Name == remoteName {
			conn = &connections[i]
			break
		}
	}

	if conn == nil {
		fmt.Printf("Error: Connection '%s' not found\n", remoteName)
		os.Exit(1)
	}

	// Upload each local source
	for _, localSource := range localSources {
		// Check if local source exists
		info, err := os.Stat(localSource)
		if err != nil {
			fmt.Printf("Error: Local source '%s' not found\n", localSource)
			os.Exit(1)
		}

		if info.IsDir() {
			if !recursive {
				fmt.Printf("Error: Local source '%s' is a directory, use -r flag to copy recursively\n", localSource)
				os.Exit(1)
			}
			// Upload directory recursively
			err = ssh.SCPUploadDirectory(conn, localSource, remotePath)
		} else {
			// Upload single file
			// If multiple sources, remotePath should be a directory
			if len(localSources) > 1 {
				// Check if remotePath exists and is a directory
				remoteFileName := filepath.Base(localSource)
				remoteFilePath := filepath.Join(remotePath, remoteFileName)
				err = ssh.SCPUploadFile(conn, localSource, remoteFilePath)
			} else {
				err = ssh.SCPUploadFile(conn, localSource, remotePath)
			}
		}

		if err != nil {
			fmt.Printf("Error uploading '%s': %v\n", localSource, err)
			os.Exit(1)
		}

		fmt.Printf("Uploaded '%s' to '%s@%s:%s'\n", localSource, conn.User, conn.Host, remotePath)
	}
}

func downloadFiles(cmd *cobra.Command, args []string) {
	// First argument is remote source
	remoteSources := args[:len(args)-1]
	localTarget := args[len(args)-1]

	// Check if local target exists
	localInfo, localErr := os.Stat(localTarget)

	// If multiple remote sources, local target must be an existing directory
	if len(remoteSources) > 1 {
		if localErr != nil || !localInfo.IsDir() {
			fmt.Printf("Error: Local target '%s' must be an existing directory when copying multiple sources\n", localTarget)
			os.Exit(1)
		}
	}

	for _, remoteSource := range remoteSources {
		// Parse remote source: remote_name:remote_path
		parts := strings.Split(remoteSource, ":")
		if len(parts) != 2 {
			fmt.Println("Error: Invalid remote source format. Use 'remote_name:remote_path'")
			os.Exit(1)
		}

		remoteName := parts[0]
		remotePath := parts[1]

		// Load connection configuration
		connections, err := config.LoadConnections()
		if err != nil {
			fmt.Println("Error loading connections:", err)
			os.Exit(1)
		}

		var conn *config.Connection
		for i, c := range connections {
			if c.Name == remoteName {
				conn = &connections[i]
				break
			}
		}

		if conn == nil {
			fmt.Printf("Error: Connection '%s' not found\n", remoteName)
			os.Exit(1)
		}

		// Determine local file path
		var localFilePath string
		if localInfo != nil && localInfo.IsDir() {
			// Local target is a directory, use remote file name
			remoteFileName := filepath.Base(remotePath)
			localFilePath = filepath.Join(localTarget, remoteFileName)
		} else {
			// Local target is a file path
			localFilePath = localTarget
		}

		// Check if remote source is a directory
		isDir, err := ssh.IsRemoteDir(conn, remotePath)
		if err != nil {
			fmt.Printf("Error checking remote path: %v\n", err)
			os.Exit(1)
		}

		if isDir {
			if !recursive {
				fmt.Printf("Error: Remote source '%s' is a directory, use -r flag to copy recursively\n", remotePath)
				os.Exit(1)
			}
			// Download directory recursively
			err = ssh.SCPDownloadDirectory(conn, remotePath, localFilePath)
		} else {
			// Download single file
			err = ssh.SCPDownloadFile(conn, remotePath, localFilePath)
		}

		if err != nil {
			fmt.Printf("Error downloading '%s': %v\n", remoteSource, err)
			os.Exit(1)
		}

		fmt.Printf("Downloaded '%s' to '%s'\n", remoteSource, localFilePath)
	}
}
