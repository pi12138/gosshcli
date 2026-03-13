package groupcmd

import (
	"gossh/internal/i18n"

	"github.com/spf13/cobra"
)

var GroupCmd = &cobra.Command{
	Use:   "group",
	Short: i18n.T("groups.short"),
}

func init() {
}
