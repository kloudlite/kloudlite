package reconciler

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
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
	IsReady   bool          `json:"isReady"`
	Resources []ResourceRef `json:"resources,omitempty"`

	CheckList           []CheckMeta      `json:"checkList,omitempty"`
	Checks              map[string]Check `json:"checks,omitempty"`
	LastReadyGeneration int64            `json:"lastReadyGeneration,omitempty"`
	LastReconcileTime   *metav1.Time     `json:"lastReconcileTime,omitempty"`
}

type Reconciler interface {
	reconcile.Reconciler
	// SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error
	SetupWithManager(mgr ctrl.Manager) error
	GetName() string
}

type NamedReconciler interface {
	GetName() string
}

type Reconciler2 struct {
	reconcile.Reconciler
	NamedReconciler
}

type Resource interface {
	client.Object
	runtime.Object
	EnsureGVK()
	GetStatus() *Status
	GetEnsuredLabels() map[string]string
	GetEnsuredAnnotations() map[string]string
}
