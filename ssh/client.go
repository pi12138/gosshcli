package ssh

import (
	"fmt"
	"gossh/config"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cheggaaa/pb/v3"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

// getAuthMethods determines the authentication methods based on the connection configuration.
func getAuthMethods(conn *config.Connection) ([]ssh.AuthMethod, error) {
	var authMethods []ssh.AuthMethod

	if conn.KeyPath != "" {
		// 1. Use private key if available
		key, err := ioutil.ReadFile(conn.KeyPath)
		if err != nil {
			return nil, fmt.Errorf("unable to read private key: %v", err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			// Try parsing with a passphrase
			fmt.Print("Enter passphrase for key: ")
			bytePassword, err := terminal.ReadPassword(int(os.Stdin.Fd()))
			fmt.Println()
			if err != nil {
				return nil, fmt.Errorf("failed to read passphrase: %v", err)
			}
			signer, err = ssh.ParsePrivateKeyWithPassphrase(key, bytePassword)
			if err != nil {
				return nil, fmt.Errorf("unable to parse private key with passphrase: %v", err)
			}
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	} else if conn.CredentialAlias != "" {
		// 2. Use password from credential store
		creds, err := config.LoadCredentials()
		if err != nil {
			return nil, fmt.Errorf("error loading credentials: %v", err)
		}
		var foundPassword string
		for _, c := range creds {
			if c.Alias == conn.CredentialAlias {
				foundPassword = c.Password
				break
			}
		}
		if foundPassword == "" {
			return nil, fmt.Errorf("could not find password for alias '%s'", conn.CredentialAlias)
		}
		authMethods = append(authMethods, ssh.Password(foundPassword))
	} else {
		// 3. Fallback to interactive password prompt
		fmt.Print("Enter password: ")
		bytePassword, err := terminal.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			fmt.Println()
			return nil, fmt.Errorf("failed to read password: %v", err)
		}
		fmt.Println()
		authMethods = append(authMethods, ssh.Password(string(bytePassword)))
	}
	return authMethods, nil
}

// newClient creates a new SSH client.
func newClient(conn *config.Connection) (*ssh.Client, error) {
	authMethods, err := getAuthMethods(conn)
	if err != nil {
		return nil, err
	}

	sshConfig := &ssh.ClientConfig{
		User:            conn.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Note: In a real-world app, you'd want to verify the host key.
	}

	fmt.Printf("Connecting to %s@%s:%d...\n", conn.User, conn.Host, conn.Port)
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", conn.Host, conn.Port), sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %s", err)
	}
	return client, nil
}

// Connect establishes an interactive SSH session.
func Connect(conn *config.Connection) {
	client, err := newClient(conn)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		fmt.Printf("Failed to create session: %s\n", err)
		os.Exit(1)
	}
	defer session.Close()

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	fd := int(os.Stdin.Fd())
	oldState, err := terminal.MakeRaw(fd)
	if err != nil {
		fmt.Printf("failed to make terminal raw: %v\n", err)
		os.Exit(1)
	}
	defer terminal.Restore(fd, oldState)

	termWidth, termHeight, err := terminal.GetSize(fd)
	if err != nil {
		termWidth = 80
		termHeight = 24
	}

	if err := session.RequestPty("xterm-256color", termHeight, termWidth, ssh.TerminalModes{}); err != nil {
		fmt.Printf("request for PTY failed: %s\n", err)
		os.Exit(1)
	}

	if err := session.Shell(); err != nil {
		fmt.Printf("failed to start shell: %s\n", err)
		os.Exit(1)
	}

	session.Wait()
}

// ExecuteRemoteCommand executes a non-interactive command on the remote server.
func ExecuteRemoteCommand(conn *config.Connection, command string) error {
	client, err := newClient(conn)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		err = fmt.Errorf("failed to create session: %s", err)
		fmt.Println(err)
		return err
	}
	defer session.Close()

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	fmt.Printf("Executing command: %s\n", command)
	if err := session.Run(command); err != nil {
		err = fmt.Errorf("failed to run command: %s", err)
		fmt.Println(err)
		return err
	}
	return nil
}

