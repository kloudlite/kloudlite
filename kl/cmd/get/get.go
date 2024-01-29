package get

import (
	"github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "get",
	Short: "get [ config | secret ] entries",
	Long:  `get config/secret entries.`,
	Example: functions.Desc(`# get config table
{cmd} get config <configname>

# get secret table
{cmd} get secret <secretname>

# get config/secret in yaml format
{cmd} get [command] <name> -o yaml

# get config/secret in json format
{cmd} get [command] <name> -o json
`),
}

func init() {
	Cmd.AddCommand(configCmd)
	Cmd.AddCommand(secretCmd)
}
