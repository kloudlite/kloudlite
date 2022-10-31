package cluster

import (
	"context"
	"strings"
	"time"

	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	extensionsv1 "operators.kloudlite.io/apis/extensions/v1"
	redpandaMsvcv1 "operators.kloudlite.io/apis/redpanda.msvc/v1"
	"operators.kloudlite.io/lib/constants"
	fn "operators.kloudlite.io/lib/functions"
	"operators.kloudlite.io/lib/logging"
	rApi "operators.kloudlite.io/lib/operator"
	stepResult "operators.kloudlite.io/lib/operator/step-result"
	"operators.kloudlite.io/operators/extensions/internal/env"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
	logger logging.Logger
	Name   string
	Env    *env.Env
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	DefaultsPatched     string = "defaults-patched"
	RedpandaTopicsReady string = "redpanda-topics-ready"
	RedpandaUserReady   string = "redpanda-user-ready"
)

// +kubebuilder:rbac:groups=extensions.kloudlite.io,resources=Clusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=extensions.kloudlite.io,resources=Clusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=extensions.kloudlite.io,resources=Clusters/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, request.NamespacedName, &extensionsv1.Cluster{})
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

	if step := req.RestartIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	// TODO: initialize all checks here
	if step := req.EnsureChecks(RedpandaTopicsReady); !step.ShouldProceed() {
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

	if step := r.reconRedpandaTopics(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconRedpandaUser(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	req.Object.Status.LastReconcileTime = metav1.Time{Time: time.Now()}
	req.Logger.Infof("RECONCILATION COMPLETE")
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod * time.Second}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) finalize(req *rApi.Request[*extensionsv1.Cluster]) stepResult.Result {
	return req.Finalize()
}

func (r *Reconciler) reconDefaults(req *rApi.Request[*extensionsv1.Cluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks

	check := rApi.Check{Generation: obj.Generation}

	kTopics := strings.Split(r.Env.DefaultCreateTopics, ",")
	if len(obj.Spec.Redpanda.Topics) != len(kTopics) {
		topics := make([]string, 0, len(kTopics))
		for i := range kTopics {
			topics = append(topics, obj.Name+"-"+kTopics[i])
		}

		obj.Spec.Redpanda.Topics = topics

		if err := r.Update(ctx, obj); err != nil {
			return req.CheckFailed(DefaultsPatched, check, err.Error())
		}
		return req.Done()
	}

	check.Status = true
	if check != checks[DefaultsPatched] {
		checks[DefaultsPatched] = check
		return req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) reconRedpandaTopics(req *rApi.Request[*extensionsv1.Cluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	idLabel := "kloudite.io/topic-for-cluster"
	var topicsList redpandaMsvcv1.TopicList
	if err := r.List(
		ctx, &topicsList, &client.ListOptions{
			LabelSelector: labels.SelectorFromValidatedSet(
				map[string]string{
					idLabel: obj.Name,
				},
			),
			Namespace: obj.Namespace,
		},
	); err != nil {
		return nil
	}

	if len(topicsList.Items) != len(obj.Spec.Redpanda.Topics) {
		for _, ktopic := range obj.Spec.Redpanda.Topics {
			if err := r.Create(
				ctx, &redpandaMsvcv1.Topic{
					ObjectMeta: metav1.ObjectMeta{
						Name:            ktopic,
						Namespace:       obj.Namespace,
						Labels:          map[string]string{idLabel: obj.Name},
						OwnerReferences: []metav1.OwnerReference{fn.AsOwner(obj, true)},
					},
					Spec: redpandaMsvcv1.TopicSpec{
						AdminSecretRef: obj.Spec.Redpanda.AdminSecretRef,
					},
				},
			); err != nil {
				req.Logger.Infof(err.Error())
				if !apiErrors.IsAlreadyExists(err) {
					return req.CheckFailed(RedpandaTopicsReady, check, err.Error())
				}
			}
		}
		if check != checks[RedpandaUserReady] {
			checks[RedpandaUserReady] = check
			return req.UpdateStatus()
		}
		return req.Done().RequeueAfter(2 * time.Second)
	}

	check.Status = true
	if check != checks[RedpandaTopicsReady] {
		checks[RedpandaTopicsReady] = check
		return req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) reconRedpandaUser(req *rApi.Request[*extensionsv1.Cluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks

	check := rApi.Check{Generation: obj.Generation}

	aclUser, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Name), &redpandaMsvcv1.ACLUser{})
	if err != nil {
		req.Logger.Infof("would be creating acl user")
	}

	if aclUser == nil {
		if err := r.Create(
			ctx, &redpandaMsvcv1.ACLUser{
				ObjectMeta: metav1.ObjectMeta{
					Name:            obj.Name,
					Namespace:       obj.Namespace,
					OwnerReferences: []metav1.OwnerReference{fn.AsOwner(obj, true)},
				},
				Spec: redpandaMsvcv1.ACLUserSpec{
					AdminSecretRef: obj.Spec.Redpanda.AdminSecretRef,
					Topics:         append(obj.Spec.Redpanda.Topics, obj.Spec.Redpanda.ExtraTopicsWithACL...),
				},
			},
		); err != nil {
			return req.CheckFailed(RedpandaUserReady, check, err.Error())
		}
	}

	check.Status = true
	if check != checks[RedpandaUserReady] {
		checks[RedpandaUserReady] = check
		return req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)

	builder := ctrl.NewControllerManagedBy(mgr).For(&extensionsv1.Cluster{})
	builder.Owns(&redpandaMsvcv1.Topic{})
	builder.Owns(&redpandaMsvcv1.ACLUser{})
	// builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
