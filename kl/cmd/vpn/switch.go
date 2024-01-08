package vpn

import (
	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:   "switch",
	Short: "switch to context",
	Long: `This command let switch between contexts.
Example:
  # switch to existing context by selecting one from list 
  kl context switch

	# switch to context with context name
	kl context switch <context_name>
	`,
	Run: func(_ *cobra.Command, _ []string) {

	},
}
