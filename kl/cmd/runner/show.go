package runner

import (
	"fmt"
	"os"
	"path"

	"github.com/kloudlite/kl/constants"
	common_util "github.com/kloudlite/kl/lib/common"
	"github.com/kloudlite/kl/lib/server"
	"github.com/kloudlite/kl/lib/ui/table"
	"github.com/kloudlite/kl/lib/ui/text"
	"github.com/kloudlite/kl/lib/util"
	"github.com/spf13/cobra"
)

var ShowCommand = &cobra.Command{
	Use:   "show",
	Short: "print your " + constants.CmdName + "-config file and current context",
	Long:  `Show kl-config`,
	Run: func(_ *cobra.Command, _ []string) {

		configFolder, err := util.GetConfigFolder()
		if err != nil {
			common_util.PrintError(err)
			return

		}
		contextFile, err := os.ReadFile(path.Join(configFolder, "config"))
		if err != nil {
			common_util.PrintError(err)
			return
		}

		file, err := os.ReadFile(server.GetConfigPath())
		if err != nil {
			common_util.PrintError(err)
			return
		}

		common_util.Log(table.HeaderText("context:"))
		common_util.Log(text.Colored("---------------------------------------", 4))
		fmt.Println(string(contextFile))

		common_util.Log(table.HeaderText("kl-config:"))
		common_util.Log(text.Colored("---------------------------------------", 4))
		fmt.Println(string(file))

	},
}
