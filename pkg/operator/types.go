package operator

import (
	"github.com/kloudlite/operator/pkg/logging"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	rawJson "github.com/kloudlite/operator/pkg/raw-json"
)

// +kubebuilder:object:generate=true
type State string

const (
	WaitingState   string = "yet-to-be-reconciled"
	RunningState   string = "under-reconcilation"
	ErroredState   string = "errored-during-reconcilation"
	CompletedState string = "finished-reconcilation"
)

// +kubebuilder:object:generate=true
type Check struct {
	Status bool `json:"status"`
	// State      State  `json:"state"`
	Message    string `json:"message,omitempty"`
	Generation int64  `json:"generation,omitempty"`
}

// +kubebuilder:object:generate=true
type ResourceRef struct {
	metav1.TypeMeta `json:",inline" graphql:"children-required"`
	Namespace       string `json:"namespace"`
	Name            string `json:"name"`
}

// +kubebuilder:object:generate=true
// +kubebuilder:printcolumn:JSONPath=".status.isReady",name=Ready,type=boolean
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

type Status struct {
	// +kubebuilder:validation:Optional
	IsReady   bool             `json:"isReady"`
	Resources []ResourceRef    `json:"resources,omitempty"`
	Message   *rawJson.RawJson `json:"message,omitempty"`

	// Messages    []ContainerMessage `json:"messages,omitempty"`

	// DisplayVars *rawJson.RawJson   `json:"displayVars,omitempty"`

	// GeneratedVars     *rawJson.RawJson   `json:"generatedVars,omitempty"`

	// Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ChildConditions   []metav1.Condition `json:"childConditions,omitempty"`
	// OpsConditions     []metav1.Condition `json:"opsConditions,omitempty"`

	Checks              map[string]Check `json:"checks,omitempty"`
	LastReadyGeneration int64            `json:"lastReadyGeneration,omitempty"`
	LastReconcileTime   *metav1.Time     `json:"lastReconcileTime,omitempty"`
}

type Reconciler interface {
	reconcile.Reconciler
	SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error
	GetName() string
}

type Resource interface {
	client.Object
	runtime.Object
	EnsureGVK()
	GetStatus() *Status
	GetEnsuredLabels() map[string]string
	GetEnsuredAnnotations() map[string]string
}
