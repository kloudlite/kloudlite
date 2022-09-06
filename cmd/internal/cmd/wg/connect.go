package wg

import (
	"fmt"
	"github.com/spf13/cobra"
	"kloudlite.io/cmd/internal/lib/common"
	"os"
	"os/exec"
)

func startServiceInBg() {
	command := exec.Command("kl", "wg", "connect")
	err := command.Start()
	if err != nil {
		fmt.Println(err)
		return
	}
	configFolder, err := common.GetConfigFolder()
	os.WriteFile(configFolder+"/wgpid", []byte(fmt.Sprintf("%d", command.Process.Pid)), 0644)
	return
}

var background bool

var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command.`,
	Run: func(cmd *cobra.Command, args []string) {
		if background {
			startServiceInBg()
			return
		}
		startService()
	},
}

func init() {
	connectCmd.Flags().BoolVar(&background, "background", false, "")
}
