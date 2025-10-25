package templates

import (
	"embed"

	"github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/template"
)

//go:embed *
var templatesDir embed.FS

type templateFile string

func (tf templateFile) Render(values any) ([]byte, error) {
	b, err := templatesDir.ReadFile(string(tf))
	if err != nil {
		return nil, err
	}
	return template.Render(b, values)
}

const (
	AWS_CreateInstanceScript templateFile = "scripts/aws/create-instance.sh"
	AWS_DeleteInstanceScript templateFile = "scripts/aws/delete-instance.sh"
	AWS_StartInstanceScript  templateFile = "scripts/aws/start-instance.sh"
	AWS_StopInstanceScript   templateFile = "scripts/aws/stop-instance.sh"

	WorkMachineHostManagerDeployment templateFile = "workmachine-host-manager-deployment.yaml.tpl"

	K3sAgentSetup templateFile = "./k3s-agent-setup.yml"
)

type K3sAgentSetupArgs struct {
	K3sURL       string
	K3sToken     string
	MachineName  string
	MachineOwner string
}
