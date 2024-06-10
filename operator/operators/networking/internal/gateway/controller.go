package gateway

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"

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
	generateWireguardKeys  string = "generate-wireguard-keys"
	setupDeploymentRBAC    string = "setup-deployment-rbac"
	setupGatewayDeployment string = "setup-gateway-deployment"
	setupMutationWebhooks  string = "setup-mutation-webhooks"
	trackLoadBalancer      string = "track-load-balancer"
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
	rApi.NewRunningCheck("finalizing", req)
	return req.Finalize()
}

func (r *Reconciler) patchDefaults(req *rApi.Request[*networkingv1.Gateway]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(patchDefaults, req)

	hasUpdate := false
	if obj.Spec.AdminNamespace == "" {
		hasUpdate = true
		obj.Spec.AdminNamespace = r.Env.GatewayAdminNamespace
	}

	if obj.Spec.WireguardKeysRef.Name == "" {
		hasUpdate = true
		obj.Spec.WireguardKeysRef.Name = "gateway-wg"
	}

	if obj.Spec.WireguardKeysRef.Namespace == "" {
		hasUpdate = true
		obj.Spec.WireguardKeysRef.Namespace = r.Env.GatewayAdminNamespace
	}

	if hasUpdate {
		if err := r.Update(ctx, obj); err != nil {
			return check.Failed(err)
		}
		return check.StillRunning(fmt.Errorf("waiting for resource to be updated"))
	}

	adminNamespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.AdminNamespace}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, adminNamespace, func() error {
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	return check.Completed()
}

func (r *Reconciler) generateWireguardKeys(req *rApi.Request[*networkingv1.Gateway]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(generateWireguardKeys, req)

	_, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.WireguardKeysRef.Namespace, obj.Spec.WireguardKeysRef.Name), &corev1.Secret{})
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
				Namespace: obj.Spec.WireguardKeysRef.Namespace,
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
			Namespace: obj.Spec.AdminNamespace,
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

	extraPeersCfg := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-wg-extra-peers", obj.Name), Namespace: obj.Spec.AdminNamespace}}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, extraPeersCfg, func() error {
		if extraPeersCfg.Data == nil {
			extraPeersCfg.Data = make(map[string]string, 1)
		}
		peers := make([]string, 0, len(obj.Spec.Peers))
		for _, peer := range obj.Spec.Peers {
			npeer := fmt.Sprintf("[Peer]\nPublicKey = %s\nAllowedIPs = %s\n PersistentKeepalive = 25\n", peer.PublicKey, strings.Join(peer.AllowedIPs, ", "))
			if peer.PublicEndpoint != nil {
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
		if peer.DNSSuffix != nil {
			gatewayDNSServers = append(gatewayDNSServers, fmt.Sprintf("%s=%s:53", *peer.DNSSuffix, peer.IP))
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
			Name:            obj.Name,
			Namespace:       obj.Spec.AdminNamespace,
			Labels:          map[string]string{"kloudlite.io/managed-by-gateway": "true"},
			OwnerReferences: []metav1.OwnerReference{fn.AsOwner(obj, true)},
		},

		ServiceAccountName: fmt.Sprintf("%s-svc-account", obj.Name),

		GatewayAdminAPIImage: "ghcr.io/kloudlite/operator/networking/cmd/ip-manager:v1.0.7-nightly",
		WebhookServerImage:   "ghcr.io/kloudlite/operator/networking/cmd/webhook:v1.0.7-nightly",

		GatewayWgSecretName:          obj.Spec.WireguardKeysRef.Name,
		GatewayGlobalIP:              obj.Spec.GlobalIP,
		GatewayDNSSuffix:             obj.Spec.DNSSuffix,
		GatewayInternalDNSNameserver: gatewayInternalDNSNameServer,
		GatewayWgExtraPeersHash:      fn.Md5([]byte(extraPeersCfg.Data["peers.conf"])),
		GatewayDNSServers:            strings.Join(gatewayDNSServers, ","),

		ClusterCIDR:              obj.Spec.ClusterCIDR,
		ServiceCIDR:              obj.Spec.SvcCIDR,
		IPManagerConfigName:      "gateway-ip-manager",
		IPManagerConfigNamespace: obj.Spec.AdminNamespace,
	})
	if err != nil {
		return check.Failed(err)
	}

	rr, err := r.yamlClient.ApplyYAML(ctx, b)
	if err != nil {
		return check.Failed(err)
	}
	req.AddToOwnedResources(rr...)

	deployment, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.AdminNamespace, obj.Name), &appsv1.Deployment{})
	if err != nil {
		return check.Failed(err)
	}

	for _, c := range deployment.Status.Conditions {
		if c.Type == appsv1.DeploymentAvailable {
			return check.Completed()
		}
	}

	return check.Failed(errors.Newf("deployment %s/%s is not available/ready yet", r.Env.GatewayAdminNamespace, obj.Name))
}

func (r *Reconciler) setupMutationWebhooks(req *rApi.Request[*networkingv1.Gateway]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(setupMutationWebhooks, req)

	webhookCertSecretName := fmt.Sprintf("%s-webhook-cert", obj.Name)
	webhookCertSecretNamespace := obj.Spec.AdminNamespace

	webhookCert, err := rApi.Get(ctx, r.Client, fn.NN(webhookCertSecretNamespace, webhookCertSecretName), &corev1.Secret{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return check.Failed(err)
		}

		caBundle, cert, key, err := GenTLSCert([]string{fmt.Sprintf("%s.%s.svc", obj.Name, obj.Spec.AdminNamespace)})
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
		NamePrefix:                obj.Name,
		Namespace:                 obj.Spec.AdminNamespace,
		OwnerReferences:           []metav1.OwnerReference{fn.AsOwner(obj, true)},
		WebhookServerImage:        "ghcr.io/kloudlite/operator/wireguard/apps/mutation-webhook:v1.0.7-nightly",
		WebhookServerCertCABundle: string(webhookCert.Data["ca.crt"]),

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

func syncDNS(ctx context.Context, registrationAddr string, dnsSuffix string, dnsAddr string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, fmt.Sprintf("%s/gateway/%s/%s", registrationAddr, dnsSuffix, dnsAddr), nil)
	if err != nil {
		return errors.NewEf(err, "creating http request")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.NewEf(err, "executing http request")
	}

	if resp.StatusCode != http.StatusOK {
		return errors.Newf("http request failed with status: %d, url: %s", resp.StatusCode, req.URL.String())
	}

	return nil
}

func (r *Reconciler) trackLoadBalancer(req *rApi.Request[*networkingv1.Gateway]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(trackLoadBalancer, req)

	svc, err := rApi.Get(ctx, r.Client, fn.NN(r.Env.GatewayAdminNamespace, fmt.Sprintf("%s-wg", obj.Name)), &corev1.Service{})
	if err != nil {
		return check.Failed(err)
	}

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
