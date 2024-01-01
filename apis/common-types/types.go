package common_types

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Storage struct {
	// +kubebuilder:validation:Pattern=[\d]+(M|G)i$
	Size string `json:"size"`

	// +kubebuilder:validation:Optional
	StorageClass string `json:"storageClass,omitempty"`
}

func (s Storage) ToInt() (int64, error) {
	quantity, err := resource.ParseQuantity(s.Size)
	if err != nil {
		fmt.Printf("could not convert storage (%s) to int\n", s.Size)
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

// func (c CloudProvider) GetStorageClass(env *env.Env, fsType FsType, region string) (string, error) { // 	switch c {
// 	case Digitalocean:
// 		{
// 			switch fsType {
// 			case Ext4:
// 				return env.DoBlockStorageExt4, nil
// 			case Xfs:
// 				return env.DoBlockStorageXFS, nil
// 			}
// 		}
// 	case Azure: {
// 		return fmt.Sprintf("kl-%s-block-%s-%s", c, fsType, region), nil
// 	}
// 	default:
// 		return "", errors.NewE(errors.Newf("unknown pair (provider=%s, fstype=%s)", c, fsType))
// 	}
// 	return "", errors.NewE(errors.Newf("unknown pair (provider=%s, fstype=%s)", c, fsType))
// }

type MsvcRef struct {
	metav1.TypeMeta `json:",inline"`
	Name            string `json:"name"`
	Namespace       string `json:"namespace"`
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

const (
	CloudProviderAWS          CloudProvider = "aws"
	CloudProviderDigitalOcean CloudProvider = "do"
	CloudProviderAzure        CloudProvider = "azure"
	CloudProviderGCP          CloudProvider = "gcp"
)
