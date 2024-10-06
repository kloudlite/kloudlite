package status

import (
	"errors"
	"fmt"
	"github.com/kloudlite/kl/cmd/connect"
	"github.com/kloudlite/kl/domain/envclient"
	"time"

	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/domain/fileclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

const (
	K3sServerNotReady = "k3s server is not ready, please wait"
)

var Cmd = &cobra.Command{
	Use:   "status",
	Short: "get status of your current context (user, account, environment, vpn status)",
	Run: func(cmd *cobra.Command, _ []string) {
		apic, err := apiclient.New()
		if err != nil {
			fn.PrintError(err)
			return
		}

		if u, err := apic.GetCurrentUser(); err == nil {
			fn.Logf("\nLogged in as %s (%s)\n",
				text.Blue(u.Name),
				text.Blue(u.Email),
			)
		}

		fc, err := fileclient.New()
		if err != nil {
			fn.PrintError(err)
			return
		}

		acc, err := fc.CurrentAccountName()
		if err == nil {
			fn.Log(fmt.Sprint(text.Bold(text.Blue("Account: ")), acc))
		}

		e, err := apic.EnsureEnv()
		if err == nil {
			fn.Log(fmt.Sprint(text.Bold(text.Blue("Environment: ")), e.Name))
		} else if errors.Is(err, fileclient.NoEnvSelected) {
			filePath := fn.ParseKlFile(cmd)
			klFile, err := fc.GetKlFile(filePath)
			if err != nil {
				fn.PrintError(err)
				return
			}
			fn.Log(fmt.Sprint(text.Bold(text.Blue("Environment: ")), klFile.DefaultEnv))
		}

		err = getK3sStatus()
		if err != nil {
			fn.Log("Local Cluster: ", text.Yellow("getting ready"))
			fn.Log("Edge Gateway Connection: ", text.Yellow("getting ready"))
			return
		}
	},
}

func getK3sStatus() error {
	fc, err := fileclient.New()
	if err != nil {
		return fn.NewE(err)
	}

	k3sTracker, err := fc.GetK3sTracker()
	if err != nil {
		return err
	}

	lastCheckedAt, err := time.Parse(time.RFC3339, k3sTracker.LastCheckedAt)
	if err != nil {
		return err
	}

	if time.Since(lastCheckedAt) > 3*time.Second {
		return fn.Error(K3sServerNotReady)
	}

	if k3sTracker.Compute {
		fn.Log("Local Cluster: ", text.Green("ready"))
	} else {
		fn.Log("Local Cluster: ", text.Yellow("getting ready"))
	}

	if k3sTracker.Gateway {
		fn.Log("Edge Gateway Connection: ", text.Green("ready"))
	} else {
		fn.Log("Edge Gateway Connection: ", text.Yellow("getting ready"))
	}

	if envclient.InsideBox() {
		if !k3sTracker.Gateway {
			fn.Log("Workspace Connection:", text.Yellow("offline"))
			return nil
		}
		if connect.ChekcWireguardConnection() {
			fn.Log("Workspace Connection:", text.Green("online"))
			return nil
		}
		fn.Log("Workspace Connection:", text.Yellow("offline"))
	}
	return nil
}
