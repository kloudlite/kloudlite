package v1

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rApi "operators.kloudlite.io/lib/operator"
)

type ContainerResource struct {
	Min string `json:"min,omitempty"`
	Max string `json:"max,omitempty"`
}

type ContainerEnv struct {
	Key     string `json:"key"`
	Value   string `json:"value,omitempty"`
	Type    string `json:"type,omitempty"`
	RefName string `json:"ref_name,omitempty"`
	RefKey  string `json:"ref_key,omitempty"`
}

type ContainerVolumeItem struct {
	Key      string `json:"key"`
	FileName string `json:"fileName"`
}

type EnvFrom struct {
	Config string `json:"config,omitempty"`
	Secret string `json:"secret,omitempty"`
}

type ContainerVolume struct {
	MountPath string                `json:"mountPath"`
	Type      ResourceType          `json:"type"`
	RefName   string                `json:"refName"`
	Items     []ContainerVolumeItem `json:"items,omitempty"`
}

type AppContainer struct {
	Name            string            `json:"name"`
	Image           string            `json:"image"`
	ImagePullPolicy string            `json:"imagePullPolicy,omitempty"`
	Command         []string          `json:"command,omitempty"`
	Args            []string          `json:"args,omitempty"`
	ResourceCpu     ContainerResource `json:"resource_cpu,omitempty"`
	ResourceMemory  ContainerResource `json:"resource_memory,omitempty"`
	Env             []ContainerEnv    `json:"env,omitempty"`
	EnvFrom         []EnvFrom         `json:"envFrom,omitempty"`
	Volumes         []ContainerVolume `json:"volumes,omitempty"`
}

type AppSvc struct {
	Port       uint16 `json:"port"`
	TargetPort uint16 `json:"target_port,omitempty"`
	Type       string `json:"type,omitempty"`
}

type HPA struct {
	MinReplicas     int `json:"minReplicas,omitempty"`
	MaxReplicas     int `json:"maxReplicas,omitempty"`
	ThresholdCpu    int `json:"thresholdCpu,omitempty"`
	ThresholdMemory int `json:"thresholdMemory,omitempty"`
}

// AppSpec defines the desired state of App
type AppSpec struct {
	Replicas   int             `json:"replicas,omitempty"`
	Services   []AppSvc        `json:"services,omitempty"`
	Containers []AppContainer  `json:"containers"`
	Volumes    []corev1.Volume `json:"volumes,omitempty"`
	Hpa        HPA             `json:"hpa,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// App is the Schema for the apps API
type App struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AppSpec     `json:"spec,omitempty"`
	Status rApi.Status `json:"status,omitempty"`
}

func (app *App) GetStatus() *rApi.Status {
	return &app.Status
}

func (app *App) GetEnsuredLabels() map[string]string {
	return map[string]string{
		fmt.Sprintf("%s/ref", GroupVersion.Group): app.Name,
	}
}

func (app *App) LogRef() string {
	return fmt.Sprintf("%s/%s/%s", app.Namespace, "App", app.Name)
}

// +kubebuilder:object:root=true

// AppList contains a list of App
type AppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []App `json:"items"`
}

func init() {
	SchemeBuilder.Register(&App{}, &AppList{})
}
