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
	Short: "switch infra context",
	Long: `This command let switch between infra contexts.
Example:
  # switch to existing infra context by selecting one from list 
  kl infra context switch

	# switch to infra context with infra context name
	kl infra context switch --name <infra context_name>
	`,
	Run: func(cmd *cobra.Command, _ []string) {
		name := ""
		if cmd.Flags().Changed("name") {
			name, _ = cmd.Flags().GetString("name")
		}

		c, err := client.GetInfraContexts()
		if err != nil {
			fn.PrintError(err)
			return
		}

		if len(c.InfraContexts) == 0 {
			fn.PrintError(fmt.Errorf("no infra context found"))
			return
		}

		if name == "" {
			n, err := fzf.FindOne(func() []string {
				var infraContexts []string
				for _, ctx := range c.InfraContexts {
					infraContexts = append(infraContexts, ctx.Name)
				}
				return infraContexts
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
			fn.PrintError(fmt.Errorf("infra context name is required"))
			return
		}

		if _, ok := c.InfraContexts[name]; !ok {
			fn.PrintError(fmt.Errorf("infra context %s not found", name))
			return
		}

		if err := client.SetActiveInfraContext(name); err != nil {
			fn.PrintError(err)
			return
		}

	},
}

func init() {
	switchCmd.Flags().StringP("name", "n", "", "infra context name")
	switchCmd.Aliases = append(switchCmd.Aliases, "sw")
}
