package v1

import (
	"github.com/kloudlite/kloudlite/operator/toolkit/reconciler"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AppSpec defines the desired state of App.
type AppSpec struct {
	Intercept *Intercept `json:"intercept,omitempty"`
	Paused    bool       `json:"paused,omitempty"`

	// +kubebuilder:default=1
	Replicas int            `json:"replicas,omitempty"`
	PodSpec  corev1.PodSpec `json:"podSpec"`
	Services []AppService   `json:"services,omitempty"`
	HPA      *AppHPA        `json:"hpa,omitempty"`
	Router   *AppRouter     `json:"router,omitempty"`
}

type Intercept struct {
	Enabled *bool `json:"enabled,omitempty"`

	ToHost       string                     `json:"toHost,omitempty"`
	PortMappings []AppInterceptPortMappings `json:"portMappings,omitempty"`
}

type AppInterceptPortMappings struct {
	Protocol   corev1.Protocol `json:"protocol"`
	AppPort    int32           `json:"appPort"`
	DevicePort int32           `json:"devicePort"`
}

// AppService creates k8s Service of type ClusterIP
type AppService struct {
	Port int32 `json:"port"`

	// +kubebuilder:default=TCP
	Protocol corev1.Protocol `json:"protocol,omitempty"`
}

type AppHPA struct {
	Enabled *bool `json:"enabled,omitempty"`

	// +kubebuilder:default=1
	MinReplicas int32 `json:"minReplicas,omitempty"`
	// +kubebuilder:default=5
	MaxReplicas int32 `json:"maxReplicas,omitempty"`
	// +kubebuilder:default=90
	ThresholdCpu int32 `json:"thresholdCpu,omitempty"`
	// +kubebuilder:default=75
	ThresholdMemory int32 `json:"thresholdMemory,omitempty"`
}

// AppRouter inspired by github.com/kloudlite/operator/apis/crds/v1.RouterSpec
type AppRouter struct {
	Enabled *bool `json:"enabled,omitempty"`

	RouterSpec `json:",inline,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".status.lastReconcileTime",name=Seen,type=date
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.operator\\.kloudlite\\.io\\/checks",name=Checks,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.operator\\.kloudlite\\.io\\/resource\\.ready",name=Ready,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.operator\\.kloudlite\\.io\\/toHost",name=Intercepted,type=string
// +kubebuilder:printcolumn:JSONPath=".spec.paused",name=Paused,type=boolean
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// App is the Schema for the apps API.
type App struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AppSpec           `json:"spec,omitempty"`
	Status reconciler.Status `json:"status,omitempty"`
}

func (r *App) EnsureGVK() {
	if r != nil {
		r.SetGroupVersionKind(GroupVersion.WithKind("App"))
	}
}

func (r *App) GetStatus() *reconciler.Status {
	return &r.Status
}

func (r *App) GetEnsuredLabels() map[string]string {
	return map[string]string{
		NameLabelKey: r.Name,
		KindLabelKey: "App",
	}
}

func (m *App) GetEnsuredAnnotations() map[string]string {
	return map[string]string{}
}

func (a *App) IsInterceptEnabled() bool {
	return a.Spec.Intercept != nil && (a.Spec.Intercept.Enabled == nil || *a.Spec.Intercept.Enabled)
}

func (a *App) IsHPAEnabled() bool {
	return a.Spec.HPA != nil && (a.Spec.HPA.Enabled == nil || *a.Spec.HPA.Enabled)
}

func (a *App) IsRouterEnabled() bool {
	return a.Spec.Router != nil && (a.Spec.Router.Enabled == nil || *a.Spec.Router.Enabled)
}

// +kubebuilder:object:root=true

// AppList contains a list of App.
type AppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []App `json:"items"`
}

func init() {
	SchemeBuilder.Register(&App{}, &AppList{})
}
