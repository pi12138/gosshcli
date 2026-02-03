package cmd

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"gossh/internal/config"
	"gossh/internal/i18n"
	"os"
	"strconv"
	"strings"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: i18n.T("add.short"),
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
		fmt.Println(i18n.T("add.error.name.user.host.required"))
		os.Exit(1)
	}
	if keyPath != "" && credAlias != "" {
		fmt.Println(i18n.T("add.error.key.password.together"))
		os.Exit(1)
	}

	if credAlias != "" && !credentialAliasExists(credAlias) {
		fmt.Println(i18n.TWith("add.error.credential.not.found", map[string]interface{}{"Alias": credAlias}))
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
		fmt.Println(i18n.TWith("add.error.adding.connection", map[string]interface{}{"Error": err}))
		os.Exit(1)
	}
	fmt.Println(i18n.TWith("add.success", map[string]interface{}{"Name": name}))
}

func runInteractiveAdd() {
	reader := bufio.NewReader(os.Stdin)
	conn := config.Connection{}

	fmt.Print(i18n.T("add.enter.name"))
	conn.Name, _ = reader.ReadString('\n')
	conn.Name = strings.TrimSpace(conn.Name)

	fmt.Print(i18n.T("add.enter.group"))
	conn.Group, _ = reader.ReadString('\n')
	conn.Group = strings.TrimSpace(conn.Group)

	fmt.Print(i18n.T("add.enter.username"))
	conn.User, _ = reader.ReadString('\n')
	conn.User = strings.TrimSpace(conn.User)

	fmt.Print(i18n.T("add.enter.host"))
	conn.Host, _ = reader.ReadString('\n')
	conn.Host = strings.TrimSpace(conn.Host)

	fmt.Print(i18n.T("add.enter.port"))
	portStr, _ := reader.ReadString('\n')
	portStr = strings.TrimSpace(portStr)
	if portStr == "" {
		conn.Port = 22
	} else {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			fmt.Println(i18n.T("add.invalid.port"))
			os.Exit(1)
		}
		conn.Port = port
	}

	fmt.Print(i18n.T("add.enter.auth.method"))
	authMethod, _ := reader.ReadString('\n')
	authMethod = strings.TrimSpace(strings.ToLower(authMethod))

	switch authMethod {
	case "key":
		fmt.Print(i18n.T("add.enter.key.path"))
		conn.KeyPath, _ = reader.ReadString('\n')
		conn.KeyPath = strings.TrimSpace(conn.KeyPath)
	case "password":
		creds, err := config.LoadCredentials()
		if err != nil || len(creds) == 0 {
			fmt.Println(i18n.T("add.no.password.credentails"))
			os.Exit(1)
		}
		fmt.Println(i18n.T("add.available.password.aliases"))
		for i, c := range creds {
			fmt.Printf("%d: %s\n", i+1, c.Alias)
		}
		fmt.Print(i18n.T("add.choose.password.alias"))
		aliasChoice, _ := reader.ReadString('\n')
		aliasChoice = strings.TrimSpace(aliasChoice)
		aliasIndex, err := strconv.Atoi(aliasChoice)
		if err != nil || aliasIndex < 1 || aliasIndex > len(creds) {
			fmt.Println(i18n.T("add.invalid.selection"))
			os.Exit(1)
		}
		conn.CredentialAlias = creds[aliasIndex-1].Alias
	case "none":
	default:
		fmt.Println(i18n.T("add.invalid.auth.method"))
	}

	if err := config.AddConnection(conn); err != nil {
		fmt.Println(i18n.TWith("add.error.adding.connection", map[string]interface{}{"Error": err}))
		os.Exit(1)
	}
	fmt.Println(i18n.TWith("add.success", map[string]interface{}{"Name": conn.Name}))
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
