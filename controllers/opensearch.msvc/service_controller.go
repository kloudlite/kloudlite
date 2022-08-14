package opensearchmsvc

import (
	"context"

	"operators.kloudlite.io/env"
	"operators.kloudlite.io/lib/logging"
	rApi "operators.kloudlite.io/lib/operator"
	stepResult "operators.kloudlite.io/lib/operator/step-result"

	"k8s.io/apimachinery/pkg/runtime"
	opensearchmsvcv1 "operators.kloudlite.io/apis/opensearch.msvc/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ServiceReconciler reconciles a Service object
type ServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	logger logging.Logger
}

func (r *ServiceReconciler) GetName() string {
	return "opensearch-service"
}

// +kubebuilder:rbac:groups=opensearch.msvc.kloudlite.io,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=opensearch.msvc.kloudlite.io,resources=services/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=opensearch.msvc.kloudlite.io,resources=services/finalizers,verbs=update

func (r *ServiceReconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, oReq.NamespacedName, &opensearchmsvcv1.Service{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
	}

	req.Logger.Infof("-------------------- NEW RECONCILATION------------------")

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

func (r *ServiceReconciler) finalize(req *rApi.Request[*opensearchmsvcv1.Service]) stepResult.Result {
	return req.Finalize()
}

func (r *ServiceReconciler) reconcileStatus(req *rApi.Request[*opensearchmsvcv1.Service]) stepResult.Result {
	return req.Done()
}

func (r *ServiceReconciler) reconcileOperations(req *rApi.Request[*opensearchmsvcv1.Service]) stepResult.Result {
	return req.Done()
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager, envVars *env.Env, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()

	r.logger = logger.WithName("opensearch")
	return ctrl.NewControllerManagedBy(mgr).
		For(&opensearchmsvcv1.Service{}).
		Complete(r)
}
