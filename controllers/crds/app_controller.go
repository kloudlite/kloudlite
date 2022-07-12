package crds

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"operators.kloudlite.io/env"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/constants"
	fn "operators.kloudlite.io/lib/functions"
	rApi "operators.kloudlite.io/lib/operator"

	"operators.kloudlite.io/lib/templates"
)

// AppReconciler reconciles a Deployment object
type AppReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	Env env.Env
}

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps/finalizers,verbs=update

func (r *AppReconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	req, _ := rApi.NewRequest(ctx, r.Client, oReq.NamespacedName, &crdsv1.App{})

	if req == nil {
		return ctrl.Result{}, nil
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.Result(), x.Err()
		}
	}

	req.Logger.Infof("-------------------- NEW RECONCILATION------------------")

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

func (r *AppReconciler) finalize(req *rApi.Request[*crdsv1.App]) rApi.StepResult {
	return req.Finalize()
}

func (r *AppReconciler) reconcileStatus(req *rApi.Request[*crdsv1.App]) rApi.StepResult {
	ctx := req.Context()
	obj := req.Object

	isReady := true
	var cs []metav1.Condition
	var childC []metav1.Condition

	// STEP: 2. sync conditions from deployments/statefulsets
	deploymentRes, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Name), &appsv1.Deployment{})
	if err != nil {
		isReady = false
		if !apiErrors.IsNotFound(err) {
			return req.FailWithStatusError(err)
		}
		cs = append(cs, conditions.New(conditions.DeploymentExists, false, conditions.NotFound, err.Error()))
	} else {
		cs = append(cs, conditions.New(conditions.DeploymentExists, true, conditions.Found))

		rConditions, err := conditions.ParseFromResource(deploymentRes, "Deployment")
		if err != nil {
			return req.FailWithStatusError(err)
		}
		childC = append(childC, rConditions...)
		rReady := meta.IsStatusConditionTrue(rConditions, "DeploymentAvailable")
		if !rReady {
			isReady = false
		}
		cs = append(
			cs, conditions.New(conditions.DeploymentReady, rReady, conditions.Empty),
		)
	}

	// STEP: 2.1: check current number of replicas
	if err := func() error {
		readyReplicas, ok := obj.Status.DisplayVars.GetInt("readyReplicas")
		if ok && readyReplicas == int(deploymentRes.Status.ReadyReplicas) {
			return nil
		}
		isReady = false
		return obj.Status.DisplayVars.Set("readyReplicas", deploymentRes.Status.ReadyReplicas)
	}(); err != nil {
		return req.FailWithStatusError(err)
	}

	// STEP: 3. service exists?
	_, err = rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Name), &corev1.Service{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.FailWithStatusError(err)
		}
		isReady = false
		cs = append(cs, conditions.New(conditions.ServiceExists, false, conditions.NotFound, err.Error()))
	} else {
		cs = append(cs, conditions.New(conditions.ServiceExists, true, conditions.Found))
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

func (r *AppReconciler) reconcileOperations(req *rApi.Request[*crdsv1.App]) rApi.StepResult {
	ctx := req.Context()
	app := req.Object

	if !controllerutil.ContainsFinalizer(app, constants.CommonFinalizer) {
		controllerutil.AddFinalizer(app, constants.CommonFinalizer)
		controllerutil.AddFinalizer(app, constants.ForegroundFinalizer)
		return rApi.NewStepResult(&ctrl.Result{}, r.Update(ctx, app))
	}

	volumes, vMounts := crdsv1.ParseVolumes(app.Spec.Containers)

	b, err := templates.Parse(
		templates.CrdsV1.App, map[string]any{
			"object":        app,
			"volumes":       volumes,
			"volume-mounts": vMounts,
			"owner-refs": []metav1.OwnerReference{
				fn.AsOwner(app, true),
			},
		},
	)
	if err != nil {
		return req.FailWithOpError(err)
	}

	if _, err := fn.KubectlApplyExec(b); err != nil {
		return req.FailWithOpError(err)
	}
	return req.Done()
}

func (r *AppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crdsv1.App{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
