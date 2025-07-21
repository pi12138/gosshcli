package cmd

import "github.com/spf13/cobra"

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage gossh configuration (import/export)",
	Long:  `A parent command for managing gossh configuration, such as importing from or exporting to a file.`,
}

func init() {
	rootCmd.AddCommand(configCmd)
}
