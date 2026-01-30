package cmd

import (
	"github.com/spf13/cobra"
	"gossh/internal/i18n"
)

var connectCmd = &cobra.Command{
	Use:   "connect [name]",
	Short: i18n.T("connect.short"),
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		connectByName(args[0])
	},
}
