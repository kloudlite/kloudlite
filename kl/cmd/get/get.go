package get

import (
	"github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "get",
	Short: "Get config or secret entries of current environment",
	Example: functions.Desc(`  {cmd} get config <configname>  		# get config entries
  {cmd} get secret <secretname>		# get secret entries
  {cmd} get [command] <name> -o <format>	# get config/secret in json/yaml format
`),
}

func init() {
	Cmd.AddCommand(configCmd)
	Cmd.AddCommand(secretCmd)
}
