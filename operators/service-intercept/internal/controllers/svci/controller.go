package svci

import (
	"context"
	"fmt"
	"os"
	"time"

	"k8s.io/client-go/tools/record"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/service-intercept/internal/cmd/webhook"
	"github.com/kloudlite/operator/operators/service-intercept/internal/env"
	"github.com/kloudlite/operator/operators/service-intercept/internal/templates"
	"github.com/kloudlite/operator/pkg/constants"
	"github.com/kloudlite/operator/pkg/errors"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	"github.com/kloudlite/operator/pkg/tls_utils"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Logger logging.Logger
	Name   string
	Env    *env.Env

	yamlClient kubectl.YAMLClient
	recorder   record.EventRecorder

	svcInterceptTemplate []byte
	templateWebhook      []byte
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	CreatedForLabel string = "kloudlite.io/created-for"

	CreateIntercept         string = "create-intercept"
	InterceptClosePerformed string = "cleanup"
	TrackInterceptedSvc     string = "tracks-intercept-svc"
)

var DeleteChecklist = []rApi.CheckMeta{
	{Name: CreateIntercept, Title: "Intercept close performed"},
}

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.Logger), r.Client, request.NamespacedName, &crdsv1.ServiceIntercept{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	req.PreReconcile()
	defer req.PostReconcile()

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureCheckList([]rApi.CheckMeta{
		{Name: CreateIntercept, Title: func() string {
			return "Intercept performed"
		}(), Hide: false},
	}); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.RestartIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.createSvcIntercept(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.trackInterceptSvc(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *Reconciler) finalize(req *rApi.Request[*crdsv1.ServiceIntercept]) stepResult.Result {
	if step := req.EnsureCheckList(DeleteChecklist); !step.ShouldProceed() {
		return step
	}

	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(InterceptClosePerformed, req)

	svc, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Name), &corev1.Service{})
	if err != nil {
		return check.Failed(err)
	}

	var podList corev1.PodList
	if err := r.List(ctx, &podList, &client.ListOptions{
		LabelSelector: labels.SelectorFromValidatedSet(fn.MapMerge(svc.Spec.Selector, map[string]string{
			CreatedForLabel: "intercept",
		})),
		Namespace: obj.Namespace,
	}); err != nil {
		return check.Failed(err)
	}

	for _, p := range podList.Items {
		if err := r.Delete(ctx, &p); err != nil {
			return check.Failed(err)
		}
	}

	if step := req.CleanupOwnedResourcesV2(check); !step.ShouldProceed() {
		return step
	}

	return req.Finalize()
}

const (
	ServiceInterceptServiceName string = "service-intercept"
)

type SetupWebhookArgs struct {
	Client          client.Client
	YAMLClient      kubectl.YAMLClient
	Env             *env.Env
	templateWebhook []byte
}

func setupAdmissionWebhook(ctx context.Context, args SetupWebhookArgs) (map[string][]byte, error) {
	certSecretName := ServiceInterceptServiceName + "-webhook-cert"
	certSecretNamespace := args.Env.KloudliteNamespace

	webhookCert, err := rApi.Get(ctx, args.Client, fn.NN(certSecretNamespace, certSecretName), &corev1.Secret{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return nil, err
		}

		caBundle, cert, key, err := tls_utils.GenTLSCert(tls_utils.GenTLSCertArgs{
			DNSNames: []string{
				fmt.Sprintf("%s.%s.svc", ServiceInterceptServiceName, args.Env.KloudliteNamespace),
				fmt.Sprintf("%s.%s.svc.cluster.local", ServiceInterceptServiceName, args.Env.KloudliteNamespace),
			},
			CertificateLabel: "service intercept webhook cert",
		})
		if err != nil {
			return nil, errors.NewEf(err, "failed to generate TLS certificates")
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
			return nil, err
		}
	}

	b, err := templates.ParseBytes(args.templateWebhook, templates.WebhookTemplateArgs{
		CaBundle:         string(webhookCert.Data["ca.crt"]),
		ServiceName:      ServiceInterceptServiceName,
		ServiceNamespace: certSecretNamespace,
		ServiceHTTPSPort: 9443,
		ServiceSelector:  args.Env.ServiceInterceptWebhookServiceSelector,

		NamespaceSelector: metav1.LabelSelector{
			MatchLabels: map[string]string{
				"kloudlite.io/gateway.enabled": "true",
			},
		},
	})
	if err != nil {
		return nil, err
	}

	if _, err := args.YAMLClient.ApplyYAML(ctx, b); err != nil {
		return nil, err
	}

	return webhookCert.Data, nil
}

