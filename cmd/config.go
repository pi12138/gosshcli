package cmd

import "github.com/spf13/cobra"

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage connection configurations",
}

func init() {
	// Add config command to root
	rootCmd.AddCommand(configCmd)

	// Add subcommands to config command
	configCmd.AddCommand(addCmd)
	configCmd.AddCommand(listCmd)
	configCmd.AddCommand(removeCmd)
}
