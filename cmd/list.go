package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"gossh/config"
	"os"
	"sort"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List local connection configurations",
	Run: func(cmd *cobra.Command, args []string) {
		filterGroup, _ := cmd.Flags().GetString("group")

		connections, err := config.LoadConnections()
		if err != nil {
			fmt.Println("Error loading connections:", err)
			os.Exit(1)
		}

		if len(connections) == 0 {
			fmt.Println("No connections configured. Use 'gossh add' to create one.")
			return
		}

		// Group connections
		groups := make(map[string][]config.Connection)
		var ungrouped []config.Connection
		for _, c := range connections {
			if filterGroup != "" && c.Group != filterGroup {
				continue // Skip if filtering and group doesn't match
			}
			if c.Group != "" {
				groups[c.Group] = append(groups[c.Group], c)
			} else {
				ungrouped = append(ungrouped, c)
			}
		}

		if len(groups) == 0 && len(ungrouped) == 0 {
			if filterGroup != "" {
				fmt.Printf("No connections found in group '%s'.\n", filterGroup)
			} else {
				fmt.Println("No connections found.")
			}
			return
		}

		// Display grouped connections
		if len(groups) > 0 {
			// Sort group names for consistent order
			var groupNames []string
			for name := range groups {
				groupNames = append(groupNames, name)
			}
			sort.Strings(groupNames)

			for _, groupName := range groupNames {
				fmt.Printf("Group: %s\n", groupName)
				for _, c := range groups[groupName] {
					printConnectionInfo(c, true)
				}
			}
		}

		// Display ungrouped connections
		if len(ungrouped) > 0 {
			if len(groups) > 0 {
				fmt.Println("\nUngrouped:")
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
	fmt.Printf("%s- %s (%s@%s:%d) (auth: %s)\n", indent, c.Name, c.User, c.Host, c.Port, authMethod)
}

func init() {
	listCmd.Flags().StringP("group", "g", "", "Filter connections by group name")
	rootCmd.AddCommand(listCmd)
}