// TestConnection attempts to establish a connection and immediately closes it.
func TestConnection(conn *config.Connection) error {
	client, err := newClient(conn)
	if err != nil {
		return err
	}
	defer client.Close()
	// If newClient succeeded, the connection is considered successful.
	return nil
}

// UploadFile uploads a local file or directory to the remote server.
func UploadFile(conn *config.Connection, localPath, remotePath string, recursive bool) error {
	return UploadFileWithOpts(conn, localPath, remotePath, recursive, false)
}

// UploadFileWithOpts uploads a local file or directory to the remote server with options.
func UploadFileWithOpts(conn *config.Connection, localPath, remotePath string, recursive, force bool) error {
	client, err := newClient(conn)
	if err != nil {
		return err
	}
	defer client.Close()

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return fmt.Errorf("failed to create SFTP client: %v", err)
	}
	defer sftpClient.Close()

	localInfo, err := os.Stat(localPath)
	if err != nil {
		return fmt.Errorf("failed to stat local path: %v", err)
	}

	if localInfo.IsDir() {
		if !recursive {
			return fmt.Errorf("'%s' is a directory, use -r flag for recursive copy", localPath)
		}
		return uploadDir(sftpClient, localPath, remotePath)
	}

	return uploadFileWithOpts(sftpClient, localPath, remotePath, force)
}

// uploadFile uploads a single file.
func uploadFile(sftpClient *sftp.Client, localPath, remotePath string) error {
	return uploadFileWithOpts(sftpClient, localPath, remotePath, false)
}

// uploadFileWithOpts uploads a single file with options.
func uploadFileWithOpts(sftpClient *sftp.Client, localPath, remotePath string, force bool) error {
	srcFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %v", err)
	}
	defer srcFile.Close()

	// Get file info to determine size for progress bar
	fileInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}
	fileSize := fileInfo.Size()

	// Check if remote path is a directory
	remoteInfo, err := sftpClient.Stat(remotePath)
	if err == nil && remoteInfo.IsDir() {
		// If remote path is a directory, append filename to it
		_, localFileName := filepath.Split(localPath)
		remotePath = filepath.Join(remotePath, localFileName)
	}

	// Check if remote file already exists
	remoteInfo, err = sftpClient.Stat(remotePath)
	if err == nil {
		// Remote file exists
		if !force {
			return fmt.Errorf("remote file '%s' already exists. Use -f flag to force overwrite", remotePath)
		}
	}

	dstFile, err := sftpClient.Create(remotePath)
	if err != nil {
		return fmt.Errorf("failed to create remote file: %v", err)
	}
	defer dstFile.Close()

	// Create progress bar
	bar := pb.Full.Start64(fileSize)
	bar.Set(pb.Bytes, true)
	barReader := bar.NewProxyReader(srcFile)

	// Copy with progress bar
	bytes, err := io.Copy(dstFile, barReader)
	bar.Finish()

	if err != nil {
		return fmt.Errorf("failed to copy file: %v", err)
	}

	fmt.Printf("Uploaded: %s -> %s (%d bytes)\n", localPath, remotePath, bytes)
	return nil
}

