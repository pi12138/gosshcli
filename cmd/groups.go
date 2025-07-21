package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"gossh/config"
	"os"
	"sort"
)

var groupsCmd = &cobra.Command{
	Use:   "groups",
	Short: "List all unique connection group names",
	Run: func(cmd *cobra.Command, args []string) {
		connections, err := config.LoadConnections()
		if err != nil {
			fmt.Println("Error loading connections:", err)
			os.Exit(1)
		}

		groupSet := make(map[string]struct{})
		for _, c := range connections {
			if c.Group != "" {
				groupSet[c.Group] = struct{}{}
			}
		}

		if len(groupSet) == 0 {
			fmt.Println("No groups found.")
			return
		}

		var groups []string
		for group := range groupSet {
			groups = append(groups, group)
		}
		sort.Strings(groups)

		fmt.Println("Available groups:")
		for _, group := range groups {
			fmt.Printf("- %s\n", group)
		}
	},
}

func init() {
	rootCmd.AddCommand(groupsCmd)
}
