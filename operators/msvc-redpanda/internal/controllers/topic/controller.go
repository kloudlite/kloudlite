package topic

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	redpandaMsvcv1 "github.com/kloudlite/operator/apis/redpanda.msvc/v1"
	"github.com/kloudlite/operator/operators/msvc-redpanda/internal/env"
	"github.com/kloudlite/operator/operators/msvc-redpanda/internal/types"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	"github.com/kloudlite/operator/pkg/redpanda"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	logger     logging.Logger
	Name       string
	Env        *env.Env
	yamlClient *kubectl.YAMLClient
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	RedpandaTopicReady string = "redpanda-topic-ready"
	DefaultsPatched    string = "defaults-patched"
)

// +kubebuilder:rbac:groups=redpanda.msvc.kloudlite.io,resources=topics,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=redpanda.msvc.kloudlite.io,resources=topics/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=redpanda.msvc.kloudlite.io,resources=topics/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, request.NamespacedName, &redpandaMsvcv1.Topic{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	req.Logger.Infof("NEW RECONCILATION")

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureChecks(DefaultsPatched, RedpandaTopicReady); !step.ShouldProceed() {
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

	if step := r.reconRedpandaTopic(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	req.Object.Status.LastReconcileTime = metav1.Time{Time: time.Now()}
	req.Logger.Infof("RECONCILATION COMPLETE")
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) reconDefaults(req *rApi.Request[*redpandaMsvcv1.Topic]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	if obj.Spec.AdminSecretRef.Namespace == "" {
		obj.Spec.AdminSecretRef.Namespace = obj.Namespace
		err := r.Update(ctx, obj)
		return req.Done().RequeueAfter(1 * time.Second).Err(err)
	}

	check.Status = true
	if check != checks[DefaultsPatched] {
		checks[DefaultsPatched] = check
		return req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) finalize(req *rApi.Request[*redpandaMsvcv1.Topic]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	topicDeleted := "topic-deleted"

	req.Logger.Infof("deleting topic")
	defer func() {
		if checks[topicDeleted].Status {
			req.Logger.Infof("topic deletion successfull")
		}
		req.Logger.Infof("still ... deleting topic")
	}()

	if step := req.EnsureChecks(topicDeleted); !step.ShouldProceed() {
		return step
	}

	check := rApi.Check{Generation: obj.Generation}

	adminScrt, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.AdminSecretRef.Namespace, obj.Spec.AdminSecretRef.Name), &corev1.Secret{})
	if err != nil {
		if apiErrors.IsNotFound(err) {
			return req.Finalize()
		}
		return req.CheckFailed(RedpandaTopicReady, check, err.Error()).Err(nil)
	}

	adminCreds, err := fn.ParseFromSecret[types.AdminUserCreds](adminScrt)
	if err != nil {
		return req.CheckFailed(RedpandaTopicReady, check, err.Error()).Err(nil)
	}

	adminCli := redpanda.NewAdminClient(adminCreds.Username, adminCreds.Password, adminCreds.KafkaBrokers, adminCreds.AdminEndpoint)

	if err := adminCli.DeleteTopic(obj.Name); err != nil {
		return req.CheckFailed(topicDeleted, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != checks[topicDeleted] {
		checks[topicDeleted] = check
		return req.UpdateStatus()
	}
	return req.Finalize()
}

func (r *Reconciler) reconRedpandaTopic(req *rApi.Request[*redpandaMsvcv1.Topic]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	adminScrt, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.AdminSecretRef.Namespace, obj.Spec.AdminSecretRef.Name), &corev1.Secret{})
	if err != nil {
		return req.CheckFailed(RedpandaTopicReady, check, err.Error()).Err(nil)
	}

	adminCreds, err := fn.ParseFromSecret[types.AdminUserCreds](adminScrt)
	if err != nil {
		return req.CheckFailed(RedpandaTopicReady, check, err.Error()).Err(nil)
	}

	adminCli := redpanda.NewAdminClient(adminCreds.Username, adminCreds.Password, adminCreds.KafkaBrokers, adminCreds.AdminEndpoint)

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
		return req.Done().RequeueAfter(2 * time.Second)
	}

	check.Status = true
	if check != checks[RedpandaTopicReady] {
		checks[RedpandaTopicReady] = check
		return req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

	//for _, topic := range strings.Split(r.Env.MustHaveTopics, ",") {
	//	name := strings.TrimSpace(topic)
	//	if err := r.Client.Create(
	//		context.TODO(), &redpandaMsvcv1.Topic{
	//			ObjectMeta: metav1.ObjectMeta{
	//				Name:      name,
	//				Namespace: "kl-core",
	//			},
	//			Spec: redpandaMsvcv1.TopicSpec{
	//				AdminSecretRef: ct.SecretRef{
	//					Name:      r.Env.AdminSecretName,
	//					Namespace: r.Env.AdminSecretNamespace,
	//				},
	//				PartitionCount: 3,
	//			},
	//		},
	//	); err != nil {
	//		if !apiErrors.IsAlreadyExists(err) {
	//			return err
	//		}
	//	}
	//}
	//
	//r.logger.Infof("ensured must have topics exists in the cluster")

	builder := ctrl.NewControllerManagedBy(mgr).For(&redpandaMsvcv1.Topic{})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
