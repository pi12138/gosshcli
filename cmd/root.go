package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"gossh/config"
	"gossh/ssh"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "gossh",
	Short: "gossh is a ssh client implemented in Go",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// connectByName is a helper function to connect to a server by its configuration name.
func connectByName(name string) {
	connections, err := config.LoadConnections()
	if err != nil {
		fmt.Println("Error loading connections:", err)
		os.Exit(1)
	}

	var conn *config.Connection
	for i, c := range connections {
		if c.Name == name {
			conn = &connections[i]
			break
		}
	}

	if conn == nil {
		fmt.Printf("Error: connection '%s' not found\n", name)
		os.Exit(1)
	}

	ssh.Connect(conn)
}
