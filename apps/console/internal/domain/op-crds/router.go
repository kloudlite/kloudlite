package op_crds

type Route struct {
	Path string `json:"path"`
	App  string `json:"app"`
	Port uint16 `json:"port"`
}

type RouterSpec struct {
	Domains     []string           `json:"domains"`
	Routes      map[string][]Route `json:"routes"`
	Annotations map[string]string  `json:"annotations,omitempty"`
	Labels      map[string]string  `json:"labels,omitempty"`
}

type RouterMetadata struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

const RouterAPIVersion = "crds.kloudlite.io/v1"
const RouterKind = "Router"

type Router struct {
	APIVersion string         `json:"apiVersion,omitempty"`
	Kind       string         `json:"kind,omitempty"`
	Metadata   RouterMetadata `json:"metadata"`
	Spec       RouterSpec     `json:"spec,omitempty"`
	Status     Status         `json:"status,omitempty"`
}
