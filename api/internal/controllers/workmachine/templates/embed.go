package templates

import (
	"embed"

	"github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/template"
	corev1 "k8s.io/api/core/v1"
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

	K3sAgentSetup templateFile = "k3s-agent-setup.yml"
)

type K3sAgentSetupArgs struct {
	K3sVersion    string
	K3sURL        string
	K3sAgentToken string
	MachineName   string
	MachineOwner  string
}

type WorkspaceHostManagerValues struct {
	Namespace       string
	WorkMachineName string
	SSHUsername string
	NodeSelector    map[string]string
	Tolerations     []corev1.Toleration
}
