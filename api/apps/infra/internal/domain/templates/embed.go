package templates

import (
	"embed"
	"path/filepath"

	"github.com/kloudlite/operator/pkg/templates"
)

//go:embed *
var templatesDir embed.FS

type templateFile string

const (
	HelmKloudliteAgent               templateFile = "./helm-charts-kloudlite-agent.yml.tpl"
	GlobalVPNKloudliteDeviceTemplate templateFile = "./global-vpn-kloudlite-device.yml.tpl"
	GatewayServiceTemplate          templateFile = "./kloudlite-gateway-svc.yml.tpl"
)

func Read(t templateFile) ([]byte, error) {
	return templatesDir.ReadFile(filepath.Join(string(t)))
}

var ParseBytes = templates.ParseBytes
