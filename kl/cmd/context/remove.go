package context

import (
	"fmt"

	"github.com/kloudlite/kl/domain/client"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "remove existing context from contexts list",
	Long: `This command let you remove existing context from contexts list.
Example:
  # remove context
  kl context [remove|rm] <context_name>

	# interactive delete context
  kl context [remove|rm]
	`,
	Run: func(cmd *cobra.Command, _ []string) {
		name := ""

		if cmd.Flags().Changed("name") {
			name, _ = cmd.Flags().GetString("name")
		}

		c, err := client.GetContexts()
		if err != nil {
			fn.PrintError(err)
			return
		}

		if len(c.Contexts) == 0 {
			fn.PrintError(fmt.Errorf("no context found"))
			return
		}

		if name == "" {
			n, err := fzf.FindOne(func() []string {
				var contexts []string
				for _, ctx := range c.Contexts {
					contexts = append(contexts, ctx.Name)
				}
				return contexts
			}(), func(item string) string {
				return item
			})
			if err != nil {
				fn.PrintError(err)
				return
			}

			name = *n
		}

		if err := client.DeleteContext(name); err != nil {
			fn.PrintError(err)
			return
		}

		fn.Log(fmt.Sprintf("context %s removed", name))
	},
}

func init() {
	removeCmd.Aliases = append(removeCmd.Aliases, "rm")
	removeCmd.Flags().StringP("name", "n", "", "context name")
}
