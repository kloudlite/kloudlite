package op_crds

import (
	createjsonpatch "github.com/snorwin/jsonpatch"
)

type ResourceSpec struct {
	Region       string            `json:"region,omitempty"`
	Services     []Service         `json:"services,omitempty"`
	Containers   []Container       `json:"containers,omitempty"`
	Replicas     int               `json:"replicas,omitempty"`
	Hpa          *HPA              `json:"hpa,omitempty"`
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}

type ResourceMetadata struct {
	Name        string            `json:"name,omitempty"`
	Namespace   string            `json:"namespace,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

type Overrides struct {
	Patches []createjsonpatch.JSONPatch `json:"patches,omitempty"`
}

type Resource struct {
	APIVersion string `json:"apiVersion,omitempty"`
	Kind       string `json:"kind,omitempty"`

	Metadata  ResourceMetadata `json:"metadata"`
	Spec      *ResourceSpec    `json:"spec,omitempty"`
	Overrides *Overrides       `json:"overrides,omitempty"`
	Enabled   *bool            `json:"enabled,omitempty"`
}
