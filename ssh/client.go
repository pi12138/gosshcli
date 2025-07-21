package ssh

import (
	"fmt"
	"gossh/config"
	"io/ioutil"
	"os"

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
