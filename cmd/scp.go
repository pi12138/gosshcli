package cmd

import (
	"fmt"
	"gossh/internal/config"
	"gossh/internal/i18n"
	"gossh/internal/ssh"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var scpCmd = &cobra.Command{
	Use:   "scp [source] [destination]",
	Short: i18n.T("scp.short"),
	Long:  i18n.T("scp.long"),
	Run:   runScp,
}

func init() {
	scpCmd.Flags().BoolP("recursive", "r", false, "Copy directories recursively")
	scpCmd.Flags().BoolP("force", "f", false, "Force overwrite of existing files")
}

func runScp(cmd *cobra.Command, args []string) {
	if len(args) != 2 {
		fmt.Println(i18n.T("scp.error.both.paths"))
		_ = cmd.Help()
		os.Exit(1)
	}

	recursive, _ := cmd.Flags().GetBool("recursive")
	force, _ := cmd.Flags().GetBool("force")
	source := args[0]
	destination := args[1]

	sourceHasColon := strings.Contains(source, ":")
	destHasColon := strings.Contains(destination, ":")

	if sourceHasColon && destHasColon {
		fmt.Println(i18n.T("scp.error.both.remote"))
		os.Exit(1)
	}

	if !sourceHasColon && !destHasColon {
		fmt.Println(i18n.T("scp.error.no.remote"))
		os.Exit(1)
	}

	connections, err := config.LoadConnections()
	if err != nil {
		fmt.Println(i18n.TWith("error.loading.connections", map[string]interface{}{"Error": err}))
		os.Exit(1)
	}

	if sourceHasColon {
		parts := strings.SplitN(source, ":", 2)
		if len(parts) != 2 {
			fmt.Println(i18n.T("scp.error.invalid.format"))
			os.Exit(1)
		}
		connName := parts[0]
		remotePath := parts[1]
		localPath := destination

		conn := findConnection(connections, connName)
		if conn == nil {
			fmt.Println(i18n.TWith("error.connection.not.found", map[string]interface{}{"Name": connName}))
			os.Exit(1)
		}

		fmt.Println(i18n.TWith("scp.downloading", map[string]interface{}{
			"User":  conn.User,
			"Host":  conn.Host,
			"Path":  remotePath,
			"Local": localPath,
		}))
		err = ssh.DownloadFileWithOpts(conn, remotePath, localPath, recursive, force)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(i18n.T("scp.download.success"))
	} else {
		parts := strings.SplitN(destination, ":", 2)
		if len(parts) != 2 {
			fmt.Println(i18n.T("scp.error.invalid.format"))
			os.Exit(1)
		}
		connName := parts[0]
		remotePath := parts[1]
		localPath := source

		conn := findConnection(connections, connName)
		if conn == nil {
			fmt.Println(i18n.TWith("error.connection.not.found", map[string]interface{}{"Name": connName}))
			os.Exit(1)
		}

		fmt.Println(i18n.TWith("scp.uploading", map[string]interface{}{
			"Local": localPath,
			"User":  conn.User,
			"Host":  conn.Host,
			"Path":  remotePath,
		}))
		err = ssh.UploadFileWithOpts(conn, localPath, remotePath, recursive, force)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(i18n.T("scp.upload.success"))
	}
}

func findConnection(connections []config.Connection, name string) *config.Connection {
	for i, c := range connections {
		if c.Name == name {
			return &connections[i]
		}
	}
	return nil
}
