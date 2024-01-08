package vpn

import (
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create new context",
	Long: `This command let create new context.
Example:
  # create new context
  kl context new

	# creating new context with name
	kl context new --name <context_name>
	`,
	Run: func(_ *cobra.Command, _ []string) {

	},
}
