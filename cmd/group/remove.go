package groupcmd

import (
	"fmt"
	"gossh/internal/config"
	"gossh/internal/i18n"
	"os"

	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <group_name> <connection_names...>",
	Short: i18n.T("groups.remove.short"),
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		groupName := args[0]
		connectionNames := args[1:]

		store, err := config.LoadStore()
		if err != nil {
			fmt.Println(i18n.TWith("error.loading.connections", map[string]interface{}{"Error": err}))
			os.Exit(1)
		}

		var targetGroup *config.Group
		var targetIndex int
		for i, g := range store.Groups {
			if g.Name == groupName {
				targetGroup = &store.Groups[i]
				targetIndex = i
				break
			}
		}

		if targetGroup == nil {
			fmt.Println(i18n.TWith("groups.not.found", map[string]interface{}{"Group": groupName}))
			os.Exit(1)
		}

		removedCount := 0
		var remainingConns []string

		toRemoveMap := make(map[string]bool)
		for _, name := range connectionNames {
			toRemoveMap[name] = true
		}

		for _, existingConn := range targetGroup.Connections {
			if toRemoveMap[existingConn] {
				removedCount++
			} else {
				remainingConns = append(remainingConns, existingConn)
			}
		}

		if removedCount == 0 {
			fmt.Println(i18n.T("groups.remove.none.found"))
			return
		}

		store.Groups[targetIndex].Connections = remainingConns

		if err := config.SaveStore(store); err != nil {
			fmt.Println(i18n.TWith("error.saving.connections", map[string]interface{}{"Error": err}))
			os.Exit(1)
		}

		fmt.Println(i18n.TWith("groups.remove.success", map[string]interface{}{
			"Count": removedCount,
			"Group": groupName,
		}))
	},
}

func init() {
	GroupCmd.AddCommand(removeCmd)
}
