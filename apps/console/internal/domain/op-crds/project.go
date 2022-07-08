package op_crds

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type ArtifactRegistry struct {
	Enabled bool `json:"enabled"`
}

type ProjectSpec struct {
	DisplayName      string           `json:"displayName,omitempty"`
	ArtifactRegistry ArtifactRegistry `json:"artifactRegistry,omitempty"`
}

type Status struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

type ProjectMetadata struct {
	Name        string            `json:"name,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

const APIVersion = "crds.kloudlite.io/v1"
const ProjectKind = "Project"

type Project struct {
	APIVersion string          `json:"apiVersion,omitempty"`
	Kind       string          `json:"kind,omitempty"`
	Metadata   ProjectMetadata `json:"metadata,omitempty"`
	Spec       ProjectSpec     `json:"spec,omitempty"`
	Status     *Status         `json:"status,omitempty"`
}
