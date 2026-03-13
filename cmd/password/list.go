package passwordcmd

import (
	"fmt"
	"gossh/internal/config"
	"gossh/internal/i18n"
	"os"

	"github.com/spf13/cobra"
)

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

func init() {
	PasswordCmd.AddCommand(listPasswordCmd)
}
