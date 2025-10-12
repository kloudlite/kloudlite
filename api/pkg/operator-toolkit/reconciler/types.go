package reconciler

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Reconciler interface {
	reconcile.Reconciler
	SetupWithManager(mgr ctrl.Manager) error
	GetName() string
}

type Resource interface {
	client.Object
	runtime.Object
	GetStatus() *Status
}
