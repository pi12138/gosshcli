package scp

import (
	"fmt"
	"gossh/config"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

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

// SCPSession represents an SCP session for file transfers
type SCPSession struct {
	client *ssh.Client
}

// NewSCPSession creates a new SCP session
func NewSCPSession(conn *config.Connection) (*SCPSession, error) {
	authMethods, err := getAuthMethods(conn)
	if err != nil {
		return nil, err
	}

	sshConfig := &ssh.ClientConfig{
		User:            conn.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", conn.Host, conn.Port), sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %s", err)
	}

	return &SCPSession{client: client}, nil
}

// Close closes the SCP session
func (s *SCPSession) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}

// UploadFile uploads a local file to the remote server
func (s *SCPSession) UploadFile(localPath, remotePath string, preserveTimes bool) error {
	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %v", err)
	}
	defer localFile.Close()

	fileInfo, err := localFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat local file: %v", err)
	}

	session, err := s.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	// Prepare SCP command
	remoteDir := path.Dir(remotePath)
	remoteFileName := path.Base(remotePath)
	
	// Create directory if it doesn't exist
	mkdirCmd := fmt.Sprintf("mkdir -p %s", remoteDir)
	if err := s.executeCommand(mkdirCmd); err != nil {
		return fmt.Errorf("failed to create remote directory: %v", err)
	}

	// SCP protocol: send file permissions, size, filename, then file content
	mode := fileInfo.Mode().Perm()
	size := fileInfo.Size()
	modTime := fileInfo.ModTime()

	scpCommand := fmt.Sprintf("scp -t %s", remotePath)
	
	stdin, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdin pipe: %v", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %v", err)
	}

	if err := session.Start(scpCommand); err != nil {
		return fmt.Errorf("failed to start scp command: %v", err)
	}

	// Wait for server acknowledgment
	ack := make([]byte, 1)
	if _, err := stdout.Read(ack); err != nil {
		return fmt.Errorf("failed to read server acknowledgment: %v", err)
	}

	if ack[0] != 0 {
		return fmt.Errorf("server rejected scp command")
	}

	// Send file metadata
	var metadata string
	if preserveTimes {
		metadata = fmt.Sprintf("C%04o %d %s\n", mode, size, remoteFileName)
	} else {
		metadata = fmt.Sprintf("C%04o %d %s\n", mode, size, remoteFileName)
	}

	if _, err := stdin.Write([]byte(metadata)); err != nil {
		return fmt.Errorf("failed to send file metadata: %v", err)
	}

	// Wait for acknowledgment
	if _, err := stdout.Read(ack); err != nil {
		return fmt.Errorf("failed to read metadata acknowledgment: %v", err)
	}

	if ack[0] != 0 {
		return fmt.Errorf("server rejected file metadata")
	}

	// Send file content
	if _, err := io.CopyN(stdin, localFile, size); err != nil {
		return fmt.Errorf("failed to send file content: %v", err)
	}

	// Send null byte to indicate end of file
	if _, err := stdin.Write([]byte{0}); err != nil {
		return fmt.Errorf("failed to send end-of-file marker: %v", err)
	}

	// Wait for final acknowledgment
	if _, err := stdout.Read(ack); err != nil {
		return fmt.Errorf("failed to read final acknowledgment: %v", err)
	}

	if ack[0] != 0 {
		return fmt.Errorf("server rejected file transfer")
	}

	// Set modification time if requested
	if preserveTimes {
		touchCmd := fmt.Sprintf("touch -t %s %s", modTime.Format("200601021504.05"), remotePath)
		if err := s.executeCommand(touchCmd); err != nil {
			fmt.Printf("Warning: failed to preserve modification time: %v\n", err)
		}
	}

	return session.Wait()
}

