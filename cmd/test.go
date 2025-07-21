package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"gossh/config"
	"gossh/ssh"
	"os"
)

var testCmd = &cobra.Command{
	Use:   "test <name>",
	Short: "Test a connection configuration",
	Long: `Tests a saved connection configuration by attempting to establish an SSH connection. 
It authenticates and then immediately disconnects, reporting success or failure.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		connectionName := args[0]

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

		// We need a way to test the connection without a full interactive session.
		// Let's create a dedicated function in the ssh package for this.
		if err := ssh.TestConnection(conn); err != nil {
			fmt.Printf("Connection test for '%s' failed: %v\n", connectionName, err)
			os.Exit(1)
		}

		fmt.Printf("Connection test for '%s' successful!\n", connectionName)
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
