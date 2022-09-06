package runner

import (
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
	"kloudlite.io/cmd/internal/common"
	"kloudlite.io/cmd/internal/constants"
	"kloudlite.io/cmd/internal/lib/server"
)

var ShowCommand = &cobra.Command{
	Use:   "show",
	Short: "print your " + constants.CMD_NAME + "-config file and current context",
	Long:  `Show kl-config`,
	Run: func(_ *cobra.Command, _ []string) {

		type Ctx struct {
			ProjectId string
			AccountId string
		}

		file, err := ioutil.ReadFile(server.GetConfigPath())
		if err != nil {
			common.PrintError(err)
			return
		}

		fmt.Println(string(file))
	},
}
