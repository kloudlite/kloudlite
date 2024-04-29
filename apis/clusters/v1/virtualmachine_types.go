package v1

import (
	"context"

	ct "github.com/kloudlite/operator/apis/common-types"
	fn "github.com/kloudlite/operator/pkg/functions"
	rApi "github.com/kloudlite/operator/pkg/operator"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

// +kubebuilder:validation:Enum=on;off;
type MachineState string

const (
	MachineStateOn  = "on"
	MachineStateOff = "off"
)

type GCPVirtualMachineConfig struct {
	Region       string `json:"region"`
	GCPProjectID string `json:"gcpProjectID" graphql:"noinput"`

	VPC *GcpVPCParams `json:"vpc,omitempty" graphql:"noinput"`

	ServiceAccount GCPServiceAccount `json:"serviceAccount"`

	// This secret will be unmarshalled into type GCPCredentials
	CredentialsRef ct.SecretRef `json:"credentialsRef"`

	AvailabilityZone string `json:"availabilityZone,omitempty"`

	MachineType string `json:"machineType"`

	PoolType GCPPoolType `json:"poolType"`

	StartupScript string `json:"startupScript"`

	AllowIncomingHttpTraffic bool `json:"allowIncomingHttpTraffic"`
	AllowSSH                 bool `json:"allowSSH"`

	BootVolumeSize int `json:"bootVolumeSize"`
}

func (g GCPVirtualMachineConfig) RetrieveCreds(ctx context.Context, kcli client.Client) (*GCPCredentials, error) {
	creds, err := rApi.Get(ctx, kcli, fn.NN(g.CredentialsRef.Namespace, g.CredentialsRef.Name), &corev1.Secret{})
	if err != nil {
		return nil, err
	}

	return fn.ParseFromSecret[GCPCredentials](creds)
}

type AwsVMVpcParams struct {
	VPCId       string `json:"vpcId"`
	VPCSubnetID string `json:"VPCSubnetID"`
}

type AWSVirtualMachineConfig struct {
	VPC *AwsVMVpcParams `json:"vpc,omitempty" graphql:"noinput"`

	// AvailabilityZone AwsAZ `json:"availabilityZone"`
	AvailabilityZone string `json:"availabilityZone"`

	NvidiaGpuEnabled bool   `json:"nvidiaGpuEnabled"`
	RootVolumeType   string `json:"rootVolumeType" graphql:"noinput"`
	RootVolumeSize   int    `json:"rootVolumeSize" graphql:"noinput"`

	IAMInstanceProfileRole *string `json:"iamInstanceProfileRole,omitempty" graphql:"noinput"`

	InstanceType string `json:"instanceType"`

	AllowIncomingHttpTraffic bool `json:"allowIncomingHttpTraffic"`
	AllowSSH                 bool `json:"allowSSH"`
}

type VirtualMachineControllerParams struct {
	TFWorkspaceName      string `json:"tfWorkspaceName"`
	TFWorkspaceNamespace string `json:"tfWorkspaceNamespace"`

	JobRef ct.NamespacedResourceRef `json:"jobRef"`
}

// VirtualMachineSpec defines the desired state of VirtualMachine
type VirtualMachineSpec struct {
	KloudliteAccount string `json:"kloudliteAccount"`

	CloudProvider ct.CloudProvider         `json:"cloudProvider"`
	AWS           *AWSVirtualMachineConfig `json:"aws,omitempty"`
	GCP           *GCPVirtualMachineConfig `json:"gcp,omitempty"`

	MachineState MachineState `json:"machineState"`

	ControllerParams VirtualMachineControllerParams `json:"controllerParams,omitempty" graphql:"noinput"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".status.lastReconcileTime",name=Seen,type=date
// +kubebuilder:printcolumn:JSONPath=".spec.machineState",name="Machine_State",type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/checks",name=Checks,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/resource\\.ready",name=Ready,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date
// VirtualMachine is the Schema for the virtualmachines API
type VirtualMachine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VirtualMachineSpec `json:"spec,omitempty"`
	Status rApi.Status        `json:"status,omitempty"`
}

func (vm *VirtualMachine) EnsureGVK() {
	if vm != nil {
		vm.SetGroupVersionKind(GroupVersion.WithKind("VirtualMachine"))
	}
}

func (vm *VirtualMachine) GetStatus() *rApi.Status {
	return &vm.Status
}

func (vm *VirtualMachine) GetEnsuredLabels() map[string]string {
	return map[string]string{}
}

func (vm *VirtualMachine) GetEnsuredAnnotations() map[string]string {
	return map[string]string{}
}

//+kubebuilder:object:root=true

// VirtualMachineList contains a list of VirtualMachine
type VirtualMachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VirtualMachine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VirtualMachine{}, &VirtualMachineList{})
}
