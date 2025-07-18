package ssh

import (
	"fmt"
	"gossh/config"
	"io/ioutil"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

func Connect(conn *config.Connection) {
	var authMethods []ssh.AuthMethod

	if conn.KeyPath != "" {
		// 1. Use private key if available
		key, err := ioutil.ReadFile(conn.KeyPath)
		if err != nil {
			fmt.Printf("Unable to read private key: %v\n", err)
			os.Exit(1)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			fmt.Printf("Unable to parse private key: %v\n", err)
			os.Exit(1)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	} else if conn.CredentialAlias != "" {
		// 2. Use password from credential store
		creds, err := config.LoadCredentials()
		if err != nil {
			fmt.Println("Error loading credentials:", err)
			os.Exit(1)
		}
		var foundPassword string
		for _, c := range creds {
			if c.Alias == conn.CredentialAlias {
				foundPassword = c.Password
				break
			}
		}
		if foundPassword == "" {
			fmt.Printf("Error: could not find password for alias '%s'.\n", conn.CredentialAlias)
			os.Exit(1)
		}
		authMethods = append(authMethods, ssh.Password(foundPassword))
	} else {
		// 3. Fallback to interactive password prompt
		fmt.Print("Enter password: ")
		bytePassword, err := terminal.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			fmt.Println("\nFailed to read password:", err)
			os.Exit(1)
		}
		fmt.Println()
		authMethods = append(authMethods, ssh.Password(string(bytePassword)))
	}

	sshConfig := &ssh.ClientConfig{
		User:            conn.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	fmt.Printf("Connecting to %s@%s:%d...\n", conn.User, conn.Host, conn.Port)
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", conn.Host, conn.Port), sshConfig)
	if err != nil {
		fmt.Printf("Failed to dial: %s\n", err)
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

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // enable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	fd := int(os.Stdin.Fd())
	oldState, err := terminal.MakeRaw(fd)
	if err != nil {
		fmt.Printf("failed to make terminal raw: %v\n", err)
		os.Exit(1)
	}
	defer terminal.Restore(fd, oldState)

	termWidth, termHeight, err := terminal.GetSize(fd)
	if err != nil {
		// Ignore error, use default size
		termWidth = 80
		termHeight = 24
	}

	if err := session.RequestPty("xterm-256color", termHeight, termWidth, modes); err != nil {
		fmt.Printf("request for PTY failed: %s\n", err)
		os.Exit(1)
	}

	if err := session.Shell(); err != nil {
		fmt.Printf("failed to start shell: %s\n", err)
		os.Exit(1)
	}

	session.Wait()
}