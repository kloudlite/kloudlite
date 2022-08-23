package v1

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"operators.kloudlite.io/lib/constants"
	fn "operators.kloudlite.io/lib/functions"
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
	RefName string `json:"refName,omitempty"`
	RefKey  string `json:"refKey,omitempty"`
}

type ContainerVolumeItem struct {
	Key      string `json:"key"`
	FileName string `json:"fileName,omitempty"`
}

type EnvFrom struct {
	// must be one of config, secret
	Type string `json:"type"`

	RefName string `json:"refName"`
}

type ContainerVolume struct {
	MountPath string                `json:"mountPath"`
	Type      ResourceType          `json:"type"`
	RefName   string                `json:"refName"`
	Items     []ContainerVolumeItem `json:"items,omitempty"`
	// SubPath   string                `json:"subPath,omitempty"`
}

type ShellProbe struct {
	Command []string `json:"command,omitempty"`
}

type HttpGetProbe struct {
	Path        string            `json:"path"`
	Port        uint              `json:"port"`
	HttpHeaders map[string]string `json:"httpHeaders,omitempty"`
}

type TcpProbe struct {
	Port uint16 `json:"port"`
}

type Probe struct {
	// +kubebuilder:validation:Enum=shell;httpGet;tcp
	Type string `json:"type"`
	// +kubebuilder:validation:Optional
	Shell *ShellProbe `json:"shell,omitempty"`
	// +kubebuilder:validation:Optional
	HttpGet *HttpGetProbe `json:"httpGet,omitempty"`
	// +kubebuilder:validation:Optional
	Tcp *TcpProbe `json:"tcp,omitempty"`

	FailureThreshold uint `json:"failureThreshold,omitempty"`
	InitialDelay     uint `json:"initialDelay,omitempty"`
	Interval         uint `json:"interval,omitempty"`
}

type AppContainer struct {
	Name  string `json:"name"`
	Image string `json:"image"`
	// +kubebuilder:default=IfNotPresent
	ImagePullPolicy string            `json:"imagePullPolicy,omitempty"`
	Command         []string          `json:"command,omitempty"`
	Args            []string          `json:"args,omitempty"`
	ResourceCpu     ContainerResource `json:"resourceCpu,omitempty"`
	ResourceMemory  ContainerResource `json:"resourceMemory,omitempty"`
	Env             []ContainerEnv    `json:"env,omitempty"`
	EnvFrom         []EnvFrom         `json:"envFrom,omitempty"`
	Volumes         []ContainerVolume `json:"volumes,omitempty"`
	LivenessProbe   *Probe            `json:"livenessProbe,omitempty"`
	ReadinessProbe  *Probe            `json:"readinessProbe,omitempty"`
}

type AppSvc struct {
	Port       uint16 `json:"port"`
	TargetPort uint16 `json:"targetPort,omitempty"`
	Type       string `json:"type,omitempty"`
	Name       string `json:"name,omitempty"`
}

type HPA struct {
	Enabled bool `json:"enabled,omitempty"`
	// +kubebuilder:default=1
	MinReplicas int `json:"minReplicas,omitempty"`
	// +kubebuilder:default=5
	MaxReplicas int `json:"maxReplicas,omitempty"`
	// +kubebuilder:default=90
	ThresholdCpu int `json:"thresholdCpu,omitempty"`
	// +kubebuilder:default=75
	ThresholdMemory int `json:"thresholdMemory,omitempty"`
}

// AppSpec defines the desired state of App
type AppSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=kloudlite-svc-account
	ServiceAccount string `json:"serviceAccount,omitempty"`
	// +kubebuilder:default=1
	// +kubebuilder:validation:Optional
	Replicas   int             `json:"replicas"`
	Services   []AppSvc        `json:"services,omitempty"`
	Containers []AppContainer  `json:"containers"`
	Volumes    []corev1.Volume `json:"volumes,omitempty"`
	Hpa        HPA             `json:"hpa,omitempty"`
	// +kubebuilder:validation:Optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
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
	m := map[string]string{
		"kloudlite.io/app.name": app.Name,
	}

	for idx := range app.Spec.Containers {
		m[fmt.Sprintf("kloudlite.io/image-%s", fn.Sha1Sum([]byte(app.Spec.Containers[idx].Image)))] = "true"
	}

	return m
}

func (app *App) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		constants.AnnotationKeys.GroupVersionKind: GroupVersion.WithKind("App").String(),
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
