package infra

import (
	"github.com/kloudlite/kl/cmd/infra/context"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "infra",
	Short: "create new infra context and manage existing infra contexts",
	Long: `Create new infra context and manage infra existing contexts
Examples:
  # creating new context
  kl infra context new

  # list all contexts
  kl infra context list

  # switch to context
  kl infra context switch <context_name>

  # remove context
  kl infra context remove <context_name>
	`,
}

func init() {
	Cmd.Aliases = append(Cmd.Aliases, "infra")

	Cmd.AddCommand(context.Cmd)
}
