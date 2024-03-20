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
	WaitingState   State = "yet-to-be-reconciled"
	RunningState   State = "under-reconcilation"
	ErroredState   State = "errored-during-reconcilation"
	CompletedState State = "finished-reconcilation"
)

// +kubebuilder:object:generate=true
type Check struct {
	Status     bool   `json:"status"`
	Message    string `json:"message,omitempty"`
	Generation int64  `json:"generation,omitempty"`

	State     State        `json:"state,omitempty"`
	StartedAt *metav1.Time `json:"startedAt,omitempty"`
	Info      string       `json:"info,omitempty"`
	Debug     string       `json:"debug,omitempty"`
	Error     string       `json:"error,omitempty"`
}

// +kubebuilder:object:generate=true
type CheckMeta struct {
	Name        string  `json:"name"`
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	Debug       bool    `json:"debug,omitempty"`
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

	CheckList           []CheckMeta      `json:"checkList,omitempty"`
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
