package types

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"
)

type CPUResource struct {
	// +kubebuilder:validation:Pattern=[\d]+m$
	Min string `json:"min"`
	// +kubebuilder:validation:Pattern=[\d]+m$
	Max string `json:"max"`
}

type MemoryResource struct {
	// +kubebuilder:validation:Pattern=[\d]+Mi$
	Min string `json:"min"`
	// +kubebuilder:validation:Pattern=[\d]+Mi$
	Max string `json:"max"`
}

// +kubebuilder:object:generate=true
type Resource struct {
	Cpu    *CPUResource    `json:"cpu,omitempty"`
	Memory *MemoryResource `json:"memory,omitempty"`
}

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

type ResourceWithStorage struct {
	Cpu     *CPUResource    `json:"cpu,omitempty"`
	Memory  *MemoryResource `json:"memory,omitempty"`
	Storage *Storage        `json:"storage,omitempty"`
}