// uploadDir recursively uploads a directory.
func uploadDir(sftpClient *sftp.Client, localPath, remotePath string) error {
	// Create remote directory
	err := sftpClient.MkdirAll(remotePath)
	if err != nil {
		return fmt.Errorf("failed to create remote directory: %v", err)
	}

	entries, err := os.ReadDir(localPath)
	if err != nil {
		return fmt.Errorf("failed to read local directory: %v", err)
	}

	for _, entry := range entries {
		localFilePath := filepath.Join(localPath, entry.Name())
		remoteFilePath := filepath.Join(remotePath, entry.Name())

		if entry.IsDir() {
			err = uploadDir(sftpClient, localFilePath, remoteFilePath)
			if err != nil {
				return err
			}
		} else {
			err = uploadFile(sftpClient, localFilePath, remoteFilePath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// DownloadFile downloads a remote file or directory to the local machine.
func DownloadFile(conn *config.Connection, remotePath, localPath string, recursive bool) error {
	return DownloadFileWithOpts(conn, remotePath, localPath, recursive, false)
}

// DownloadFileWithOpts downloads a remote file or directory to the local machine with options.
func DownloadFileWithOpts(conn *config.Connection, remotePath, localPath string, recursive, force bool) error {
	client, err := newClient(conn)
	if err != nil {
		return err
	}
	defer client.Close()

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return fmt.Errorf("failed to create SFTP client: %v", err)
	}
	defer sftpClient.Close()

	remoteInfo, err := sftpClient.Stat(remotePath)
	if err != nil {
		return fmt.Errorf("failed to stat remote path: %v", err)
	}

	if remoteInfo.IsDir() {
		if !recursive {
			return fmt.Errorf("'%s' is a directory, use -r flag for recursive copy", remotePath)
		}
		return downloadDir(sftpClient, remotePath, localPath)
	}

	return downloadFileWithOpts(sftpClient, remotePath, localPath, force)
}

// downloadFile downloads a single file.
func downloadFile(sftpClient *sftp.Client, remotePath, localPath string) error {
	return downloadFileWithOpts(sftpClient, remotePath, localPath, false)
}

// downloadFileWithOpts downloads a single file with options.
func downloadFileWithOpts(sftpClient *sftp.Client, remotePath, localPath string, force bool) error {
	srcFile, err := sftpClient.Open(remotePath)
	if err != nil {
		return fmt.Errorf("failed to open remote file: %v", err)
	}
	defer srcFile.Close()

	// Get remote file info to determine size for progress bar
	fileInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get remote file info: %v", err)
	}
	fileSize := fileInfo.Size()

	// Check if local path is a directory
	localInfo, err := os.Stat(localPath)
	if err == nil && localInfo.IsDir() {
		// If local path is a directory, append filename to it
		remoteFileName := filepath.Base(remotePath)
		localPath = filepath.Join(localPath, remoteFileName)
	}

	// Check if local file already exists
	localInfo, err = os.Stat(localPath)
	if err == nil {
		// Local file exists
		if !force {
			return fmt.Errorf("local file '%s' already exists. Use -f flag to force overwrite", localPath)
		}
	}

	dstFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %v", err)
	}
	defer dstFile.Close()

	// Create progress bar
	bar := pb.Full.Start64(fileSize)
	bar.Set(pb.Bytes, true)
	barReader := bar.NewProxyReader(srcFile)

	// Copy with progress bar
	bytes, err := io.Copy(dstFile, barReader)
	bar.Finish()

	if err != nil {
		return fmt.Errorf("failed to copy file: %v", err)
	}

	fmt.Printf("Downloaded: %s -> %s (%d bytes)\n", remotePath, localPath, bytes)
	return nil
}

// downloadDir recursively downloads a directory.
func downloadDir(sftpClient *sftp.Client, remotePath, localPath string) error {
	// Create local directory
	err := os.MkdirAll(localPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create local directory: %v", err)
	}

	entries, err := sftpClient.ReadDir(remotePath)
	if err != nil {
		return fmt.Errorf("failed to read remote directory: %v", err)
	}

	for _, entry := range entries {
		remoteFilePath := filepath.Join(remotePath, entry.Name())
		localFilePath := filepath.Join(localPath, entry.Name())

		if entry.IsDir() {
			err = downloadDir(sftpClient, remoteFilePath, localFilePath)
			if err != nil {
				return err
			}
		} else {
			err = downloadFile(sftpClient, remoteFilePath, localFilePath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
