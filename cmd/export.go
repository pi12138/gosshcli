package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"gossh/internal/config"
	"gossh/internal/i18n"
	"os"
)

type ExportData struct {
	Connections []config.Connection `json:"connections"`
	Credentials []config.Credential `json:"credentials"`
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: i18n.T("export.short"),
	Long:  i18n.T("export.long"),
	Run: func(cmd *cobra.Command, args []string) {
		connections, err := config.LoadConnections()
		if err != nil {
			if !os.IsNotExist(err) {
				fmt.Println(i18n.TWith("export.error.loading.connections", map[string]interface{}{"Error": err}))
				os.Exit(1)
			}
		}

		credentials, err := config.LoadCredentials()
		if err != nil {
			if !os.IsNotExist(err) {
				fmt.Println(i18n.TWith("export.error.loading.credentials", map[string]interface{}{"Error": err}))
				os.Exit(1)
			}
		}

		exportData := ExportData{
			Connections: connections,
			Credentials: credentials,
		}

		jsonData, err := json.MarshalIndent(exportData, "", "  ")
		if err != nil {
			fmt.Println(i18n.TWith("export.error.creating.json", map[string]interface{}{"Error": err}))
			os.Exit(1)
		}

		fmt.Println(string(jsonData))
	},
}

func init() {
	configCmd.AddCommand(exportCmd)
}
