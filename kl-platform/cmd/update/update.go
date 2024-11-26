package update

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "update",
	Short: "update the platform",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("init the platform")
	},
}

func init() {
}
