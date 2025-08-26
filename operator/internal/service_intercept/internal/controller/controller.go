package controller

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/codingconcepts/env"
	fn "github.com/kloudlite/kloudlite/operator/toolkit/functions"
	"github.com/kloudlite/kloudlite/operator/toolkit/kubectl"
	"github.com/kloudlite/kloudlite/operator/toolkit/reconciler"
	v1 "github.com/kloudlite/operator/api/v1"
	"github.com/kloudlite/operator/common"
	"github.com/kloudlite/operator/internal/service_intercept/internal/templates"
	"github.com/kloudlite/operator/internal/service_intercept/internal/webhook"
	"github.com/kloudlite/operator/pkg/errors"
	"github.com/kloudlite/operator/pkg/tls_utils"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

type Env struct {
	MaxConcurrentReconciles int    `env:"MAX_CONCURRENT_RECONCILES" default:"5"`
	KloudliteNamespace      string `env:"KLOUDLITE_NAMESPACE" default:"kloudlite"`

	DevMode                         bool   `env:"DEV" default:"false"`
	DevWebhookProxy                 bool   `env:"DEV_WEBHOOK_PROXY" default:"false"`
	DevWebhookProxyServiceName      string `env:"DEV_WEBHOOK_PROXY_SERVICE_NAME"`
	DevWebhookProxyServiceNamespace string `env:"DEV_WEBHOOK_PROXY_SERVICE_NAMESPACE"`
}

const (
	serviceInterceptorPodLabelKey   = v1.ProjectDomain + "/pod.role"
	serviceInterceptorPodLabelValue = "service-interceptor"
)

// Reconciler reconciles a ServiceIntercept object
type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme

	env Env

	YAMLClient kubectl.YAMLClient

	svcInterceptTemplate []byte
	templateWebhook      []byte
}

func (r *Reconciler) GetName() string {
	return v1.ProjectDomain + "/service-intercept-controller"
}

const (
	CreatedForLabel string = "kloudlite.io/created-for"

	ServiceInterceptServiceName string = "service-intercept"
)

// +kubebuilder:rbac:groups=kloudlite.io,resources=serviceintercepts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kloudlite.io,resources=serviceintercepts/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kloudlite.io,resources=serviceintercepts/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := reconciler.NewRequest(ctx, r.Client, request.NamespacedName, &v1.ServiceIntercept{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	req.PreReconcile()
	defer req.PostReconcile()

	return reconciler.ReconcileSteps(req, []reconciler.Step[*v1.ServiceIntercept]{
		{
			Name:     "setup-service-intercept",
			Title:    "Setup Service Intercept",
			OnCreate: r.createSvcIntercept,
			OnDelete: r.cleanupSvcIntercept,
		},
		{
			Name:     "track-service-pods",
			Title:    "Track Pods for intercepted service",
			OnCreate: r.setupTrackServicePods,
			OnDelete: nil,
		},
	})
}

func (r *Reconciler) createSvcIntercept(check *reconciler.Check[*v1.ServiceIntercept], obj *v1.ServiceIntercept) reconciler.StepResult {
	podname := obj.Name + "-intercept"
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: podname, Namespace: obj.Namespace}}

	svciGenerationKey := "kloudlite.io/service-intercept.generation"

	if err := r.Get(check.Context(), client.ObjectKeyFromObject(pod), pod); err != nil {
		if apiErrors.IsNotFound(err) {
			if obj.Spec.ToHost == "" {
				return check.Failed(fmt.Errorf("no address configured on service intercept, failed to intercept")).NoRequeue()
			}

			svc, err := reconciler.Get(check.Context(), r.Client, fn.NN(obj.Namespace, obj.Spec.ServiceRef.Name), &corev1.Service{})
			if err != nil {
				return check.Failed(err)
			}

			svcPortMap := make(map[string]corev1.ServicePort, len(svc.Spec.Ports))

			for _, svcPort := range svc.Spec.Ports {
				svcPortMap[fmt.Sprintf("%s/%s", svcPort.Protocol, svcPort.Port)] = svcPort
			}

			for _, pm := range obj.Spec.PortMappings {
				if _, ok := svcPortMap[fmt.Sprintf("%s/%s", pm.Protocol, pm.ServicePort)]; !ok {
					return check.Failed(fmt.Errorf("%s/%s port is not configured on service %s/%s", pm.Protocol, pm.ServicePort, obj.Namespace, obj.Spec.ServiceRef.Name))
				}
			}

			b, err := templates.ParseBytes(r.svcInterceptTemplate, templates.ServiceInterceptPodSpecParams{
				TargetHost:   obj.Spec.ToHost,
				PortMappings: obj.Spec.PortMappings,
			})
			if err != nil {
				return check.Failed(err).NoRequeue()
			}

			if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, pod, func() error {
				pod.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
				pod.SetLabels(fn.MapMerge(pod.GetLabels(), svc.Spec.Selector, map[string]string{
					serviceInterceptorPodLabelKey: serviceInterceptorPodLabelValue,
				}))
				pod.SetAnnotations(fn.MapMerge(pod.GetAnnotations(), map[string]string{svciGenerationKey: fmt.Sprintf("%d", obj.Generation)}))

				if err := yaml.Unmarshal(b, &pod.Spec); err != nil {
					return fmt.Errorf("failed to unmarshal into pod.spec: %w", err)
				}

				return nil
			}); err != nil {
				return check.Failed(err)
			}

			obj.Status.Selector = svc.Spec.Selector
			if err := r.Status().Update(check.Context(), obj); err != nil {
				return check.Failed(err)
			}
		}
		return check.Abort("waiting for service intercept pod to start")
	}

	if pod.Annotations[svciGenerationKey] != fmt.Sprintf("%d", obj.Generation) {
		if err := r.Delete(check.Context(), pod); err != nil {
			return check.Failed(err)
		}
		return check.Abort("waiting for previous generation pod to be deleted")
	}

	if pod.Status.Phase != corev1.PodRunning {
		return check.Errored(fmt.Errorf("waiting for pod to start running"))
	}

	return check.Passed()
}

