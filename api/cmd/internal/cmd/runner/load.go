package runner

import (
	"github.com/spf13/cobra"
	"kloudlite.io/cmd/internal/common"
	"kloudlite.io/cmd/internal/lib/server"
)

var LoadCommand = &cobra.Command{
	Use:   "load",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		loadEnv()
	},
}

func loadEnv() {
	_, err := server.GenerateEnv()
	if err != nil {
		common.PrintError(err)
		return
	}

}
