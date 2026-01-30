package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"gossh/internal/config"
	"gossh/internal/i18n"
	"os"
)

var removeCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: i18n.T("remove.short"),
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		if err := config.RemoveConnection(name); err != nil {
			fmt.Println(i18n.TWith("remove.error.removing", map[string]interface{}{"Error": err}))
			os.Exit(1)
		}
		fmt.Println(i18n.TWith("remove.success", map[string]interface{}{"Name": name}))
	},
}

func init() {
}
