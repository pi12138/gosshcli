package cmd

import (
	"github.com/spf13/cobra"
	"gossh/internal/i18n"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: i18n.T("config.short"),
}

func init() {
	configCmd.AddCommand(addCmd)
	configCmd.AddCommand(listCmd)
	configCmd.AddCommand(removeCmd)
}
