package v1

import (
	common_types "github.com/kloudlite/operator/apis/common-types"
	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AwsSubnet struct {
	AvailabilityZone AwsAZ  `json:"availabilityZone"`
	CIDR             string `json:"cidr"`
}

// AwsVPCSpec defines the desired state of AwsVPC
type AwsVPCSpec struct {
	CredentialsRef common_types.SecretRef      `json:"credentialsRef"`
	CredentialKeys CloudProviderCredentialKeys `json:"credentialKeys" graphql:"noinput"`

	Region AwsRegion `json:"region"`

	CIDR          string      `json:"cidr,omitempty"`
	PublicSubnets []AwsSubnet `json:"publicSubnets,omitempty"`

	Output *common_types.SecretRef `json:"output,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".status.lastReconcileTime",name=Last_Reconciled_At,type=date
// +kubebuilder:printcolumn:JSONPath=".metadata.labels.kloudlite\\.io\\/region",name=AwsRegion,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/resource\\.ready",name=Ready,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/checks",name=Checks,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// AwsVPC is the Schema for the awsvpcs API
type AwsVPC struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AwsVPCSpec  `json:"spec,omitempty"`
	Status rApi.Status `json:"status,omitempty"`
}

func (b *AwsVPC) EnsureGVK() {
	if b != nil {
		b.SetGroupVersionKind(GroupVersion.WithKind("AwsVPC"))
	}
}

func (b *AwsVPC) GetStatus() *rApi.Status {
	return &b.Status
}

func (b *AwsVPC) GetEnsuredLabels() map[string]string {
	return map[string]string{
		constants.RegionKey: string(b.Spec.Region),
	}
}

func (b *AwsVPC) GetEnsuredAnnotations() map[string]string {
	return map[string]string{}
}

//+kubebuilder:object:root=true

// AwsVPCList contains a list of AwsVPC
type AwsVPCList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AwsVPC `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AwsVPC{}, &AwsVPCList{})
}
