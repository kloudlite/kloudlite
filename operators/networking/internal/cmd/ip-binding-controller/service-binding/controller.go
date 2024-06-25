package service_binding

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	networkingv1 "github.com/kloudlite/operator/apis/networking/v1"
	"github.com/kloudlite/operator/operators/networking/internal/cmd/ip-binding-controller/env"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	corev1 "k8s.io/api/core/v1"
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
  logger := r.logger.With("service", request.String())

  logger.Debug("1. starts reconciling")
  svc, err := rApi.Get(ctx, r.Client, request.NamespacedName, &corev1.Service{})
  if err != nil {
    return ctrl.Result{}, client.IgnoreNotFound(err)
  }

	if svc.GetDeletionTimestamp() != nil {
		return ctrl.Result{}, nil
	}

  logger.Debug("2. checking whether the service is a cluster IP service ?")
	if svc.Spec.Type != corev1.ServiceTypeClusterIP {
		return ctrl.Result{}, nil
	}

  logger.Debug("3. fetching service binding")
  svcBinding, err := rApi.Get(ctx, r.Client, fn.NN(svc.GetNamespace(), strings.ReplaceAll(svc.GetLabels()["kloudlite.io/servicebinding.ip"], ".", "-")), &networkingv1.ServiceBinding{})
  if err != nil {
    return ctrl.Result{}, client.IgnoreNotFound(err)
  }

  logger.Debug("4. updating service binding with ports, and service IP")
	svcBinding.Spec.ServiceIP = &svc.Spec.ClusterIP
	svcBinding.Spec.Ports = svc.Spec.Ports

	svcHost := fmt.Sprintf("%s.%s.%s", svc.Name, svc.Namespace, r.Env.GatewayDNSSuffix)

	ann := svcBinding.GetAnnotations()
	if ann == nil {
		ann = make(map[string]string, 1)
	}
	ann["kloudlite.io/global.hostname"] = svcHost
	svcBinding.SetAnnotations(ann)

	if err := r.Update(ctx, svcBinding); err != nil {
	  return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	r2, err := http.NewRequestWithContext(ctx, http.MethodPut, fmt.Sprintf("%s/service/%s", r.Env.GatewayAdminApiAddr, svcBinding.Name), nil)
	if err != nil {
	  return ctrl.Result{}, err
	}

  r.logger.Debug("5. registering this service", "url", r2.URL.String())
	if _, err := http.DefaultClient.Do(r2); err != nil {
	  return ctrl.Result{}, err
	}
  r.logger.Debug("FINISH: registering this service", "url", r2.URL.String())

	svcDnsReq, err := http.NewRequestWithContext(ctx, http.MethodPut, fmt.Sprintf("%s/service/%s/%s", r.Env.ServiceDNSHttpAddr, svcHost, svcBinding.Spec.GlobalIP), nil)
	if err != nil {
	  return ctrl.Result{}, err
	}

  r.logger.Debug("START: updating DNS records for service", "url", svcDnsReq.URL.String())
	if _, err := http.DefaultClient.Do(svcDnsReq); err != nil {
	  return ctrl.Result{}, err
	}
  r.logger.Debug("FINISH: updated service DNS for service", "url", svcDnsReq.URL.String())

  r.logger.Info("service successfully reconciled", "service-binding", svcBinding.Name)
  return ctrl.Result{}, nil
}

func (r *Reconciler) finalize(req *rApi.Request[*networkingv1.ServiceBinding]) stepResult.Result {
	rApi.NewRunningCheck("finalizing", req)
	return req.Finalize()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logging.NewSlogLogger(logging.SlogOptions{
		Prefix:        r.Name,
		ShowTimestamp: false,
		ShowCaller:    true,
		LogLevel:      slog.LevelInfo,
	})
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig(), kubectl.YAMLClientOpts{Logger: logger})

	builder := ctrl.NewControllerManagedBy(mgr).For(&networkingv1.ServiceBinding{})
	builder.Watches(&corev1.Service{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
	  for k := range obj.GetLabels() {
	    if strings.HasPrefix(k, "kloudlite.io/") {
	      return []reconcile.Request{{NamespacedName: fn.NN(obj.GetNamespace(), obj.GetName())}}
	    }
	  }
		return nil
	}))

	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
