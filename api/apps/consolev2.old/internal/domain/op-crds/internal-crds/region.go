package internal_crds

type RegionSpec struct {
	Name string `json:"name,omitempty"`
}

type RegionMetadata struct {
	Name        string            `json:"name,omitempty"`
	Namespace   string            `json:"namespace,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

const RegionAPIVersion = "crds.kloudlite.io/v1"
const RegionKind = "Region"

type Region struct {
	APIVersion string `json:"apiVersion,omitempty"`
	Kind       string `json:"kind,omitempty"`

	Metadata RegionMetadata `json:"metadata"`
	Spec     RegionSpec     `json:"spec,omitempty"`
}
