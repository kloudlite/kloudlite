package pod_pinger

import (
	"context"
	"fmt"
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

	probing "github.com/prometheus-community/pro-bing"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Env        *env.Env
	logger     logging.Logger
	Name       string
	yamlClient kubectl.YAMLClient
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	bindService string = "bind-service"
)

func getJobSvcAccountName() string {
	return "job-svc-account"
}

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=lifecycles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=lifecycles/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=lifecycles/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	pod, err := rApi.Get(ctx, r.Client, request.NamespacedName, &corev1.Pod{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	v, ok := pod.GetLabels()[constants.KloudliteGatewayEnabledLabel]
	if !ok {
		if err := r.Delete(ctx, pod); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}
	if v != "true" {
		return ctrl.Result{}, nil
	}

	var pblist networkingv1.PodBindingList
	if err := r.List(ctx, &pblist, &client.ListOptions{
		LabelSelector: apiLabels.SelectorFromValidatedSet(map[string]string{"kloudlite.io/podbinding.reservation": fmt.Sprintf("%s.%s", pod.GetNamespace(), pod.GetName())}),
	}); err != nil {
		return ctrl.Result{}, err
	}

	if len(pblist.Items) != 1 {
		return ctrl.Result{}, nil
	}

	pinger, err := probing.NewPinger(pblist.Items[0].Spec.GlobalIP)
	if err != nil {
		r.logger.Errorf(err, "failed to create pinger for %s", v)
		return ctrl.Result{}, err
	}
	pinger.Count = 1
	pinger.Timeout = 500 * time.Millisecond
	if err = pinger.RunWithContext(ctx); err != nil {
		r.logger.Errorf(err, "failed to ping %s", v)
		if err := r.Delete(ctx, pod); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	}

	r.logger.Debugf("ping success for %s, requeing after 5s", v)
	return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig(), kubectl.YAMLClientOpts{Logger: r.logger})

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
