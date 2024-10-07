package use

import (
	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/kloudlite/kl/k3s"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"
	"github.com/spf13/cobra"
)

var teamCmd = &cobra.Command{
	Use:   "team",
	Short: "use team",
	Run: func(_ *cobra.Command, _ []string) {
		if err := UseTeam(); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func UseTeam() error {
	apic, err := apiclient.New()

	if err != nil {
		return fn.NewE(err)
	}
	teams, err := apic.ListTeams()
	if err != nil {
		return fn.NewE(err)
	}

	var selectedTeam *apiclient.Team

	if len(teams) == 0 {
		return fn.Error("no teams found")
	} else if len(teams) == 1 {
		selectedTeam = &teams[0]
	} else {
		selectedTeam, err = fzf.FindOne(teams, func(item apiclient.Team) string {
			return item.Metadata.Name
		}, fzf.WithPrompt("Select team to use >"))
		if err != nil {
			return err
		}
	}

	data, err := fileclient.GetExtraData()
	if err != nil {
		return fn.NewE(err)
	}

	data.SelectedTeam = selectedTeam.Metadata.Name

	err = fileclient.SaveExtraData(data)
	if err != nil {
		return fn.NewE(err)
	}

	k, err := k3s.NewClient()
	if err != nil {
		return err
	}
	if err = k.CreateClustersTeams(selectedTeam.Metadata.Name); err != nil {
		return fn.NewE(err)
	}
	fn.Log("Selected team is ", selectedTeam.Metadata.Name)
	return nil
}
