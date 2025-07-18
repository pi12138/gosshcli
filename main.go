package main

import (
	"fmt"
	"gossh/cmd"
	"gossh/config"
	"gossh/ssh"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gossh",
	Short: "gossh is a ssh client implemented in Go",
}

var connectCmd = &cobra.Command{
	Use:   "connect [name]",
	Short: "Establish a remote ssh connection to a saved configuration",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		connectByName(args[0])
	},
}

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

var removeCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove a connection configuration",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		if err := config.RemoveConnection(name); err != nil {
			fmt.Println("Error removing connection:", err)
			os.Exit(1)
		}
		fmt.Printf("Connection '%s' removed successfully.\n", name)
	},
}

func init() {
	addCmd.Flags().StringP("name", "n", "", "Connection name (required)")
	addCmd.Flags().StringP("user", "u", "", "Username (required)")
	addCmd.Flags().StringP("host", "H", "", "Host address (required)")
	addCmd.Flags().IntP("port", "p", 22, "Port number")
	addCmd.Flags().StringP("key", "k", "", "Path to private key")
	addCmd.Flags().StringP("use-password", "P", "", "Use a saved password by its alias")

	rootCmd.AddCommand(cmd.PasswordCmd)
	rootCmd.AddCommand(connectCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(removeCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func connectByName(name string) {
	connections, err := config.LoadConnections()
	if err != nil {
		fmt.Println("Error loading connections:", err)
		os.Exit(1)
	}

	var conn *config.Connection
	for i, c := range connections {
		if c.Name == name {
			conn = &connections[i]
			break
		}
	}

	if conn == nil {
		fmt.Printf("Error: connection '%s' not found\n", name)
		os.Exit(1)
	}

	ssh.Connect(conn)
}
