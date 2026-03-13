package groupcmd

import (
	"fmt"
	"gossh/internal/config"
	"gossh/internal/i18n"
	"os"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <group_name>",
	Short: i18n.T("groups.delete.short"),
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		groupName := args[0]
		purge, _ := cmd.Flags().GetBool("purge")

		store, err := config.LoadStore()
		if err != nil {
			fmt.Println(i18n.TWith("error.loading.connections", map[string]interface{}{"Error": err}))
			os.Exit(1)
		}

		var newGroups []config.Group
		var targetGroup *config.Group
		for i, g := range store.Groups {
			if g.Name == groupName {
				targetGroup = &store.Groups[i]
			} else {
				newGroups = append(newGroups, g)
			}
		}

		if targetGroup == nil {
			fmt.Println(i18n.TWith("groups.not.found", map[string]interface{}{"Group": groupName}))
			os.Exit(1)
		}

		store.Groups = newGroups

		if purge && len(targetGroup.Connections) > 0 {
			var newConnections []config.Connection
			toPurgeMap := make(map[string]bool)
			for _, name := range targetGroup.Connections {
				toPurgeMap[name] = true
			}

			for _, c := range store.Connections {
				if !toPurgeMap[c.Name] {
					newConnections = append(newConnections, c)
				}
			}
			store.Connections = newConnections
			
			// Also need to remove purged connections from other groups they might belong to
			for i := range store.Groups {
				var remainingConns []string
				for _, name := range store.Groups[i].Connections {
					if !toPurgeMap[name] {
						remainingConns = append(remainingConns, name)
					}
				}
				store.Groups[i].Connections = remainingConns
			}

			fmt.Println(i18n.TWith("groups.delete.purge.success", map[string]interface{}{
				"Group": groupName,
				"Count": len(targetGroup.Connections),
			}))
		} else {
			fmt.Println(i18n.TWith("groups.delete.success", map[string]interface{}{"Group": groupName}))
		}

		if err := config.SaveStore(store); err != nil {
			fmt.Println(i18n.TWith("error.saving.connections", map[string]interface{}{"Error": err}))
			os.Exit(1)
		}
	},
}

func init() {
	GroupCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().BoolP("purge", "p", false, i18n.T("groups.delete.flag.purge"))
}
