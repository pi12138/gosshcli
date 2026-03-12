package cmd

import (
	"fmt"
	"gossh/internal/config"
	"gossh/internal/i18n"
	"gossh/internal/ssh"
	"os"
	"strings"
	"sync"

	"github.com/spf13/cobra"
)

var scpCmd = &cobra.Command{
	Use:   "scp [source] [destination...]",
	Short: i18n.T("scp.short"),
	Long:  i18n.T("scp.long"),
	Args:  cobra.MinimumNArgs(2),
	Run:   runScp,
}

func init() {
	scpCmd.Flags().BoolP("recursive", "r", false, "Copy directories recursively")
	scpCmd.Flags().BoolP("force", "f", false, "Force overwrite of existing files")
	scpCmd.Flags().BoolP("parallel", "p", false, "Enable parallel transfer to multiple destinations")
	scpCmd.Flags().BoolP("quiet", "q", false, "Suppress progress bars")
}

func runScp(cmd *cobra.Command, args []string) {
	recursive, _ := cmd.Flags().GetBool("recursive")
	force, _ := cmd.Flags().GetBool("force")
	parallel, _ := cmd.Flags().GetBool("parallel")
	quiet, _ := cmd.Flags().GetBool("quiet")

	source := args[0]
	destinations := args[1:]

	sourceHasColon := strings.Contains(source, ":")

	// Check for multi-device transfer requirement
	if len(destinations) > 1 && !parallel {
		fmt.Println("Error: Multiple destinations detected. Use -p or --parallel to enable multi-device transfer.")
		os.Exit(1)
	}

	connections, err := config.LoadConnections()
	if err != nil {
		fmt.Println(i18n.TWith("error.loading.connections", map[string]interface{}{"Error": err}))
		os.Exit(1)
	}

	if sourceHasColon {
		// Download mode: scp host:path local_path
		if len(destinations) > 1 {
			fmt.Println("Error: Downloading from multiple sources to one local path is not supported yet.")
			os.Exit(1)
		}
		
		parts := strings.SplitN(source, ":", 2)
		connName := parts[0]
		remotePath := parts[1]
		localPath := destinations[0]

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
		err = ssh.DownloadFileWithOpts(conn, remotePath, localPath, recursive, force, quiet)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(i18n.T("scp.download.success"))
	} else {
		// Upload mode: scp local_path host1:path host2:path ...
		if parallel {
			quiet = true // Force quiet in parallel mode to avoid bar conflicts
			var wg sync.WaitGroup
			errChan := make(chan error, len(destinations))

			for _, dest := range destinations {
				if !strings.Contains(dest, ":") {
					fmt.Printf("Error: Invalid destination format '%s'. Expected 'host:path'\n", dest)
					continue
				}

				wg.Add(1)
				go func(d string) {
					defer wg.Done()
					parts := strings.SplitN(d, ":", 2)
					connName := parts[0]
					remotePath := parts[1]

					conn := findConnection(connections, connName)
					if conn == nil {
						errChan <- fmt.Errorf("connection '%s' not found", connName)
						return
					}

					fmt.Printf("Uploading to %s:%s...\n", conn.Host, remotePath)
					err := ssh.UploadFileWithOpts(conn, source, remotePath, recursive, force, quiet)
					if err != nil {
						errChan <- fmt.Errorf("[%s] %v", conn.Name, err)
					}
				}(dest)
			}

			wg.Wait()
			close(errChan)

			hasError := false
			for err := range errChan {
				fmt.Printf("Task Error: %v\n", err)
				hasError = true
			}
			if !hasError {
				fmt.Println(i18n.T("scp.upload.success"))
			} else {
				os.Exit(1)
			}
		} else {
			// Single target upload
			destination := destinations[0]
			if !strings.Contains(destination, ":") {
				fmt.Println(i18n.T("scp.error.no.remote"))
				os.Exit(1)
			}

			parts := strings.SplitN(destination, ":", 2)
			connName := parts[0]
			remotePath := parts[1]

			conn := findConnection(connections, connName)
			if conn == nil {
				fmt.Println(i18n.TWith("error.connection.not.found", map[string]interface{}{"Name": connName}))
				os.Exit(1)
			}

			fmt.Println(i18n.TWith("scp.uploading", map[string]interface{}{
				"Local": source,
				"User":  conn.User,
				"Host":  conn.Host,
				"Path":  remotePath,
			}))
			err = ssh.UploadFileWithOpts(conn, source, remotePath, recursive, force, quiet)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(i18n.T("scp.upload.success"))
		}
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

