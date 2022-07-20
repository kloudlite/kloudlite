package op_crds

type Service struct {
	Port       int    `json:"port,omitempty"`
	TargetPort int    `json:"targetPort,omitempty"`
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
	RefName *string `json:"refName,omitempty"`
	RefKey  *string `json:"refKey,omitempty"`
}

type Container struct {
	Name           string     `json:"name,omitempty"`
	Image          *string    `json:"image,omitempty"`
	ResourceCpu    *Limit     `json:"resourceCpu,omitempty"`
	ResourceMemory *Limit     `json:"resourceMemory,omitempty"`
	Env            []EnvEntry `json:"env,omitempty"`
}

type HPA struct {
	MinReplicas     int `json:"minReplicas,omitempty"`
	MaxReplicas     int `json:"maxReplicas,omitempty"`
	ThresholdCpu    int `json:"thresholdCpu,omitempty"`
	ThresholdMemory int `json:"thresholdMemory,omitempty"`
}

type AppSpec struct {
	Services     []Service         `json:"services,omitempty"`
	Containers   []Container       `json:"containers,omitempty"`
	Replicas     int               `json:"replicas,omitempty"`
	Hpa          *HPA              `json:"hpa,omitempty"`
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}

type AppMetadata struct {
	Name        string            `json:"name,omitempty"`
	Namespace   string            `json:"namespace,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

const AppAPIVersion = "crds.kloudlite.io/v1"
const AppKind = "App"

type App struct {
	APIVersion string `json:"apiVersion,omitempty"`
	Kind       string `json:"kind,omitempty"`

	Metadata AppMetadata `json:"metadata"`
	Spec     AppSpec     `json:"spec,omitempty"`
	Status   *Status     `json:"status,omitempty"`
}
