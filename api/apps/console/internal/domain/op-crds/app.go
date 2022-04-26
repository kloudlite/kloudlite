package op_crds

type Service struct {
	Port       int    `json:"port,omitempty"`
	TargetPort int    `json:"target_port,omitempty"`
	Type       string `json:"type,omitempty"`
}

type Limit struct {
	Min string `json:"min,omitempty"`
	Max string `json:"max,omitempty"`
}

type EnvEntry struct {
	Key     string  `json:"key,omitempty"`
	Value   *string `json:"value,omitempty"`
	Type    string  `json:"type,omitempty"`
	RefName *string `json:"ref_name,omitempty"`
	RefKey  *string `json:"ref_key,omitempty"`
}

type Container struct {
	Name           string     `json:"name,omitempty"`
	Image          *string    `json:"image,omitempty"`
	ResourceCpu    Limit      `json:"resource_cpu,omitempty"`
	ResourceMemory Limit      `json:"resource_memory,omitempty"`
	Env            []EnvEntry `json:"env,omitempty"`
}

type AppSpec struct {
	Services   []Service   `json:"services,omitempty"`
	Containers []Container `json:"containers,omitempty"`
	Replicas   int         `json:"replicas,omitempty"`
}

type AppMetadata struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

const AppAPIVersion = "crds.kloudlite.io/v1"
const AppKind = "App"

type App struct {
	APIVersion string      `json:"apiVersion,omitempty"`
	Kind       string      `json:"kind,omitempty"`
	Metadata   AppMetadata `json:"metadata"`
	Spec       AppSpec     `json:"spec,omitempty"`
	Status     Status      `json:"status,omitempty"`
}
