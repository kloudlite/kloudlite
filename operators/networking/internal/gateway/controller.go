package gateway

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	ct "github.com/kloudlite/operator/apis/common-types"
	networkingv1 "github.com/kloudlite/operator/apis/networking/v1"
	"github.com/kloudlite/operator/operators/networking/internal/env"
	"github.com/kloudlite/operator/operators/networking/internal/gateway/templates"
	"github.com/kloudlite/operator/pkg/constants"
	"github.com/kloudlite/operator/pkg/errors"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/yaml"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Env        *env.Env
	logger     logging.Logger
	Name       string
	yamlClient kubectl.YAMLClient

	templateDeployment []byte
	templateWebhook    []byte
	templateRBAC       []byte
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	bindService            string = "bind-service"
	patchDefaults          string = "patch-defaults"
	createGatewayNamespace string = "create-gateway-namespace"
	generateWireguardKeys  string = "generate-wireguard-keys"
	setupDeploymentRBAC    string = "setup-deployment-rbac"
	setupGatewayDeployment string = "setup-gateway-deployment"
	setupMutationWebhooks  string = "setup-mutation-webhooks"
	trackLoadBalancer      string = "track-load-balancer"
)

const (
	gatewayLocalOverrideConfigName      = "gateway-local-overrides"
	gatewayLocalOverrideConfigNamespace = "kloudlite"
)

const (
	// Read more @ https://en.wikipedia.org/wiki/Reserved_IP_addresses
	gatewayInternalDNSNameServer string = "198.18.0.53"
)

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=lifecycles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=lifecycles/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=lifecycles/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &networkingv1.Gateway{})
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

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureCheckList([]rApi.CheckMeta{
		{Name: patchDefaults, Title: "Patch Defaults", Debug: true},
		{Name: createGatewayNamespace, Title: "Create Gateway Namespace"},
		{Name: generateWireguardKeys, Title: "Generate Wireguard Keys"},
		{Name: setupDeploymentRBAC, Title: "Setup Deployment RBAC"},
		{Name: setupGatewayDeployment, Title: "Setup Gateway Device Deployment"},
		{Name: setupMutationWebhooks, Title: "Setup Mutation Webhooks for kloudlite systems"},
		{Name: trackLoadBalancer, Title: "Track Load Balancer"},
	}); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.patchDefaults(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.createGatewayNamespace(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.generateWireguardKeys(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.setupDeploymentRBAC(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.setupGatewayDeployment(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.setupMutationWebhooks(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.trackLoadBalancer(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *Reconciler) finalize(req *rApi.Request[*networkingv1.Gateway]) stepResult.Result {
	const finalizing = "finalizing"

	if len(req.Object.Status.Checks) != 1 {
		req.Object.Status.Checks = nil
	}

	if step := req.EnsureCheckList([]rApi.CheckMeta{
		{Name: finalizing, Title: "Cleaning up Gateway"},
	}); !step.ShouldProceed() {
		return step
	}

	check := rApi.NewRunningCheck(finalizing, req)

	if step := req.CleanupOwnedResourcesV2(check); !step.ShouldProceed() {
		return step
	}

	return req.Finalize()
}

func getAccountName(obj *networkingv1.Gateway) *string {
	if v, ok := obj.GetLabels()[constants.AccountNameKey]; ok && v != "" {
		return &v
	}
	return nil
}

func (r *Reconciler) patchDefaults(req *rApi.Request[*networkingv1.Gateway]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(patchDefaults, req)

	hasUpdate := false

	if obj.Spec.TargetNamespace == "" {
		hasUpdate = true
		// obj.Spec.TargetNamespace = fmt.Sprintf("kl-gateway-%s", obj.Name)
		obj.Spec.TargetNamespace = "kl-gateway"
	}

	if obj.Spec.WireguardKeysRef.Name == "" {
		hasUpdate = true
		obj.Spec.WireguardKeysRef.Name = obj.Spec.TargetNamespace
	}

	if obj.Spec.LocalOverrides == nil {
		cm, err := rApi.Get(ctx, r.Client, fn.NN(gatewayLocalOverrideConfigNamespace, gatewayLocalOverrideConfigName), &corev1.ConfigMap{})
		if err != nil {
			if !apiErrors.IsNotFound(err) {
				return check.Failed(err)
			}
		}

		if cm != nil {
			hasUpdate = true
			obj.Spec.LocalOverrides = &ct.SecretRef{
				Name:      gatewayLocalOverrideConfigName,
				Namespace: gatewayLocalOverrideConfigNamespace,
			}
		}
	}

	if hasUpdate {
		if err := r.Update(ctx, obj); err != nil {
			return check.Failed(err)
		}
		return check.StillRunning(fmt.Errorf("waiting for resource to be updated"))
	}

	return check.Completed()
}

func (r *Reconciler) createGatewayNamespace(req *rApi.Request[*networkingv1.Gateway]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(createGatewayNamespace, req)

	ns, err := rApi.Get(ctx, r.Client, fn.NN("", obj.Spec.TargetNamespace), &corev1.Namespace{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return check.Failed(err)
		}

		if err := r.Create(ctx, &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: obj.Spec.TargetNamespace,
				Labels: map[string]string{
					constants.KloudliteNamespaceForGateway: obj.Name,
				},
			},
		}); err != nil {
			return check.Failed(err)
		}
	}

	v, ok := ns.Labels[constants.KloudliteNamespaceForGateway]
	if !ok {
		fn.MapSet(&ns.Labels, constants.KloudliteNamespaceForGateway, obj.Name)
		if err := r.Update(ctx, ns); err != nil {
			return check.Failed(err)
		}
		return check.StillRunning(fmt.Errorf("waiting for namespace to be updated with label"))
	}

	if v != obj.Name {
		return check.Failed(fmt.Errorf("namespace %s/%s is not labeled with %s=%s, it might be because namespace already belongs to some other gateway", obj.Spec.TargetNamespace, obj.Name, constants.KloudliteNamespaceForGateway, obj.Name))
	}

	return check.Completed()
}

func (r *Reconciler) generateWireguardKeys(req *rApi.Request[*networkingv1.Gateway]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(generateWireguardKeys, req)

	_, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.TargetNamespace, obj.Spec.WireguardKeysRef.Name), &corev1.Secret{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return check.Failed(err)
		}

		key, err := wgtypes.GenerateKey()
		if err != nil {
			return check.Failed(err)
		}

		if err := r.Create(ctx, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      obj.Spec.WireguardKeysRef.Name,
				Namespace: obj.Spec.TargetNamespace,
			},
			Data: map[string][]byte{
				"private_key": []byte(key.String()),
				"public_key":  []byte(key.PublicKey().String()),
			},
		}); err != nil {
			return check.Failed(err)
		}
	}

	return check.Completed()
}

