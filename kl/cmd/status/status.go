package status

import (
	"errors"
	"fmt"
	"github.com/go-ping/ping"
	"github.com/kloudlite/kl/constants"
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
			fn.Log("Compute attached: ", text.Yellow("not ready"))
			fn.Log("Gateway attached: ", text.Yellow("not ready"))
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
		fn.Log("Compute attached: ", text.Green("ready"))
	} else {
		fn.Log("Compute attached: ", text.Yellow("not ready"))
	}

	if k3sTracker.Gateway {
		fn.Log("Gateway attached: ", text.Green("ready"))
	} else {
		fn.Log("Gateway attached: ", text.Yellow("not ready"))
	}

	if !k3sTracker.Gateway {
		fn.Log("Workspace status:", text.Yellow("offline"))
		return nil
	}

	if envclient.InsideBox() {
		pinger, err := ping.NewPinger(constants.KLDNS)
		if err != nil {
			return err
		}
		pinger.Count = 1
		pinger.Timeout = 2 * time.Second
		if err := pinger.Run(); err != nil {
			fn.Log("Workspace status:", text.Yellow("offline"))
			return nil
		}
		stats := pinger.Statistics()
		if stats.PacketsRecv == 0 {
			fn.Log("Workspace status:", text.Yellow("offline"))
			return nil
		}
		fn.Log("Workspace status:", text.Green("online"))
	}
	return nil
}
