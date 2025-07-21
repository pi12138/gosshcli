package cmd

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"gossh/config"
	"os"
	"strconv"
	"strings"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new connection configuration",
	Run: func(cmd *cobra.Command, args []string) {
		interactive, _ := cmd.Flags().GetBool("interactive")

		if interactive {
			runInteractiveAdd()
		} else {
			runFlagBasedAdd(cmd)
		}
	},
}

func runFlagBasedAdd(cmd *cobra.Command) {
	name, _ := cmd.Flags().GetString("name")
	group, _ := cmd.Flags().GetString("group")
	user, _ := cmd.Flags().GetString("user")
	host, _ := cmd.Flags().GetString("host")
	port, _ := cmd.Flags().GetInt("port")
	keyPath, _ := cmd.Flags().GetString("key")
	credAlias, _ := cmd.Flags().GetString("use-password")

	if name == "" || user == "" || host == "" {
		fmt.Println("Error: --name, --user, and --host are required for non-interactive mode.")
		os.Exit(1)
	}
	if keyPath != "" && credAlias != "" {
		fmt.Println("Error: --key and --use-password flags cannot be used together.")
		os.Exit(1)
	}

	if credAlias != "" && !credentialAliasExists(credAlias) {
		fmt.Printf("Error: credential with alias '%s' not found. Use 'gossh password add %s' to create it.\n", credAlias, credAlias)
		os.Exit(1)
	}

	conn := config.Connection{
		Name:            name,
		Group:           group,
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
}

func runInteractiveAdd() {
	reader := bufio.NewReader(os.Stdin)
	conn := config.Connection{}

	fmt.Print("Enter connection name: ")
	conn.Name, _ = reader.ReadString('\n')
	conn.Name = strings.TrimSpace(conn.Name)

	fmt.Print("Enter group (optional): ")
	conn.Group, _ = reader.ReadString('\n')
	conn.Group = strings.TrimSpace(conn.Group)

	fmt.Print("Enter username: ")
	conn.User, _ = reader.ReadString('\n')
	conn.User = strings.TrimSpace(conn.User)

	fmt.Print("Enter host address: ")
	conn.Host, _ = reader.ReadString('\n')
	conn.Host = strings.TrimSpace(conn.Host)

	fmt.Print("Enter port (default 22): ")
	portStr, _ := reader.ReadString('\n')
	portStr = strings.TrimSpace(portStr)
	if portStr == "" {
		conn.Port = 22
	} else {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			fmt.Println("Invalid port number.")
			os.Exit(1)
		}
		conn.Port = port
	}

	fmt.Print("Authentication method (key, password, none): ")
	authMethod, _ := reader.ReadString('\n')
	authMethod = strings.TrimSpace(strings.ToLower(authMethod))

	switch authMethod {
	case "key":
		fmt.Print("Enter path to private key: ")
		conn.KeyPath, _ = reader.ReadString('\n')
		conn.KeyPath = strings.TrimSpace(conn.KeyPath)
	case "password":
		creds, err := config.LoadCredentials()
		if err != nil || len(creds) == 0 {
			fmt.Println("No saved password credentials found. Please add one first using 'gossh password add <alias>'.")
			os.Exit(1)
		}
		fmt.Println("Available password aliases:")
		for i, c := range creds {
			fmt.Printf("%d: %s\n", i+1, c.Alias)
		}
		fmt.Print("Choose a password alias: ")
		aliasChoice, _ := reader.ReadString('\n')
		aliasChoice = strings.TrimSpace(aliasChoice)
		aliasIndex, err := strconv.Atoi(aliasChoice)
		if err != nil || aliasIndex < 1 || aliasIndex > len(creds) {
			fmt.Println("Invalid selection.")
			os.Exit(1)
		}
		conn.CredentialAlias = creds[aliasIndex-1].Alias
	case "none":
		// Do nothing, will use interactive password prompt on connect
	default:
		fmt.Println("Invalid authentication method. Assuming 'none'.")
	}

	if err := config.AddConnection(conn); err != nil {
		fmt.Println("Error adding connection:", err)
		os.Exit(1)
	}
	fmt.Printf("Connection '%s' added successfully.\n", conn.Name)
}

func credentialAliasExists(alias string) bool {
	creds, err := config.LoadCredentials()
	if err != nil {
		return false
	}
	for _, c := range creds {
		if c.Alias == alias {
			return true
		}
	}
	return false
}

func init() {
	addCmd.Flags().StringP("name", "n", "", "Connection name")
	addCmd.Flags().StringP("group", "g", "", "Group name for the connection")
	addCmd.Flags().StringP("user", "u", "", "Username")
	addCmd.Flags().StringP("host", "H", "", "Host address")
	addCmd.Flags().IntP("port", "p", 22, "Port number")
	addCmd.Flags().StringP("key", "k", "", "Path to private key")
	addCmd.Flags().StringP("use-password", "P", "", "Use a saved password by its alias")
	addCmd.Flags().BoolP("interactive", "i", false, "Use interactive mode to add a new connection")
}