func (r *Reconciler) cleanupSvcIntercept(check *reconciler.Check[*v1.ServiceIntercept], obj *v1.ServiceIntercept) reconciler.StepResult {
	podname := obj.Name + "-intercept"
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: podname, Namespace: obj.Namespace}}

	if err := fn.DeleteAndWait(check.Context(), r.Client, pod); err != nil {
		return check.Errored(err)
	}

	return check.Passed()
}

func (r *Reconciler) setupTrackServicePods(check *reconciler.Check[*v1.ServiceIntercept], obj *v1.ServiceIntercept) reconciler.StepResult {
	svc, err := reconciler.Get(check.Context(), r.Client, fn.NN(obj.Namespace, obj.Spec.ServiceRef.Name), &corev1.Service{})
	if err != nil {
		return check.Failed(fmt.Errorf("failed to get service: %w", err))
	}

	if !fn.IsOwner(obj, fn.AsOwner(svc, true)) {
		obj.SetOwnerReferences(append(obj.GetOwnerReferences(), fn.AsOwner(svc, true)))
		if err := r.Update(check.Context(), obj); err != nil {
			return check.Failed(err)
		}
		return check.Abort("waiting for reconciliation").RequeueAfter(1 * time.Second)
	}

	var podList corev1.PodList
	if err := r.List(check.Context(), &podList, &client.ListOptions{
		LabelSelector: labels.SelectorFromValidatedSet(svc.Spec.Selector),
		Namespace:     obj.Namespace,
	}); err != nil {
		return check.Failed(err)
	}

	for _, p := range podList.Items {
		if p.Name == obj.Name+"-intercept" {
			continue
		}
		// if cf := p.Labels[CreatedForLabel]; cf == "intercept" {
		// 	continue
		// }

		check.Logger().Info("pod", "name", p.GetName())

		if err := r.Delete(check.Context(), &p); err != nil {
			return check.Failed(err)
		}
	}

	return check.Passed()
}

type SetupWebhookArgs struct {
	Client          client.Client
	YAMLClient      kubectl.YAMLClient
	Env             *Env
	templateWebhook []byte
}

