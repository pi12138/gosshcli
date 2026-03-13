package passwordcmd

import (
	"gossh/internal/i18n"

	"github.com/spf13/cobra"
)

var PasswordCmd = &cobra.Command{
	Use:   "password",
	Short: i18n.T("password.short"),
}

func init() {
}
