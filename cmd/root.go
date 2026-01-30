package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"gossh/internal/config"
	"gossh/internal/i18n"
	"gossh/internal/ssh"
	"os"
)

var version string

var rootCmd *cobra.Command

func init() {
	rootCmd = &cobra.Command{
		Use:     "gossh",
		Short:   i18n.T("root.short"),
		Version: version,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if r := cmd.Root(); r != nil {
				r.Short = i18n.T("root.short")
				for _, c := range r.Commands() {
					switch c.Name() {
					case "config":
						c.Short = i18n.T("config.short")
					case "connect":
						c.Short = i18n.T("connect.short")
					case "password":
						c.Short = i18n.T("password.short")
					case "exec":
						c.Short = i18n.T("exec.short")
						c.Long = i18n.T("exec.long")
					case "scp":
						c.Short = i18n.T("scp.short")
						c.Long = i18n.T("scp.long")
					case "test":
						c.Short = i18n.T("test.short")
						c.Long = i18n.T("test.long")
					case "groups":
						c.Short = i18n.T("groups.short")
					case "help":
						c.Short = i18n.T("root.help")
					}

					for _, sc := range c.Commands() {
						if c.Name() == "config" {
							switch sc.Name() {
							case "add":
								sc.Short = i18n.T("add.short")
							case "list":
								sc.Short = i18n.T("list.short")
							case "remove":
								sc.Short = i18n.T("remove.short")
							case "import":
								sc.Short = i18n.T("import.short")
								sc.Long = i18n.T("import.long")
							case "export":
								sc.Short = i18n.T("export.short")
								sc.Long = i18n.T("export.long")
							}
						}
						if c.Name() == "password" {
							switch sc.Name() {
							case "add":
								sc.Short = i18n.T("password.add.short")
							case "list":
								sc.Short = i18n.T("password.list.short")
							case "remove":
								sc.Short = i18n.T("password.remove.short")
							}
						}
					}
				}
			}
			return nil
		},
	}

	rootCmd.SetVersionTemplate(`{{printf "%s\n" .Version}}`)

	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{
		Use:    "help [command]",
		Short:  "Help about any command",
		Long:   `Help provides help for any command in the application.`,
		Hidden: true,
	})

	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(connectCmd)
	rootCmd.AddCommand(PasswordCmd)
	rootCmd.AddCommand(execCmd)
	rootCmd.AddCommand(scpCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(groupsCmd)
	rootCmd.AddCommand(copyCmd)
}

func Execute() {

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func RootCommand() *cobra.Command {
	return rootCmd
}

// connectByName is a helper function to connect to a server by its configuration name.
func connectByName(name string) {
	connections, err := config.LoadConnections()
	if err != nil {
		fmt.Println(i18n.TWith("error.loading.connections", map[string]interface{}{"Error": err}))
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
		fmt.Println(i18n.TWith("error.connection.not.found", map[string]interface{}{"Name": name}))
		os.Exit(1)
	}

	ssh.Connect(conn)
}
