package serverless

import (
	"context"
	"encoding/json"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/constants"
	fn "operators.kloudlite.io/lib/functions"
	"operators.kloudlite.io/lib/kubectl"
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

func (r *LambdaReconciler) GetName() string {
	return "lambda"
}

const (
	KnativeServingExists conditions.Type = "KnativeServingExists"
	KnativeServingReady  conditions.Type = "KnativeServingReady"
)

func parseServingConditions(obj *unstructured.Unstructured) ([]metav1.Condition, error) {
	b, err := json.Marshal(obj.Object)
	if err != nil {
		return nil, err
	}
	var j struct {
		Status struct {
			Conditions []metav1.Condition `json:"conditions,omitempty"`
		} `json:"status"`
	}
	if err := json.Unmarshal(b, &j); err != nil {
		return nil, err
	}
	return j.Status.Conditions, nil
}

// +kubebuilder:rbac:groups=serverless.kloudlite.io,resources=lambdas,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=serverless.kloudlite.io,resources=lambdas/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=serverless.kloudlite.io,resources=lambdas/finalizers,verbs=update

func (r *LambdaReconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(ctx, r.Client, oReq.NamespacedName, &serverlessv1.Lambda{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// STEP: cleaning up last run, clearing opsConditions
	if len(req.Object.Status.OpsConditions) > 0 {
		req.Object.Status.OpsConditions = []metav1.Condition{}
		return ctrl.Result{RequeueAfter: 0}, r.Status().Update(ctx, req.Object)
	}

	if x := r.handleRestart(req); !x.ShouldProceed() {
		return x.Result(), x.Err()
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.Result(), x.Err()
		}
	}

	req.Logger.Infof("--------------------NEW RECONCILATION------------------")

	if x := req.EnsureLabelsAndAnnotations(); !x.ShouldProceed() {
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

func (r *LambdaReconciler) handleRestart(req *rApi.Request[*serverlessv1.Lambda]) rApi.StepResult {
	obj := req.Object
	ctx := req.Context()

	annotations := obj.GetAnnotations()
	if _, ok := req.Object.GetAnnotations()[constants.AnnotationKeys.Restart]; ok {
		req.Logger.Infof("resource came for restarting")

		exitCode, err := kubectl.Restart(kubectl.Deployments, req.Object.GetNamespace(), req.Object.GetEnsuredLabels())
		if exitCode != 0 {
			req.Logger.Error(err)
			// failed to restart, with non-zero exit code
		}
		patch := client.MergeFrom(req.Object.DeepCopy())
		delete(annotations, constants.AnnotationKeys.Restart)
		obj.SetAnnotations(annotations)
		if err := r.Patch(ctx, obj, patch); err != nil {
			return req.FailWithOpError(err)
		}
	}
	return req.Next()
}

func (r *LambdaReconciler) finalize(req *rApi.Request[*serverlessv1.Lambda]) rApi.StepResult {
	return req.Finalize()
}

func (r *LambdaReconciler) reconcileStatus(req *rApi.Request[*serverlessv1.Lambda]) rApi.StepResult {
	ctx := req.Context()
	obj := req.Object

	isReady := true
	var cs []metav1.Condition
	var childC []metav1.Condition

	// STEP: 1. sync conditions from Knative Serving
	knativeRes, err := rApi.Get(
		ctx, r.Client, fn.NN(obj.Namespace, obj.Name), fn.NewUnstructured(constants.KnativeServiceType),
	)

	if err != nil {
		isReady = false
		if !apiErrors.IsNotFound(err) {
			return req.FailWithStatusError(err)
		}
		cs = append(cs, conditions.New(KnativeServingExists, false, conditions.NotFound, err.Error()))
	} else {
		cs = append(cs, conditions.New(KnativeServingExists, true, conditions.Found))

		ksConditions, err := parseServingConditions(knativeRes)
		if err != nil {
			return req.FailWithStatusError(err)
		}
		childC = append(childC, ksConditions...)
		rReady := meta.IsStatusConditionTrue(ksConditions, "Ready")
		if !rReady {
			isReady = false
		}
		cs = append(
			cs, conditions.New(KnativeServingReady, rReady, conditions.Empty),
		)
	}

	// STEP: 5. patch aggregated conditions
	nConditionsC, hasUpdatedC, err := conditions.Patch(obj.Status.ChildConditions, childC)
	if err != nil {
		return req.FailWithStatusError(err)
	}

	nConditions, hasSUpdated, err := conditions.Patch(obj.Status.Conditions, cs)
	if err != nil {
		return req.FailWithStatusError(err)
	}

	if !hasUpdatedC && !hasSUpdated && isReady == obj.Status.IsReady {
		return req.Next()
	}

	obj.Status.IsReady = isReady
	obj.Status.Conditions = nConditions
	obj.Status.ChildConditions = nConditionsC
	obj.Status.OpsConditions = []metav1.Condition{}
	return rApi.NewStepResult(&ctrl.Result{}, r.Status().Update(ctx, obj))
}

func (r *LambdaReconciler) reconcileOperations(req *rApi.Request[*serverlessv1.Lambda]) rApi.StepResult {
	ctx := req.Context()
	obj := req.Object

	// STEP: 1. add finalizers if needed
	if !controllerutil.ContainsFinalizer(obj, constants.CommonFinalizer) {
		controllerutil.AddFinalizer(obj, constants.CommonFinalizer)
		controllerutil.AddFinalizer(obj, constants.ForegroundFinalizer)

		return rApi.NewStepResult(&ctrl.Result{}, r.Update(ctx, obj))
	}

	// STEP: 3. apply CRs of helm/custom controller
	if errP := func() error {
		b, err := templates.Parse(
			templates.ServerlessLambda, map[string]any{
				"object": obj,
				"owner-refs": []metav1.OwnerReference{
					fn.AsOwner(obj, true),
				},
			},
		)

		if err != nil {
			return err
		}

		if _, err := fn.KubectlApplyExec(b); err != nil {
			return err
		}
		return nil
	}(); errP != nil {
		req.FailWithOpError(errP)
	}

	return req.Done()
}

// SetupWithManager sets up the controller with the Manager.
func (r *LambdaReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&serverlessv1.Lambda{}).
		Owns(fn.NewUnstructured(constants.KnativeServiceType)).
		Complete(r)
}
