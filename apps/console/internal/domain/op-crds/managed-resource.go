package op_crds

type ManagedResourceSpec struct {
	Kind               string            `json:"kind,omitempty"`
	ApiVersion         string            `json:"apiVersion"`
	ManagedServiceName string            `json:"managedSvcName"`
	Inputs             map[string]string `json:"inputs,omitempty"`
}

type ManagedResourceMetadata struct {
	Name        string            `json:"name,omitempty"`
	Namespace   string            `json:"namespace,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

const ManagedResourceAPIVersion = "crds.kloudlite.io/v1"
const ManagedResourceKind = "ManagedResource"

type ManagedResource struct {
	APIVersion string                  `json:"apiVersion,omitempty"`
	Kind       string                  `json:"kind,omitempty"`
	Metadata   ManagedResourceMetadata `json:"metadata"`
	Spec       ManagedResourceSpec     `json:"spec,omitempty"`
}
