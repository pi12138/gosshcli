package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"gossh/internal/config"
	"gossh/internal/i18n"
	"gossh/internal/ssh"
	"os"
)

var testCmd = &cobra.Command{
	Use:   "test <name>",
	Short: i18n.T("test.short"),
	Long:  i18n.T("test.long"),
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		connectionName := args[0]

		conn, err := config.ResolveConnection(connectionName)
		if err != nil {
			fmt.Println(i18n.TWith("error.connection.not.found", map[string]interface{}{"Name": connectionName}))
			os.Exit(1)
		}

		if err := ssh.TestConnection(conn); err != nil {
			fmt.Println(i18n.TWith("test.failed", map[string]interface{}{"Name": connectionName, "Error": err}))
			os.Exit(1)
		}

		fmt.Println(i18n.TWith("test.success", map[string]interface{}{"Name": connectionName}))
	},
}
