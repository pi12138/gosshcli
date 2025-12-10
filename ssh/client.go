package ssh

import (
	"fmt"
	"gossh/config"
	"io"
	"io/ioutil"
	"os"
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

// scpCommand creates an SCP command string for the given operation.
func scpCommand(operation string, path string, recursive bool) string {
	cmd := "scp -t"
	if recursive {
		cmd += " -r"
	}
	cmd += " " + path
	return cmd
}

// getSCPSession creates an SSH session configured for SCP operations.
func getSCPSession(conn *config.Connection) (*ssh.Client, *ssh.Session, io.WriteCloser, io.Reader, error) {
	client, err := newClient(conn)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, nil, nil, nil, fmt.Errorf("failed to create session: %s", err)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		session.Close()
		client.Close()
		return nil, nil, nil, nil, fmt.Errorf("failed to get stdin pipe: %s", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		stdin.Close()
		session.Close()
		client.Close()
		return nil, nil, nil, nil, fmt.Errorf("failed to get stdout pipe: %s", err)
	}

	return client, session, stdin, stdout, nil
}

// sendSCPCommand sends an SCP command and waits for the remote response.
func sendSCPCommand(w io.Writer, r io.Reader, command string) error {
	_, err := fmt.Fprintln(w, command)
	if err != nil {
		return err
	}
	return readSCPResponse(r)
}

// readSCPResponse reads the response from the SCP server.
func readSCPResponse(r io.Reader) error {
	buf := make([]byte, 1)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return err
	}
	if buf[0] == 0 {
		// Success
		return nil
	}
	if buf[0] != 1 {
		return fmt.Errorf("unknown SCP response code: %d", buf[0])
	}
	// Error message follows
	msg, err := readSCPError(r)
	if err != nil {
		return err
	}
	return fmt.Errorf("SCP error: %s", msg)
}

// readSCPError reads an error message from the SCP server.
func readSCPError(r io.Reader) (string, error) {
	var msg strings.Builder
	buf := make([]byte, 1)
	for {
		_, err := io.ReadFull(r, buf)
		if err != nil {
			return "", err
		}
		if buf[0] == '\n' {
			break
		}
		msg.Write(buf)
	}
	return msg.String(), nil
}

// SCPUploadFile uploads a single file to a remote server using SCP.
func SCPUploadFile(conn *config.Connection, localPath string, remotePath string) error {
	client, session, stdin, stdout, err := getSCPSession(conn)
	if err != nil {
		return err
	}
	defer client.Close()
	defer session.Close()
	defer stdin.Close()

	// Start the SCP command on the remote server
	if err := session.Start(scpCommand("-t", remotePath, false)); err != nil {
		return fmt.Errorf("failed to start SCP command: %s", err)
	}

	// Wait for SCP server to be ready
	if err := readSCPResponse(stdout); err != nil {
		return err
	}

	// Get file info
	fileInfo, err := os.Stat(localPath)
	if err != nil {
		return fmt.Errorf("failed to get file info: %s", err)
	}

	// Send file header
	fileMode := strconv.FormatUint(uint64(fileInfo.Mode()&0777), 8)
	fileSize := strconv.FormatInt(fileInfo.Size(), 10)
	fileName := filepath.Base(localPath)
	header := fmt.Sprintf("C%s %s %s\n", fileMode, fileSize, fileName)
	if err := sendSCPCommand(stdin, stdout, header); err != nil {
		return err
	}

	// Open and send file content
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %s", err)
	}
	defer file.Close()

	if _, err := io.Copy(stdin, file); err != nil {
		return fmt.Errorf("failed to send file content: %s", err)
	}

	// Send terminating zero byte
	if _, err := stdin.Write([]byte{0}); err != nil {
		return fmt.Errorf("failed to send terminating byte: %s", err)
	}

	// Wait for final response
	if err := readSCPResponse(stdout); err != nil {
		return err
	}

	// Wait for session to complete
	if err := session.Wait(); err != nil {
		return fmt.Errorf("SCP session failed: %s", err)
	}

	return nil
}

