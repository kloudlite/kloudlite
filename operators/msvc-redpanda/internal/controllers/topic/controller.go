package topic

import (
	"context"
	"github.com/kloudlite/operator/operators/msvc-redpanda/internal/types"
	"time"

	redpandaMsvcv1 "github.com/kloudlite/operator/apis/redpanda.msvc/v1"
	"github.com/kloudlite/operator/operators/msvc-redpanda/internal/env"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	"github.com/kloudlite/operator/pkg/redpanda"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Logger     logging.Logger
	Name       string
	Env        *env.Env
	yamlClient *kubectl.YAMLClient
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	RedpandaTopicReady string = "redpanda-topic-ready"
)

// +kubebuilder:rbac:groups=redpanda.msvc.kloudlite.io,resources=topics,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=redpanda.msvc.kloudlite.io,resources=topics/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=redpanda.msvc.kloudlite.io,resources=topics/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.Logger), r.Client, request.NamespacedName, &redpandaMsvcv1.Topic{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	req.LogPreReconcile()
	defer req.LogPostReconcile()

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconRedpandaTopic(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	req.Object.Status.LastReconcileTime = &metav1.Time{Time: time.Now()}
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) finalize(req *rApi.Request[*redpandaMsvcv1.Topic]) stepResult.Result {
	obj, checks := req.Object, req.Object.Status.Checks
	topicDeleted := "topic-deleted"
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(topicDeleted)
	defer req.LogPostCheck(topicDeleted)

	adminCli, err := r.getAdminClient(req)
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.CheckFailed(topicDeleted, check, err.Error())
		}
		return req.Finalize()
	}

	if err := adminCli.DeleteTopic(obj.Name); err != nil {
		return req.CheckFailed(topicDeleted, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != checks[topicDeleted] {
		checks[topicDeleted] = check
		req.UpdateStatus()
	}
	return req.Finalize()
}

func (r *Reconciler) getAdminClient(req *rApi.Request[*redpandaMsvcv1.Topic]) (redpanda.AdminClient, error) {
	ctx, obj := req.Context(), req.Object

	var rpkAdmin redpandaMsvcv1.Admin
	if err := r.Get(ctx, fn.NN("", obj.Spec.RedpandaAdmin), &rpkAdmin); err != nil {
		return nil, err
	}

	if rpkAdmin.Spec.AuthFlags == nil || !rpkAdmin.Spec.AuthFlags.Enabled {
		return redpanda.NewAdminClient(rpkAdmin.Spec.AdminEndpoint, rpkAdmin.Spec.KafkaBrokers, nil), nil
	}

	adminScrt, err := rApi.Get(ctx, r.Client, fn.NN(rpkAdmin.Spec.AuthFlags.TargetSecret.Namespace, rpkAdmin.Spec.AuthFlags.TargetSecret.Name), &corev1.Secret{})
	if err != nil {
		return nil, err
	}

	adminCreds, err := fn.ParseFromSecret[types.AdminUserCreds](adminScrt)
	if err != nil {
		return nil, err
	}

	return redpanda.NewAdminClient(rpkAdmin.Spec.AdminEndpoint, rpkAdmin.Spec.KafkaBrokers, &redpanda.AdminAuthOpts{
		Username: adminCreds.Username,
		Password: adminCreds.Password,
	}), nil
}

func (r *Reconciler) reconRedpandaTopic(req *rApi.Request[*redpandaMsvcv1.Topic]) stepResult.Result {
	obj, checks := req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(RedpandaTopicReady)
	defer req.LogPostCheck(RedpandaTopicReady)

	adminCli, err := r.getAdminClient(req)
	if err != nil {
		return req.CheckFailed(RedpandaTopicReady, check, err.Error()).Err(nil)
	}

	tExists, err := adminCli.TopicExists(obj.Name)
	if err != nil {
		req.Logger.Infof("will be creating now")
		return req.CheckFailed(RedpandaTopicReady, check, err.Error())
	}

	if !tExists {
		if err := adminCli.CreateTopic(obj.Name, obj.Spec.PartitionCount); err != nil {
			return req.CheckFailed(RedpandaTopicReady, check, err.Error())
		}
		checks[RedpandaTopicReady] = check
		return req.Done().RequeueAfter(500 * time.Millisecond)
	}

	check.Status = true
	if check != checks[RedpandaTopicReady] {
		checks[RedpandaTopicReady] = check
		req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.Logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

	builder := ctrl.NewControllerManagedBy(mgr).For(&redpandaMsvcv1.Topic{})
	builder.Owns(&corev1.Secret{})

	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
