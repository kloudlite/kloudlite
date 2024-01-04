package runner

import (
	"fmt"

	"github.com/kloudlite/kl/domain/client"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/table"
	"github.com/kloudlite/kl/pkg/ui/text"
	"sigs.k8s.io/yaml"

	"github.com/kloudlite/kl/constants"
	"github.com/spf13/cobra"
)

var ShowCommand = &cobra.Command{
	Use:   "show",
	Short: "print your " + constants.CmdName + "-config file and current context",
	Long:  `Show kl-config`,
	Run: func(_ *cobra.Command, _ []string) {

		k, err := client.GetContextFile()
		if err != nil {
			fn.PrintError(err)
			return
		}

		k.Session = "*********"
		contextFile, err := yaml.Marshal(k)
		if err != nil {
			fn.PrintError(err)
			return
		}

		kfile, err := client.GetKlFile(nil)
		if err != nil {
			fn.PrintError(err)
			return
		}

		yamlFile, err := yaml.Marshal(kfile)
		if err != nil {
			fn.PrintError(err)
			return
		}

		fn.Log(table.HeaderText("context:"))
		fn.Log(text.Colored("---------------------------------------", 4))
		fmt.Println(string(contextFile))

		fn.Log(table.HeaderText("kl-config:"))
		fn.Log(text.Colored("---------------------------------------", 4))
		fmt.Println(string(yamlFile))
	},
}
