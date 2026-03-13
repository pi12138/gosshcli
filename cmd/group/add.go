package groupcmd

import (
	"fmt"
	"gossh/internal/config"
	"gossh/internal/i18n"
	"os"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <group_name> [connection_names...]",
	Short: i18n.T("groups.add.short"),
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		groupName := args[0]
		connectionNames := args[1:]

		if len(connectionNames) == 0 {
			fmt.Println(i18n.T("groups.add.no.connections"))
			return
		}

		store, err := config.LoadStore()
		if err != nil {
			fmt.Println(i18n.TWith("error.loading.connections", map[string]interface{}{"Error": err}))
			os.Exit(1)
		}

		// Verify that the requested connections actually exist
		validConns := make(map[string]bool)
		for _, c := range store.Connections {
			validConns[c.Name] = true
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

		addedCount := 0
		var notFoundNames []string

		for _, name := range connectionNames {
			if !validConns[name] {
				notFoundNames = append(notFoundNames, name)
				continue
			}

			// Check if already in group
			alreadyIn := false
			for _, existingConn := range targetGroup.Connections {
				if existingConn == name {
					alreadyIn = true
					break
				}
			}

			if !alreadyIn {
				store.Groups[targetIndex].Connections = append(store.Groups[targetIndex].Connections, name)
				addedCount++
			}
		}

		if addedCount > 0 {
			if err := config.SaveStore(store); err != nil {
				fmt.Println(i18n.TWith("error.saving.connections", map[string]interface{}{"Error": err}))
				os.Exit(1)
			}
			fmt.Println(i18n.TWith("groups.add.success", map[string]interface{}{
				"Count": addedCount,
				"Group": groupName,
			}))
		}

		if len(notFoundNames) > 0 {
			fmt.Println(i18n.TWith("groups.add.not.found", map[string]interface{}{
				"Names": fmt.Sprintf("%v", notFoundNames),
			}))
		}
	},
}

func init() {
	GroupCmd.AddCommand(addCmd)
}
