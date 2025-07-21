package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"gossh/config"
	"os"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new connection configuration",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		user, _ := cmd.Flags().GetString("user")
		host, _ := cmd.Flags().GetString("host")
		port, _ := cmd.Flags().GetInt("port")
		keyPath, _ := cmd.Flags().GetString("key")
		credAlias, _ := cmd.Flags().GetString("use-password")

		if name == "" || user == "" || host == "" {
			fmt.Println("Error: --name, --user, and --host are required")
			os.Exit(1)
		}
		if keyPath != "" && credAlias != "" {
			fmt.Println("Error: --key and --use-password flags cannot be used together")
			os.Exit(1)
		}

		if credAlias != "" {
			creds, err := config.LoadCredentials()
			if err != nil {
				fmt.Println("Error loading credentials:", err)
				os.Exit(1)
			}
			found := false
			for _, c := range creds {
				if c.Alias == credAlias {
					found = true
					break
				}
			}
			if !found {
				fmt.Printf("Error: credential with alias '%s' not found. Use 'gossh password add %s' to create it.\n", credAlias, credAlias)
				os.Exit(1)
			}
		}

		conn := config.Connection{
			Name:            name,
			User:            user,
			Host:            host,
			Port:            port,
			KeyPath:         keyPath,
			CredentialAlias: credAlias,
		}

		if err := config.AddConnection(conn); err != nil {
			fmt.Println("Error adding connection:", err)
			os.Exit(1)
		}
		fmt.Printf("Connection '%s' added successfully.\n", name)
	},
}

func init() {
	addCmd.Flags().StringP("name", "n", "", "Connection name (required)")
	addCmd.Flags().StringP("user", "u", "", "Username (required)")
	addCmd.Flags().StringP("host", "H", "", "Host address (required)")
	addCmd.Flags().IntP("port", "p", 22, "Port number")
	addCmd.Flags().StringP("key", "k", "", "Path to private key")
	addCmd.Flags().StringP("use-password", "P", "", "Use a saved password by its alias")
	rootCmd.AddCommand(addCmd)
}
