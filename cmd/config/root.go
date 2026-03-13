package configcmd

import (
	"github.com/spf13/cobra"
	"gossh/internal/i18n"
)

var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: i18n.T("config.short"),
}

func init() {
}
