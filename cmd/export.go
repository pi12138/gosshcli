package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"gossh/config"
	"os"
)

// ExportData is the structure for the exported JSON file.
type ExportData struct {
	Connections []config.Connection `json:"connections"`
	Credentials []config.Credential `json:"credentials"`
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export your connections and credentials to stdout",
	Long: `Exports all your saved connection configurations and encrypted credentials 
into a single JSON object printed to standard output. 
You can redirect this output to a file for backup.
Example: gossh config export > gossh_backup.json`,
	Run: func(cmd *cobra.Command, args []string) {
		connections, err := config.LoadConnections()
		if err != nil {
			// It's okay if the file doesn't exist, just export an empty list.
			if !os.IsNotExist(err) {
				fmt.Println("Error loading connections:", err)
				os.Exit(1)
			}
		}

		credentials, err := config.LoadCredentials()
		if err != nil {
			if !os.IsNotExist(err) {
				fmt.Println("Error loading credentials:", err)
				os.Exit(1)
			}
		}

		exportData := ExportData{
			Connections: connections,
			Credentials: credentials,
		}

		jsonData, err := json.MarshalIndent(exportData, "", "  ")
		if err != nil {
			fmt.Println("Error creating JSON data for export:", err)
			os.Exit(1)
		}

		fmt.Println(string(jsonData))
	},
}

func init() {
	configCmd.AddCommand(exportCmd)
}
