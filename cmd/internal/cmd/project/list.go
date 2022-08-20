package project

import (
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list projects under selected account and select one",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command.
`,
	Run: func(cmd *cobra.Command, args []string) {
		TriggerSelectProject()
	},
}
