package cluster

import (
	"fmt"
	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/kloudlite/kl/k3s"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "clean the cluster",
	Run: func(cmd *cobra.Command, args []string) {
		if err := cleanCluster(cmd); err != nil {
			fmt.Println(err)
			return
		}
	},
}

func cleanCluster(cmd *cobra.Command) error {
	fc, err := fileclient.New()
	if err != nil {
		return err
	}

	apic, err := apiclient.New()
	if err != nil {
		return err
	}

	k3sClient, err := k3s.NewClient(cmd)
	if err != nil {
		return err
	}

	data, err := fileclient.GetExtraData()
	if err != nil {
		return fn.NewE(err)
	}

	fn.Printf(text.Yellow(fmt.Sprintf("this will delete k3s cluster for team %s and all its data and volumes. Do you want to continue? (y/N): ", data.SelectedTeam)))
	if !fn.Confirm("Y", "N") {
		return nil
	}

	clusters, err := apic.GetClustersOfTeam(data.SelectedTeam)
	if err != nil {
		return fn.NewE(err)
	}
	wgConfig, err := fc.GetWGConfig()
	if err != nil {
		return fn.NewE(err)
	}
	for _, c := range clusters {
		if c.Metadata.Labels["kloudlite.io/local-uuid"] == wgConfig.UUID {
			if err := apic.DeleteCluster(data.SelectedTeam, c.Metadata.Name); err != nil {
				return fn.NewE(err)
			}
			if err = fc.DeleteClusterData(data.SelectedTeam); err != nil {
				return fn.NewE(err)
			}
			if err = k3sClient.RemoveClusterVolume(c.Metadata.Name); err != nil {
				return fn.NewE(err)
			}
			fn.Log(fmt.Sprintf("cluster %s of team %s deleted", c.Metadata.Name, data.SelectedTeam))
			break
		}
	}
	return nil

}