// DownloadFile downloads a remote file to the local system
func (s *SCPSession) DownloadFile(remotePath, localPath string, preserveTimes bool) error {
	// Create local directory if it doesn't exist
	localDir := filepath.Dir(localPath)
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return fmt.Errorf("failed to create local directory: %v", err)
	}

	session, err := s.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	scpCommand := fmt.Sprintf("scp -f %s", remotePath)
	
	stdin, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdin pipe: %v", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %v", err)
	}

	if err := session.Start(scpCommand); err != nil {
		return fmt.Errorf("failed to start scp command: %v", err)
	}

	// Send acknowledgment to start transfer
	if _, err := stdin.Write([]byte{0}); err != nil {
		return fmt.Errorf("failed to send initial acknowledgment: %v", err)
	}

	// Read file metadata
	metadata := make([]byte, 1024)
	n, err := stdout.Read(metadata)
	if err != nil {
		return fmt.Errorf("failed to read file metadata: %v", err)
	}

	metadataStr := string(metadata[:n])
	parts := strings.Fields(metadataStr)
	if len(parts) < 3 {
		return fmt.Errorf("invalid file metadata format")
	}

	// Parse file permissions and size
	var mode uint32
	var size int64
	if _, err := fmt.Sscanf(parts[0], "C%04o", &mode); err != nil {
		return fmt.Errorf("failed to parse file permissions: %v", err)
	}

	size, err = strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse file size: %v", err)
	}

	filename := parts[2]
	if strings.HasSuffix(filename, "\n") {
		filename = strings.TrimSuffix(filename, "\n")
	}

	// Send acknowledgment
	if _, err := stdin.Write([]byte{0}); err != nil {
		return fmt.Errorf("failed to send metadata acknowledgment: %v", err)
	}

	// Create local file
	localFile, err := os.OpenFile(localPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(mode))
	if err != nil {
		return fmt.Errorf("failed to create local file: %v", err)
	}
	defer localFile.Close()

	// Read file content
	if _, err := io.CopyN(localFile, stdout, size); err != nil {
		return fmt.Errorf("failed to download file content: %v", err)
	}

	// Send final acknowledgment
	if _, err := stdin.Write([]byte{0}); err != nil {
		return fmt.Errorf("failed to send final acknowledgment: %v", err)
	}

	// Set modification time if requested and available
	if preserveTimes {
		// Note: Standard SCP doesn't preserve timestamps in the protocol
		// This would require additional commands or using SFTP
		fmt.Println("Note: SCP protocol doesn't preserve timestamps by default")
	}

	return session.Wait()
}

// UploadDirectory uploads a local directory to the remote server recursively
func (s *SCPSession) UploadDirectory(localPath, remotePath string, preserveTimes bool) error {
	return filepath.Walk(localPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(localPath, path)
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if relPath == "." {
			return nil
		}

		remoteFilePath := filepath.Join(remotePath, relPath)
		remoteFilePath = filepath.ToSlash(remoteFilePath) // Convert to Unix path separators

		if info.IsDir() {
			// Create remote directory
			mkdirCmd := fmt.Sprintf("mkdir -p %s", remoteFilePath)
			return s.executeCommand(mkdirCmd)
		} else {
			// Upload file
			return s.UploadFile(path, remoteFilePath, preserveTimes)
		}
	})
}

// DownloadDirectory downloads a remote directory to the local system recursively
func (s *SCPSession) DownloadDirectory(remotePath, localPath string, preserveTimes bool) error {
	// First, get directory listing
	listCmd := fmt.Sprintf("find %s -type f", remotePath)
	output, err := s.ExecuteCommand(listCmd)
	if err != nil {
		return fmt.Errorf("failed to list remote directory: %v", err)
	}

	files := strings.Split(strings.TrimSpace(output), "\n")
	for _, file := range files {
		file = strings.TrimSpace(file)
		if file == "" {
			continue
		}

		// Calculate relative path
		relPath, err := filepath.Rel(remotePath, file)
		if err != nil {
			return err
		}

		localFilePath := filepath.Join(localPath, relPath)

		// Create local directory
		if err := os.MkdirAll(filepath.Dir(localFilePath), 0755); err != nil {
			return fmt.Errorf("failed to create local directory: %v", err)
		}

		// Download file
		if err := s.DownloadFile(file, localFilePath, preserveTimes); err != nil {
			return fmt.Errorf("failed to download file %s: %v", file, err)
		}
	}

	return nil
}

// executeCommand executes a remote command
func (s *SCPSession) executeCommand(command string) error {
	session, err := s.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	return session.Run(command)
}

// ExecuteCommand executes a remote command and returns output
func (s *SCPSession) ExecuteCommand(command string) (string, error) {
	session, err := s.client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	output, err := session.Output(command)
	return string(output), err
}