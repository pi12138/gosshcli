package groupcmd

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
	Use:   "exec <group_name> <command>",
	Short: i18n.T("groups.exec.short"),
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		groupName := args[0]
		command := strings.Join(args[1:], " ")

		store, err := config.LoadStore()
		if err != nil {
			fmt.Println(i18n.TWith("error.loading.connections", map[string]interface{}{"Error": err}))
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
		for _, cName := range targetGroup.Connections {
			wg.Add(1)
			go func(name string) {
				defer wg.Done()
				fmt.Printf("[%s] ---\n", name)
				conn, err := config.ResolveConnection(name)
				if err != nil {
					fmt.Printf("[%s] Error resolving connection: %v\n", name, err)
					return
				}
				if err := ssh.ExecuteRemoteCommand(conn, command); err != nil {
					fmt.Printf("[%s] Error: %v\n", name, err)
				}
				fmt.Printf("[%s] Done\n", name)
			}(cName)
		}
		wg.Wait()
	},
}

func init() {
	GroupCmd.AddCommand(execCmd)
}
