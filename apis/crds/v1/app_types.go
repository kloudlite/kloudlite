package v1

import (
	"fmt"

	"github.com/kloudlite/operator/pkg/constants"
	jsonPatch "github.com/kloudlite/operator/pkg/json-patch"
	rApi "github.com/kloudlite/operator/pkg/operator"
	"github.com/kloudlite/operator/pkg/templates"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ContainerResource struct {
	Min string `json:"min,omitempty"`
	Max string `json:"max,omitempty"`
}

type ContainerEnv struct {
	Key      string         `json:"key"`
	Value    string         `json:"value,omitempty"`
	Type     ConfigOrSecret `json:"type,omitempty"`
	RefName  string         `json:"refName,omitempty"`
	RefKey   string         `json:"refKey,omitempty"`
	Optional *bool          `json:"optional,omitempty"`
}

type ContainerVolumeItem struct {
	Key      string `json:"key"`
	FileName string `json:"fileName,omitempty"`
}

type EnvFrom struct {
	Type    ConfigOrSecret `json:"type"`
	RefName string         `json:"refName"`
}

type ContainerVolume struct {
	MountPath string                `json:"mountPath"`
	Type      ConfigOrSecret        `json:"type"`
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

func (ac AppContainer) ToYAML() []byte {
	b, err := templates.ParseBytes([]byte(`
- name: {{.Name}}
  image: {{.Image}}
  imagePullPolicy: {{.ImagePullPolicy}}

  {{- if .Command }}
  command: {{.Command | toYAML | nindent 4 }}
  {{- end}}

  {{- if .Args }}
  args: {{.Args | toYAML | nindent 4}}
  {{- end }}

  {{- if .EnvFrom }}
  envFrom:
  {{- range .EnvFrom }}
    {{call .ToYAML }}
  {{- end }}
  {{- end }}

  {{- if .Env }}
  env:
    {{- range .Env }}
    {{call .ToYAML}}
    {{- end }}
  {{- end }}

  {{- if or .ResourceCpu .ResourceMemory }}
  resources:
  {{- if and .ResourceCpu.Min .ResourceMemory.Min }}
    requests:
      cpu: {{ .ResourceCpu.Min }}
      memory: {{ .ResourceMemory.Min }}
  {{- end }}
  {{- if and .ResourceCpu.Max .ResourceMemory.Max }}
    limits:
      cpu: {{ .ResourceCpu.Max }}
      memory: {{ .ResourceMemory.Max }}
  {{- end }}
  {{- end }}

  {{- if $volumeMounts }}
  {{- $vMounts := index $volumeMounts $idx }}
  {{- if $vMounts }}
  volumeMounts: {{- $vMounts | toYAML | nindent 4 }}
  {{- end}}
  {{- end }}

  {{- if .LivenessProbe }}
  {{- with .LivenessProbe}}
  livenessProbe:
    failureThreshold: {{.FailureThreshold | default 3}}
    initialDelaySeconds: {{.InitialDelay | default 2}}
    periodSeconds: {{.Interval | default 10 }}

    {{- if eq .Type "shell"}}
    exec:
      command: {{ .Shell | toYAML | nindent 8 }}
    {{- end }}

    {{- if eq .Type "httpGet"}}
    httpGet: {{.HttpGet | toYAML | nindent 6}}
    {{- end }}

    {{- if eq .Type "httpHeaders"}}
    tcpProbe: {{.Tcp | toYAML | nindent 6}}
    {{- end}}
  {{- end }}
  {{- end}}

  {{- if .ReadinessProbe }}
  {{- with .ReadinessProbe}}
  readinessProbe:
    failureThreshold: {{.FailureThreshold | default 3}}
    initialDelaySeconds: {{.InitialDelay | default 2}}
    periodSeconds: {{.Interval | default 10 }}

    {{- if eq .Type "shell"}}
    exec:
      command: {{ .Shell | toYAML | nindent 8 }}
    {{- end }}

    {{- if eq .Type "httpGet"}}
    httpGet: {{.HttpGet | toYAML | nindent 6}}
    {{- end }}

    {{- if eq .Type "httpHeaders"}}
    tcpProbe: {{.Tcp | toYAML | nindent 6}}
    {{- end}}
  {{- end }}
  {{- end}}
`), ac)
	if err != nil {
		return nil
	}
	return b
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
	DisplayName string `json:"displayName,omitempty"`

	Region string `json:"region,omitempty"`

	Intercept *Intercept `json:"intercept,omitempty"`
	Freeze    bool       `json:"freeze,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=kloudlite-svc-account
	ServiceAccount string `json:"serviceAccount,omitempty"`
	// +kubebuilder:default=1
	Replicas   int            `json:"replicas,omitempty"`
	Services   []AppSvc       `json:"services,omitempty"`
	Containers []AppContainer `json:"containers"`
	// +kubebuilder:validation:Optional
	// Volumes []corev1.Volume `json:"volumes,omitempty"`

	Hpa *HPA `json:"hpa,omitempty"`

	// +kubebuilder:validation:Optional
	NodeSelector map[string]string   `json:"nodeSelector,omitempty"`
	Tolerations  []corev1.Toleration `json:"tolerations,omitempty"`
}

type Intercept struct {
	Enabled bool `json:"enabled"`
	// +kubebuilder:validation:MinLength=1
	ToDevice string `json:"toDevice"`
}

type JsonPatch struct {
	Applied bool                       `json:"applied,omitempty"`
	Patches []jsonPatch.PatchOperation `json:"patches,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".status.lastReconcileTime",name=Last_Reconciled_At,type=date
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/resource\\.ready",name=Ready,type=string
// +kubebuilder:printcolumn:JSONPath=".status.displayVars.intercepted",name=Intercepted,type=string
// +kubebuilder:printcolumn:JSONPath=".status.displayVars.frozen",name=Frozen,type=boolean
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// App is the Schema for the apps API
type App struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec AppSpec `json:"spec"`
	// +kubebuilder:default=true
	Enabled *bool `json:"enabled,omitempty"`

	Status rApi.Status `json:"status,omitempty" graphql:"noinput"`
}

func (app *App) EnsureGVK() {
	if app != nil {
		app.SetGroupVersionKind(GroupVersion.WithKind("App"))
	}
}

func (app *App) GetStatus() *rApi.Status {
	return &app.Status
}

func (app *App) GetEnsuredLabels() map[string]string {
	m := map[string]string{
		"kloudlite.io/app.name": app.Name,
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