func (r *Reconciler) trackInterceptSvc(req *rApi.Request[*crdsv1.ServiceIntercept]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(TrackInterceptedSvc, req)

	svc, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Name), &corev1.Service{})
	if err != nil {
		return check.Failed(err)
	}

	if !fn.IsOwner(obj, fn.AsOwner(svc, true)) {
		obj.SetOwnerReferences(append(obj.GetOwnerReferences(), fn.AsOwner(svc, true)))
		if err := r.Update(ctx, obj); err != nil {
			return check.Failed(err)
		}
		return check.StillRunning(fmt.Errorf("waiting for reconcilation")).RequeueAfter(1 * time.Second)
	}

	if !fn.MapEqual(obj.Status.Selector, svc.Spec.Selector) {
		obj.Status.Selector = svc.Spec.Selector
		if err := r.Status().Update(ctx, obj); err != nil {
			return check.Failed(err)
		}
		return check.StillRunning(fmt.Errorf("waiting for reconcilation")).RequeueAfter(1 * time.Second)
	}

	var podList corev1.PodList
	if err := r.List(ctx, &podList, &client.ListOptions{
		LabelSelector: labels.SelectorFromValidatedSet(svc.Spec.Selector),
		Namespace:     obj.Namespace,
	}); err != nil {
		return check.Failed(err)
	}

	for _, p := range podList.Items {
		if cf := p.Labels[CreatedForLabel]; cf == "intercept" {
			continue
		}

		if err := r.Delete(ctx, &p); err != nil {
			return check.Failed(err)
		}
	}

	return check.Completed()
}

func (r *Reconciler) createSvcIntercept(req *rApi.Request[*crdsv1.ServiceIntercept]) stepResult.Result {
	ctx, obj := req.Context(), req.Object

	check := rApi.NewRunningCheck(CreateIntercept, req)

	podname := obj.Name + "-intercept"
	podns := obj.Namespace
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: podname, Namespace: podns}}

	svciGenerationLabel := "kloudlite.io/svci-generation"

	if err := r.Get(ctx, client.ObjectKeyFromObject(pod), pod); err != nil {
		if apiErrors.IsNotFound(err) {

			if obj.Spec.ToAddr == "" {
				return check.Failed(fmt.Errorf("no address configured on service intercept, failed to intercept")).NoRequeue()
			}

			svc, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Name), &corev1.Service{})
			if err != nil {
				return check.Failed(err)
			}

			udpPorts := make(map[int32]corev1.ServicePort)
			tcpPorts := make(map[int32]corev1.ServicePort)

			for _, port := range svc.Spec.Ports {
				switch port.Protocol {
				case corev1.ProtocolTCP:
					tcpPorts[port.Port] = port
				case corev1.ProtocolUDP:
					udpPorts[port.Port] = port
				}
			}

			tcpPortMappings := make(map[uint16]uint16)
			udpPortMappings := make(map[uint16]uint16)

			for _, pm := range obj.Spec.PortMappings {
				if _, ok := tcpPorts[int32(pm.ServicePort)]; ok {
					tcpPortMappings[pm.ServicePort] = pm.DevicePort
				}

				if _, ok := udpPorts[int32(pm.ServicePort)]; ok {
					udpPortMappings[pm.ServicePort] = pm.DevicePort
				}
			}

			b, err := templates.ParseBytes(r.svcInterceptTemplate, map[string]any{
				"name":      podname,
				"namespace": podns,
				"labels": fn.MapMerge(
					map[string]string{
						svciGenerationLabel: fmt.Sprintf("%d", obj.Generation),
						CreatedForLabel:     "intercept",
					},
					svc.Spec.Selector,
				),
				"owner-references": []metav1.OwnerReference{fn.AsOwner(obj, true)},
				"device-host":      obj.Spec.ToAddr,

				"tcp-port-mappings": tcpPortMappings,
				"udp-port-mappings": udpPortMappings,
			})
			if err != nil {
				return check.Failed(err).NoRequeue()
			}

			if _, err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
				return check.Failed(err).NoRequeue()
			}

			return check.StillRunning(fmt.Errorf("waiting for intercept pod to start"))
		}
	}

	if pod.Labels[svciGenerationLabel] != fmt.Sprintf("%d", obj.Generation) {
		if err := r.Delete(ctx, pod); err != nil {
			return check.Failed(err)
		}
		return check.StillRunning(fmt.Errorf("waiting for previous generation pod to be deleted"))
	}

	if pod.Status.Phase != corev1.PodRunning {
		return check.StillRunning(fmt.Errorf("waiting for pod to start running"))
	}

	return check.Completed()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.Logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig(), kubectl.YAMLClientOpts{Logger: r.Logger})
	r.recorder = mgr.GetEventRecorderFor(r.GetName())

	var err error
	r.svcInterceptTemplate, err = templates.Read(templates.SvcIntercept)
	if err != nil {
		return err
	}

	templateWebhook, err := templates.Read(templates.WebhookTemplate)
	if err != nil {
		return err
	}

	webhookCertData, err := setupAdmissionWebhook(context.TODO(), SetupWebhookArgs{
		Client:          r.Client,
		YAMLClient:      r.yamlClient,
		Env:             r.Env,
		templateWebhook: templateWebhook,
	})
	if err != nil {
		return err
	}

	if err := os.WriteFile("/tmp/tls.crt", webhookCertData["tls.crt"], 0o666); err != nil {
		return err
	}
	if err := os.WriteFile("/tmp/tls.key", webhookCertData["tls.key"], 0o666); err != nil {
		return err
	}

	go webhook.Run(webhook.RunArgs{
		Addr:            ":9443",
		LogLevel:        "info",
		KubeRestConfig:  mgr.GetConfig(),
		CreatedForLabel: CreatedForLabel,
		TLSCertFile:     "/tmp/tls.crt",
		TLSKeyFile:      "/tmp/tls.key",
	})

	builder := ctrl.NewControllerManagedBy(mgr).For(&crdsv1.ServiceIntercept{})

	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.Owns(&corev1.Pod{})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
