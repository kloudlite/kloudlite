package serverless

import (
	"context"
	"encoding/json"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/constants"
	fn "operators.kloudlite.io/lib/functions"
	rApi "operators.kloudlite.io/lib/operator"
	"operators.kloudlite.io/lib/templates"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

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

	req.Logger.Info("--------------------NEW RECONCILATION------------------")

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
	ctx := req.Context()
	lambdaSvc := req.Object

	isReady := true
	publicUrl := ""
	var cs []metav1.Condition

	knativeSvc, err := rApi.Get(
		ctx, r.Client, fn.NN(lambdaSvc.Namespace, lambdaSvc.Name), fn.NewUnstructured(constants.KnativeServiceType),
	)
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.FailWithStatusError(err)
		}
		isReady = false
		knativeSvc = nil
	}

	if err := func() error {
		if knativeSvc != nil {
			if _, ok := knativeSvc.Object["status"]; !ok {
				cs = append(cs, conditions.New("Initializing", true, ""))
				return nil
			}

			var j struct {
				Status struct {
					Url        string             `json:"url,omitempty"`
					Conditions []metav1.Condition `json:"conditions,omitempty"`
				} `json:"status"`
			}

			b, err := json.Marshal(knativeSvc.Object)
			if err != nil {
				return err
			}

			if err := json.Unmarshal(b, &j); err != nil {
				return err
			}

			status := knativeSvc.Object["status"].(map[string]any)
			publicUrl = j.Status.Url
			rApi.SetLocal(req, "Url", status["url"])

			cs = append(cs, j.Status.Conditions...)
			if meta.IsStatusConditionFalse(j.Status.Conditions, "Ready") {
				isReady = false
			}
		}
		return nil
	}(); err != nil {
		return req.FailWithStatusError(err)
	}

	newConditions, hasUpdated, err := conditions.Patch(lambdaSvc.Status.Conditions, cs)
	if err != nil {
		return req.FailWithStatusError(err)
	}

	if !hasUpdated && isReady == lambdaSvc.Status.IsReady {
		return req.Next()
	}

	lambdaSvc.Status.IsReady = isReady
	if err := lambdaSvc.Status.DisplayVars.Set("url", publicUrl); err != nil {
		return req.FailWithStatusError(err)
	}
	lambdaSvc.Status.Conditions = newConditions
	lambdaSvc.Status.OpsConditions = []metav1.Condition{}
	if err := r.Status().Update(ctx, lambdaSvc); err != nil {
		return req.FailWithStatusError(err)
	}
	return req.Done()
}

func (r *LambdaReconciler) reconcileOperations(req *rApi.Request[*serverlessv1.Lambda]) rApi.StepResult {
	ctx := req.Context()
	lambdaSvc := req.Object

	if !controllerutil.ContainsFinalizer(lambdaSvc, constants.CommonFinalizer) {
		controllerutil.AddFinalizer(lambdaSvc, constants.CommonFinalizer)
		if err := r.Update(ctx, lambdaSvc); err != nil {
			return req.FailWithOpError(err)
		}
		return req.Done()
	}

	volumes, vMounts := crdsv1.ParseVolumes(lambdaSvc.Spec.Containers)
	pObj, err := templates.ParseObject(
		templates.ServerlessLambda, map[string]any{
			"obj":          lambdaSvc,
			"volumes":      volumes,
			"volumeMounts": vMounts,
		},
	)
	if err != nil {
		return req.FailWithOpError(err)
	}
	pObj.SetOwnerReferences(
		[]metav1.OwnerReference{
			fn.AsOwner(lambdaSvc, true),
		},
	)

	if err := fn.KubectlApply(ctx, r.Client, pObj); err != nil {
		return req.FailWithOpError(err)
	}

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
