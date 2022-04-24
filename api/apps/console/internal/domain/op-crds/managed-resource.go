package op_crds

type ManagedResourceSpec struct {
	Type           string            `json:"type"`
	ManagedService string            `json:"managedSvc"`
	Inputs         map[string]string `json:"inputs,omitempty"`
}

type ManagedResourceMetadata struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

const ManagedResourceAPIVersion = "crds.kloudlite.io/v1"
const ManagedResourceKind = "ManagedResource"

type ManagedResource struct {
	APIVersion string                  `json:"apiVersion,omitempty"`
	Kind       string                  `json:"kind,omitempty"`
	Metadata   ManagedResourceMetadata `json:"metadata"`
	Spec       ManagedResourceSpec     `json:"spec,omitempty"`
	Status     Status                  `json:"status,omitempty"`
}
