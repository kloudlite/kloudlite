package common_types

import (
	"fmt"
	"github.com/kloudlite/operator/pkg/errors"
	"k8s.io/apimachinery/pkg/api/resource"
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

// +kubebuilder:object:generate=true

type Resources struct {
	Cpu CpuT `json:"cpu"`
	// +kubebuilder:validation:Pattern=[\d]+Mi$
	Memory string `json:"memory"`

	Storage *Storage `json:"storage,omitempty"`
}

type FsType string

const (
	Ext4 FsType = "ext4"
	Xfs  FsType = "xfs"
)

type CloudProvider struct {
	// +kubebuilder:validation:Enum=do;aws;gcp;azure;k3s-local
	Cloud  string `json:"cloud"`
	Region string `json:"region"`
	// +kubebuilder:validation:Optional
	Account string `json:"account,omitempty"`
}

func (c CloudProvider) GetStorageClass(fsType FsType) (string, error) {
	// return fmt.Sprintf("kl-%s-block-%s-%s", c.Cloud, fsType, c.Region), nil
	switch c.Cloud {
	case "k3s-local":
		return "local-path", nil
	case "do":
		{
			return fmt.Sprintf("kl-%s-block-%s-%s", c.Cloud, fsType, c.Region), nil
		}
	case "azure":
		{
			return fmt.Sprintf("kl-%s-block-%s-%s", c.Cloud, fsType, c.Region), nil
		}
	}
	return "", errors.Newf("no storage class found, unknown pair (provider=%s, fstype=%s)", c, fsType)
}

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
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
}

type SecretRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
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
