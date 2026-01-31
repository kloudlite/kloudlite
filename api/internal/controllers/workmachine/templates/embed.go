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

	WorkMachineHostManagerPod templateFile = "workmachine-host-manager-pod.yaml.tpl"

	K3sAgentSetup    templateFile = "k3s-agent-setup.yml"     // Cloud-init format for Azure/GCP
	K3sAgentSetupAWS templateFile = "k3s-agent-setup-aws.yml" // Bash script for AWS (handles NVMe device detection)
)

type K3sAgentSetupArgs struct {
	K3sVersion      string
	K3sURL          string
	K3sAgentToken   string
	MachineName     string
	MachineOwner    string
	HostedSubdomain string // e.g., "mega.khost.dev" - used for registry mirror config
	BtrfsDevice     string // Device path for BTRFS storage (e.g., /dev/xvdf for AWS, /dev/disk/azure/scsi1/lun0 for Azure)
}

type WorkspaceHostManagerValues struct {
	Namespace        string
	WorkMachineName  string
	TargetNamespace  string
	SSHUsername      string
	HostManagerImage string
}
