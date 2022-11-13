package artifacts

import (
	"context"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"operators.kloudlite.io/env"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/errors"
	fn "operators.kloudlite.io/lib/functions"
	"operators.kloudlite.io/lib/harbor"
	"operators.kloudlite.io/lib/logging"
	rApi "operators.kloudlite.io/lib/operator"
	stepResult "operators.kloudlite.io/lib/operator/step-result"
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
	env       *env.Env
	harborCli *harbor.Client
	logger    logging.Logger
	Name      string
}

func (r *HarborProjectReconciler) GetName() string {
	return r.Name
}

const (
	ProjectExists conditions.Type = "harbor.project/Exists"
	WebhookExists conditions.Type = "harbor.project/WebhookExists"
)

const (
	KeyWebhook string = "webhook"
	KeyProject string = "project"
)

// +kubebuilder:rbac:groups=artifacts.kloudlite.io,resources=harborprojects,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=artifacts.kloudlite.io,resources=harborprojects/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=artifacts.kloudlite.io,resources=harborprojects/finalizers,verbs=update

func (r *HarborProjectReconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, oReq.NamespacedName, &artifactsv1.HarborProject{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
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

func (r *HarborProjectReconciler) finalize(req *rApi.Request[*artifactsv1.HarborProject]) stepResult.Result {
	if err := r.harborCli.DeleteProject(req.Context(), req.Object.Name); err != nil {
		return req.FailWithOpError(err)
	}
	return req.Finalize()
}

func (r *HarborProjectReconciler) reconcileStatus(req *rApi.Request[*artifactsv1.HarborProject]) stepResult.Result {
	ctx := req.Context()
	obj := req.Object

	var cs []metav1.Condition
	isReady := true

	var project harbor.Project
	obj.Status.DisplayVars.Get(KeyProject, &project)

	if &project == nil {
		isReady = false
		cs = append(cs, conditions.New(ProjectExists, false, conditions.NotFound))
	} else {
		exists, err := r.harborCli.CheckIfProjectExists(ctx, obj.Name)
		if err != nil {
			isReady = false
			return req.FailWithStatusError(errors.NewEf(err, "checking if project exists"))
		}

		if exists {
			cs = append(cs, conditions.New(ProjectExists, true, conditions.Found))
		} else {
			isReady = false
			cs = append(cs, conditions.New(ProjectExists, false, conditions.NotFound))
		}
	}

	// check if webhook added
	var webhook harbor.Webhook
	obj.Status.DisplayVars.Get(KeyWebhook, &webhook)

	if &webhook == nil {
		isReady = false
		cs = append(cs, conditions.New(WebhookExists, false, conditions.NotFound))
	} else {
		exists, err := r.harborCli.CheckWebhookExists(ctx, "", &webhook)
		if err != nil {
			isReady = false
			// cs = append(cs, conditions.New(WebhookExists, false, conditions.NotFound, err.Message()))
			return req.FailWithStatusError(errors.NewEf(err, "checking if webhook exists"))
		}
		if exists {
			cs = append(cs, conditions.New(WebhookExists, true, conditions.Found))
		} else {
			isReady = false
			cs = append(cs, conditions.New(WebhookExists, false, conditions.NotFound))
		}
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

func (r *HarborProjectReconciler) reconcileOperations(req *rApi.Request[*artifactsv1.HarborProject]) stepResult.Result {
	ctx := req.Context()
	obj := req.Object

	if !fn.ContainsFinalizers(obj, constants.CommonFinalizer, constants.ForegroundFinalizer) {
		controllerutil.AddFinalizer(obj, constants.CommonFinalizer)
		controllerutil.AddFinalizer(obj, constants.ForegroundFinalizer)
		if err := r.Update(ctx, obj); err != nil {
			return req.FailWithOpError(err)
		}
		return req.Done()
	}

	if meta.IsStatusConditionFalse(obj.Status.Conditions, ProjectExists.String()) {
		project, err := r.harborCli.CreateProject(ctx, obj.Name)
		if err != nil {
			return req.FailWithOpError(errors.NewEf(err, "creating harbor project"))
		}

		if err := obj.Status.DisplayVars.Set(KeyProject, project); err != nil {
			return req.FailWithOpError(err).Err(nil)
		}

		if err := r.Status().Update(ctx, obj); err != nil {
			return req.FailWithOpError(err)
		}

		return req.Done().RequeueAfter(0)
	}

	if meta.IsStatusConditionFalse(obj.Status.Conditions, WebhookExists.String()) {
		webhook, err := r.harborCli.CreateWebhook(ctx, obj.Name)
		if err != nil {
			return req.FailWithOpError(err)
		}
		if err := obj.Status.DisplayVars.Set(KeyWebhook, webhook); err != nil {
			return req.FailWithOpError(err)
		}
		if err := r.Status().Update(ctx, obj); err != nil {
			return req.FailWithOpError(err)
		}
		if webhook != nil {
			req.Logger.Infof("webook: %+v\n", *webhook)
		}
	}

	obj.Status.OpsConditions = []metav1.Condition{}
	// obj.Status.Generation = obj.Generation
	if err := r.Status().Update(ctx, obj); err != nil {
		return req.FailWithOpError(err)
	}
	return req.Next()
}

// SetupWithManager sets up the controllers with the Manager.
func (r *HarborProjectReconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.env = envVars
	r.logger = logger.WithName(r.Name)

	harborCli, err := harbor.NewClient(
		harbor.Args{
			HarborAdminUsername: r.env.HarborAdminUsername,
			HarborAdminPassword: r.env.HarborAdminPassword,
			HarborRegistryHost:  r.env.HarborImageRegistryHost,
			WebhookAddr:         r.env.HarborWebhookAddr,
		},
	)

	if err != nil {
		return err
	}
	r.harborCli = harborCli

	return ctrl.NewControllerManagedBy(mgr).
		For(&artifactsv1.HarborProject{}).
		Complete(r)
}