// SCPUploadDirectory recursively uploads a directory to a remote server using SCP.
func SCPUploadDirectory(conn *config.Connection, localPath string, remotePath string) error {
	client, session, stdin, stdout, err := getSCPSession(conn)
	if err != nil {
		return err
	}
	defer client.Close()
	defer session.Close()
	defer stdin.Close()

	// Start the SCP command on the remote server
	if err := session.Start(scpCommand("-t", remotePath, true)); err != nil {
		return fmt.Errorf("failed to start SCP command: %s", err)
	}

	// Wait for SCP server to be ready
	if err := readSCPResponse(stdout); err != nil {
		return err
	}

	// Upload directory contents
	if err := uploadDirectoryRecursive(stdin, stdout, localPath, filepath.Base(localPath)); err != nil {
		return err
	}

	// Send end of directory marker
	if _, err := fmt.Fprintln(stdin, "E"); err != nil {
		return err
	}

	// Wait for final response
	if err := readSCPResponse(stdout); err != nil {
		return err
	}

	// Wait for session to complete
	if err := session.Wait(); err != nil {
		return fmt.Errorf("SCP session failed: %s", err)
	}

	return nil
}

// uploadDirectoryRecursive recursively uploads a directory.
func uploadDirectoryRecursive(w io.Writer, r io.Reader, localPath, remoteName string) error {
	// Get directory info
	dirInfo, err := os.Stat(localPath)
	if err != nil {
		return fmt.Errorf("failed to get directory info: %s", err)
	}

	// Send directory header
	dirMode := strconv.FormatUint(uint64(dirInfo.Mode()&0777|040000), 8) // Add S_IFDIR (040000)
	if err := sendSCPCommand(w, r, fmt.Sprintf("D%s 0 %s\n", dirMode, remoteName)); err != nil {
		return err
	}

	// Read directory contents
	entries, err := os.ReadDir(localPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %s", err)
	}

	// Upload each entry
	for _, entry := range entries {
		subLocalPath := filepath.Join(localPath, entry.Name())
		subRemoteName := entry.Name()

		if entry.IsDir() {
			// Recursively upload subdirectory
			if err := uploadDirectoryRecursive(w, r, subLocalPath, subRemoteName); err != nil {
				return err
			}
			// Send end of directory marker
			if _, err := fmt.Fprintln(w, "E"); err != nil {
				return err
			}
			// Wait for response
			if err := readSCPResponse(r); err != nil {
				return err
			}
		} else {
			// Upload file
			fileInfo, err := entry.Info()
			if err != nil {
				return fmt.Errorf("failed to get file info: %s", err)
			}

			// Send file header
			fileMode := strconv.FormatUint(uint64(fileInfo.Mode()&0777), 8)
			fileSize := strconv.FormatInt(fileInfo.Size(), 10)
			header := fmt.Sprintf("C%s %s %s\n", fileMode, fileSize, subRemoteName)
			if err := sendSCPCommand(w, r, header); err != nil {
				return err
			}

			// Open and send file content
			file, err := os.Open(subLocalPath)
			if err != nil {
				return fmt.Errorf("failed to open local file: %s", err)
			}

			if _, err := io.Copy(w, file); err != nil {
				file.Close()
				return fmt.Errorf("failed to send file content: %s", err)
			}
			file.Close()

			// Send terminating zero byte
			if _, err := w.Write([]byte{0}); err != nil {
				return fmt.Errorf("failed to send terminating byte: %s", err)
			}

			// Wait for response
			if err := readSCPResponse(r); err != nil {
				return err
			}
		}
	}

	return nil
}

