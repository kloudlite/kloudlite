package mongodbclustermsvc

import (
	"context"
	rApi "operators.kloudlite.io/lib/operator"

	"k8s.io/apimachinery/pkg/runtime"
	mongodbclustermsvcv1 "operators.kloudlite.io/apis/mongodb-cluster.msvc/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DatabaseReconciler reconciles a Database object
type DatabaseReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=mongodb-cluster.msvc.kloudlite.io,resources=databases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mongodb-cluster.msvc.kloudlite.io,resources=databases/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mongodb-cluster.msvc.kloudlite.io,resources=databases/finalizers,verbs=update

func (r *DatabaseReconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(ctx, r.Client, oReq.NamespacedName, &mongodbclustermsvcv1.Database{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.Result(), x.Err()
		}
	}

	req.Logger.Info("----------------[Type: mongodbclustermsvcv1.Database] NEW RECONCILATION ----------------")

	if x := req.EnsureLabels(); !x.ShouldProceed() {
		return x.Result(), x.Err()
	}

	if x := r.reconcileStatus(req); !x.ShouldProceed() {
		return x.Result(), x.Err()
	}

	if x := r.reconcileOperations(req); !x.ShouldProceed() {
		return x.Result(), x.Err()
	}

	return ctrl.Result{}, nil
}

func (r *DatabaseReconciler) finalize(req *rApi.Request[*mongodbclustermsvcv1.Database]) rApi.StepResult {
	return req.Finalize()
}

func (r *DatabaseReconciler) reconcileStatus(req *rApi.Request[*mongodbclustermsvcv1.Database]) rApi.StepResult {
	return req.Done()
}

func (r *DatabaseReconciler) reconcileOperations(req *rApi.Request[*mongodbclustermsvcv1.Database]) rApi.StepResult {
	return req.Done()
}

// SetupWithManager sets up the controller with the Manager.
func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mongodbclustermsvcv1.Database{}).
		Complete(r)
}
