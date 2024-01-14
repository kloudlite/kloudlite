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
	Short: "Create new context",
	Long: `This command let create new context.
Example:
  # create new context
  kl context new

	# creating new context with name
	kl context new --name <context_name>
	`,
	Run: func(cmd *cobra.Command, _ []string) {
		name := ""
		accountName := ""
		deviceName := ""

		name = fn.ParseStringFlag(cmd, "name")
		accountName = fn.ParseStringFlag(cmd, "account")
		deviceName = fn.ParseStringFlag(cmd, "device")

		if name == "" {
			var err error
			name, err = input.Prompt(input.Options{
				Placeholder: "Enter context name",
				CharLimit:   15,
				Password:    false,
			})

			if err != nil {
				fn.PrintError(err)
				return
			}
		}

		if name == "" {
			fn.PrintError(fmt.Errorf("context name is required"))
			return
		}

		ctxs, err := client.GetContexts()
		if err != nil {
			fn.PrintError(err)
			return
		}

		if _, ok := ctxs.Contexts[name]; ok {
			fn.PrintError(fmt.Errorf("context %s already exists", name))
			return
		}

		a, err := server.SelectAccount(accountName)
		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := client.WriteContextFile(client.Context{
			AccountName: a.Metadata.Name,
			Name:        name,
		}); err != nil {
			fn.PrintError(err)
			return
		}

		if err := client.SetActiveContext(name); err != nil {
			fn.PrintError(err)
			return
		}

		d, err := server.EnsureDevice([]fn.Option{
			fn.MakeOption("accountName", a.Metadata.Name),
			fn.MakeOption("deviceName", deviceName),
		}...)

		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := client.WriteContextFile(client.Context{
			AccountName: a.Metadata.Name,
			Name:        name,
			DeviceName:  d,
		}); err != nil {
			fn.PrintError(err)
			return
		}

		fn.Log(fmt.Sprintf("Context %s created", name))
	},
}

func init() {
	newCmd.Flags().StringP("name", "n", "", "context name")
	newCmd.Flags().StringP("account", "a", "", "account name")
	newCmd.Flags().StringP("device", "d", "", "device name")
}
