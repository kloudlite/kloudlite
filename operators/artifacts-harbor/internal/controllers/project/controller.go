package project

import (
	"context"
	"net/http"
	"time"

	artifactsv1 "github.com/kloudlite/operator/apis/artifacts/v1"
	"github.com/kloudlite/operator/operators/artifacts-harbor/internal/env"
	"github.com/kloudlite/operator/pkg/constants"
	"github.com/kloudlite/operator/pkg/errors"
	"github.com/kloudlite/operator/pkg/harbor"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Reconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	HarborCli *harbor.Client
	logger    logging.Logger
	Name      string
	Env       *env.Env
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	DefaultsPatched    string = "defaults-patched"
	HarborProjectReady string = "harbor-project-ready"
	WebhookReady       string = "webhook-ready"
)

// +kubebuilder:rbac:groups=artifacts.kloudlite.io,resources=harborprojects,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=artifacts.kloudlite.io,resources=harborprojects/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=artifacts.kloudlite.io,resources=harborprojects/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &artifactsv1.HarborProject{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	req.LogPreReconcile()
	defer req.LogPostReconcile()

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureChecks(DefaultsPatched, HarborProjectReady, WebhookReady); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconDefaults(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconHarborProject(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconWebhook(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	req.Object.Status.LastReconcileTime = &metav1.Time{Time: time.Now()}
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) finalize(req *rApi.Request[*artifactsv1.HarborProject]) stepResult.Result {
	return req.Finalize()
}

func (r *Reconciler) reconDefaults(req *rApi.Request[*artifactsv1.HarborProject]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(DefaultsPatched)
	defer req.LogPostCheck(DefaultsPatched)

	hasUpdated := false

	if obj.Spec.Project == nil {
		obj.Spec.Project = &harbor.Project{Name: obj.Name}
		hasUpdated = true
	}

	if hasUpdated {
		if err := r.Update(ctx, obj); err != nil {
			return req.CheckFailed(DefaultsPatched, check, err.Error())
		}
		return req.Next().RequeueAfter(100 * time.Millisecond)
	}

	check.Status = true
	if check != obj.Status.Checks[DefaultsPatched] {
		obj.Status.Checks[DefaultsPatched] = check
		req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) reconHarborProject(req *rApi.Request[*artifactsv1.HarborProject]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(HarborProjectReady)
	defer req.LogPostCheck(HarborProjectReady)

	exists, err := r.HarborCli.CheckIfProjectExists(ctx, obj.Spec.Project.Name)
	if err != nil {
		return req.CheckFailed(HarborProjectReady, check, err.Error())
	}

	if !exists {
		project, err := r.HarborCli.CreateProject(ctx, obj.Spec.Project.Name)
		if err != nil {
			return req.CheckFailed(HarborProjectReady, check, err.Error())
		}
		obj.Spec.Project = project
		if err := r.Update(ctx, obj); err != nil {
			return nil
		}
		return req.Done().RequeueAfter(100 * time.Millisecond)
	}

	if obj.Spec.Project.Location == "" {
		req.Logger.Infof("project location is empty, going to query harbor for it")
		project, err := r.HarborCli.GetProject(ctx, obj.Spec.Project.Name)
		if err != nil {
			return req.CheckFailed(HarborProjectReady, check, err.Error())
		}
		obj.Spec.Project = project
		if err := r.Update(ctx, obj); err != nil {
			return req.CheckFailed(HarborProjectReady, check, err.Error())
		}
		return req.Done().RequeueAfter(100 * time.Millisecond)
	}

	check.Status = true
	if check != obj.Status.Checks[HarborProjectReady] {
		obj.Status.Checks[HarborProjectReady] = check
		req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) reconWebhook(req *rApi.Request[*artifactsv1.HarborProject]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(WebhookReady)
	defer req.LogPostCheck(WebhookReady)

	webhook, err := r.HarborCli.FindWebhookByName(ctx, obj.Name, r.Env.HarborWebhookName)
	if err != nil {
		httpErr, ok := err.(*errors.HttpError)
		if !ok || httpErr.Code != http.StatusNotFound {
			return req.CheckFailed(WebhookReady, check, err.Error()).Err(nil)
		}
		req.Logger.Infof("webhook does not exist, will be creating now ...")
	}

	if webhook != nil && (obj.Spec.Webhook == nil || obj.Spec.Webhook.Location == "") {
		if err := r.HarborCli.DeleteWebhook(ctx, obj.Spec.Project.Name, webhook.Id); err != nil {
			return req.CheckFailed(WebhookReady, check, err.Error()).Err(nil)
		}
		webhook = nil
	}

	if webhook == nil {
		webhook, err := r.HarborCli.CreateWebhook(
			ctx, obj.Name, harbor.WebhookIn{
				Name:        r.Env.HarborWebhookName,
				Endpoint:    r.Env.HarborWebhookEndpoint,
				Events:      []harbor.Event{harbor.PushArtifact},
				AuthzSecret: r.Env.HarborWebhookAuthz,
			},
		)
		if err != nil {
			return req.CheckFailed(WebhookReady, check, err.Error())
		}
		obj.Spec.Webhook = webhook
		if err := r.Update(ctx, obj); err != nil {
			return req.CheckFailed(WebhookReady, check, err.Error())
		}
		return req.Done().RequeueAfter(100 * time.Millisecond)
	}

	check.Status = true
	if check != obj.Status.Checks[WebhookReady] {
		obj.Status.Checks[WebhookReady] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)

	builder := ctrl.NewControllerManagedBy(mgr).For(&artifactsv1.HarborProject{})
	builder.Owns(&corev1.Secret{})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