// SCPDownloadFile downloads a single file from a remote server using SCP.
func SCPDownloadFile(conn *config.Connection, remotePath string, localPath string) error {
	client, session, stdin, stdout, err := getSCPSession(conn)
	if err != nil {
		return err
	}
	defer client.Close()
	defer session.Close()
	defer stdin.Close()

	// Start the SCP command on the remote server
	if err := session.Start(fmt.Sprintf("scp -f %s", remotePath)); err != nil {
		return fmt.Errorf("failed to start SCP command: %s", err)
	}

	// Wait for SCP server to be ready
	if err := readSCPResponse(stdout); err != nil {
		return err
	}

	// Read SCP commands from server
	buf := make([]byte, 1)
	_, err = io.ReadFull(stdout, buf)
	if err != nil {
		return fmt.Errorf("failed to read SCP command: %s", err)
	}

	if buf[0] != 'C' {
		if buf[0] == 1 {
			// Error message follows
			msg, _ := readSCPError(stdout)
			return fmt.Errorf("SCP error: %s", msg)
		}
		return fmt.Errorf("unknown SCP command: %c", buf[0])
	}

	// Read file header: "C<mode> <size> <name>\n"
	header, err := readSCPHeader(stdout)
	if err != nil {
		return fmt.Errorf("failed to read file header: %s", err)
	}

	parts := strings.Fields(header)
	if len(parts) < 3 {
		return fmt.Errorf("invalid file header: %s", header)
	}

	fileMode, _ := strconv.ParseUint(parts[0], 8, 32)
	fileSize, _ := strconv.ParseInt(parts[1], 10, 64)

	// Send acknowledgment
	if err := sendSCPCommand(stdin, stdout, ""); err != nil {
		return err
	}

	// Create parent directories if they don't exist
	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return fmt.Errorf("failed to create parent directories: %s", err)
	}

	// Open local file for writing
	file, err := os.OpenFile(localPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(fileMode))
	if err != nil {
		return fmt.Errorf("failed to open local file: %s", err)
	}
	defer file.Close()

	// Read file content
	_, err = io.CopyN(file, stdout, fileSize)
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to read file content: %s", err)
	}

	// Read terminating zero byte
	_, err = io.ReadFull(stdout, buf)
	if err != nil {
		return fmt.Errorf("failed to read terminating byte: %s", err)
	}

	// Send acknowledgment
	if err := sendSCPCommand(stdin, stdout, ""); err != nil {
		return err
	}

	// Wait for session to complete
	if err := session.Wait(); err != nil {
		return fmt.Errorf("SCP session failed: %s", err)
	}

	return nil
}

// SCPDownloadDirectory recursively downloads a directory from a remote server using SCP.
func SCPDownloadDirectory(conn *config.Connection, remotePath string, localPath string) error {
	client, session, stdin, stdout, err := getSCPSession(conn)
	if err != nil {
		return err
	}
	defer client.Close()
	defer session.Close()
	defer stdin.Close()

	// Start the SCP command on the remote server
	if err := session.Start(fmt.Sprintf("scp -rf %s", remotePath)); err != nil {
		return fmt.Errorf("failed to start SCP command: %s", err)
	}

	// Wait for SCP server to be ready
	if err := readSCPResponse(stdout); err != nil {
		return err
	}

	// Create the local directory if it doesn't exist
	if err := os.MkdirAll(localPath, 0755); err != nil {
		return fmt.Errorf("failed to create local directory: %s", err)
	}

	// Process SCP commands
	if err := downloadDirectoryRecursive(stdin, stdout, localPath); err != nil {
		return err
	}

	// Wait for session to complete
	if err := session.Wait(); err != nil {
		return fmt.Errorf("SCP session failed: %s", err)
	}

	return nil
}

