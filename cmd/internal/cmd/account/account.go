package account

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "account",
	Short: "list | use  account",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command.
`,
}
