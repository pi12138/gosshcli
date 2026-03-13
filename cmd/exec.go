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

var execCmd = &cobra.Command{
	Use:   "exec [name] [command]",
	Short: i18n.T("exec.short"),
	Long:  i18n.T("exec.long"),
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		groupName, _ := cmd.Flags().GetString("group")

		if groupName == "" && len(args) < 2 {
			fmt.Println(i18n.T("exec.error.usage"))
			_ = cmd.Help()
			os.Exit(1)
		}

		store, err := config.LoadStore()
		if err != nil {
			fmt.Println(i18n.TWith("error.loading.connections", map[string]interface{}{"Error": err}))
			os.Exit(1)
		}

		if groupName != "" {
			command := strings.Join(args, " ")
			if command == "" {
				fmt.Println(i18n.T("exec.error.command.required"))
				os.Exit(1)
			}

			var targetGroup *config.Group
			for i, g := range store.Groups {
				if g.Name == groupName {
					targetGroup = &store.Groups[i]
					break
				}
			}

			if targetGroup == nil || len(targetGroup.Connections) == 0 {
				fmt.Println(i18n.TWith("groups.no.connections", map[string]interface{}{"Group": groupName}))
				return
			}

			fmt.Println(i18n.TWith("groups.executing", map[string]interface{}{
				"Command": command,
				"Group":   groupName,
				"Count":   len(targetGroup.Connections),
			}))

			var wg sync.WaitGroup
			for _, connName := range targetGroup.Connections {
				wg.Add(1)
				go func(cName string) {
					defer wg.Done()
					resolvedConn, err := config.ResolveConnection(cName)
					if err != nil {
						fmt.Printf("[%s] Resolve Error: %v\n", cName, err)
						return
					}
					fmt.Printf("[%s] ---\n", cName)
					if err := ssh.ExecuteRemoteCommand(resolvedConn, command); err != nil {
						fmt.Printf("[%s] Error: %v\n", cName, err)
					}
					fmt.Printf("[%s] Done\n", cName)
				}(connName)
			}
			wg.Wait()
		} else {
			connectionName := args[0]
			command := strings.Join(args[1:], " ")

			resolvedConn, err := config.ResolveConnection(connectionName)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			if err := ssh.ExecuteRemoteCommand(resolvedConn, command); err != nil {
				os.Exit(1)
			}
		}
	},
}

func init() {
	execCmd.Flags().StringP("group", "g", "", i18n.T("exec.flag.group"))
}
