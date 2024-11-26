package start

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"

	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kloudlite/kl-platform/domain/fileclient"
	"github.com/kloudlite/kloudlite/kl-platform/pkg/k3s"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "start",
	Short: "start the kloudlite platform",
	Run: func(cmd *cobra.Command, args []string) {
		if err := Run(cmd, args); err != nil {
			fn.PrintError(err)
		}
	},
}

func Run(cmd *cobra.Command, args []string) error {

	fc, err := fileclient.New(cmd)
	if err != nil {
		return err
	}

	cf, err := fc.GetConfigFile()
	if err != nil {
		return err
	}

	r, err := http.Get("https://github.com/kloudlite/helm-charts/releases/latest/download/crds-all.yml")

	if err != nil {
		return err
	}

	crds, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	staticPath := "/var/lib/rancher/k3s/server/manifests"
	if err := os.MkdirAll(staticPath, 0755); err != nil {
		return err
	}

	if err := os.WriteFile(path.Join(staticPath, "kloudlite-crds.yml"), crds, 0644); err != nil {
		return err
	}

	if err := os.WriteFile(path.Join(staticPath, "kloudlite.yaml"), []byte(fmt.Sprintf(`
apiVersion: v1
kind: Namespace
metadata:
  name: kloudlite
---
apiVersion: helm.cattle.io/v1
kind: HelmChart
metadata:
  name: kloudlite-platform
  namespace: kloudlite
  annotations:
    redeploy: "2" # Increment this value to trigger re-application
spec:
  repo: https://kloudlite.github.io/helm-charts
  chart: kloudlite-platform
  targetNamespace: kloudlite
  valuesContent: |-
    baseDomain: %s
`, cf.KlConfig.BaseDomain)), 0644); err != nil {
		return err
	}

	k := k3s.NewK3s(cmd.Context())
	if err := k.StartServer(cf.K3sArgs); err != nil {
		return err
	}

	return nil
}

func init() {
}
