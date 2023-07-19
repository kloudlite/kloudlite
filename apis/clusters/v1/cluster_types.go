package v1

import (
	common_types "github.com/kloudlite/operator/apis/common-types"
	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type HelmValuesWgOperatorConfBasicAuth struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

type HelmValuesWgOperatorConfNameserver struct {
	Endpoint  string                            `json:"endpoint"`
	BasicAuth HelmValuesWgOperatorConfBasicAuth `json:"basicAuth"`
}

type HelmValuesWgOperatorConf struct {
	Nameserver HelmValuesWgOperatorConfNameserver `json:"nameserver"`

	BaseDomain string  `json:"baseDomain"`
	PodCidr    *string `json:"podCidr,omitempty"`
	SvcCidr    *string `json:"svcCidrf,omitempty"`
}

type HelmValuesOperatorsWgOperator struct {
	Image         *string                  `json:"image,omitempty"`
	Configuration HelmValuesWgOperatorConf `json:"configuration"`
}

type HelmValuesOperatorsResourceWatcher struct {
	Image *string `json:"image,omitempty"`
}

type HelmValuesOperators struct {
	ResourceWatcher *HelmValuesOperatorsResourceWatcher `json:"resourceWatcher,omitempty"`
	WgOperator      HelmValuesOperatorsWgOperator       `json:"wgOperator"`
}

type HelmValuesAgent struct {
	Image *string `json:"image,omitempty"`
}

type HelmValues struct {
	ClusterToken          string  `json:"clusterToken"`
	AccessToken           *string `json:"accessToken,omitempty"`
	MessageOfficeGRPCAddr *string `json:"messageOfficeGRPCAddr,omitempty"`

	Agent     *HelmValuesAgent    `json:"agent,omitempty"`
	Operators HelmValuesOperators `json:"operators"`
}

// ClusterSpec defines the desired state of Cluster
// For now considered basis on AWS Specific
type ClusterSpec struct {
	Region      string `json:"region"`
	AccountName string `json:"accountName"`

	CredentialsRef common_types.SecretRef `json:"credentialsRef"`

	// +kubebuilder:validation:Enum=dev;HA
	AvailablityMode string `json:"availabilityMode"`
	// +kubebuilder:validation:Enum=aws;do;gcp;azure
	CloudProvider string `json:"cloudProvider"`

	NodeIps []string `json:"nodeIps,omitempty"`
	VPC     *string  `json:"vpc,omitempty"`

	HelmValues HelmValues `json:"helmValues"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// Cluster is the Schema for the clusters API
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec `json:"spec,omitempty"`
	Status rApi.Status `json:"status,omitempty"`
}

func (b *Cluster) EnsureGVK() {
	if b != nil {
		b.SetGroupVersionKind(GroupVersion.WithKind("Cluster"))
	}
}

func (b *Cluster) GetStatus() *rApi.Status {
	return &b.Status
}

func (b *Cluster) GetEnsuredLabels() map[string]string {
	return map[string]string{}
}

func (b *Cluster) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		constants.GVKKey: GroupVersion.WithKind("Cluster").String(),
	}
}

//+kubebuilder:object:root=true

// ClusterList contains a list of Cluster
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cluster{}, &ClusterList{})
}
