package add

import (
	"fmt"

	"github.com/spf13/cobra"
	"kloudlite.io/cmd/internal/lib/server"
)

var addMresCommand = &cobra.Command{
	Use:   "mres",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		SelectMreses()
	},
}

func SelectMreses() {
	configs, err := server.GetMreses()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(configs)
}
