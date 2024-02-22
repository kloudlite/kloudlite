package secret_replicator

import (
	"context"
	"strings"
	"time"

	"github.com/kloudlite/operator/operators/config-secret-replicator/internal/env"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiLabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Env        *env.Env
	logger     logging.Logger
	Name       string
	yamlClient kubectl.YAMLClient
	recorder   record.EventRecorder
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	SecretReplicatorFinalizer = "kloudlite.io/finalizer-secret-replicator"
)

// +kubebuilder:rbac:groups=k8s.io/core/v1,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=k8s.io/core/v1,resources=secrets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=k8s.io/core/v1,resources=secrets/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	r.logger.Debugf("starting reconcile for %s", request.NamespacedName)
	secret, err := rApi.Get(ctx, r.Client, request.NamespacedName, &corev1.Secret{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if _, ok := secret.GetAnnotations()[constants.ReplicationEnableKey]; !ok {
		// ignoring, as secret is not set to be replicated
		return ctrl.Result{}, nil
	}

	if secret.GetDeletionTimestamp() != nil {
		if x := r.finalize(ctx, secret); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	if step := r.replicatesecret(ctx, secret); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	// req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *Reconciler) finalize(ctx context.Context, secret *corev1.Secret) stepResult.Result {
	r.logger.Debugf("[check:START]finalizing secret %s/%s", secret.Namespace, secret.Name)
	if step := r.dereplicatesecret(ctx, secret); !step.ShouldProceed() {
		return step
	}

	if updated := controllerutil.RemoveFinalizer(secret, SecretReplicatorFinalizer); updated {
		if err := r.Update(ctx, secret); err != nil {
			return stepResult.New().RequeueAfter(500 * time.Millisecond)
		}
	}

	r.logger.Debugf("[check:END] finalized secret %s/%s", secret.Namespace, secret.Name)
	return stepResult.New().Continue(true)
}

func (r *Reconciler) replicatesecret(ctx context.Context, secret *corev1.Secret) stepResult.Result {
	fail := func(err error) stepResult.Result {
		// r.recorder.Event(secret, "Warning", "SecretReplicator", err.Error())
		return stepResult.New().Err(err)
	}

	r.logger.Infof("[check:START] Replicating secret %s/%s", secret.Namespace, secret.Name)

	v, ok := secret.GetAnnotations()[constants.ReplicationEnableKey]
	if ok && v != constants.ReplicationEnableValueTrue {
		return r.dereplicatesecret(ctx, secret)
	}

	if updated := controllerutil.AddFinalizer(secret, SecretReplicatorFinalizer); updated {
		if err := r.Update(ctx, secret); err != nil {
			return stepResult.New().RequeueAfter(500 * time.Millisecond)
		}
	}

	var nslist corev1.NamespaceList
	if err := r.List(ctx, &nslist); err != nil {
		return fail(err)
	}

	excludeNamespaces := map[string]struct{}{}
	if v, ok := secret.GetAnnotations()[constants.ReplicationExcludeNsKey]; ok {
		for _, s := range strings.Split(v, ",") {
			excludeNamespaces[s] = struct{}{}
		}
	}

	for i := range nslist.Items {
		namespace := nslist.Items[i].GetName()

		if _, ok := excludeNamespaces[namespace]; ok {
			continue
		}

		if namespace == secret.Namespace {
			continue
		}

		scrt := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: secret.Name, Namespace: namespace}, Type: secret.Type}
		if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, scrt, func() error {
			if scrt.Labels == nil {
				scrt.Labels = make(map[string]string, 2)
			}
			scrt.Labels[constants.ReplicationFromNameKey] = secret.Name
			scrt.Labels[constants.ReplicationFromNamespaceKey] = secret.Namespace

			scrt.Data = secret.Data
			return nil
		}); err != nil {
			return fail(err)
		}
	}

	r.logger.Infof("[check:END] Replicated secret %s/%s", secret.Namespace, secret.Name)
	return stepResult.New().Continue(true)
}

func (r *Reconciler) dereplicatesecret(ctx context.Context, secret *corev1.Secret) stepResult.Result {
	fail := func(err error) stepResult.Result {
		// r.recorder.Event(secret, "Warning", "secretReplicator", err.Error())
		return stepResult.New().Err(err)
	}

	r.logger.Infof("DeReplicating secret %s/%s", secret.Namespace, secret.Name)

	var cfglist corev1.SecretList
	if err := r.List(ctx, &cfglist, &client.ListOptions{
		LabelSelector: apiLabels.SelectorFromValidatedSet(map[string]string{
			constants.ReplicationFromNameKey:      secret.Name,
			constants.ReplicationFromNamespaceKey: secret.Namespace,
		}),
	}); err != nil {
		return fail(err)
	}

	cfgmaps := make([]*corev1.Secret, len(cfglist.Items))
	for i := range cfglist.Items {
		cfgmaps[i] = &cfglist.Items[i]
	}

	if err := fn.DeleteAndWait(ctx, r.logger, r.Client, cfgmaps...); err != nil {
		return fail(err)
	}

	r.logger.Infof("DeReplicated secret %s/%s", secret.Namespace, secret.Name)

	return stepResult.New().Continue(true)
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig(), kubectl.YAMLClientOpts{Logger: r.logger})
	r.recorder = mgr.GetEventRecorderFor(r.GetName())

	builder := ctrl.NewControllerManagedBy(mgr).For(&corev1.Node{}).Named(r.Name)
	// builder := ctrl.NewControllerManagedBy(mgr).Named(r.Name)
	builder.Watches(&corev1.Secret{}, handler.EnqueueRequestsFromMapFunc(func(_ context.Context, obj client.Object) []reconcile.Request {
		if _, ok := obj.GetAnnotations()[constants.ReplicationEnableKey]; ok {
			return []reconcile.Request{{NamespacedName: fn.NN(obj.GetNamespace(), obj.GetName())}}
		}

		name, hasName := obj.GetLabels()[constants.ReplicationFromNameKey]
		namespace, hasNamespace := obj.GetLabels()[constants.ReplicationFromNamespaceKey]

		if hasName && hasNamespace {
			return []reconcile.Request{{NamespacedName: fn.NN(namespace, name)}}
		}

		return nil
	}))

	builder.WithOptions(controller.Options{
		MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles,
		RateLimiter:             workqueue.NewItemFastSlowRateLimiter(100*time.Millisecond, 500*time.Millisecond, 10),
	})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
