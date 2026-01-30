package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"gossh/internal/config"
	"gossh/internal/i18n"
	"gossh/internal/ssh"
	"os"
	"path/filepath"
	"strings"
)

var copyCmd = &cobra.Command{
	Use:   "copy [user@host]",
	Short: i18n.T("copy.short"),
	Long:  i18n.T("copy.long"),
	Run:   runCopy,
}

func runCopy(cmd *cobra.Command, args []string) {
	connName, _ := cmd.Flags().GetString("name")
	port, _ := cmd.Flags().GetInt("port")

	var conn *config.Connection

	if connName != "" {
		if len(args) > 0 {
			fmt.Println(i18n.T("copy.error.args.with.name"))
			os.Exit(1)
		}
		connections, err := config.LoadConnections()
		if err != nil {
			fmt.Println(i18n.TWith("error.loading.connections", map[string]interface{}{"Error": err}))
			os.Exit(1)
		}
		for i, c := range connections {
			if c.Name == connName {
				conn = &connections[i]
				break
			}
		}
		if conn == nil {
			fmt.Println(i18n.TWith("error.connection.not.found", map[string]interface{}{"Name": connName}))
			os.Exit(1)
		}
	} else {
		if len(args) != 1 {
			fmt.Println(i18n.T("copy.error.no.args"))
			_ = cmd.Help()
			os.Exit(1)
		}
		parts := strings.Split(args[0], "@")
		if len(parts) != 2 {
			fmt.Println(i18n.T("copy.error.invalid.format"))
			os.Exit(1)
		}
		conn = &config.Connection{
			Name: "temp-copy",
			User: parts[0],
			Host: parts[1],
			Port: port,
		}
	}

	pubKeyPath, _ := cmd.Flags().GetString("pubkey")
	if pubKeyPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(i18n.TWith("ssh.error.reading.key", map[string]interface{}{"Error": err}))
			os.Exit(1)
		}
		pubKeyPath = filepath.Join(home, ".ssh", "id_rsa.pub")
	}

	pubKey, err := os.ReadFile(pubKeyPath)
	if err != nil {
		fmt.Println(i18n.TWith("copy.error.reading.key", map[string]interface{}{"Path": pubKeyPath, "Error": err}))
		os.Exit(1)
	}

	remoteCmd := fmt.Sprintf("mkdir -p ~/.ssh && chmod 700 ~/.ssh && echo '%s' >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys", strings.TrimSpace(string(pubKey)))

	fmt.Println(i18n.TWith("copy.attempting", map[string]interface{}{"User": conn.User, "Host": conn.Host}))

	tempConn := *conn
	tempConn.KeyPath = ""
	tempConn.CredentialAlias = ""

	err = ssh.ExecuteRemoteCommand(&tempConn, remoteCmd)
	if err != nil {
		fmt.Println(i18n.T("copy.failed"))
		os.Exit(1)
	}

	fmt.Println(i18n.T("copy.success"))
}

func init() {
	home, err := os.UserHomeDir()
	defaultPubKeyPath := ""
	if err == nil {
		defaultPubKeyPath = filepath.Join(home, ".ssh", "id_rsa.pub")
	}

	copyCmd.Flags().StringP("name", "n", "", i18n.T("copy.flag.name"))
	copyCmd.Flags().IntP("port", "p", 22, i18n.T("copy.flag.port"))
	copyCmd.Flags().StringP("pubkey", "i", defaultPubKeyPath, i18n.T("copy.flag.pubkey"))
}
