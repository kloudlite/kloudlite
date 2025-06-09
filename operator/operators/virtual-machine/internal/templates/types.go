package templates

import (
	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type VMJobVars struct {
	JobMetadata metav1.ObjectMeta

	NodeSelector map[string]string
	Tolerations  []corev1.Toleration

	CloudProvider string

	JobImage           string
	JobImagePullPolicy string

	TFWorkspaceName      string
	TFWorkspaceNamespace string

	ValuesJSON string
}

type TFVarsGcpVM struct {
	AllowIncomingHttpTraffic bool                         `json:"allow_incoming_http_traffic"`
	AllowSsh                 bool                         `json:"allow_ssh"`
	AvailabilityZone         string                       `json:"availability_zone"`
	BootvolumeSize           float64                      `json:"bootvolume_size"`
	BootvolumeType           string                       `json:"bootvolume_type"`
	GcpCredentialsJson       string                       `json:"gcp_credentials_json"`
	GcpProjectID             string                       `json:"gcp_project_id"`
	GcpRegion                string                       `json:"gcp_region"`
	Labels                   map[string]string            `json:"labels"`
	MachineState             clustersv1.MachineState      `json:"machine_state"`
	MachineType              string                       `json:"machine_type"`
	NamePrefix               string                       `json:"name_prefix"`
	Network                  string                       `json:"network"`
	ProvisionMode            string                       `json:"provision_mode"`
	ServiceAccount           clustersv1.GCPServiceAccount `json:"service_account"`
	StartupScript            string                       `json:"startup_script"`
	VmName                   string                       `json:"vm_name"`
}
