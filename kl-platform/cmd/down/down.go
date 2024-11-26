package down

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "down",
	Short: "stop the platform",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("stop the platform")
	},
}

func init() {
}