func (r *Reconciler) setupDeploymentRBAC(req *rApi.Request[*networkingv1.Gateway]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(setupDeploymentRBAC, req)

	b, err := templates.ParseBytes(r.templateRBAC, templates.GatewayRBACTemplateArgs{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-svc-account", obj.Name),
			Namespace: obj.Spec.TargetNamespace,
		},
	})
	if err != nil {
		return check.Failed(err)
	}

	rr, err := r.yamlClient.ApplyYAML(ctx, b)
	if err != nil {
		return check.Failed(err)
	}

	req.AddToOwnedResources(rr...)

	return check.Completed()
}

func (r *Reconciler) setupGatewayDeployment(req *rApi.Request[*networkingv1.Gateway]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(setupGatewayDeployment, req)

	var localOverridesCfg corev1.ConfigMap
	if obj.Spec.LocalOverrides != nil {
		if err := r.Get(ctx, fn.NN(obj.Spec.LocalOverrides.Namespace, obj.Spec.LocalOverrides.Name), &localOverridesCfg); err != nil {
			if !apiErrors.IsNotFound(err) {
				return check.Failed(err)
			}
		}
	}

	var lo networkingv1.LocalOverrides
	if v, ok := localOverridesCfg.Data["peers"]; ok {
		if err := yaml.Unmarshal([]byte(v), &lo.Peers); err != nil {
			return check.Failed(err)
		}
	}

	extraPeersCfg := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-wg-extra-peers", obj.Name), Namespace: obj.Spec.TargetNamespace}}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, extraPeersCfg, func() error {
		if extraPeersCfg.Data == nil {
			extraPeersCfg.Data = make(map[string]string, 1)
		}
		peers := make([]string, 0, len(obj.Spec.Peers))
		for _, peer := range obj.Spec.Peers {
			npeer := fmt.Sprintf("[Peer]\nPublicKey = %s\nAllowedIPs = %s\n PersistentKeepalive = 25\n", peer.PublicKey, strings.Join(peer.AllowedIPs, ", "))
			if peer.PublicEndpoint != nil && *peer.PublicEndpoint != "" {
				npeer = fmt.Sprintf("%s\nEndpoint = %s\n", npeer, *peer.PublicEndpoint)
			}
			peers = append(peers, npeer)
		}

		for _, peer := range lo.Peers {
			npeer := fmt.Sprintf("[Peer]\nPublicKey = %s\nAllowedIPs = %s\n PersistentKeepalive = 25\n", peer.PublicKey, strings.Join(peer.AllowedIPs, ", "))
			if peer.PublicEndpoint != nil && *peer.PublicEndpoint != "" {
				npeer = fmt.Sprintf("%s\nEndpoint = %s\n", npeer, *peer.PublicEndpoint)
			}
			peers = append(peers, npeer)
		}

		extraPeersCfg.Data["peers.conf"] = strings.Join(peers, "\n")
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	gatewayDNSServers := make([]string, 0, len(obj.Spec.Peers)+2)
	for _, peer := range obj.Spec.Peers {
		if peer.IP == nil {
			continue
		}
		if *peer.IP == obj.Spec.GlobalIP {
			continue
		}
		if peer.DNSSuffix != nil {
			gatewayDNSServers = append(gatewayDNSServers, fmt.Sprintf("%s=%s:53", *peer.DNSSuffix, *peer.IP))
		}
		if peer.DNSHostname == "kloudlite-global-vpn-device.device.local" {
			gatewayDNSServers = append(gatewayDNSServers, fmt.Sprintf("%s=%s:53", "device.local", *peer.IP))
		}
	}

	gatewayDNSServers = append(gatewayDNSServers, fmt.Sprintf("%s=%s:53", obj.Spec.DNSSuffix, obj.Spec.GlobalIP))

	dnsService, err := rApi.Get(ctx, r.Client, fn.NN("kube-system", "kube-dns"), &corev1.Service{})
	if err != nil {
		return check.Failed(errors.NewEf(err, "failed to find kube-dns service"))
	}

	gatewayDNSServers = append(gatewayDNSServers, fmt.Sprintf("%s=%s:53", "svc.cluster.local", dnsService.Spec.ClusterIP))

	b, err := templates.ParseBytes(r.templateDeployment, templates.GatewayDeploymentArgs{
		ObjectMeta: metav1.ObjectMeta{
			Name:      obj.Name,
			Namespace: obj.Spec.TargetNamespace,
			Labels: map[string]string{
				"kloudlite.io/managed-by-gateway": "true",
				"kloudlite.io/gateway.name":       obj.Name,
			},
			OwnerReferences: []metav1.OwnerReference{fn.AsOwner(obj, true)},
		},
		ServiceAccountName:           fmt.Sprintf("%s-svc-account", obj.Name),
		GatewayWgSecretName:          obj.Spec.WireguardKeysRef.Name,
		GatewayGlobalIP:              obj.Spec.GlobalIP,
		GatewayDNSSuffix:             obj.Spec.DNSSuffix,
		GatewayInternalDNSNameserver: gatewayInternalDNSNameServer,
		GatewayWgExtraPeersHash:      fn.Md5([]byte(extraPeersCfg.Data["peers.conf"])),
		GatewayDNSServers:            strings.Join(gatewayDNSServers, ","),
		GatewayServiceType:           obj.Spec.ServiceType,
		GatewayNodePort:              fn.DefaultIfNil(obj.Spec.NodePort),
		ClusterCIDR:                  obj.Spec.ClusterCIDR,
		ServiceCIDR:                  obj.Spec.SvcCIDR,
		IPManagerConfigName:          "gateway-ip-manager",
		IPManagerConfigNamespace:     obj.Spec.TargetNamespace,

		ImageWebhookServer:       r.Env.ImageWebhookServer,
		ImageIPManager:           r.Env.ImageIPManager,
		ImageIPBindingController: r.Env.ImageIPBindingController,
		ImageDNS:                 r.Env.ImageDNS,
		ImageLogsProxy:           r.Env.ImageLogsProxy,
	})
	if err != nil {
		return check.Failed(err)
	}

	fmt.Printf("deployment:\n\n%s\n\n", b)

	rr, err := r.yamlClient.ApplyYAML(ctx, b)
	if err != nil {
		return check.Failed(err)
	}
	req.AddToOwnedResources(rr...)

	deployment, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.TargetNamespace, obj.Name), &appsv1.Deployment{})
	if err != nil {
		return check.Failed(err)
	}

	for _, c := range deployment.Status.Conditions {
		if c.Type == appsv1.DeploymentAvailable {
			return check.Completed()
		}
	}

	return check.Failed(errors.Newf("deployment %s/%s is not available/ready yet", obj.Spec.TargetNamespace, obj.Name))
}

