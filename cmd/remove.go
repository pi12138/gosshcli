package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"gossh/config"
	"os"
)

var removeCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove a connection configuration",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		if err := config.RemoveConnection(name); err != nil {
			fmt.Println("Error removing connection:", err)
			os.Exit(1)
		}
		fmt.Printf("Connection '%s' removed successfully.\n", name)
	},
}

func init() {
}
