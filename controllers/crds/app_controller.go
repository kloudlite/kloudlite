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
	"operators.kloudlite.io/lib/kubectl"
	"operators.kloudlite.io/lib/logging"
	stepResult "operators.kloudlite.io/lib/operator/step-result"
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
	Env    *env.Env
	Logger logging.Logger
}

func (r *AppReconciler) GetName() string {
	return "app"
}

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps/finalizers,verbs=update

func (r *AppReconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.Logger), r.Client, oReq.NamespacedName, &crdsv1.App{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if step := r.handleRestart(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	req.Logger.Infof("-------------------- NEW RECONCILATION------------------")

	if x := req.EnsureLabelsAndAnnotations(); !x.ShouldProceed() {
		return x.ReconcilerResponse()
	}

	if x := r.reconcileStatus(req); !x.ShouldProceed() {
		return x.ReconcilerResponse()
	}

	if x := r.reconcileOperations(req); !x.ShouldProceed() {
		// // return x.ReconcilerResponse()
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *AppReconciler) handleRestart(req *rApi.Request[*crdsv1.App]) stepResult.Result {
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

func (r *AppReconciler) finalize(req *rApi.Request[*crdsv1.App]) stepResult.Result {
	return req.Finalize()
}

func (r *AppReconciler) reconcileStatus(req *rApi.Request[*crdsv1.App]) stepResult.Result {
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

		// TODO: conditions from pod
		// conditions.FromPod(ctx, r.Client, constants.DeploymentType)
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
	if err := r.Status().Update(ctx, obj); err != nil {
		return req.FailWithStatusError(err)
	}
	return req.Done()
}

func (r *AppReconciler) reconcileOperations(req *rApi.Request[*crdsv1.App]) stepResult.Result {
	ctx := req.Context()
	app := req.Object

	if !fn.ContainsFinalizers(app, constants.CommonFinalizer, constants.ForegroundFinalizer) {
		controllerutil.AddFinalizer(app, constants.CommonFinalizer)
		controllerutil.AddFinalizer(app, constants.ForegroundFinalizer)
		if err := r.Update(ctx, app); err != nil {
			return req.FailWithOpError(err)
		}
		return req.Done()
	}

	volumes, vMounts := crdsv1.ParseVolumes(app.Spec.Containers)

	b, err := templates.Parse(
		templates.CrdsV1.App, map[string]any{
			"object":        app,
			"volumes":       volumes,
			"volume-mounts": vMounts,
			"freeze":        app.GetLabels()[constants.LabelKeys.Freeze] == "true",
			"owner-refs": []metav1.OwnerReference{
				fn.AsOwner(app, true),
			},
		},
	)

	if err != nil {
		// this error won't be fixed in runtime
		return req.FailWithOpError(err).Err(nil)
	}

	if err := fn.KubectlApplyExec(ctx, b); err != nil {
		return req.FailWithOpError(err).Err(nil)
	}

	app.Status.OpsConditions = []metav1.Condition{}
	if err := r.Status().Update(ctx, app); err != nil {
		return req.FailWithOpError(err)
	}
	return req.Next()
}

func (r *AppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Logger = r.Logger.WithName("app")
	return ctrl.NewControllerManagedBy(mgr).
		For(&crdsv1.App{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
