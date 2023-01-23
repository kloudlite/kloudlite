package op_crds

type SecretMetadata struct {
	Name        string            `json:"name,omitempty"`
	Namespace   string            `json:"namespace,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

const SecretAPIVersion = "crds.kloudlite.io/v1"
const SecretKind = "Secret"

type Secret struct {
	APIVersion string            `json:"apiVersion,omitempty"`
	Kind       string            `json:"kind,omitempty"`
	Metadata   SecretMetadata    `json:"metadata,omitempty"`
	Data       map[string][]byte `json:"data,omitempty"`
	StringData map[string]any    `json:"stringData,omitempty"`
}
