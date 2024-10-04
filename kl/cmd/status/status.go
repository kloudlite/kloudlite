package status

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/domain/fileclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
	"net/http"
)

const (
	StatusFailed = "failed to get status"
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
			return
		}

	},
}

func getK3sStatus() error {
	deployments := []struct {
		name      string
		namespace string
	}{
		{"default", "kl-gateway"},
		{"kl-agent", "kloudlite"},
		{"kl-agent-operator", "kloudlite"},
	}

	for _, d := range deployments {
		isReady, err := checkDeploymentStatus(d.name, d.namespace)
		if err != nil {
			return err
		}
		status := text.Green("ready")
		if !isReady {
			status = text.Yellow("not ready")
		}
		fn.Log(fmt.Sprintf("%s: %s", d.name, status))
	}

	return nil
}

func checkDeploymentStatus(name, namespace string) (bool, error) {
	url := fmt.Sprintf("http://%s:8080/apis/apps/v1/namespaces/%s/deployments/%s", constants.K3sServerIp, namespace, name)
	resp, err := http.Get(url)
	if err != nil {
		return false, fn.NewE(err, StatusFailed)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fn.NewE(fmt.Errorf("unexpected status code: %d", resp.StatusCode), StatusFailed)
	}

	var data struct {
		Status struct {
			Conditions []struct {
				Type   string `json:"type"`
				Status string `json:"status"`
			} `json:"conditions"`
		} `json:"status"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	for _, c := range data.Status.Conditions {
		if c.Type == "Available" && c.Status == "True" {
			return true, nil
		}
	}

	return false, nil
}
