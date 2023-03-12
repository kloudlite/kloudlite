package op_crds

type ProjectSpec struct {
	DisplayName string `json:"displayName,omitempty"`
	AccountRef  string `json:"accountRef,omitempty"`
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
}
