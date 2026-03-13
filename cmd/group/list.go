package groupcmd

import (
	"fmt"
	"gossh/internal/config"
	"gossh/internal/i18n"
	"os"
	"sort"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list [group_name]",
	Short: i18n.T("groups.list.short"),
	Run: func(cmd *cobra.Command, args []string) {
		store, err := config.LoadStore()
		if err != nil {
			fmt.Println(i18n.TWith("error.loading.connections", map[string]interface{}{"Error": err}))
			os.Exit(1)
		}

		detail, _ := cmd.Flags().GetBool("detail")

		// Case 1: List details for a specific group
		if len(args) > 0 {
			showGroupDetails(args[0], store)
			return
		}

		// Case 2: List all groups with details
		if detail {
			showAllGroupsWithDetails(store)
			return
		}

		// Case 3: Summary list (default)
		showGroupsSummary(store)
	},
}

func showGroupsSummary(store config.ConfigStore) {
	if len(store.Groups) == 0 {
		fmt.Println(i18n.T("groups.none"))
		return
	}

	var groupNames []string
	for _, g := range store.Groups {
		groupNames = append(groupNames, g.Name)
	}
	sort.Strings(groupNames)

	fmt.Println(i18n.T("groups.available"))
	for _, name := range groupNames {
		for _, g := range store.Groups {
			if g.Name == name {
				fmt.Printf("- %s (%d connections)\n", name, len(g.Connections))
				break
			}
		}
	}
}

func showGroupDetails(groupName string, store config.ConfigStore) {
	var targetGroup *config.Group
	for i, g := range store.Groups {
		if g.Name == groupName {
			targetGroup = &store.Groups[i]
			break
		}
	}

	if targetGroup == nil || len(targetGroup.Connections) == 0 {
		fmt.Println(i18n.TWith("groups.no.connections", map[string]interface{}{"Group": groupName}))
		return
	}

	fmt.Println(i18n.TWith("groups.connections.in.group", map[string]interface{}{"Group": groupName}))
	for _, cName := range targetGroup.Connections {
		cPtr, err := config.ResolveConnection(cName)
		if err == nil {
			printConnDetail(*cPtr)
		} else {
			fmt.Printf("  - %s (Error resolving: %v)\n", cName, err)
		}
	}
}

func showAllGroupsWithDetails(store config.ConfigStore) {
	if len(store.Groups) == 0 {
		fmt.Println(i18n.T("groups.none"))
		return
	}

	var groupNames []string
	for _, g := range store.Groups {
		groupNames = append(groupNames, g.Name)
	}
	sort.Strings(groupNames)

	for _, groupName := range groupNames {
		fmt.Println(i18n.TWith("list.group", map[string]interface{}{"Group": groupName}))
		var targetGroup *config.Group
		for i, g := range store.Groups {
			if g.Name == groupName {
				targetGroup = &store.Groups[i]
				break
			}
		}

		if targetGroup != nil && len(targetGroup.Connections) > 0 {
			for _, cName := range targetGroup.Connections {
				cPtr, err := config.ResolveConnection(cName)
				if err == nil {
					printConnDetail(*cPtr)
				} else {
					fmt.Printf("  - %s (Error resolving: %v)\n", cName, err)
				}
			}
		} else {
			fmt.Println("  (No connections)")
		}
		fmt.Println()
	}
}

func printConnDetail(c config.Connection) {
	authMethod := "interactive"
	if c.KeyPath != "" {
		authMethod = "key"
	} else if c.CredentialAlias != "" {
		authMethod = fmt.Sprintf("password (alias: %s)", c.CredentialAlias)
	}
	fmt.Println(i18n.TWith("list.connection.info", map[string]interface{}{
		"Indent":     "  ",
		"Name":       c.Name,
		"User":       c.User,
		"Host":       c.Host,
		"Port":       c.Port,
		"AuthMethod": authMethod,
	}))
}

func init() {
	GroupCmd.AddCommand(listCmd)
	listCmd.Flags().BoolP("detail", "d", false, i18n.T("groups.list.flag.detail"))
}
