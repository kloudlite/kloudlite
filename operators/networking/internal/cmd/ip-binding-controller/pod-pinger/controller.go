package pod_pinger

import (
	"context"
	"strings"
	"time"

	"github.com/kloudlite/operator/operators/networking/internal/cmd/ip-binding-controller/env"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	corev1 "k8s.io/api/core/v1"
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

	v, ok := pod.GetLabels()["kloudlite.io/podbinding.ip"]
	if !ok {
		// kill this POD
		if err := r.Delete(ctx, pod); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	pinger, err := probing.NewPinger(v)
	if err != nil {
		r.logger.Errorf(err, "failed to create pinger for %s", v)
		return ctrl.Result{}, err
	}
	pinger.Count = 1
	pinger.Timeout = 500 * time.Millisecond
	err = pinger.Run() // Blocks until finished.
	if err != nil {
		r.logger.Errorf(err, "failed to ping %s", v)
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

	builder := ctrl.NewControllerManagedBy(mgr).For(&corev1.Namespace{})
	builder.Watches(&corev1.Pod{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
		if strings.HasPrefix(obj.GetNamespace(), "env-") {
			return []reconcile.Request{{NamespacedName: fn.NN(obj.GetNamespace(), obj.GetName())}}
		}
		return nil
	}))

	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
