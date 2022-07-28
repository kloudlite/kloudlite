package mysqlclustermsvc

import (
	"context"
	"k8s.io/apimachinery/pkg/runtime"
	mysqlCluster "operators.kloudlite.io/apis/mysql-cluster.msvc/v1"
	rApi "operators.kloudlite.io/lib/operator"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DatabaseReconciler reconciles a Database object
type DatabaseReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=mysql-cluster.msvc.kloudlite.io,resources=databases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mysql-cluster.msvc.kloudlite.io,resources=databases/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mysql-cluster.msvc.kloudlite.io,resources=databases/finalizers,verbs=update

func (r *DatabaseReconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(ctx, r.Client, oReq.NamespacedName, &mysqlCluster.Database{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
	}

	req.Logger.Infof("----------------[Type: mysqlCluster.Database] NEW RECONCILATION ----------------")

	if x := req.EnsureLabelsAndAnnotations(); !x.ShouldProceed() {
		return x.ReconcilerResponse()
	}

	if x := r.reconcileStatus(req); !x.ShouldProceed() {
		return x.ReconcilerResponse()
	}

	if x := r.reconcileOperations(req); !x.ShouldProceed() {
		return x.ReconcilerResponse()
	}

	return ctrl.Result{}, nil
}

func (r *DatabaseReconciler) finalize(req *rApi.Request[*mysqlCluster.Database]) rApi.StepResult {
	return req.Finalize()
}

func (r *DatabaseReconciler) reconcileStatus(req *rApi.Request[*mysqlCluster.Database]) rApi.StepResult {
	return req.Done()
}

func (r *DatabaseReconciler) reconcileOperations(req *rApi.Request[*mysqlCluster.Database]) rApi.StepResult {
	return req.Done()
}

func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mysqlCluster.Database{}).
		Complete(r)
}
