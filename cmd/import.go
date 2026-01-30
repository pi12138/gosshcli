package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"gossh/internal/config"
	"gossh/internal/i18n"
	"os"
	"strings"
)

var importCmd = &cobra.Command{
	Use:   "import <file-path>",
	Short: i18n.T("import.short"),
	Long:  i18n.T("import.long"),
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		force, _ := cmd.Flags().GetBool("force")

		fileData, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Println(i18n.TWith("import.error.reading.file", map[string]interface{}{"File": filePath, "Error": err}))
			os.Exit(1)
		}

		var importData ExportData
		if err := json.Unmarshal(fileData, &importData); err != nil {
			fmt.Println(i18n.TWith("import.error.parsing.json", map[string]interface{}{"Error": err}))
			os.Exit(1)
		}

		if !force {
			fmt.Println(i18n.TWith("import.confirm", map[string]interface{}{
				"Connections": len(importData.Connections),
				"Credentials": len(importData.Credentials),
			}))
			fmt.Println(i18n.T("import.warning"))
			fmt.Print(i18n.T("import.confirm.prompt"))

			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))

			if response != "y" && response != "yes" {
				fmt.Println(i18n.T("import.cancelled"))
				os.Exit(0)
			}
		}

		if err := config.SaveConnections(importData.Connections); err != nil {
			fmt.Println(i18n.TWith("import.error.saving.connections", map[string]interface{}{"Error": err}))
			os.Exit(1)
		}

		if err := config.SaveCredentials(importData.Credentials); err != nil {
			fmt.Println(i18n.TWith("import.error.saving.credentials", map[string]interface{}{"Error": err}))
			os.Exit(1)
		}

		fmt.Println(i18n.T("import.success"))
	},
}

func init() {
	importCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt before overwriting")
	configCmd.AddCommand(importCmd)
}
