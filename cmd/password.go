package cmd

import (
	"fmt"
	"gossh/config"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

var PasswordCmd = &cobra.Command{
	Use:   "password",
	Short: "Manage saved passwords",
}

var addPasswordCmd = &cobra.Command{
	Use:   "add [alias]",
	Short: "Add a new password with an alias",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		alias := args[0]

		fmt.Print("Enter password: ")
		bytePassword, err := terminal.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			fmt.Println("\nFailed to read password:", err)
			os.Exit(1)
		}
		fmt.Println()

		cred := config.Credential{
			Alias:    alias,
			Password: string(bytePassword),
		}

		if err := config.AddCredential(cred); err != nil {
			fmt.Println("Error adding credential:", err)
			os.Exit(1)
		}
		fmt.Printf("Credential '%s' added successfully.\n", alias)
	},
}

var listPasswordCmd = &cobra.Command{
	Use:   "list",
	Short: "List saved password aliases",
	Run: func(cmd *cobra.Command, args []string) {
		credentials, err := config.LoadCredentials()
		if err != nil {
			fmt.Println("Error loading credentials:", err)
			os.Exit(1)
		}
		if len(credentials) == 0 {
			fmt.Println("No credentials saved.")
			return
		}
		fmt.Println("Saved credential aliases:")
		for _, c := range credentials {
			fmt.Printf("- %s\n", c.Alias)
		}
	},
}

var removePasswordCmd = &cobra.Command{
	Use:   "remove [alias]",
	Short: "Remove a saved password",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		alias := args[0]
		if err := config.RemoveCredential(alias); err != nil {
			fmt.Println("Error removing credential:", err)
			os.Exit(1)
		}
		fmt.Printf("Credential '%s' removed successfully.\n", alias)
	},
}

func init() {
	PasswordCmd.AddCommand(addPasswordCmd)
	PasswordCmd.AddCommand(listPasswordCmd)
	PasswordCmd.AddCommand(removePasswordCmd)
	rootCmd.AddCommand(PasswordCmd)
}
