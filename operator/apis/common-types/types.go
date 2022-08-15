package common_types

import (
	"fmt"
	"strconv"
	"strings"

	"operators.kloudlite.io/lib/errors"
)

type Storage struct {
	// +kubebuilder:default="5Gi"
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern=[\d]+Gi$
	Size string `json:"size"`

	// +kubebuilder:validation:Optional
	StorageClass string `json:"storageClass,omitempty"`
}

func (s Storage) ToInt() int {
	sp := strings.Split(s.Size, "Gi")
	sGb, _ := strconv.ParseInt(sp[0], 0, 32)
	return int(sGb)
}

type cpuTT struct {
	// +kubebuilder:validation:Pattern=[\d]+m$
	Min string `json:"min"`
	// +kubebuilder:validation:Pattern=[\d]+m$
	Max string `json:"max"`
}

type Resources struct {
	Cpu cpuTT `json:"cpu"`
	// +kubebuilder:validation:Pattern=[\d]+Mi$
	Memory string `json:"memory"`
}

type FsType string

const (
	Ext4 FsType = "ext4"
	Xfs  FsType = "xfs"
)

// const (
// 	Digitalocean NodeProvider = "do"
// 	Aws          NodeProvider = "aws"
// 	Gcp          NodeProvider = "gcp"
// 	Azure        NodeProvider = "azure"
// 	K3sLocal     NodeProvider = "k3s-local"
// )

type NodeProvider struct {
	// +kubebuilder:validation:Enum=do;aws;gcp;azure;k3s-local
	Cloud  string `json:"cloud"`
	Region string `json:"region"`
	// +kubebuilder:validation:Optional
	Account string `json:"account,omitempty"`
}

func (c NodeProvider) GetStorageClass(fsType FsType) (string, error) {
	// return fmt.Sprintf("kl-%s-block-%s-%s", c.Cloud, fsType, c.Region), nil
	switch c.Cloud {
	case "do":
		{
			return fmt.Sprintf("kl-%s-block-%s-%s", c.Cloud, fsType, c.Region), nil
		}
	case "azure":
		{
			return fmt.Sprintf("kl-%s-block-%s-%s", c.Cloud, fsType, c.Region), nil
		}
	}
	return "", errors.Newf("unknown pair (provider=%s, fstype=%s)", c, fsType)
}

// func (c NodeProvider) GetStorageClass(env *env.Env, fsType FsType, region string) (string, error) {
// 	switch c {
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
