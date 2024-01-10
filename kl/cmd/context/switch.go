package context

import (
	"fmt"

	"github.com/kloudlite/kl/domain/client"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"
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
	kl context switch --name <context_name>
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

		if name == "" {
			fn.PrintError(fmt.Errorf("context name is required"))
			return
		}

		if _, ok := c.Contexts[name]; !ok {
			fn.PrintError(fmt.Errorf("context %s not found", name))
			return
		}

		if err := client.SetActiveContext(name); err != nil {
			fn.PrintError(err)
			return
		}

	},
}

func init() {
	switchCmd.Flags().StringP("name", "n", "", "context name")
	switchCmd.Aliases = append(switchCmd.Aliases, "sw")
}