func (r *Reconciler) setupMutationWebhooks(req *rApi.Request[*networkingv1.Gateway]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(setupMutationWebhooks, req)

	webhookCertSecretName := fmt.Sprintf("%s-webhook-cert", obj.Name)
	webhookCertSecretNamespace := obj.Spec.TargetNamespace

	webhookCert, err := rApi.Get(ctx, r.Client, fn.NN(webhookCertSecretNamespace, webhookCertSecretName), &corev1.Secret{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return check.Failed(err)
		}

		caBundle, cert, key, err := GenTLSCert([]string{fmt.Sprintf("%s.%s.svc", obj.Name, obj.Spec.TargetNamespace)})
		if err != nil {
			return check.Failed(err)
		}

		webhookCert = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:            webhookCertSecretName,
				Namespace:       webhookCertSecretNamespace,
				OwnerReferences: []metav1.OwnerReference{fn.AsOwner(obj, true)},
			},
			Data: map[string][]byte{
				"ca.crt":  caBundle,
				"tls.crt": cert,
				"tls.key": key,
			},
		}
		if err := r.Create(ctx, webhookCert); err != nil {
			return check.Failed(err)
		}
	}

	b, err := templates.ParseBytes(r.templateWebhook, templates.WebhookTemplateArgs{
		NamePrefix: func() string {
			if v := getAccountName(obj); v != nil {
				return fmt.Sprintf("%s-%s", *v, obj.Name)
			}
			return obj.Name
		}(),
		Namespace:                 obj.Spec.TargetNamespace,
		OwnerReferences:           []metav1.OwnerReference{fn.AsOwner(obj, true)},
		WebhookServerImage:        "ghcr.io/kloudlite/operator/wireguard/apps/mutation-webhook:v1.0.7-nightly",
		WebhookServerCertCABundle: string(webhookCert.Data["ca.crt"]),

		WebhookNamespaceSelector: func() map[string]string {
			selector := map[string]string{
				constants.KloudliteGatewayEnabledLabel: "true",
			}
			if v := getAccountName(obj); v != nil {
				selector[constants.AccountNameKey] = *v
			}

			return selector
		}(),
		// WebhookNamespaceSelectorKey:   constants.KloudliteGatewayEnabledLabel,
		// WebhookNamespaceSelectorValue: "true",
		ServiceName: obj.Name,
	})
	if err != nil {
		return check.Failed(err)
	}

	rr, err := r.yamlClient.ApplyYAML(ctx, b)
	if err != nil {
		return check.Failed(err)
	}

	req.AddToOwnedResources(rr...)

	return check.Completed()
}

