package service_binding

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
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
	apiLabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	apiErrors "k8s.io/apimachinery/pkg/api/errors"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Env        *env.Env
	logger     *slog.Logger
	Name       string
	yamlClient kubectl.YAMLClient
}

const (
	svcNetworkingProxyTo       = "kloudlite.io/networking.proxy.to"
	serviceBindingIPAnnotation = "kloudlite.io/servicebinding.ip"
)

func (r *Reconciler) GetName() string {
	return r.Name
}

var ErrResourceNotFound error = fmt.Errorf("not found")

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=lifecycles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=lifecycles/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=lifecycles/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	r.logger.Info("reconciling", "resource", request.NamespacedName)

	switch {
	case strings.HasPrefix(request.Name, "service/"):
		{
			resp, err := r.reconcileService(ctx, ctrl.Request{NamespacedName: fn.NN(request.Namespace, request.Name[len("service/"):])})
			if err != nil {
				return ctrl.Result{}, client.IgnoreNotFound(err)
			}

			return resp, nil
		}
	default:
		{
			resp, err := r.reconcileServiceBinding(ctx, request)
			if err != nil {
				if !apiErrors.IsNotFound(err) {
					return ctrl.Result{}, err
				}
			}

			return resp, nil
		}
	}
}

func (r *Reconciler) reconcileServiceBinding(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	logger := r.logger.With("service-binding", request.String())

	svcb, err := rApi.Get(ctx, r.Client, request.NamespacedName, &networkingv1.ServiceBinding{})
	if err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("starts reconciling")
	defer logger.Info("finished reconciling")

	if svcb.GetDeletionTimestamp() != nil {
		return ctrl.Result{}, nil
	}

	// if svcb.Spec.Hostname == "" {
	// 	return ctrl.Result{}, nil
	// }

	if v, ok := svcb.Annotations["kloudlite.io/global.hostname"]; ok {
		svcDnsReq, err := http.NewRequestWithContext(ctx, http.MethodPut, fmt.Sprintf("%s/service/%s/%s", r.Env.ServiceDNSHttpAddr, v, svcb.Spec.GlobalIP), nil)
		if err != nil {
			return ctrl.Result{}, err
		}

		r.logger.Debug("START: updating DNS records for service", "url", svcDnsReq.URL.String())
		if _, err := http.DefaultClient.Do(svcDnsReq); err != nil {
			return ctrl.Result{}, err
		}
		r.logger.Debug("FINISH: updated service DNS for service", "url", svcDnsReq.URL.String())
	}

	svcbUpdateRequest, err := http.NewRequestWithContext(ctx, http.MethodPut, fmt.Sprintf("%s/service-binding/%s/%s", r.Env.GatewayAdminApiAddr, svcb.Namespace, svcb.Name), nil)
	if err != nil {
		return ctrl.Result{}, err
	}
	resp, err := http.DefaultClient.Do(svcbUpdateRequest)
	if err != nil {
		return ctrl.Result{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return ctrl.Result{}, fmt.Errorf("bad response, expected %d, got %d", http.StatusOK, resp.StatusCode)
	}

	r.logger.Info("service binding successfully reconciled", "service-binding", fmt.Sprintf("%s/%s", svcb.GetNamespace(), svcb.GetName()))
	return ctrl.Result{}, nil
}

func (r *Reconciler) reconcileService(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	logger := r.logger.With("service", request.String())

	svc, err := rApi.Get(ctx, r.Client, request.NamespacedName, &corev1.Service{})
	if err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("starts reconciling")
	defer func() {
		logger.Info("finished reconciling")
	}()

	// hasNamedPort := false
	// for i := range svc.Spec.Ports {
	// 	if svc.Spec.Ports[i].TargetPort.Type == intstr.String {
	// 		hasNamedPort = true
	// 		break
	// 	}
	// }
	//
	// if hasNamedPort {
	// 	var podsList corev1.PodList
	// 	if err := r.List(ctx, &podsList, &client.ListOptions{
	// 		Limit:         1,
	// 		LabelSelector: apiLabels.SelectorFromValidatedSet(svc.Spec.Selector),
	// 	}); err != nil {
	// 		return ctrl.Result{}, err
	// 	}
	//
	// 	for _, pod := range podsList.Items {
	// 		for _, c := range pod.Spec.Containers {
	// 			for _, cp := range c.Ports {
	// 				for i := range svc.Spec.Ports {
	// 					if svc.Spec.Ports[i].TargetPort.String() == cp.Name {
	// 						svc.Spec.Ports[i].TargetPort = intstr.FromInt32(cp.ContainerPort)
	// 					}
	// 				}
	// 			}
	// 		}
	// 	}
	// }

	if svc.GetDeletionTimestamp() != nil {
		return ctrl.Result{}, nil
	}

	var sblist networkingv1.ServiceBindingList
	if err := r.List(ctx, &sblist, &client.ListOptions{
		Limit: 1,
		LabelSelector: apiLabels.SelectorFromValidatedSet(map[string]string{
			"kloudlite.io/servicebinding.reservation": fmt.Sprintf("%s.%s", svc.GetNamespace(), svc.GetName()),
		}),
	}); err != nil {
		logger.Error("failed to find service binding")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if len(sblist.Items) == 0 {
		logger.Info("service binding not found, creating now")

		createSvcBindingReq, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/service/%s/%s", r.Env.GatewayAdminApiAddr, svc.GetNamespace(), svc.GetName()), nil)
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

		return ctrl.Result{RequeueAfter: 500 * time.Millisecond}, nil
	}

	ns, err := rApi.Get(ctx, r.Client, fn.NN("", svc.GetNamespace()), &corev1.Namespace{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	kloudliteDNSHost := func() string {
		if v, ok := svc.GetLabels()[constants.KloudliteDNSHostname]; ok {
			return v
		}

		if v, ok := ns.Labels[constants.KloudliteNamespaceForEnvironment]; ok {
			return fmt.Sprintf("%s.%s", svc.GetName(), v)
		}

		if v, ok := ns.Labels[constants.KloudliteNamespaceForClusterManagedService]; ok {
			return v
		}

		return ""
	}()

	sb := &sblist.Items[0]

	v, hasServiceBindingIPAnn := svc.GetAnnotations()[serviceBindingIPAnnotation]
	r.logger.Info("service binding ann check", "has service binding IP ann", hasServiceBindingIPAnn, "value", v, "service binding IP", sb.Spec.GlobalIP, slog.Group("service", "namespace", svc.GetNamespace(), "name", svc.GetName(), "ann", svc.GetAnnotations()))
	if !hasServiceBindingIPAnn || v != sb.Spec.GlobalIP {
		fn.MapSet(&svc.Annotations, serviceBindingIPAnnotation, sb.Spec.GlobalIP)
		r.logger.Info("setting service binding IP annotation on service")
		if err := r.Update(ctx, svc); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: 1 * time.Second}, nil
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, sb, func() error {
		if sb.Generation == 0 {
			return fmt.Errorf("must not be triggered to create service binding")
		}

		sb.Spec.Ports = svc.Spec.Ports

		switch {
		case svc.Annotations[svcNetworkingProxyTo] != "":
			sb.Spec.ServiceIP = fn.New(svc.Annotations[svcNetworkingProxyTo])
		case svc.Spec.Type == corev1.ServiceTypeExternalName:
			sb.Spec.ServiceIP = &svc.Spec.ExternalName
		default:
			sb.Spec.ServiceIP = &svc.Spec.ClusterIP
		}

		if kloudliteDNSHost != "" {
			sb.Spec.Hostname = kloudliteDNSHost
		}

		svcHost := fmt.Sprintf("%s.%s.%s", svc.Name, svc.Namespace, r.Env.GatewayDNSSuffix)

		ann := sb.GetAnnotations()
		if ann == nil {
			ann = make(map[string]string, 1)
		}
		ann["kloudlite.io/global.hostname"] = svcHost
		sb.SetAnnotations(ann)

		return nil
	}); err != nil {
		return ctrl.Result{}, err
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
		ShowCaller:    true,
		ShowDebugLogs: false,
	})
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig(), kubectl.YAMLClientOpts{Logger: logger})

	builder := ctrl.NewControllerManagedBy(mgr).For(&networkingv1.ServiceBinding{})
	builder.Watches(&corev1.Service{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
		if obj.GetLabels()[constants.KloudliteGatewayEnabledLabel] != "true" {
			return nil
		}

		return []reconcile.Request{{NamespacedName: fn.NN(obj.GetNamespace(), fmt.Sprintf("service/%s", obj.GetName()))}}
	}))

	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
