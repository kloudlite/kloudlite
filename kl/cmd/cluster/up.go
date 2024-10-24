package cluster

import (
	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/kloudlite/kl/k3s"
	"github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"github.com/spf13/cobra"
	"os"
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Starts the k3s server",
	Long:  `Starts the k3s server`,
	Run: func(cmd *cobra.Command, _ []string) {
		if err := startK3sServer(cmd); err != nil {
			functions.PrintError(err)
			return
		}
	},
}

func startK3sServer(cmd *cobra.Command) error {
	defer spinner.Client.UpdateMessage("starting k3s server")()
	fc, err := fileclient.New()
	if err != nil {
		return functions.NewE(err)
	}
	currentTeam, err := fc.CurrentTeamName()
	if err != nil {
		return functions.NewE(err)
	}
	extraData, err := fileclient.GetExtraData()
	if (err != nil && os.IsNotExist(err)) || extraData.SelectedTeam == "" {
		extraData.SelectedTeam = currentTeam
		if err := fileclient.SaveExtraData(extraData); err != nil {
			return functions.NewE(err)
		}
	} else if err != nil {
		return functions.NewE(err)
	}

	k, err := k3s.NewClient(cmd)
	if err != nil {
		return functions.NewE(err)
	}
	if err = k.CreateClustersTeams(extraData.SelectedTeam); err != nil {
		return functions.NewE(err)
	}
	functions.Log("k3s server started. It will usually take a minute to come online")
	return nil
}
