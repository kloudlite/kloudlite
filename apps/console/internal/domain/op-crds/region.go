package op_crds

type RegionMetadata struct {
	Name        string            `json:"name,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

type RegionSpec struct {
	Name    string  `json:"name"`
	Account *string `json:"account,omitempty"`
}

const RegionAPIVersion = "management.kloudlite.io/v1"
const RegionKind = "Region"

type Region struct {
	APIVersion string         `json:"apiVersion,omitempty"`
	Kind       string         `json:"kind,omitempty"`
	Metadata   RegionMetadata `json:"metadata,omitempty"`
	Spec       RegionSpec     `json:"spec,omitempty"`
}
