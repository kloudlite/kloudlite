package op_crds

type ContainerResource struct {
	Min string `json:"min"`
	Max string `json:"max"`
}

type ContainerEnv struct {
	Key     string `json:"key"`
	Value   string `json:"value,omitempty"`
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

type ImageFromGit struct {
}

type AppContainer struct {
	Name            string            `json:"name"`
	Image           string            `json:"image"`
	ImagePullPolicy string            `json:"imagePullPolicy"`
	Command         []string          `json:"command,omitempty"`
	Args            []string          `json:"args,omitempty"`
	ResourceCpu     ContainerResource `json:"resourceCpu"`
	ResourceMemory  ContainerResource `json:"resourceMemory"`
	Env             []ContainerEnv    `json:"env,omitempty"`
	Volumes         []ContainerVolume `json:"volumes,omitempty"`
}

type AppSvc struct {
	Port       uint16 `json:"port"`
	TargetPort uint16 `json:"targetPort,omitempty"`
	Type       string `json:"type,omitempty"`
}

// AppSpec defines the desired state of App
type AppSpec struct {
	Services   []AppSvc       `json:"services,omitempty"`
	Containers []AppContainer `json:"containers"`
}

type App struct {
	Name      string  `json:"name"`
	NameSpace string  `json:"nameSpace"`
	Spec      AppSpec `json:"spec,omitempty"`
	Status    Status  `json:"status,omitempty"`
}
