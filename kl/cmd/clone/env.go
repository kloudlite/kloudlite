package clone

import (
	"fmt"
	"github.com/kloudlite/kl/cmd/box/boxpkg"
	"github.com/kloudlite/kl/cmd/box/boxpkg/hashctrl"
	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/kloudlite/kl/k3s"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
	"os"
)

var cloneCmd = &cobra.Command{
	Use:   "env",
	Short: "Switch to a different environment",
	Run: func(cmd *cobra.Command, args []string) {
		if err := envClone(cmd, args); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func envClone(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fn.Error("env name is required")
	}
	envName := args[0]

	fc, err := fileclient.New()
	if err != nil {
		return err
	}

	apic, err := apiclient.New()
	if err != nil {
		return err
	}

	klFile, err := fc.GetKlFile("")
	if err != nil {
		return err
	}

	isValidName, err := checkEnvNameAvailability(apic, fc, envName)
	if err != nil {
		return err
	}

	if !isValidName {
		return fn.Error("env name is not available")
	}

	cluster, err := selectCluster(apic, fc)
	if err != nil {
		return err
	}
	env, err := cloneEnv(apic, fc, envName, cluster.Metadata.Name)
	if err != nil {
		return err
	}

	if klFile.DefaultEnv == "" {
		klFile.DefaultEnv = env.Metadata.Name
		if err := fc.WriteKLFile(*klFile); err != nil {
			return err
		}
	}
	fn.Log(text.Bold(text.Green("\nSelected Environment:")),
		text.Blue(fmt.Sprintf("\n%s (%s)", env.DisplayName, env.Metadata.Name)),
	)

	wpath, err := os.Getwd()
	if err != nil {
		return err
	}
	if err := hashctrl.SyncBoxHash(apic, fc, wpath); err != nil {
		return err
	}

	c, err := boxpkg.NewClient(cmd, nil)
	if err != nil {
		return err
	}

	if err := c.ConfirmBoxRestart(); err != nil {
		return err
	}
	return nil
}

func cloneEnv(apic apiclient.ApiClient, fc fileclient.FileClient, newEnvName string, clusterName string) (*apiclient.Env, error) {
	currentTeam, err := fc.CurrentTeamName()
	if err != nil {
		return nil, fn.NewE(err)
	}

	oldEnv, err := apic.EnsureEnv()
	if err != nil {
		return nil, fn.NewE(err)
	}

	env, err := apic.CloneEnv(currentTeam, oldEnv.Name, newEnvName, clusterName)
	if err != nil {
		return nil, fn.NewE(err)
	}

	err = apic.RemoveAllIntercepts()
	if err != nil {
		return nil, fn.NewE(err)
	}

	k3sClient, err := k3s.NewClient()
	if err != nil {
		return nil, fn.NewE(err)
	}
	if err = k3sClient.RemoveAllIntercepts(); err != nil {
		return nil, fn.NewE(err)
	}

	persistSelectedEnv := func(e fileclient.Env) error {
		err := fc.SelectEnv(e)
		if err != nil {
			return fn.NewE(err)
		}
		return nil
	}

	if err := persistSelectedEnv(fileclient.Env{
		Name: env.Metadata.Name,
		SSHPort: func() int {
			if oldEnv == nil {
				return 0
			}
			return oldEnv.SSHPort
		}(),
	}); err != nil {
		return nil, fn.NewE(err)
	}
	return env, nil
}

func selectCluster(apic apiclient.ApiClient, fc fileclient.FileClient) (*apiclient.BYOKCluster, error) {
	currentTeam, err := fc.CurrentTeamName()
	if err != nil {
		return nil, fn.NewE(err)
	}

	clusters, err := apic.ListBYOKClusters(currentTeam)
	if err != nil {
		return nil, fn.NewE(err)
	}

	cluster, err := fzf.FindOne(
		clusters,
		func(clus apiclient.BYOKCluster) string {
			return fmt.Sprintf("%s (%s)", clus.DisplayName, clus.Metadata.Name)
		},
		fzf.WithPrompt("Select Cluster > "),
	)

	if err != nil {
		return nil, fn.NewE(err)
	}

	return cluster, nil

}

func checkEnvNameAvailability(apic apiclient.ApiClient, fc fileclient.FileClient, envName string) (bool, error) {
	currentTeam, err := fc.CurrentTeamName()
	if err != nil {
		return false, fn.NewE(err)
	}

	return apic.CheckEnvName(currentTeam, envName)
}
