package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"gossh/config"
	"io/ioutil"
	"os"
	"strings"
)

var importCmd = &cobra.Command{
	Use:   "import <file-path>",
	Short: "Import connections and credentials from a file",
	Long: `Imports configurations from a JSON file created by 'gossh config export'.
This will overwrite your existing configurations.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		force, _ := cmd.Flags().GetBool("force")

		fileData, err := ioutil.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Error reading import file '%s': %v\n", filePath, err)
			os.Exit(1)
		}

		var importData ExportData
		if err := json.Unmarshal(fileData, &importData); err != nil {
			fmt.Printf("Error parsing JSON from import file: %v\n", err)
			os.Exit(1)
		}

		if !force {
			fmt.Printf("You are about to import %d connections and %d credentials.\n", len(importData.Connections), len(importData.Credentials))
			fmt.Println("This will OVERWRITE your current configurations.")
			fmt.Print("Are you sure you want to continue? (y/N): ")

			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))

			if response != "y" && response != "yes" {
				fmt.Println("Import cancelled.")
				os.Exit(0)
			}
		}

		if err := config.SaveConnections(importData.Connections); err != nil {
			fmt.Println("Error saving imported connections:", err)
			os.Exit(1)
		}

		if err := config.SaveCredentials(importData.Credentials); err != nil {
			fmt.Println("Error saving imported credentials:", err)
			os.Exit(1)
		}

		fmt.Println("Configuration imported successfully.")
	},
}

func init() {
	importCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt before overwriting")
	configCmd.AddCommand(importCmd)
}
