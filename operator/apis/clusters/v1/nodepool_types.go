package v1

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ct "github.com/kloudlite/operator/apis/common-types"
	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
)

// +kubebuilder:validation:Enum=ec2;spot;
type AWSPoolType string

const (
	AWSPoolTypeEC2  AWSPoolType = "ec2"
	AWSPoolTypeSpot AWSPoolType = "spot"
)

type AwsEC2PoolConfig struct {
	InstanceType string               `json:"instanceType"`
	Nodes        map[string]NodeProps `json:"nodes,omitempty"`
}

type AwsSpotPoolConfig struct {
	SpotFleetTaggingRoleName string               `json:"spotFleetTaggingRoleName" graphql:"noinput"`
	CpuNode                  *AwsSpotCpuNode      `json:"cpuNode,omitempty"`
	GpuNode                  *AwsSpotGpuNode      `json:"gpuNode,omitempty"`
	Nodes                    map[string]NodeProps `json:"nodes,omitempty"`
}

type AWSNodePoolConfig struct {
	ImageId          string `json:"imageId" graphql:"noinput"`
	ImageSSHUsername string `json:"imageSSHUsername" graphql:"noinput"`
	AvailabilityZone string `json:"availabilityZone"`

	NvidiaGpuEnabled bool   `json:"nvidiaGpuEnabled"`
	RootVolumeType   string `json:"rootVolumeType" graphql:"noinput"`
	RootVolumeSize   int    `json:"rootVolumeSize" graphql:"noinput"`

	IAMInstanceProfileRole *string `json:"iamInstanceProfileRole,omitempty" graphql:"noinput"`

	PoolType AWSPoolType `json:"poolType"`

	EC2Pool  *AwsEC2PoolConfig  `json:"ec2Pool,omitempty"`
	SpotPool *AwsSpotPoolConfig `json:"spotPool,omitempty"`
}

type InfrastuctureAsCode struct {
	StateS3BucketName     string `json:"stateS3BucketName"`
	StateS3BucketRegion   string `json:"stateS3BucketRegion"`
	StateS3BucketFilePath string `json:"stateS3BucketFilePath"`

	CloudProviderAccessKey ct.SecretKeyRef `json:"cloudProviderAccessKey"`
	CloudProviderSecretKey ct.SecretKeyRef `json:"cloudProviderSecretKey"`

	JobName      string `json:"jobName,omitempty"`
	JobNamespace string `json:"jobNamespace,omitempty"`
}

type OperatorVars struct {
	JobName      string `json:"jobName"`
	JobNamespace string `json:"jobNamespace"`
}

type NodePoolSpec struct {
	// +kubebuilder:validation:Minimum=0
	MaxCount int `json:"maxCount"`
	// +kubebuilder:validation:Minimum=0
	MinCount int `json:"minCount"`
	// +kubebuilder:validation:Minimum=0
	TargetCount int `json:"targetCount"`

	IAC InfrastuctureAsCode `json:"iac" graphql:"noinput"`

	CloudProvider ct.CloudProvider   `json:"cloudProvider"`
	AWS           *AWSNodePoolConfig `json:"aws,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.nodepool-min-target-max",name=Min/Target/Max,type=string
// +kubebuilder:printcolumn:JSONPath=".status.lastReconcileTime",name=Last_Reconciled_At,type=date
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/resource\\.ready",name=Ready,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// NodePool is the Schema for the nodepools API
type NodePool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodePoolSpec `json:"spec"`
	Status rApi.Status  `json:"status,omitempty" graphql:"noinput"`
}

func (n *NodePool) EnsureGVK() {
	if n != nil {
		n.SetGroupVersionKind(GroupVersion.WithKind("NodePool"))
	}
}

func (n *NodePool) GetStatus() *rApi.Status {
	return &n.Status
}

func (n *NodePool) GetEnsuredLabels() map[string]string {
	return map[string]string{
		constants.NodePoolNameKey: n.Name,
	}
}

func (n *NodePool) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		constants.GVKKey:          GroupVersion.WithKind("NodePool").String(),
		"nodepool-min-target-max": fmt.Sprintf("%d/%d/%d", n.Spec.MinCount, n.Spec.TargetCount, n.Spec.MaxCount),
	}
}

//+kubebuilder:object:root=true

// NodePoolList contains a list of NodePool
type NodePoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodePool `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NodePool{}, &NodePoolList{})
}
