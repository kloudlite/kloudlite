package runner

import (
	"fmt"
	"io/ioutil"
	"path"

	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/lib/common"
	"github.com/kloudlite/kl/lib/common/ui/color"
	"github.com/kloudlite/kl/lib/common/ui/table"
	"github.com/kloudlite/kl/lib/server"
	"github.com/spf13/cobra"
)

var ShowCommand = &cobra.Command{
	Use:   "show",
	Short: "print your " + constants.CMD_NAME + "-config file and current context",
	Long:  `Show kl-config`,
	Run: func(_ *cobra.Command, _ []string) {

		configFolder, err := common.GetConfigFolder()
		if err != nil {
			common.PrintError(err)
			return

		}
		contextFile, err := ioutil.ReadFile(path.Join(configFolder, "config"))
		if err != nil {
			common.PrintError(err)
			return
		}

		file, err := ioutil.ReadFile(server.GetConfigPath())
		if err != nil {
			common.PrintError(err)
			return
		}

		common.Log(table.HeaderText("context:"))
		common.Log(color.Text("---------------------------------------", 4))
		fmt.Println(string(contextFile))

		common.Log(table.HeaderText("kl-config:"))
		common.Log(color.Text("---------------------------------------", 4))
		fmt.Println(string(file))

	},
}
