package ssh

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"gossh/internal/config"
	"gossh/internal/i18n"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cheggaaa/pb/v3"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

// executeRemoteCommandWithOutput executes a command on the remote server and returns its output.
func executeRemoteCommandWithOutput(client *ssh.Client, command string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return "", fmt.Errorf("command failed: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// getRemoteFileHash tries to get the SHA256 hash of a remote file.
// First tries to execute sha256sum command, falls back to SFTP if that fails.
func getRemoteFileHash(client *ssh.Client, remotePath string, sftpClient *sftp.Client) (string, error) {
	command := fmt.Sprintf("sha256sum %s", remotePath)
	output, err := executeRemoteCommandWithOutput(client, command)
	if err == nil {
		parts := strings.Fields(output)
		if len(parts) > 0 {
			return parts[0], nil
		}
	}

	hash, err := calculateFileHashViaSFTP(remotePath, sftpClient)
	if err != nil {
		return "", fmt.Errorf("failed to get remote file hash: %w", err)
	}

	return hash, nil
}

// calculateFileHash calculates the SHA256 hash of a local file.
func calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// calculateFileHashViaSFTP calculates the SHA256 hash of a remote file via SFTP.
func calculateFileHashViaSFTP(remotePath string, sftpClient *sftp.Client) (string, error) {
	if sftpClient == nil {
		return "", fmt.Errorf("sftp client is required")
	}

	file, err := sftpClient.Open(remotePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// filesAreEqual checks if two files are identical by comparing size and SHA256 hash.
func filesAreEqual(localPath, remotePath string, client *ssh.Client, sftpClient *sftp.Client) (bool, error) {
	localInfo, err := os.Stat(localPath)
	if err != nil {
		return false, err
	}

	remoteInfo, err := sftpClient.Stat(remotePath)
	if err != nil {
		return false, err
	}

	if localInfo.Size() != remoteInfo.Size() {
		return false, nil
	}

	localHash, err := calculateFileHash(localPath)
	if err != nil {
		return false, err
	}

	remoteHash, err := getRemoteFileHash(client, remotePath, sftpClient)
	if err != nil {
		return false, err
	}

	return localHash == remoteHash, nil
}

// getAuthMethods determines the authentication methods based on the connection configuration.
func getAuthMethods(conn *config.Connection) ([]ssh.AuthMethod, error) {
	var authMethods []ssh.AuthMethod

	if conn.KeyPath != "" {
		key, err := os.ReadFile(conn.KeyPath)
		if err != nil {
			return nil, i18n.ErrorWith("ssh.error.reading.key", map[string]interface{}{"Error": err}, err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			fmt.Print(i18n.T("ssh.enter.passphrase"))
			bytePassword, err := terminal.ReadPassword(int(os.Stdin.Fd()))
			fmt.Println()
			if err != nil {
				return nil, i18n.ErrorWith("ssh.error.reading.passphrase", map[string]interface{}{"Error": err}, err)
			}
			signer, err = ssh.ParsePrivateKeyWithPassphrase(key, bytePassword)
			if err != nil {
				return nil, i18n.ErrorWith("ssh.error.parsing.key", map[string]interface{}{"Error": err}, err)
			}
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	} else if conn.CredentialAlias != "" {
		creds, err := config.LoadCredentials()
		if err != nil {
			return nil, i18n.ErrorWith("ssh.error.loading.credentials", map[string]interface{}{"Error": err}, err)
		}
		var foundPassword string
		for _, c := range creds {
			if c.Alias == conn.CredentialAlias {
				foundPassword = c.Password
				break
			}
		}
		if foundPassword == "" {
			return nil, i18n.ErrorWith("ssh.error.password.not.found", map[string]interface{}{"Alias": conn.CredentialAlias}, fmt.Errorf("password not found"))
		}
		authMethods = append(authMethods, ssh.Password(foundPassword))
	} else {
		fmt.Print(i18n.T("ssh.enter.password"))
		bytePassword, err := terminal.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			fmt.Println()
			return nil, i18n.ErrorWith("ssh.error.reading.password", map[string]interface{}{"Error": err}, err)
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
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	fmt.Println(i18n.TWith("ssh.connecting", map[string]interface{}{
		"User": conn.User,
		"Host": conn.Host,
		"Port": conn.Port,
	}))
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", conn.Host, conn.Port), sshConfig)
	if err != nil {
		return nil, i18n.ErrorWith("ssh.error.dialing", map[string]interface{}{"Error": err}, err)
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
		fmt.Println(i18n.TWith("ssh.error.session", map[string]interface{}{"Error": err}))
		os.Exit(1)
	}
	defer session.Close()

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	fd := int(os.Stdin.Fd())
	oldState, err := terminal.MakeRaw(fd)
	if err != nil {
		fmt.Println(i18n.TWith("ssh.error.raw.terminal", map[string]interface{}{"Error": err}))
		os.Exit(1)
	}
	defer terminal.Restore(fd, oldState)

	termWidth, termHeight, err := terminal.GetSize(fd)
	if err != nil {
		termWidth = 80
		termHeight = 24
	}

	if err := session.RequestPty("xterm-256color", termHeight, termWidth, ssh.TerminalModes{}); err != nil {
		fmt.Println(i18n.TWith("ssh.error.pty", map[string]interface{}{"Error": err}))
		os.Exit(1)
	}

	if err := session.Shell(); err != nil {
		fmt.Println(i18n.TWith("ssh.error.shell", map[string]interface{}{"Error": err}))
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

	fmt.Println(i18n.TWith("ssh.executing.command", map[string]interface{}{"Command": command}))
	if err := session.Run(command); err != nil {
		err = i18n.ErrorWith("ssh.error.running.command", map[string]interface{}{"Error": err}, err)
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
	return UploadFileWithOpts(conn, localPath, remotePath, recursive, false, false)
}

// UploadFileWithOpts uploads a local file or directory to the remote server with options.
func UploadFileWithOpts(conn *config.Connection, localPath, remotePath string, recursive, force, checksum bool) error {
	client, err := newClient(conn)
	if err != nil {
		return err
	}
	defer client.Close()

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return i18n.ErrorWith("ssh.error.sftp.client", map[string]interface{}{"Error": err}, err)
	}
	defer sftpClient.Close()

	localInfo, err := os.Stat(localPath)
	if err != nil {
		return i18n.ErrorWith("ssh.error.stat.local", map[string]interface{}{"Error": err}, err)
	}

	if localInfo.IsDir() {
		if !recursive {
			return i18n.ErrorWith("ssh.error.directory.recursive", map[string]interface{}{"Path": localPath}, fmt.Errorf("is directory"))
		}
		return uploadDir(sftpClient, client, conn, localPath, remotePath, checksum)
	}

	return uploadFileWithOpts(sftpClient, client, conn, localPath, remotePath, force, checksum)
}

// uploadFile uploads a single file.
func uploadFile(sftpClient *sftp.Client, client *ssh.Client, conn *config.Connection, localPath, remotePath string, checksum bool) error {
	return uploadFileWithOpts(sftpClient, client, conn, localPath, remotePath, false, checksum)
}

// uploadFileWithOpts uploads a single file with options.
func uploadFileWithOpts(sftpClient *sftp.Client, client *ssh.Client, conn *config.Connection, localPath, remotePath string, force, checksum bool) error {
	srcFile, err := os.Open(localPath)
	if err != nil {
		return i18n.ErrorWith("ssh.error.open.local.file", map[string]interface{}{"Error": err}, err)
	}
	defer srcFile.Close()

	fileInfo, err := srcFile.Stat()
	if err != nil {
		return i18n.ErrorWith("ssh.error.get.file.info", map[string]interface{}{"Error": err}, err)
	}
	fileSize := fileInfo.Size()

	remoteInfo, err := sftpClient.Stat(remotePath)
	if err == nil && remoteInfo.IsDir() {
		_, localFileName := filepath.Split(localPath)
		remotePath = filepath.Join(remotePath, localFileName)
	}

	remoteInfo, err = sftpClient.Stat(remotePath)
	if err == nil {
		if checksum {
			equal, err := filesAreEqual(localPath, remotePath, client, sftpClient)
			if err != nil {
				return i18n.ErrorWith("ssh.error.checking.file", map[string]interface{}{"Error": err}, err)
			}
			if equal {
				fmt.Println(i18n.TWith("ssh.file.unchanged", map[string]interface{}{
					"Local":  localPath,
					"Remote": remotePath,
				}))
				return nil
			}
		}
		if !force {
			return i18n.ErrorWith("ssh.error.remote.exists", map[string]interface{}{"Path": remotePath}, fmt.Errorf("file exists"))
		}
	}

	dstFile, err := sftpClient.Create(remotePath)
	if err != nil {
		return i18n.ErrorWith("ssh.error.create.remote.file", map[string]interface{}{"Error": err}, err)
	}
	defer dstFile.Close()

	bar := pb.Full.Start64(fileSize)
	bar.Set(pb.Bytes, true)
	barReader := bar.NewProxyReader(srcFile)

	bytes, err := io.Copy(dstFile, barReader)
	bar.Finish()

	if err != nil {
		return i18n.ErrorWith("ssh.error.copying.file", map[string]interface{}{"Error": err}, err)
	}

	fmt.Println(i18n.TWith("ssh.uploaded", map[string]interface{}{
		"Local":  localPath,
		"Remote": remotePath,
		"Bytes":  bytes,
	}))
	return nil
}

// uploadDir recursively uploads a directory.
func uploadDir(sftpClient *sftp.Client, client *ssh.Client, conn *config.Connection, localPath, remotePath string, checksum bool) error {
	err := sftpClient.MkdirAll(remotePath)
	if err != nil {
		return i18n.ErrorWith("ssh.error.create.remote.dir", map[string]interface{}{"Error": err}, err)
	}

	entries, err := os.ReadDir(localPath)
	if err != nil {
		return i18n.ErrorWith("ssh.error.read.local.dir", map[string]interface{}{"Error": err}, err)
	}

	for _, entry := range entries {
		localFilePath := filepath.Join(localPath, entry.Name())
		remoteFilePath := filepath.Join(remotePath, entry.Name())

		if entry.IsDir() {
			err = uploadDir(sftpClient, client, conn, localFilePath, remoteFilePath, checksum)
			if err != nil {
				return err
			}
		} else {
			err = uploadFile(sftpClient, client, conn, localFilePath, remoteFilePath, checksum)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// DownloadFile downloads a remote file or directory to the local machine.
func DownloadFile(conn *config.Connection, remotePath, localPath string, recursive bool) error {
	return DownloadFileWithOpts(conn, remotePath, localPath, recursive, false, false)
}

// DownloadFileWithOpts downloads a remote file or directory to the local machine with options.
func DownloadFileWithOpts(conn *config.Connection, remotePath, localPath string, recursive, force, checksum bool) error {
	client, err := newClient(conn)
	if err != nil {
		return err
	}
	defer client.Close()

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return i18n.ErrorWith("ssh.error.sftp.client", map[string]interface{}{"Error": err}, err)
	}
	defer sftpClient.Close()

	remoteInfo, err := sftpClient.Stat(remotePath)
	if err != nil {
		return i18n.ErrorWith("ssh.error.stat.remote", map[string]interface{}{"Error": err}, err)
	}

	if remoteInfo.IsDir() {
		if !recursive {
			return i18n.ErrorWith("ssh.error.directory.recursive", map[string]interface{}{"Path": remotePath}, fmt.Errorf("is directory"))
		}
		return downloadDir(sftpClient, client, conn, remotePath, localPath, checksum)
	}

	return downloadFileWithOpts(sftpClient, client, conn, remotePath, localPath, force, checksum)
}

// downloadFile downloads a single file.
func downloadFile(sftpClient *sftp.Client, client *ssh.Client, conn *config.Connection, remotePath, localPath string, checksum bool) error {
	return downloadFileWithOpts(sftpClient, client, conn, remotePath, localPath, false, checksum)
}

// downloadFileWithOpts downloads a single file with options.
func downloadFileWithOpts(sftpClient *sftp.Client, client *ssh.Client, conn *config.Connection, remotePath, localPath string, force, checksum bool) error {
	srcFile, err := sftpClient.Open(remotePath)
	if err != nil {
		return i18n.ErrorWith("ssh.error.open.remote.file", map[string]interface{}{"Error": err}, err)
	}
	defer srcFile.Close()

	fileInfo, err := srcFile.Stat()
	if err != nil {
		return i18n.ErrorWith("ssh.error.get.file.info", map[string]interface{}{"Error": err}, err)
	}
	fileSize := fileInfo.Size()

	localInfo, err := os.Stat(localPath)
	if err == nil && localInfo.IsDir() {
		remoteFileName := filepath.Base(remotePath)
		localPath = filepath.Join(localPath, remoteFileName)
	}

	localInfo, err = os.Stat(localPath)
	if err == nil {
		if checksum {
			equal, err := filesAreEqual(localPath, remotePath, client, sftpClient)
			if err != nil {
				return i18n.ErrorWith("ssh.error.checking.file", map[string]interface{}{"Error": err}, err)
			}
			if equal {
				fmt.Println(i18n.TWith("ssh.file.unchanged", map[string]interface{}{
					"Remote": remotePath,
					"Local":  localPath,
				}))
				return nil
			}
		}
		if !force {
			return i18n.ErrorWith("ssh.error.local.exists", map[string]interface{}{"Path": localPath}, fmt.Errorf("file exists"))
		}
	}

	dstFile, err := os.Create(localPath)
	if err != nil {
		return i18n.ErrorWith("ssh.error.create.local.file", map[string]interface{}{"Error": err}, err)
	}
	defer dstFile.Close()

	bar := pb.Full.Start64(fileSize)
	bar.Set(pb.Bytes, true)
	barReader := bar.NewProxyReader(srcFile)

	bytes, err := io.Copy(dstFile, barReader)
	bar.Finish()

	if err != nil {
		return i18n.ErrorWith("ssh.error.copying.file", map[string]interface{}{"Error": err}, err)
	}

	fmt.Println(i18n.TWith("ssh.downloaded", map[string]interface{}{
		"Remote": remotePath,
		"Local":  localPath,
		"Bytes":  bytes,
	}))
	return nil
}

// downloadDir recursively downloads a directory.
func downloadDir(sftpClient *sftp.Client, client *ssh.Client, conn *config.Connection, remotePath, localPath string, checksum bool) error {
	err := os.MkdirAll(localPath, 0755)
	if err != nil {
		return i18n.ErrorWith("ssh.error.create.local.dir", map[string]interface{}{"Error": err}, err)
	}

	entries, err := sftpClient.ReadDir(remotePath)
	if err != nil {
		return i18n.ErrorWith("ssh.error.read.remote.dir", map[string]interface{}{"Error": err}, err)
	}

	for _, entry := range entries {
		remoteFilePath := filepath.Join(remotePath, entry.Name())
		localFilePath := filepath.Join(localPath, entry.Name())

		if entry.IsDir() {
			err = downloadDir(sftpClient, client, conn, remoteFilePath, localFilePath, checksum)
			if err != nil {
				return err
			}
		} else {
			err = downloadFile(sftpClient, client, conn, remoteFilePath, localFilePath, checksum)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
