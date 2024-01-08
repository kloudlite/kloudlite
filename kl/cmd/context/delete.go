package context

import (
	"github.com/spf13/cobra"
)

var delCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete existing context",
	Long: `This command let you delete existing context.
Example:
  # delete context
  kl context [delete|del] <context_name>

	# interactive delete context
  kl context [delete|del]
	`,
	Run: func(_ *cobra.Command, _ []string) {

	},
}
