package op_crds

type ConfigMetadata struct {
	Name        string            `json:"name,omitempty"`
	Namespace   string            `json:"namespace,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

const ConfigAPIVersion = "v1"
const ConfigKind = "ConfigMap"

type Config struct {
	APIVersion string         `json:"apiVersion,omitempty"`
	Kind       string         `json:"kind,omitempty"`
	Metadata   ConfigMetadata `json:"metadata,omitempty"`
	Data       map[string]any `json:"data,omitempty"`
}
