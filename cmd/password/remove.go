package passwordcmd

import (
	"fmt"
	"gossh/internal/config"
	"gossh/internal/i18n"
	"os"

	"github.com/spf13/cobra"
)

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
	PasswordCmd.AddCommand(removePasswordCmd)
}
