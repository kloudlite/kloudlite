package op_crds

type Route struct {
	Path string `json:"path"`
	App  string `json:"app"`
	Port uint16 `json:"port"`
}

type RouterSpec struct {
	Domains []string `json:"domains"`
	Routes  []Route  `json:"routes"`
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
