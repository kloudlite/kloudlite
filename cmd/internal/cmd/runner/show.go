package runner

import (
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
	"kloudlite.io/cmd/internal/common"
	"kloudlite.io/cmd/internal/constants"
)

var ShowCommand = &cobra.Command{
	Use:   "show",
	Short: "see your " + constants.CMD_NAME + "-config file",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command.`,
	Run: func(cmd *cobra.Command, args []string) {

		file, err := ioutil.ReadFile(".kl.yml")
		if err != nil {
			common.PrintError(err)
			return
		}

		fmt.Println(string(file))

	},
}
