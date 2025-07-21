package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"gossh/config"
	"gossh/ssh"
	"os"
	"strings"
)

var execCmd = &cobra.Command{
	Use:   "exec <name> <command>",
	Short: "Execute a command on a remote server",
	Long: `Execute a command on a remote server without starting an interactive session. 
The command should be provided as a single string argument.`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		connectionName := args[0]
		command := strings.Join(args[1:], " ")

		connections, err := config.LoadConnections()
		if err != nil {
			fmt.Println("Error loading connections:", err)
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
			fmt.Printf("Error: connection '%s' not found\n", connectionName)
			os.Exit(1)
		}

		if err := ssh.ExecuteRemoteCommand(conn, command); err != nil {
			// The error is already printed in the ssh package, just exit.
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(execCmd)
}
