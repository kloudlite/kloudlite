package reconciler

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// +kubebuilder:object:generate=true
//type ResourceRef struct {
//	metav1.TypeMeta `json:",inline" graphql:"children-required"`
//	Namespace       string `json:"namespace"`
//	Name            string `json:"name"`
//}

type Reconciler interface {
	reconcile.Reconciler
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

type CustomResource interface {
	EnsureGVK()
	GetStatus() *Status
}
