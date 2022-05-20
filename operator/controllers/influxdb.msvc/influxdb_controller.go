package influxdbmsvc

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	influxdbmsvcv1 "operators.kloudlite.io/apis/influxdb.msvc/v1"
)

// InfluxDBReconciler reconciles a InfluxDB object
type InfluxDBReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=influxdb.msvc.kloudlite.io,resources=influxdbs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=influxdb.msvc.kloudlite.io,resources=influxdbs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=influxdb.msvc.kloudlite.io,resources=influxdbs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the InfluxDB object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *InfluxDBReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *InfluxDBReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&influxdbmsvcv1.InfluxDB{}).
		Complete(r)
}
