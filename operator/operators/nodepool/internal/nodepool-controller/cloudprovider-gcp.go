package nodepool_controller

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	fn "github.com/kloudlite/operator/pkg/functions"
	rApi "github.com/kloudlite/operator/pkg/operator"
	corev1 "k8s.io/api/core/v1"
)

type GCPServiceAccount struct {
	Enabled bool     `json:"enabled"`
	Email   *string  `json:"email,omitempty"`
	Scopes  []string `json:"scopes,omitempty"`
}

type GCPWorkerValues struct {
	GCPProjectID       string `json:"gcp_project_id"`
	GCPRegion          string `json:"gcp_region"`
	GCPCredentialsJSON string `json:"gcp_credentials_json"`

	NamePrefix   string `json:"name_prefix"`
	NodepoolName string `json:"nodepool_name"`

	ProvisionMode string `json:"provision_mode"`

	AvailabilityZone string `json:"availability_zone"`

	Network        string `json:"network"`
	BootVolumeType string `json:"bootvolume_type"`
	BootVolumeSize int    `json:"bootvolume_size"`

	ServiceAccount GCPServiceAccount `json:"service_account"`

	Nodes map[string]any `json:"nodes"`

	NodeLabels  map[string]string `json:"node_labels"`
	MachineType string            `json:"machine_type"`

	K3sServerPublicDNSHost string   `json:"k3s_server_public_dns_host"`
	K3sJoinToken           string   `json:"k3s_join_token"`
	K3sExtraAgentArgs      []string `json:"k3s_extra_agent_args"`

	ClusterInternalDNSHost   string            `json:"cluster_internal_dns_host"`
	SaveSSHKeyToPath         string            `json:"save_ssh_key_to_path"`
	KloudliteRelease         string            `json:"kloudlite_release"`
	LabelCloudproviderRegion string            `json:"label_cloudprovider_region"`
	Labels                   map[string]string `json:"labels"`

	AllowIncomingHttpTraffic bool `json:"allow_incoming_http_traffic"`
}

func (r *Reconciler) GCPJobValuesJson(obj *clustersv1.NodePool, nodesMap map[string]clustersv1.NodeProps) (string, error) {
	if obj.Spec.GCP == nil {
		return "", fmt.Errorf(".spec.gcp is nil")
	}

	creds, err := rApi.Get(context.TODO(), r.Client, fn.NN(obj.Spec.GCP.Credentials.Namespace, obj.Spec.GCP.Credentials.Name), &corev1.Secret{})
	if err != nil {
		return "", err
	}

	gcpCreds, err := fn.ParseFromSecret[clustersv1.GCPCredentials](creds)
	if err != nil {
		return "", err
	}

	values := GCPWorkerValues{
		GCPProjectID:       obj.Spec.GCP.GCPProjectID,
		GCPRegion:          obj.Spec.GCP.Region,
		GCPCredentialsJSON: base64.StdEncoding.EncodeToString([]byte(gcpCreds.ServiceAccountJSON)),

		NamePrefix:       "np",
		NodepoolName:     obj.Name,
		ProvisionMode:    string(obj.Spec.GCP.PoolType),
		AvailabilityZone: obj.Spec.GCP.AvailabilityZone,
		Network: func() string {
			if obj.Spec.GCP.VPC != nil {
				return obj.Spec.GCP.VPC.Name
			}
			return "default"
		}(),
		BootVolumeType: obj.Spec.GCP.BootVolumeType,
		BootVolumeSize: obj.Spec.GCP.BootVolumeSize,
		ServiceAccount: GCPServiceAccount{
			Enabled: obj.Spec.GCP.ServiceAccount.Enabled,
			Email:   obj.Spec.GCP.ServiceAccount.Email,
			Scopes:  obj.Spec.GCP.ServiceAccount.Scopes,
		},
		Nodes: func() map[string]any {
			m := make(map[string]any, len(nodesMap))
			for k, v := range nodesMap {
				m[k] = v
			}
			return m
		}(),
		NodeLabels:               obj.Spec.NodeLabels,
		MachineType:              obj.Spec.GCP.MachineType,
		K3sServerPublicDNSHost:   r.Env.K3sServerPublicHost,
		K3sJoinToken:             r.Env.K3sJoinToken,
		K3sExtraAgentArgs:        []string{"--snapshotter", "stargz"},
		ClusterInternalDNSHost:   "cluster.local",
		SaveSSHKeyToPath:         "",
		KloudliteRelease:         r.Env.KloudliteRelease,
		LabelCloudproviderRegion: r.Env.CloudProviderRegion,
		Labels: map[string]string{
			"kloudlite-account": r.Env.AccountName,
			"nodepool-name":     obj.Name,
			"kloudlite-cluster": r.Env.ClusterName,
		},
		AllowIncomingHttpTraffic: false,
	}

	b, err := json.Marshal(values)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