// downloadDirectoryRecursive recursively downloads a directory.
func downloadDirectoryRecursive(w io.Writer, r io.Reader, localPath string) error {
	for {
		// Read command byte
		buf := make([]byte, 1)
		_, err := io.ReadFull(r, buf)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("failed to read SCP command: %s", err)
		}

		command := buf[0]

		if command == 'E' {
			// End of directory
			// Send acknowledgment
			if err := sendSCPCommand(w, r, ""); err != nil {
				return err
			}
			return nil
		} else if command == 'C' {
			// File
			header, err := readSCPHeader(r)
			if err != nil {
				return fmt.Errorf("failed to read file header: %s", err)
			}

			parts := strings.Fields(header)
			if len(parts) < 3 {
				return fmt.Errorf("invalid file header: %s", header)
			}

			fileMode, _ := strconv.ParseUint(parts[0], 8, 32)
			fileSize, _ := strconv.ParseInt(parts[1], 10, 64)
			fileName := parts[2]

			// Send acknowledgment
			if err := sendSCPCommand(w, r, ""); err != nil {
				return err
			}

			// Open local file for writing
			filePath := filepath.Join(localPath, fileName)
			file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(fileMode))
			if err != nil {
				return fmt.Errorf("failed to open local file: %s", err)
			}

			// Read file content
			_, err = io.CopyN(file, r, fileSize)
			file.Close()
			if err != nil && err != io.EOF {
				return fmt.Errorf("failed to read file content: %s", err)
			}

			// Read terminating zero byte
			_, err = io.ReadFull(r, buf)
			if err != nil {
				return fmt.Errorf("failed to read terminating byte: %s", err)
			}

			// Send acknowledgment
			if err := sendSCPCommand(w, r, ""); err != nil {
				return err
			}
		} else if command == 'D' {
			// Directory
			header, err := readSCPHeader(r)
			if err != nil {
				return fmt.Errorf("failed to read directory header: %s", err)
			}

			parts := strings.Fields(header)
			if len(parts) < 3 {
				return fmt.Errorf("invalid directory header: %s", header)
			}

			dirMode, _ := strconv.ParseUint(parts[0], 8, 32)
			dirName := parts[2]

			// Create the directory
			subLocalPath := filepath.Join(localPath, dirName)
			if err := os.MkdirAll(subLocalPath, os.FileMode(dirMode)); err != nil {
				return fmt.Errorf("failed to create directory: %s", err)
			}

			// Send acknowledgment
			if err := sendSCPCommand(w, r, ""); err != nil {
				return err
			}

			// Recursively download the directory
			if err := downloadDirectoryRecursive(w, r, subLocalPath); err != nil {
				return err
			}
		} else if command == 1 {
			// Error
			msg, _ := readSCPError(r)
			return fmt.Errorf("SCP error: %s", msg)
		} else {
			return fmt.Errorf("unknown SCP command: %c", command)
		}
	}
}

// readSCPHeader reads an SCP header line (until newline).
func readSCPHeader(r io.Reader) (string, error) {
	var header strings.Builder
	buf := make([]byte, 1)
	for {
		_, err := io.ReadFull(r, buf)
		if err != nil {
			return "", err
		}
		if buf[0] == '\n' {
			break
		}
		header.Write(buf)
	}
	return header.String(), nil
}

// IsRemoteDir checks if a remote path is a directory.
func IsRemoteDir(conn *config.Connection, remotePath string) (bool, error) {
	client, err := newClient(conn)
	if err != nil {
		return false, err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return false, fmt.Errorf("failed to create session: %s", err)
	}
	defer session.Close()

	// Use test -d to check if path is a directory
	cmd := fmt.Sprintf("test -d %q", remotePath)
	if err := session.Run(cmd); err != nil {
		// If the command fails with exit code 1, the path is not a directory
		if exitErr, ok := err.(*ssh.ExitError); ok && exitErr.ExitStatus() == 1 {
			return false, nil
		}
		// Other errors
		return false, fmt.Errorf("failed to check remote path: %s", err)
	}

	// Command succeeded, path is a directory
	return true, nil
}
