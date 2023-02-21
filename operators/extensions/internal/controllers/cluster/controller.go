package cluster

import (
	"context"
	"fmt"
	"time"

	ct "github.com/kloudlite/operator/apis/common-types"
	extensionsv1 "github.com/kloudlite/operator/apis/extensions/v1"
	redpandaMsvcv1 "github.com/kloudlite/operator/apis/redpanda.msvc/v1"
	"github.com/kloudlite/operator/operators/extensions/internal/env"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
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
	RedpandaTopicsReady          string = "redpanda-topics-ready"
	RedpandaUserReady            string = "redpanda-user-ready"
	AggregatedClusterKubeConfigs string = "accumulated-cluster-kube-configs"
	GrafanaReady                 string = "grafana-ready"
)

const KloudliteNS string = "kl-core"

// +kubebuilder:rbac:groups=extensions.kloudlite.io,resources=Clusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=extensions.kloudlite.io,resources=Clusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=extensions.kloudlite.io,resources=Clusters/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &extensionsv1.Cluster{})
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
	defer func() {
		req.Logger.Infof("RECONCILATION COMPLETE (isReady=%v)", req.Object.Status.IsReady)
	}()

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureChecks(RedpandaTopicsReady, RedpandaUserReady, AggregatedClusterKubeConfigs); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconRedpandaTopics(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.aggregateClusterKubeconfigs(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	req.Object.Status.LastReconcileTime = &metav1.Time{Time: time.Now()}
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) finalize(req *rApi.Request[*extensionsv1.Cluster]) stepResult.Result {
	return req.Finalize()
}

func (r *Reconciler) reconRedpandaTopics(req *rApi.Request[*extensionsv1.Cluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	clusterTopics := []string{
		fmt.Sprintf("%s-incoming", obj.Name),
	}

	for _, topic := range clusterTopics {
		_, err := rApi.Get(ctx, r.Client, fn.NN(KloudliteNS, topic), &redpandaMsvcv1.Topic{})
		if err != nil {
			if apiErrors.IsNotFound(err) {
				if err := r.Create(
					ctx, &redpandaMsvcv1.Topic{
						ObjectMeta: metav1.ObjectMeta{
							Name:            topic,
							Namespace:       KloudliteNS,
							Annotations:     map[string]string{"kloudlite.io/created-by-cluster": obj.Name},
							OwnerReferences: []metav1.OwnerReference{fn.AsOwner(obj, true)},
						},
						Spec: redpandaMsvcv1.TopicSpec{
							AdminSecretRef: ct.SecretRef{
								Name:      r.Env.RedpandaSecretName,
								Namespace: r.Env.RedpandaSecretNamespace,
							},
						},
					},
				); err != nil {
					return req.CheckFailed(RedpandaTopicsReady, check, err.Error())
				}
				return req.Done().RequeueAfter(1 * time.Second)
			}
			return req.CheckFailed(RedpandaUserReady, check, err.Error()).Err(nil)
		}
	}

	check.Status = true
	if check != checks[RedpandaTopicsReady] {
		checks[RedpandaTopicsReady] = check
		return req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) aggregateClusterKubeconfigs(req *rApi.Request[*extensionsv1.Cluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	var kubeConfigsList corev1.SecretList
	if err := r.List(
		ctx, &kubeConfigsList, &client.ListOptions{
			LabelSelector: labels.SelectorFromValidatedSet(
				map[string]string{"kloudlite.io/cluster-config": "true"},
			),
			Namespace: KloudliteNS,
		},
	); err != nil {
		return req.CheckFailed(AggregatedClusterKubeConfigs, check, err.Error()).Err(nil)
	}

	secretName := "aggregated-kubeconfigs"
	var scrt corev1.Secret
	if err := r.Get(ctx, fn.NN(KloudliteNS, secretName), &scrt); err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.CheckFailed(AggregatedClusterKubeConfigs, check, err.Error()).Err(nil)
		}
		scrt.SetNamespace(KloudliteNS)
		scrt.SetName(secretName)
		scrt.SetLabels(map[string]string{"kloudlite.io/aggregated-kubeconfig": "true"})
	}

	if len(scrt.Data) < len(kubeConfigsList.Items) {
		if _, err := controllerutil.CreateOrUpdate(
			ctx, r.Client, &scrt, func() error {
				for _, item := range kubeConfigsList.Items {
					if scrt.Data == nil {
						scrt.Data = map[string][]byte{}
					}
					scrt.Data[item.Name] = item.Data["kubeConfig"]
				}
				return nil
			},
		); err != nil {
			return req.CheckFailed(AggregatedClusterKubeConfigs, check, err.Error()).Err(nil)
		}
	}

	check.Status = true
	if check != checks[AggregatedClusterKubeConfigs] {
		checks[AggregatedClusterKubeConfigs] = check
		return req.UpdateStatus()
	}
	return req.Next()
}

// TODO: (nxtcoder17), need to add RedpandaACL that takes control over created redpanda topics, (as of now `admin` works)

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)

	builder := ctrl.NewControllerManagedBy(mgr).For(&extensionsv1.Cluster{})
	builder.Owns(&redpandaMsvcv1.Topic{})
	builder.Owns(&redpandaMsvcv1.ACLUser{})
	builder.Watches(
		&source.Kind{Type: &corev1.Secret{}}, handler.EnqueueRequestsFromMapFunc(
			func(obj client.Object) []reconcile.Request {
				if obj.GetLabels()["kloudlite.io/cluster-config"] == "true" {
					return []reconcile.Request{{fn.NN(obj.GetNamespace(), obj.GetName())}}
				}
				return nil
			},
		),
	)

	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
