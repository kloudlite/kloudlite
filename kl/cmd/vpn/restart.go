package vpn

import (
	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "list",
	Short: "listing all contexts",
	Long: `This command let you list all contexts.
Example:
  # list all contexts
  kl context list
	`,
	Run: func(_ *cobra.Command, _ []string) {

	},
}
