package groupcmd

import (
	"fmt"
	"gossh/internal/config"
	"gossh/internal/i18n"
	"os"

	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create <group_name>",
	Short: i18n.T("groups.create.short"),
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		groupName := args[0]

		store, err := config.LoadStore()
		if err != nil {
			fmt.Println(i18n.TWith("error.loading.connections", map[string]interface{}{"Error": err}))
			os.Exit(1)
		}

		for _, g := range store.Groups {
			if g.Name == groupName {
				fmt.Println(i18n.TWith("groups.create.exists", map[string]interface{}{"Group": groupName}))
				os.Exit(1)
			}
		}

		newGroup := config.Group{
			Name:        groupName,
			Connections: []string{},
		}
		store.Groups = append(store.Groups, newGroup)

		if err := config.SaveStore(store); err != nil {
			fmt.Println(i18n.TWith("error.saving.connections", map[string]interface{}{"Error": err}))
			os.Exit(1)
		}

		fmt.Println(i18n.TWith("groups.create.success", map[string]interface{}{"Group": groupName}))
	},
}

func init() {
	GroupCmd.AddCommand(createCmd)
}
