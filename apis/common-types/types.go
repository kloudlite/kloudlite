package common_types

import (
	"operators.kloudlite.io/env"
	"operators.kloudlite.io/lib/errors"
	"strconv"
	"strings"
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

// CloudProvider +kubebuilder:validation:Enum=do;aws;gcp;azure;k3s-local
type CloudProvider string

const (
	Digitalocean CloudProvider = "do"
	Aws          CloudProvider = "aws"
	Gcp          CloudProvider = "gcp"
	Azure        CloudProvider = "azure"
	K3sLocal     CloudProvider = "k3s-local"
)

type FsType string

const (
	Ext4 FsType = "ext4"
	Xfs  FsType = "xfs"
)

func (c CloudProvider) GetStorageClass(env *env.Env, fsType FsType) (string, error) {
	switch c {
	case Digitalocean:
		{
			switch fsType {
			case Ext4:
				return env.DoBlockStorageExt4, nil
			case Xfs:
				return env.DoBlockStorageXFS, nil
			}
		}
	default:
		return "", errors.NewE(errors.Newf("unknown pair (provider=%s, fstype=%s)", c, fsType))
	}
	return "", errors.NewE(errors.Newf("unknown pair (provider=%s, fstype=%s)", c, fsType))
}
