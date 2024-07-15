package service_binding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	networkingv1 "github.com/kloudlite/operator/apis/networking/v1"
	"github.com/kloudlite/operator/operators/networking/internal/cmd/ip-binding-controller/env"
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

	logger.Info("starts reconciling")
	defer func() {
		logger.Info("finished reconciling")
	}()

	if svc.GetDeletionTimestamp() != nil {
		return ctrl.Result{}, nil
	}

	logger.Debug("2. checking whether the service is a cluster IP service ?")
	if svc.Spec.Type != corev1.ServiceTypeClusterIP {
		return ctrl.Result{}, nil
	}

	logger.Debug("3. fetching service binding")

	var sblist networkingv1.ServiceBindingList
	if err := r.List(ctx, &sblist, &client.ListOptions{
		Limit: 1,
		LabelSelector: apiLabels.SelectorFromValidatedSet(map[string]string{
			"kloudlite.io/servicebinding.reservation": fmt.Sprintf("%s.%s", svc.GetNamespace(), svc.GetName()),
		}),
	}); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if len(sblist.Items) == 0 {
		logger.Info("service binding not found, re-triggering service")

		b, err := json.Marshal(svc.Spec.Ports)
		if err != nil {
			return ctrl.Result{}, err
		}

		createSvcBindingReq, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/service/%s/%s", r.Env.GatewayAdminApiAddr, svc.GetNamespace(), svc.GetName()), bytes.NewReader(b))
		if err != nil {
			return ctrl.Result{}, err
		}

		resp, err := http.DefaultClient.Do(createSvcBindingReq)
		if err != nil {
			return ctrl.Result{}, err
		}

		if resp.StatusCode != http.StatusOK {
			return ctrl.Result{}, fmt.Errorf("unexpected response from ip-manager, got=%d, expected=%d", resp.StatusCode, http.StatusOK)
		}

		return ctrl.Result{RequeueAfter: 1 * time.Second}, nil
		// return r.retriggerService(ctx, svc)
	}

	svcBinding := &sblist.Items[0]

	logger.Debug("4. updating service binding with ports, and service IP")
	svcBinding.Spec.ServiceIP = &svc.Spec.ClusterIP
	svcBinding.Spec.Ports = svc.Spec.Ports

	ns, err := rApi.Get(ctx, r.Client, fn.NN("", svc.GetNamespace()), &corev1.Namespace{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	klDNSHostname := func() string {
		if v, ok := svc.GetLabels()[constants.KloudliteDNSHostname]; ok {
			return v
		}

		if v, ok := ns.Labels[constants.KloudliteNamespaceForEnvironment]; ok {
			return v
		}

		if v, ok := ns.Labels[constants.KloudliteNamespaceForClusterManagedService]; ok {
			return v
		}

		return ""
	}()

	if klDNSHostname != "" {
		svcBinding.Spec.Hostname = klDNSHostname
	}

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

	svcBinding.Status.IsReady = true
	svcBinding.Status.LastReconcileTime = &metav1.Time{Time: time.Now()}
	if err := r.Status().Update(ctx, svcBinding); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

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

func (r *Reconciler) retriggerService(ctx context.Context, svc *corev1.Service) (ctrl.Result, error) {
	r.logger.Info("service does not have a service binding IP, retriggering webhook", "service", svc.Name)
	lb := svc.GetLabels()
	fn.MapSet(&lb, "kloudlite.io/webhook.trigger", "true")
	svc.SetLabels(lb)
	if err := r.Update(ctx, svc); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	return ctrl.Result{}, nil
}

func (r *Reconciler) finalize(req *rApi.Request[*networkingv1.ServiceBinding]) stepResult.Result {
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

	builder := ctrl.NewControllerManagedBy(mgr).For(&corev1.Node{})
	builder.Watches(&corev1.Service{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
		if obj.GetLabels()[constants.KloudliteGatewayEnabledLabel] != "true" {
			return nil
		}

		return []reconcile.Request{{NamespacedName: fn.NN(obj.GetNamespace(), obj.GetName())}}
	}))

	builder.Watches(&corev1.Namespace{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
		if obj.GetLabels()[constants.KloudliteGatewayEnabledLabel] != "true" {
			return nil
		}

		var svclist corev1.ServiceList
		if err := r.List(ctx, &svclist, client.InNamespace(obj.GetName())); err != nil {
			return nil
		}

		rr := make([]reconcile.Request, 0, len(svclist.Items))
		for _, svc := range svclist.Items {
			rr = append(rr, reconcile.Request{NamespacedName: fn.NN(svc.GetNamespace(), svc.GetName())})
		}
		return rr
	}))

	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
