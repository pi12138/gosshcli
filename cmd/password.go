package cmd

import (
	"fmt"
	"gossh/internal/config"
	"gossh/internal/i18n"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

var PasswordCmd = &cobra.Command{
	Use:   "password",
	Short: i18n.T("password.short"),
}

var addPasswordCmd = &cobra.Command{
	Use:   "add [alias]",
	Short: i18n.T("password.add.short"),
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		alias := args[0]

		fmt.Print(i18n.T("password.enter.password"))
		bytePassword, err := terminal.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			fmt.Println(i18n.TWith("password.error.reading", map[string]interface{}{"Error": err}))
			os.Exit(1)
		}
		fmt.Println()

		cred := config.Credential{
			Alias:    alias,
			Password: string(bytePassword),
		}

		if err := config.AddCredential(cred); err != nil {
			fmt.Println(i18n.TWith("password.error.adding", map[string]interface{}{"Error": err}))
			os.Exit(1)
		}
		fmt.Println(i18n.TWith("password.add.success", map[string]interface{}{"Alias": alias}))
	},
}

var listPasswordCmd = &cobra.Command{
	Use:   "list",
	Short: i18n.T("password.list.short"),
	Run: func(cmd *cobra.Command, args []string) {
		credentials, err := config.LoadCredentials()
		if err != nil {
			fmt.Println(i18n.TWith("password.list.error", map[string]interface{}{"Error": err}))
			os.Exit(1)
		}
		if len(credentials) == 0 {
			fmt.Println(i18n.T("password.list.none"))
			return
		}
		fmt.Println(i18n.T("password.list.aliases"))
		for _, c := range credentials {
			fmt.Printf("- %s\n", c.Alias)
		}
	},
}

var removePasswordCmd = &cobra.Command{
	Use:   "remove [alias]",
	Short: i18n.T("password.remove.short"),
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		alias := args[0]
		if err := config.RemoveCredential(alias); err != nil {
			fmt.Println(i18n.TWith("password.remove.error", map[string]interface{}{"Error": err}))
			os.Exit(1)
		}
		fmt.Println(i18n.TWith("password.remove.success", map[string]interface{}{"Alias": alias}))
	},
}

func init() {
	PasswordCmd.AddCommand(addPasswordCmd)
	PasswordCmd.AddCommand(listPasswordCmd)
	PasswordCmd.AddCommand(removePasswordCmd)
}
