package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"gossh/internal/config"
	"gossh/internal/i18n"
	"os"
	"sort"
)

var groupsCmd = &cobra.Command{
	Use:   "groups",
	Short: i18n.T("groups.short"),
	Run: func(cmd *cobra.Command, args []string) {
		connections, err := config.LoadConnections()
		if err != nil {
			fmt.Println(i18n.TWith("error.loading.connections", map[string]interface{}{"Error": err}))
			os.Exit(1)
		}

		groupSet := make(map[string]struct{})
		for _, c := range connections {
			if c.Group != "" {
				groupSet[c.Group] = struct{}{}
			}
		}

		if len(groupSet) == 0 {
			fmt.Println(i18n.T("groups.none"))
			return
		}

		var groups []string
		for group := range groupSet {
			groups = append(groups, group)
		}
		sort.Strings(groups)

		fmt.Println(i18n.T("groups.available"))
		for _, group := range groups {
			fmt.Printf("- %s\n", group)
		}
	},
}
