package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"gossh/internal/config"
	"gossh/internal/i18n"
	"gossh/internal/ssh"
	"os"
	"strings"
)

var execCmd = &cobra.Command{
	Use:   "exec <name> <command>",
	Short: i18n.T("exec.short"),
	Long:  i18n.T("exec.long"),
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		connectionName := args[0]
		command := strings.Join(args[1:], " ")

		connections, err := config.LoadConnections()
		if err != nil {
			fmt.Println(i18n.TWith("error.loading.connections", map[string]interface{}{"Error": err}))
			os.Exit(1)
		}

		var conn *config.Connection
		for i, c := range connections {
			if c.Name == connectionName {
				conn = &connections[i]
				break
			}
		}

		if conn == nil {
			fmt.Println(i18n.TWith("error.connection.not.found", map[string]interface{}{"Name": connectionName}))
			os.Exit(1)
		}

		if err := ssh.ExecuteRemoteCommand(conn, command); err != nil {
			os.Exit(1)
		}
	},
}
