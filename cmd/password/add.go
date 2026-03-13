package passwordcmd

import (
	"fmt"
	"gossh/internal/config"
	"gossh/internal/i18n"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

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

func init() {
	PasswordCmd.AddCommand(addPasswordCmd)
}
