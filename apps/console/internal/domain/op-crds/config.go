package op_crds

type ConfigMetadata struct {
	Name        string            `json:"name,omitempty"`
	Namespace   string            `json:"namespace,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

const ConfigAPIVersion = "crds.kloudlite.io/v1"
const ConfigKind = "Config"

type Config struct {
	APIVersion string            `json:"apiVersion,omitempty"`
	Kind       string            `json:"kind,omitempty"`
	Metadata   ConfigMetadata    `json:"metadata,omitempty"`
	Data       map[string]string `json:"data,omitempty"`
}
