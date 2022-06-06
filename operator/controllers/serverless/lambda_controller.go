package serverless

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fn "operators.kloudlite.io/lib/functions"
	rApi "operators.kloudlite.io/lib/operator"

	"k8s.io/apimachinery/pkg/runtime"
	serverlessv1 "operators.kloudlite.io/apis/serverless/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// LambdaReconciler reconciles a Lambda object
type LambdaReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=serverless.kloudlite.io,resources=lambdas,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=serverless.kloudlite.io,resources=lambdas/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=serverless.kloudlite.io,resources=lambdas/finalizers,verbs=update

func (r *LambdaReconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	req := rApi.NewRequest(ctx, r.Client, oReq.NamespacedName, &serverlessv1.Lambda{})

	if req == nil {
		return ctrl.Result{}, nil
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.Result(), x.Err()
		}
	}

	req.Logger.Info("-------------------- NEW RECONCILATION------------------")

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

func (r *LambdaReconciler) finalize(req *rApi.Request[*serverlessv1.Lambda]) rApi.StepResult {
	return req.Finalize()
}

func (r *LambdaReconciler) reconcileStatus(req *rApi.Request[*serverlessv1.Lambda]) rApi.StepResult {
	return req.Done()
}

func (r *LambdaReconciler) reconcileOperations(req *rApi.Request[*serverlessv1.Lambda]) rApi.StepResult {
	return req.Done()
}

// SetupWithManager sets up the controller with the Manager.
func (r *LambdaReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&serverlessv1.Lambda{}).
		Owns(
			fn.NewUnstructured(
				metav1.TypeMeta{
					Kind:       "Service",
					APIVersion: "serving.knative.dev/v1",
				},
			),
		).
		Complete(r)
}
