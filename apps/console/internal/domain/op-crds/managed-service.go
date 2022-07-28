package op_crds

type MsvcKind struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
}

type ManagedServiceSpec struct {
	CloudProvider string            `json:"cloudProvider,omitempty"`
	MsvcKind      MsvcKind          `json:"msvcKind,omitempty"`
	Inputs        map[string]any    `json:"inputs,omitempty"`
	NodeSelector  map[string]string `json:"nodeSelector,omitempty"`
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
}
