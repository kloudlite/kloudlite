package context

import (
	"fmt"
	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/input"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create new infra context",
	Long: `This command let create new infra context.
Example:
  # create new infra context
  kl infra context new

	# creating new infra context with name
	kl infra context new --name <infra_context_name>
	`,
	Run: func(cmd *cobra.Command, _ []string) {
		name := ""
		accountName := ""
		clusterName := ""
		if cmd.Flags().Changed("name") {
			name, _ = cmd.Flags().GetString("name")
		}
		if cmd.Flags().Changed("account") {
			accountName, _ = cmd.Flags().GetString("account")
		}
		if cmd.Flags().Changed("cluster") {
			clusterName, _ = cmd.Flags().GetString("cluster")
		}

		if name == "" {
			var err error
			name, err = input.Prompt(input.Options{
				Placeholder: "Enter infra context name",
				CharLimit:   15,
				Password:    false,
			})

			if err != nil {
				fn.PrintError(err)
				return
			}
		}

		if name == "" {
			fn.PrintError(fmt.Errorf("infra context name is required"))
			return
		}

		infraCtxs, err := client.GetInfraContexts()
		if err != nil {
			fn.PrintError(err)
			return
		}

		if _, ok := infraCtxs.InfraContexts[name]; ok {
			fn.PrintError(fmt.Errorf("infra context %s already exists", name))
			return
		}

		a, err := server.SelectAccount(accountName)
		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := client.WriteInfraContextFile(client.InfraContext{
			AccountName: a.Metadata.Name,
			Name:        name,
		}); err != nil {
			fn.PrintError(err)
			return
		}

		if err := client.SetActiveInfraContext(name); err != nil {
			fn.PrintError(err)
			return
		}

		c, err := server.EnsureCluster([]fn.Option{
			fn.MakeOption("accountName", a.Metadata.Name),
			fn.MakeOption("clusterName", clusterName),
		}...)

		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := client.WriteInfraContextFile(client.InfraContext{
			AccountName: a.Metadata.Name,
			Name:        name,
			ClusterName: c,
		}); err != nil {
			fn.PrintError(err)
			return
		}

		fn.Log(fmt.Sprintf("Infra Context %s created", name))
	},
}

func init() {
	newCmd.Flags().StringP("name", "n", "", "infra context name")
	newCmd.Flags().StringP("account", "a", "", "account name")
	newCmd.Flags().StringP("cluster", "d", "", "cluster name")
}
