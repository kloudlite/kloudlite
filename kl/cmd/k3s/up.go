package k3s

import (
	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/kloudlite/kl/k3s"
	"github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"github.com/spf13/cobra"
)

var UpCmd = &cobra.Command{
	Use:   "up",
	Short: "Starts the k3s server",
	Long:  `Starts the k3s server`,
	Run: func(_ *cobra.Command, _ []string) {
		if err := startK3sServer(); err != nil {
			functions.PrintError(err)
			return
		}
	},
}

func startK3sServer() error {
	defer spinner.Client.UpdateMessage("starting k3s server")()
	data, err := fileclient.GetExtraData()
	if err != nil {
		return functions.NewE(err)
	}
	if data.SelectedAccount == "" {
		return functions.Error("No account selected")
	}
	k, err := k3s.NewClient()
	if err != nil {
		return err
	}
	if err = k.CreateClustersAccounts(data.SelectedAccount); err != nil {
		return functions.NewE(err)
	}
	functions.Log("k3s server started")
	return nil
}
