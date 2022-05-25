package redisclustermsvc

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	redisclustermsvcv1 "operators.kloudlite.io/apis/redis-cluster.msvc/v1"
)

// KeyPrefixReconciler reconciles a KeyPrefix object
type KeyPrefixReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=redis-cluster.msvc.kloudlite.io,resources=keyprefixes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redis-cluster.msvc.kloudlite.io,resources=keyprefixes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redis-cluster.msvc.kloudlite.io,resources=keyprefixes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the KeyPrefix object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *KeyPrefixReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KeyPrefixReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&redisclustermsvcv1.KeyPrefix{}).
		Complete(r)
}
