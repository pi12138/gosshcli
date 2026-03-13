package groupcmd

import (
	"fmt"
	"gossh/internal/config"
	"gossh/internal/i18n"
	"os"

	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:   "set <group_name>",
	Short: i18n.T("groups.set.short"),
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		groupName := args[0]
		user, _ := cmd.Flags().GetString("user")
		port, _ := cmd.Flags().GetInt("port")
		keyPath, _ := cmd.Flags().GetString("key")
		credAlias, _ := cmd.Flags().GetString("use-password")

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

		modified := false
		if cmd.Flags().Changed("user") {
			store.Groups[targetIndex].User = user
			modified = true
		}
		if cmd.Flags().Changed("port") {
			store.Groups[targetIndex].Port = port
			modified = true
		}
		if cmd.Flags().Changed("key") {
			store.Groups[targetIndex].KeyPath = keyPath
			modified = true
		}
		if cmd.Flags().Changed("use-password") {
			store.Groups[targetIndex].CredentialAlias = credAlias
			modified = true
		}

		if !modified {
			fmt.Println(i18n.T("groups.set.no.changes"))
			return
		}

		if err := config.SaveStore(store); err != nil {
			fmt.Println(i18n.TWith("error.saving.connections", map[string]interface{}{"Error": err}))
			os.Exit(1)
		}

		fmt.Println(i18n.TWith("groups.set.success", map[string]interface{}{"Group": groupName}))
	},
}

func init() {
	GroupCmd.AddCommand(setCmd)
	setCmd.Flags().StringP("user", "u", "", i18n.T("groups.set.flag.user"))
	setCmd.Flags().IntP("port", "p", 0, i18n.T("groups.set.flag.port"))
	setCmd.Flags().StringP("key", "k", "", i18n.T("groups.set.flag.key"))
	setCmd.Flags().StringP("use-password", "P", "", i18n.T("groups.set.flag.use-password"))
}
