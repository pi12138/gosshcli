package cmd

import (
	"fmt"
	configcmd "gossh/cmd/config"
	groupcmd "gossh/cmd/group"
	passwordcmd "gossh/cmd/password"
	internalconfig "gossh/internal/config"
	"gossh/internal/i18n"
	"gossh/internal/ssh"
	"os"

	"github.com/spf13/cobra"
)

var version string

var rootCmd = &cobra.Command{
	Use:     "gossh",
	Short:   i18n.T("root.short"),
	Version: version,
}

func init() {
	rootCmd.SetVersionTemplate(`{{printf "%s\n" .Version}}`)

	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{
		Use:    "help [command]",
		Short:  "Help about any command",
		Long:   `Help provides help for any command in the application.`,
		Hidden: true,
	})

	rootCmd.AddCommand(configcmd.ConfigCmd)
	rootCmd.AddCommand(connectCmd)
	rootCmd.AddCommand(passwordcmd.PasswordCmd)
	rootCmd.AddCommand(execCmd)
	rootCmd.AddCommand(scpCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(groupcmd.GroupCmd)
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
	conn, err := internalconfig.ResolveConnection(name)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ssh.Connect(conn)
}
