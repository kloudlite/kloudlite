package domain

type ContainerResource struct {
	Min string `json:"min"`
	Max string `json:"max"`
}

type ContainerEnv struct {
	Key     string `json:"key"`
	Value   string `json:"value,omitemtpy"`
	Type    string `json:"type,omitempty"`
	RefName string `json:"refName,omitempty"`
	RefKey  string `json:"refKey,omitempty"`
}

type ContainerVolumeItem struct {
	Key      string `json:"key"`
	FileName string `json:"fileName"`
}

type ContainerVolume struct {
	Name      string                `json:"name"`
	MountPath string                `json:"mountPath"`
	Type      string                `json:"type"`
	RefName   string                `json:"refName"`
	Items     []ContainerVolumeItem `json:"items"`
}

type AppContainer struct {
	Name            string            `json:"name"`
	Image           string            `json:"image"`
	ImagePullPolicy string            `json:"imagePullPolicy"`
	Command         []string          `json:"command,omitempty"`
	Args            []string          `json:"args,omitempty"`
	ResourceCpu     ContainerResource `json:"resourceCpu"`
	ResourceMemory  ContainerResource `json:"resourceMemory"`
	Env             []ContainerEnv    `json:"env,omitemtpy"`
	Volumes         []ContainerVolume `json:"volumes,omitempty"`
}

type AppSvc struct {
	Port       uint16 `json:"port"`
	TargetPort uint16 `json:"targetPort"`
	Type       string `json:"type"`
}

type App struct {
	Name       string         `json:"name"`
	Namespace  string         `json:"namespace"`
	Services   []AppSvc       `json:"services"`
	Containers []AppContainer `json:"containers"`
}
