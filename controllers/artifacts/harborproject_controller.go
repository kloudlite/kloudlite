package artifacts

import (
	"context"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"operators.kloudlite.io/env"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/errors"
	"operators.kloudlite.io/lib/harbor"
	rApi "operators.kloudlite.io/lib/operator"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"k8s.io/apimachinery/pkg/runtime"
	artifactsv1 "operators.kloudlite.io/apis/artifacts/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// HarborProjectReconciler reconciles a HarborProject object
type HarborProjectReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Env       *env.Env
	harborCli *harbor.Client
}

func (r *HarborProjectReconciler) GetName() string {
	return "artifact-harbor-project"
}

const (
	HarborProjectExists           conditions.Type = "HarborProjectExists"
	HarborProjectStorageAllocated conditions.Type = "HarborProjectStorageAllocated"
)

const (
	KeyHarborAllocatedStorage string = "allocated-size-in-gb"
)

func convertGBToBytes(gb int) int {
	return gb * 1e6
}

// +kubebuilder:rbac:groups=artifacts.kloudlite.io,resources=harborprojects,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=artifacts.kloudlite.io,resources=harborprojects/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=artifacts.kloudlite.io,resources=harborprojects/finalizers,verbs=update

func (r *HarborProjectReconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(ctx, r.Client, oReq.NamespacedName, &artifactsv1.HarborProject{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
	}

	// STEP: cleaning up last run, clearing opsConditions
	if len(req.Object.Status.OpsConditions) > 0 {
		req.Object.Status.OpsConditions = []metav1.Condition{}
		return ctrl.Result{RequeueAfter: 0}, r.Status().Update(ctx, req.Object)
	}

	req.Logger.Infof("----------------[Type: artifactsv1.HarborProject] NEW RECONCILATION ----------------")

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

func (r *HarborProjectReconciler) finalize(req *rApi.Request[*artifactsv1.HarborProject]) rApi.StepResult {
	if err := r.harborCli.DeleteProject(req.Context(), req.Object.Name); err != nil {
		return req.FailWithOpError(err)
	}
	return req.Finalize()
}

func (r *HarborProjectReconciler) reconcileStatus(req *rApi.Request[*artifactsv1.HarborProject]) rApi.StepResult {
	ctx := req.Context()
	obj := req.Object

	var cs []metav1.Condition
	isReady := true

	// STEP: if harbor project has been created ?
	ok, err := r.harborCli.CheckIfProjectExists(ctx, obj.Name)
	if err != nil {
		isReady = false
		cs = append(cs, conditions.New(HarborProjectExists, false, conditions.NotFound, err.Error()))
	}
	if !ok {
		isReady = false
		cs = append(cs, conditions.New(HarborProjectExists, false, conditions.NotFound))
	} else {
		cs = append(cs, conditions.New(HarborProjectExists, true, conditions.Found))
	}

	// STEP: if asked storage has been allocated ?
	allocatedStorage, ok := obj.Status.DisplayVars.GetInt(KeyHarborAllocatedStorage)
	if !ok || obj.Spec.SizeInGB != allocatedStorage {
		isReady = false
		cs = append(cs, conditions.New(HarborProjectStorageAllocated, false, conditions.NotReconciledYet))
	} else {
		cs = append(cs, conditions.New(HarborProjectStorageAllocated, true, conditions.Found))
	}

	nConditions, hasUpdated, err := conditions.Patch(obj.Status.Conditions, cs)
	if err != nil {
		return req.FailWithStatusError(err)
	}

	if !hasUpdated && isReady == obj.Status.IsReady {
		return req.Next()
	}

	obj.Status.Conditions = nConditions
	obj.Status.IsReady = isReady

	if err := r.Status().Update(ctx, obj); err != nil {
		return req.FailWithStatusError(err)
	}

	return req.Done()
}

func (r *HarborProjectReconciler) reconcileOperations(req *rApi.Request[*artifactsv1.HarborProject]) rApi.StepResult {
	ctx := req.Context()
	obj := req.Object

	if !controllerutil.ContainsFinalizer(obj, constants.CommonFinalizer) {
		controllerutil.AddFinalizer(obj, constants.CommonFinalizer)
		controllerutil.AddFinalizer(obj, constants.ForegroundFinalizer)
		if err := r.Update(ctx, obj); err != nil {
			return req.FailWithOpError(err)
		}
		return req.Done()
	}

	if meta.IsStatusConditionFalse(obj.Status.Conditions, HarborProjectExists.String()) {
		if err := func() error {
			// 2 GB default storage size
			if obj.Spec.SizeInGB == 0 {
				obj.Spec.SizeInGB = r.Env.HarborProjectStorageSize
			}

			if !r.Env.HarborQuoteEnabled {
				obj.Spec.SizeInGB = 0
			}

			if err := r.harborCli.CreateProject(ctx, obj.Name, convertGBToBytes(obj.Spec.SizeInGB)); err != nil {
				return errors.NewEf(err, "creating harbor project")
			}

			return obj.Status.DisplayVars.Set(KeyHarborAllocatedStorage, obj.Spec.SizeInGB)
		}(); err != nil {
			return req.FailWithOpError(err)
		}
		return rApi.NewStepResult(&ctrl.Result{RequeueAfter: 0}, nil)
	}

	// TODO: it should not be called until harbor quota issue gets fixed
	if meta.IsStatusConditionFalse(obj.Status.Conditions, HarborProjectStorageAllocated.String()) {
		if err := func() error {
			if err := r.harborCli.SetProjectQuota(ctx, obj.Name, obj.Spec.SizeInGB); err != nil {
				return err
			}
			if err := obj.Status.DisplayVars.Set(KeyHarborAllocatedStorage, obj.Spec.SizeInGB); err != nil {
				return err
			}
			return r.Status().Update(ctx, obj)
		}(); err != nil {
			return req.FailWithOpError(err)
		}
		return rApi.NewStepResult(&ctrl.Result{RequeueAfter: 0}, nil)
	}

	return req.Done()
}

// SetupWithManager sets up the controller with the Manager.
func (r *HarborProjectReconciler) SetupWithManager(mgr ctrl.Manager) error {
	harborCli, err := harbor.NewClient(
		harbor.Args{
			HarborAdminUsername: r.Env.HarborAdminUsername,
			HarborAdminPassword: r.Env.HarborAdminPassword,
			HarborRegistryHost:  r.Env.HarborImageRegistryHost,
		},
	)
	if err != nil {
		return nil
	}
	r.harborCli = harborCli

	return ctrl.NewControllerManagedBy(mgr).
		For(&artifactsv1.HarborProject{}).
		Complete(r)
}