func (r *Reconciler) trackLoadBalancer(req *rApi.Request[*networkingv1.Gateway]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(trackLoadBalancer, req)

	svc, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.TargetNamespace, fmt.Sprintf("%s-wg", obj.Name)), &corev1.Service{})
	if err != nil {
		return check.Failed(err)
	}

	switch obj.Spec.ServiceType {
	case networkingv1.GatewayServiceTypeLoadBalancer:
		{
			if svc.Spec.Type != corev1.ServiceTypeLoadBalancer {
				return check.Failed(fmt.Errorf("failed to find a loadbalancer service"))
			}

			hosts := make([]string, 0, len(svc.Status.LoadBalancer.Ingress))
			for _, ingress := range svc.Status.LoadBalancer.Ingress {
				hosts = append(hosts, ingress.IP)
			}

			var port *int32
			for i := range svc.Spec.Ports {
				if svc.Spec.Ports[i].Name == "wireguard" {
					port = &svc.Spec.Ports[i].Port
				}
			}

			if port == nil {
				return check.Failed(fmt.Errorf("failed to find a nodeport for our gateway service"))
			}

			if obj.Spec.LoadBalancer == nil {
				obj.Spec.LoadBalancer = &networkingv1.GatewayLoadBalancer{}
			}

			hasUpdate := false
			if !reflect.DeepEqual(obj.Spec.LoadBalancer.Hosts, hosts) {
				hasUpdate = true
				obj.Spec.LoadBalancer.Hosts = hosts
			}
			if obj.Spec.LoadBalancer.Port != *port {
				hasUpdate = true
				obj.Spec.LoadBalancer.Port = *port
			}

			if hasUpdate {
				if err := r.Update(ctx, obj); err != nil {
					return check.Failed(err)
				}
			}
		}
	case networkingv1.GatewayServiceTypeNodePort:
		{
			var port *int32
			for i := range svc.Spec.Ports {
				if svc.Spec.Ports[i].Name == "wireguard" {
					port = &svc.Spec.Ports[i].Port
				}
			}

			if obj.Spec.NodePort == nil || *obj.Spec.NodePort != *port {
				obj.Spec.NodePort = port
				if err := r.Update(ctx, obj); err != nil {
					return check.Failed(err)
				}
			}
		}
	default:
		{
			return check.Failed(fmt.Errorf("unsupported service type: %s", obj.Spec.ServiceType))
		}
	}

	return check.Completed()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig(), kubectl.YAMLClientOpts{Logger: r.logger})

	var err error
	r.templateDeployment, err = templates.Read(templates.GatewayDeploymentTemplate)
	if err != nil {
		return err
	}

	r.templateWebhook, err = templates.Read(templates.WebhookTemplate)
	if err != nil {
		return err
	}

	r.templateRBAC, err = templates.Read(templates.GatewayDeploymentRBACTemplate)
	if err != nil {
		return err
	}

	builder := ctrl.NewControllerManagedBy(mgr).For(&networkingv1.Gateway{})
	builder.Owns(&appsv1.Deployment{})
	builder.Owns(&corev1.Service{})
	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
