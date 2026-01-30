package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"gossh/internal/config"
	"gossh/internal/i18n"
	"os"
	"sort"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: i18n.T("list.short"),
	Run: func(cmd *cobra.Command, args []string) {
		filterGroup, _ := cmd.Flags().GetString("group")

		connections, err := config.LoadConnections()
		if err != nil {
			fmt.Println(i18n.TWith("error.loading.connections", map[string]interface{}{"Error": err}))
			os.Exit(1)
		}

		if len(connections) == 0 {
			fmt.Println(i18n.T("list.no.connections"))
			return
		}

		groups := make(map[string][]config.Connection)
		var ungrouped []config.Connection
		for _, c := range connections {
			if filterGroup != "" && c.Group != filterGroup {
				continue
			}
			if c.Group != "" {
				groups[c.Group] = append(groups[c.Group], c)
			} else {
				ungrouped = append(ungrouped, c)
			}
		}

		if len(groups) == 0 && len(ungrouped) == 0 {
			if filterGroup != "" {
				fmt.Println(i18n.TWith("list.no.connections.group", map[string]interface{}{"Group": filterGroup}))
			} else {
				fmt.Println(i18n.T("list.no.connections.found"))
			}
			return
		}

		if len(groups) > 0 {
			var groupNames []string
			for name := range groups {
				groupNames = append(groupNames, name)
			}
			sort.Strings(groupNames)

			for _, groupName := range groupNames {
				fmt.Println(i18n.TWith("list.group", map[string]interface{}{"Group": groupName}))
				for _, c := range groups[groupName] {
					printConnectionInfo(c, true)
				}
			}
		}

		if len(ungrouped) > 0 {
			if len(groups) > 0 {
				fmt.Println("\n" + i18n.T("list.ungrouped"))
			}
			for _, c := range ungrouped {
				printConnectionInfo(c, false)
			}
		}
	},
}

func printConnectionInfo(c config.Connection, isGrouped bool) {
	authMethod := "interactive"
	if c.KeyPath != "" {
		authMethod = "key"
	} else if c.CredentialAlias != "" {
		authMethod = fmt.Sprintf("password (alias: %s)", c.CredentialAlias)
	}
	indent := ""
	if isGrouped {
		indent = "  "
	}
	fmt.Println(i18n.TWith("list.connection.info", map[string]interface{}{
		"Indent":     indent,
		"Name":       c.Name,
		"User":       c.User,
		"Host":       c.Host,
		"Port":       c.Port,
		"AuthMethod": authMethod,
	}))
}

func init() {
	listCmd.Flags().StringP("group", "g", "", "Filter connections by group name")
}