func setupAdmissionWebhook(ctx context.Context, args SetupWebhookArgs) (tlsCert, tlsKey []byte, err error) {
	certSecretName := ServiceInterceptServiceName + "-webhook-cert"
	certSecretNamespace := args.Env.KloudliteNamespace

	webhookCert, err := reconciler.Get(ctx, args.Client, fn.NN(certSecretNamespace, certSecretName), &corev1.Secret{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return nil, nil, err
		}

		caBundle, cert, key, err := tls_utils.GenTLSCert(tls_utils.GenTLSCertArgs{
			DNSNames: []string{
				fmt.Sprintf("%s.%s.svc", ServiceInterceptServiceName, args.Env.KloudliteNamespace),
				fmt.Sprintf("%s.%s.svc.cluster.local", ServiceInterceptServiceName, args.Env.KloudliteNamespace),
			},
			CertificateLabel: "service intercept webhook cert",
		})
		if err != nil {
			return nil, nil, errors.NewEf(err, "failed to generate TLS certificates")
		}

		webhookCert = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      certSecretName,
				Namespace: certSecretNamespace,
			},
			Data: map[string][]byte{
				"ca.crt":  caBundle,
				"tls.crt": cert,
				"tls.key": key,
			},
		}
		if err := args.Client.Create(ctx, webhookCert); err != nil {
			return nil, nil, err
		}
	}

	if args.Env.DevWebhookProxy {
		caBundle, cert, key, err := tls_utils.GenTLSCert(tls_utils.GenTLSCertArgs{
			DNSNames: []string{
				fmt.Sprintf("%s.%s.svc", args.Env.DevWebhookProxyServiceName, args.Env.DevWebhookProxyServiceNamespace),
				fmt.Sprintf("%s.%s.svc.cluster.local", args.Env.DevWebhookProxyServiceName, args.Env.DevWebhookProxyServiceNamespace),
			},
			CertificateLabel: "service intercept webhook cert",
		})
		if err != nil {
			return nil, nil, errors.NewEf(err, "failed to generate TLS certificates")
		}

		if _, err := controllerutil.CreateOrUpdate(ctx, args.Client, webhookCert, func() error {
			webhookCert.Data = map[string][]byte{
				"ca.crt":  caBundle,
				"tls.crt": cert,
				"tls.key": key,
			}

			return nil
		}); err != nil {
			return nil, nil, err
		}
	}

	b, err := templates.ParseBytes(args.templateWebhook, templates.WebhookTemplateArgs{
		WebhookProxy: templates.WebhookProxy{
			Enabled:          args.Env.DevWebhookProxy,
			ServiceName:      args.Env.DevWebhookProxyServiceName,
			ServiceNamespace: args.Env.DevWebhookProxyServiceNamespace,
		},

		CaBundle:         string(webhookCert.Data["ca.crt"]),
		ServiceName:      ServiceInterceptServiceName,
		ServiceNamespace: certSecretNamespace,
		ServiceHTTPSPort: 9443,
		ServiceSelector: map[string]string{
			v1.ProjectDomain + "/type": "service-intercept",
		},

		NamespaceSelector: metav1.LabelSelector{
			MatchLabels: map[string]string{
				"kloudlite.io/gateway.enabled": "true",
			},
		},
		InterceptorPodLabelKey:   serviceInterceptorPodLabelKey,
		InterceptorPodLabelValue: serviceInterceptorPodLabelValue,
	})
	if err != nil {
		return nil, nil, err
	}

	if _, err := args.YAMLClient.ApplyYAML(ctx, b); err != nil {
		return nil, nil, err
	}

	return webhookCert.Data["tls.crt"], webhookCert.Data["tls.key"], nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()

	if err := env.Set(&r.env); err != nil {
		return fmt.Errorf("failed to set env: %w", err)
	}

	if r.YAMLClient == nil {
		return fmt.Errorf("yaml client must be set")
	}

	var err error
	r.svcInterceptTemplate, err = templates.Read(templates.SvcIntercept)
	if err != nil {
		return err
	}

	templateWebhook, err := templates.Read(templates.WebhookTemplate)
	if err != nil {
		return err
	}

	// INFO: running service intercept mutating webhook as manager.Runnable
	mgr.Add(manager.RunnableFunc(func(ctx context.Context) error {
		tlsCert, tlsKey, err := setupAdmissionWebhook(context.TODO(), SetupWebhookArgs{
			Client:          r.Client,
			YAMLClient:      r.YAMLClient,
			Env:             &r.env,
			templateWebhook: templateWebhook,
		})
		if err != nil {
			return err
		}

		tlsCertPath, tlsKeyPath := "/tmp/tls.crt", "/tmp/tls.key"

		if err := os.WriteFile(tlsCertPath, tlsCert, 0o666); err != nil {
			return err
		}

		if err := os.WriteFile(tlsKeyPath, tlsKey, 0o666); err != nil {
			return err
		}

		mw := webhook.MutationWebhook{
			Debug:      r.env.DevMode,
			KubeClient: mgr.GetClient(),
			Scheme:     mgr.GetScheme(),
			ShouldIgnorePod: func(pod *corev1.Pod) bool {
				return pod.Labels[serviceInterceptorPodLabelKey] == serviceInterceptorPodLabelValue
			},
		}

		handler, err := mw.Handler()
		if err != nil {
			return err
		}

		server := &http.Server{Addr: ":9443", Handler: handler}

		go func() {
			<-ctx.Done()
			_ = server.Shutdown(context.Background())
		}()

		common.PrintMetadataBanner(common.Metadata{
			Name:    "service intercept webhook",
			Message: "HTTP Server Running @ :9443",
		})
		return server.ListenAndServeTLS(tlsCertPath, tlsKeyPath)
	}))

	builder := ctrl.NewControllerManagedBy(mgr).For(&v1.ServiceIntercept{}).Named(r.GetName())

	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.env.MaxConcurrentReconciles})
	builder.Owns(&corev1.Pod{})
	builder.Owns(&corev1.Service{})
	builder.WithEventFilter(reconciler.ReconcileFilter(mgr.GetEventRecorderFor(r.GetName())))
	return builder.Complete(r)
}
