package op_crds

type ManagedServiceSpec struct {
	Type   string            `json:"type"`
	Inputs map[string]string `json:"inputs,omitempty"`
}

type ManagedServiceMetadata struct {
	Name        string            `json:"name,omitempty"`
	Namespace   string            `json:"namespace,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

const ManagedServiceAPIVersion = "crds.kloudlite.io/v1"
const ManagedServiceKind = "ManagedService"

type ManagedService struct {
	APIVersion string                 `json:"apiVersion,omitempty"`
	Kind       string                 `json:"kind,omitempty"`
	Metadata   ManagedServiceMetadata `json:"metadata"`
	Spec       ManagedServiceSpec     `json:"spec,omitempty"`
	Status     Status                 `json:"status,omitempty"`
}
