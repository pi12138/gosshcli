package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"gossh/config"
	"os"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List local connection configurations",
	Run: func(cmd *cobra.Command, args []string) {
		connections, err := config.LoadConnections()
		if err != nil {
			fmt.Println("Error loading connections:", err)
			os.Exit(1)
		}
		if len(connections) == 0 {
			fmt.Println("No connections configured. Use 'gossh add' to create one.")
			return
		}
		fmt.Println("Available connections:")
		for _, c := range connections {
			authMethod := "interactive"
			if c.KeyPath != "" {
				authMethod = "key"
			} else if c.CredentialAlias != "" {
				authMethod = fmt.Sprintf("password (alias: %s)", c.CredentialAlias)
			}
			fmt.Printf("- %s (%s@%s:%d) (auth: %s)\n", c.Name, c.User, c.Host, c.Port, authMethod)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
