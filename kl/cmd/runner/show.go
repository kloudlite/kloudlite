package runner

import (
	"fmt"
	"github.com/kloudlite/kl/domain/client"
	common_util "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/table"
	"github.com/kloudlite/kl/pkg/ui/text"
	"os"
	"path"

	"github.com/kloudlite/kl/constants"
	"github.com/spf13/cobra"
)

var ShowCommand = &cobra.Command{
	Use:   "show",
	Short: "print your " + constants.CmdName + "-config file and current context",
	Long:  `Show kl-config`,
	Run: func(_ *cobra.Command, _ []string) {

		configFolder, err := client.GetConfigFolder()
		if err != nil {
			common_util.PrintError(err)
			return

		}
		contextFile, err := os.ReadFile(path.Join(configFolder, "config"))
		if err != nil {
			common_util.PrintError(err)
			return
		}

		file, err := os.ReadFile(client.GetConfigPath())
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
