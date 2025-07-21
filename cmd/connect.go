package cmd

import (
	"github.com/spf13/cobra"
)

var connectCmd = &cobra.Command{
	Use:   "connect [name]",
	Short: "Establish a remote ssh connection to a saved configuration",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		connectByName(args[0])
	},
}

func init() {
	rootCmd.AddCommand(connectCmd)
}
