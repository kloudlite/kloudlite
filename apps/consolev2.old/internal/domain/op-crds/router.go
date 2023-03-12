package op_crds

type Route struct {
	Path   string `json:"path"`
	App    string `json:"app,omitempty"`
	Lambda string `json:"lambda,omitempty"`
	Port   uint16 `json:"port"`
}

type RouterSpec struct {
	Region  string   `json:"region"`
	Domains []string `json:"domains"`
	Https   struct {
		Enabled       bool `json:"enabled"`
		ForceRedirect bool `json:"forceRedirect"`
	} `json:"https"`
	Routes []Route `json:"routes"`
}

type RouterMetadata struct {
	Name        string            `json:"name,omitempty"`
	Namespace   string            `json:"namespace,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

const RouterAPIVersion = "crds.kloudlite.io/v1"
const RouterKind = "Router"

type Router struct {
	APIVersion string         `json:"apiVersion,omitempty"`
	Kind       string         `json:"kind,omitempty"`
	Metadata   RouterMetadata `json:"metadata"`
	Spec       RouterSpec     `json:"spec,omitempty"`
}
