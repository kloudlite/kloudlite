package common_types

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:validation:Pattern=[\d]+(M|G)i$
type StorageSize string

type Storage struct {
	Size StorageSize `json:"size"`

	// +kubebuilder:validation:Optional
	StorageClass string `json:"storageClass,omitempty"`
}

func (s StorageSize) ToInt() (int64, error) {
	quantity, err := resource.ParseQuantity(string(s))
	if err != nil {
		fmt.Printf("could not convert storage (%s) to int\n", s)
		return -1, nil
	}
	size, ok := quantity.AsInt64()
	if !ok {
		return -1, nil
	}
	return size, nil
}

// +kubebuilder:object:generate=true

type CpuT struct {
	// +kubebuilder:validation:Pattern=[\d]+m$
	Min string `json:"min"`
	// +kubebuilder:validation:Pattern=[\d]+m$
	Max string `json:"max"`
}

type MemoryT struct {
	// +kubebuilder:validation:Pattern=[\d]+Mi$
	Min string `json:"min"`
	// +kubebuilder:validation:Pattern=[\d]+Mi$
	Max string `json:"max"`
}

// +kubebuilder:object:generate=true

type Resources struct {
	Cpu    CpuT    `json:"cpu"`
	Memory MemoryT `json:"memory"`

	Storage *Storage `json:"storage,omitempty"`
}

type FsType string

const (
	Ext4 FsType = "ext4"
	Xfs  FsType = "xfs"
)

// +kubebuilder:object:generate=true
type MsvcRef struct {
	metav1.TypeMeta `json:",inline"`
	Name            string `json:"name"`
	Namespace       string `json:"namespace"`
	// ClusterName     *string `json:"clusterName,omitempty"`
	// SharedSecret    *string `json:"sharedSecret,omitempty" graphql:"ignore"`
}

type SecretRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

type ConfigMapRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type SecretKeyRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
	Key       string `json:"key"`
}

type ConfigRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

// +kubebuilder:object:generate=true

type Output struct {
	SecretRef *SecretRef `json:"secretRef,omitempty"`
	ConfigRef *ConfigRef `json:"configRef,omitempty"`
}

type MinMaxFloat struct {
	// +kubebuilder:validation:Pattern="^[0-9]+([.][0-9]{1,2})?$"
	Min string `json:"min"`
	// +kubebuilder:validation:Pattern="^[0-9]+([.][0-9]{1,2})?$"
	Max string `json:"max"`
}

type MinMaxInt struct {
	// +kubebuilder:validation:Minimum=0
	Min int `json:"min"`
	// +kubebuilder:validation:Minimum=0
	Max int `json:"max"`
}

// +kubebuilder:validation:Enum=aws;do;azure;gcp
type CloudProvider string

func (c CloudProvider) String() string {
	return string(c)
}

const (
	CloudProviderUnknown      CloudProvider = "unknown"
	CloudProviderAWS          CloudProvider = "aws"
	CloudProviderDigitalOcean CloudProvider = "digitalocean"
	CloudProviderAzure        CloudProvider = "azure"
	CloudProviderGCP          CloudProvider = "gcp"
)

// +kubebuilder:object:generate=true
type NodeSelectorAndTolerations struct {
	NodeSelector map[string]string   `json:"nodeSelector,omitempty"`
	Tolerations  []corev1.Toleration `json:"tolerations,omitempty"`
}

type LocalObjectReference struct {
	// .metadata.name of the resource
	Name string `json:"name"`
}

type NamespacedResourceRef struct {
	// .metadata.name of the resource
	Name string `json:"name"`

	// .metadata.namespace of the resource
	Namespace string `json:"namespace"`
}

type ManagedResourceOutput struct {
	// refers to a k8s secret that exists in the same namespace as managed resource
	CredentialsRef LocalObjectReference `json:"credentialsRef"`
}

type ManagedServiceOutput struct {
	// refers to a k8s secret that exists in the same namespace as managed service
	CredentialsRef LocalObjectReference `json:"credentialsRef"`
}
