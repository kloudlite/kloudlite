package expose

import (
	"os"

	"github.com/kloudlite/kl/cmd/box/boxpkg"
	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "sync ports",
	Long: `
This command will sync ports to your kl-config file.
`,
	Example: ` 
  kl expose sync
`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := sync(cmd, args); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func sync(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return functions.NewE(err)
	}
	klFile, err := client.GetKlFile("")
	if err != nil {
		return functions.NewE(err)
	}
	containerWorkspacePath := cwd
	if val, ok := os.LookupEnv("KL_WORKSPACE"); ok {
		containerWorkspacePath = val
	}

	c, err := boxpkg.NewClient(cmd, args)
	if err != nil {
		return functions.NewE(err)
	}

	if err = c.SyncProxy(boxpkg.ProxyConfig{
		ExposedPorts:        klFile.Ports,
		TargetContainerPath: containerWorkspacePath,
	}); err != nil {
		return fn.NewE(err)
	}
	return nil
}
