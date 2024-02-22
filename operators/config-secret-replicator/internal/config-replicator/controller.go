package config_replicator

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kloudlite/operator/operators/config-secret-replicator/internal/env"
	"github.com/kloudlite/operator/pkg/constants"
	"github.com/kloudlite/operator/pkg/errors"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
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
	ConfigmapReplicatorFinalizer = "kloudlite.io/finalizer-configmap-replicator"
)

// +kubebuilder:rbac:groups=k8s.io/core/v1,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=k8s.io/core/v1,resources=configmaps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=k8s.io/core/v1,resources=configmaps/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	r.logger.Debugf("starting reconcile for %s", request.NamespacedName)
	configmap, err := rApi.Get(ctx, r.Client, request.NamespacedName, &corev1.ConfigMap{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if _, ok := configmap.GetAnnotations()[constants.ReplicationEnableKey]; !ok {
		// ignoring, as configmap is not set to be replicated
		return ctrl.Result{}, nil
	}

	if configmap.GetDeletionTimestamp() != nil {
		if x := r.finalize(ctx, configmap); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	// if updated := controllerutil.AddFinalizer(configmap, ConfigmapReplicatorFinalizer); updated {
	// 	if err := r.Update(ctx, configmap); err != nil {
	// 		return ctrl.Result{RequeueAfter: 500 * time.Millisecond}, nil
	// 	}
	// }

	if step := r.replicateConfigmap(ctx, configmap); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	// req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *Reconciler) finalize(ctx context.Context, configmap *corev1.ConfigMap) stepResult.Result {
	if step := r.dereplicateConfigmap(ctx, configmap); !step.ShouldProceed() {
		return step
	}

	if updated := controllerutil.RemoveFinalizer(configmap, ConfigmapReplicatorFinalizer); updated {
		if err := r.Update(ctx, configmap); err != nil {
			return stepResult.New().RequeueAfter(500 * time.Millisecond)
		}
	}

	return stepResult.New().Continue(true)
}

func (r *Reconciler) replicateConfigmap(ctx context.Context, configmap *corev1.ConfigMap) stepResult.Result {
	fail := func(err error) stepResult.Result {
		// r.recorder.Event(configmap, "Warning", "ConfigmapReplicator", err.Error())
		return stepResult.New().Err(err)
	}
	r.logger.Infof("[check:START] Replicating configmap %s/%s", configmap.Namespace, configmap.Name)

	v, ok := configmap.GetAnnotations()[constants.ReplicationEnableKey]
	if ok && v != constants.ReplicationEnableValueTrue {
		return r.dereplicateConfigmap(ctx, configmap)
	}

	if updated := controllerutil.AddFinalizer(configmap, ConfigmapReplicatorFinalizer); updated {
		if err := r.Update(ctx, configmap); err != nil {
			return stepResult.New().RequeueAfter(500 * time.Millisecond)
			// return ctrl.Result{}, err
			// return ctrl.Result{RequeueAfter: 500 * time.Millisecond}, nil
		}
	}

	var nslist corev1.NamespaceList
	if err := r.List(ctx, &nslist); err != nil {
		return fail(err)
	}

	excludeNamespaces := map[string]struct{}{}
	if v, ok := configmap.GetAnnotations()[constants.ReplicationExcludeNsKey]; ok {
		for _, s := range strings.Split(v, ",") {
			excludeNamespaces[s] = struct{}{}
		}
	}

	for i := range nslist.Items {
		namespace := nslist.Items[i].GetName()

		if _, ok := excludeNamespaces[namespace]; ok {
			continue
		}

		if namespace == configmap.Namespace {
			continue
		}

		cm, err := rApi.Get(ctx, r.Client, fn.NN(namespace, configmap.Name), &corev1.ConfigMap{})
		if err != nil {
			if !apiErrors.IsNotFound(err) {
				return fail(err)
			}

			if err := r.Create(ctx, &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      configmap.Name,
					Namespace: namespace,
					Labels: map[string]string{
						constants.ReplicationFromNameKey:      configmap.Name,
						constants.ReplicationFromNamespaceKey: configmap.Namespace,
					},
				},
				Data:       configmap.Data,
				BinaryData: configmap.BinaryData,
			}); err != nil {
				return fail(errors.NewEf(err, "replicating secret"))
			}
		}

		if cm != nil && (cm.Labels[constants.ReplicationFromNameKey] != configmap.Name || cm.Labels[constants.ReplicationFromNamespaceKey] != configmap.Namespace) {
			return fail(fmt.Errorf("configmap %s, already exists in namespace %s, will not replicate, as it would cause data loss", configmap.Name, namespace)).Err(nil)
		}
	}

	r.logger.Infof("[check:END] Replicated configmap %s/%s", configmap.Namespace, configmap.Name)
	return stepResult.New().Continue(true)
}

func (r *Reconciler) dereplicateConfigmap(ctx context.Context, configmap *corev1.ConfigMap) stepResult.Result {
	fail := func(err error) stepResult.Result {
		// r.recorder.Event(configmap, "Warning", "ConfigmapReplicator", err.Error())
		return stepResult.New().Err(err)
	}

	r.logger.Infof("[check:START] DeReplicating configmap %s/%s", configmap.Namespace, configmap.Name)

	var cfglist corev1.ConfigMapList
	if err := r.List(ctx, &cfglist, &client.ListOptions{
		LabelSelector: apiLabels.SelectorFromValidatedSet(map[string]string{
			constants.ReplicationFromNameKey:      configmap.Name,
			constants.ReplicationFromNamespaceKey: configmap.Namespace,
		}),
	}); err != nil {
		return fail(err)
	}

	cfgmaps := make([]*corev1.ConfigMap, len(cfglist.Items))
	for i := range cfglist.Items {
		cfgmaps[i] = &cfglist.Items[i]
	}

	if err := fn.DeleteAndWait(ctx, r.logger, r.Client, cfgmaps...); err != nil {
		return fail(err)
	}

	r.logger.Infof("[check:END] DeReplicated configmap %s/%s", configmap.Namespace, configmap.Name)

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
	builder.Watches(&corev1.ConfigMap{}, handler.EnqueueRequestsFromMapFunc(func(_ context.Context, obj client.Object) []reconcile.Request {
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
