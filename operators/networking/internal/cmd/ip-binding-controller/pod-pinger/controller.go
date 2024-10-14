package pod_pinger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"

	networkingv1 "github.com/kloudlite/operator/apis/networking/v1"
	"github.com/kloudlite/operator/operators/networking/internal/cmd/ip-binding-controller/env"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	corev1 "k8s.io/api/core/v1"
	apiLabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Env        *env.Env
	logger     *slog.Logger
	Name       string
	yamlClient kubectl.YAMLClient
}

const KloudlitePodReconcileAfter = "kloudlite.io/pod.reconcile.after"

func (r *Reconciler) GetName() string {
	return r.Name
}

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=lifecycles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=lifecycles/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=lifecycles/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	pod, err := rApi.Get(ctx, r.Client, request.NamespacedName, &corev1.Pod{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger := r.logger.With("pod", fmt.Sprintf("%s/%s", pod.GetNamespace(), pod.GetName()))

	requeue := func(err error, after ...time.Duration) (ctrl.Result, error) {
		if err != nil {
			logger.Debug("[end] reconciling, got", "err", err)
			return ctrl.Result{}, err
		}

		duration := 5 * time.Second
		if len(after) > 0 {
			duration = after[0]
		}

		logger.Debug("[end] reconciling, requeing", "after", duration)
		return ctrl.Result{RequeueAfter: duration}, nil
	}

	deletePod := func(reason string) (ctrl.Result, error) {
		logger.Info("deleting pod", "reason", reason)
		if err := r.Delete(ctx, pod); err != nil {
			return requeue(err)
		}
		return requeue(nil)
	}

	if pod.GetDeletionTimestamp() != nil {
		r.logger.Info("pod is deleting, ignoring check")
		return ctrl.Result{}, nil
	}

	logger.Debug("[start] reconciling")

	v, ok := pod.GetLabels()[constants.KloudliteGatewayEnabledLabel]
	if !ok {
		return deletePod("pod might have missed mutation webhook")
	}
	if v != "true" {
		logger.Debug("[end] pod opted out of being registered with gateway")
		return requeue(nil)
	}

	initContainerFound := false
	initContainerReady := false

	for _, v := range pod.Spec.InitContainers {
		if v.Name == "kloudlite-wg" {
			initContainerFound = true
		}
	}

	if !initContainerFound {
		return deletePod("wg init container not found")
	}

	for _, v := range pod.Status.InitContainerStatuses {
		if v.Name == "kloudlite-wg" {
			initContainerReady = v.Ready
		}
	}

	if !initContainerReady {
		logger.Info("init container is not ready, yet")
		return ctrl.Result{}, nil
	}

	ra, ok := pod.Annotations[KloudlitePodReconcileAfter]
	if !ok {
		ann := pod.GetAnnotations()

		after := 5 * time.Second

		fn.MapSet(&ann, KloudlitePodReconcileAfter, time.Now().Add(after).Format(time.RFC3339))
		pod.SetAnnotations(ann)
		if err := r.Update(ctx, pod); err != nil {
			return ctrl.Result{}, nil
		}
		return requeue(nil, after)
	}

	reconcileAfter, err := time.Parse(time.RFC3339, ra)
	if err != nil {
		return requeue(err)
	}

	if time.Now().Before(reconcileAfter) {
		diff := time.Until(reconcileAfter)
		r.logger.Debug("got here early, reconcilation scheduled", "after", int64(diff.Seconds()))
		return requeue(nil, diff)
	}

	var pblist networkingv1.PodBindingList
	if err := r.List(ctx, &pblist, &client.ListOptions{
		LabelSelector: apiLabels.SelectorFromValidatedSet(map[string]string{"kloudlite.io/podbinding.reservation": fmt.Sprintf("%s.%s", pod.GetNamespace(), pod.GetName())}),
	}); err != nil {
		return requeue(err)
	}

	if len(pblist.Items) == 0 {
		return deletePod("recreating pod, as there are no podbindings for it")
	}

	if len(pblist.Items) != 1 {
		return requeue(fmt.Errorf("multiple pod bindings with same reservation found, exiting"))
	}

	if out, err := exec.CommandContext(ctx, "timeout", "5", "ping", "-c", "1", pblist.Items[0].Spec.GlobalIP).CombinedOutput(); err != nil {
		logger.Error("failed to ping", "global-ip", pblist.Items[0].Spec.GlobalIP, "output", string(out))
		return deletePod(fmt.Sprintf("ping failed, got err: %s", err.Error()))
	}

	logger.Debug("ping success, requeing after 5s")

	return requeue(nil)
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logging.NewSlogLogger(logging.SlogOptions{
		Prefix:        r.Name,
		ShowCaller:    true,
		ShowDebugLogs: strings.ToLower(os.Getenv("LOG_DEBUG")) == "true",
	})
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig(), kubectl.YAMLClientOpts{Logger: logger.WithName(r.Name)})

	builder := ctrl.NewControllerManagedBy(mgr).For(&networkingv1.PodBinding{})
	builder.Watches(&corev1.Pod{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
		if obj.GetLabels()[constants.KloudliteGatewayEnabledLabel] != "true" {
			return nil
		}

		return []reconcile.Request{{NamespacedName: fn.NN(obj.GetNamespace(), obj.GetName())}}
	}))

	builder.Watches(&corev1.Namespace{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
		if obj.GetLabels()[constants.KloudliteGatewayEnabledLabel] != "true" {
			return nil
		}

		var podlist corev1.PodList
		if err := r.List(ctx, &podlist, client.InNamespace(obj.GetName())); err != nil {
			return nil
		}

		rr := make([]reconcile.Request, 0, len(podlist.Items))
		for _, pod := range podlist.Items {
			rr = append(rr, reconcile.Request{NamespacedName: fn.NN(pod.GetNamespace(), pod.GetName())})
		}
		return rr
	}))

	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
