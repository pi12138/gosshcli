package configcmd

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
	groupStr, _ := cmd.Flags().GetString("group")
	user, _ := cmd.Flags().GetString("user")
	host, _ := cmd.Flags().GetString("host")
	port, _ := cmd.Flags().GetInt("port")
	keyPath, _ := cmd.Flags().GetString("key")
	credAlias, _ := cmd.Flags().GetString("use-password")

	if name == "" || user == "" || host == "" {
		fmt.Println(i18n.T("add.error.name.user.host.required"))
		os.Exit(1)
	}

	var groups []string
	if groupStr != "" {
		for _, g := range strings.Split(groupStr, ",") {
			g = strings.TrimSpace(g)
			if g != "" {
				groups = append(groups, g)
			}
		}
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
		User:            user,
		Host:            host,
		Port:            port,
		KeyPath:         keyPath,
		CredentialAlias: credAlias,
	}

	store, err := config.LoadStore()
	if err != nil {
		fmt.Println(i18n.TWith("error.loading.connections", map[string]interface{}{"Error": err}))
		os.Exit(1)
	}

	for _, c := range store.Connections {
		if c.Name == conn.Name {
			fmt.Println(i18n.TWith("add.error.adding.connection", map[string]interface{}{"Error": "connection already exists"}))
			os.Exit(1)
		}
	}

	store.Connections = append(store.Connections, conn)

	for _, gName := range groups {
		foundGroup := false
		for i := range store.Groups {
			if store.Groups[i].Name == gName {
				store.Groups[i].Connections = append(store.Groups[i].Connections, conn.Name)
				foundGroup = true
				break
			}
		}
		if !foundGroup {
			store.Groups = append(store.Groups, config.Group{
				Name:        gName,
				Connections: []string{conn.Name},
			})
		}
	}

	if err := config.SaveStore(store); err != nil {
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

	var groups []string
	fmt.Print(i18n.T("add.enter.groups"))
	groupStr, _ := reader.ReadString('\n')
	groupStr = strings.TrimSpace(groupStr)
	if groupStr != "" {
		for _, g := range strings.Split(groupStr, ",") {
			g = strings.TrimSpace(g)
			if g != "" {
				groups = append(groups, g)
			}
		}
	}

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

	store, err := config.LoadStore()
	if err != nil {
		fmt.Println(i18n.TWith("error.loading.connections", map[string]interface{}{"Error": err}))
		os.Exit(1)
	}

	for _, c := range store.Connections {
		if c.Name == conn.Name {
			fmt.Println(i18n.TWith("add.error.adding.connection", map[string]interface{}{"Error": "connection already exists"}))
			os.Exit(1)
		}
	}

	store.Connections = append(store.Connections, conn)

	for _, gName := range groups {
		foundGroup := false
		for i := range store.Groups {
			if store.Groups[i].Name == gName {
				store.Groups[i].Connections = append(store.Groups[i].Connections, conn.Name)
				foundGroup = true
				break
			}
		}
		if !foundGroup {
			store.Groups = append(store.Groups, config.Group{
				Name:        gName,
				Connections: []string{conn.Name},
			})
		}
	}

	if err := config.SaveStore(store); err != nil {
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
	ConfigCmd.AddCommand(addCmd)
	addCmd.Flags().StringP("name", "n", "", i18n.T("add.flag.name"))
	addCmd.Flags().StringP("group", "g", "", i18n.T("add.flag.group"))
	addCmd.Flags().StringP("user", "u", "", i18n.T("add.flag.user"))
	addCmd.Flags().StringP("host", "H", "", i18n.T("add.flag.host"))
	addCmd.Flags().IntP("port", "p", 22, i18n.T("add.flag.port"))
	addCmd.Flags().StringP("key", "k", "", i18n.T("add.flag.key"))
	addCmd.Flags().StringP("use-password", "P", "", i18n.T("add.flag.use-password"))
	addCmd.Flags().BoolP("interactive", "i", false, i18n.T("add.flag.interactive"))
}
