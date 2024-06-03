package gateway

import (
	"context"
	"fmt"

	networkingv1 "github.com/kloudlite/operator/apis/networking/v1"
	"github.com/kloudlite/operator/operators/networking/internal/env"
	"github.com/kloudlite/operator/operators/networking/internal/gateway/templates"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
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
	}

	return check.Completed()
}

func (r *Reconciler) generateWireguardKeys(req *rApi.Request[*networkingv1.Gateway]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(generateWireguardKeys, req)

	scrt := &corev1.Secret{}
	if err := r.Get(ctx, fn.NN(obj.Spec.WireguardKeysRef.Namespace, obj.Spec.WireguardKeysRef.Name), scrt); err != nil {
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
			Namespace: r.Env.GatewayAdminNamespace,
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

	b, err := templates.ParseBytes(r.templateDeployment, templates.GatewayDeploymentArgs{
		ObjectMeta: metav1.ObjectMeta{Name: obj.Name, Namespace: r.Env.GatewayAdminNamespace, Labels: map[string]string{"kloudlite.io/managed-by-gateway": "true"}},

		ServiceAccountName: fmt.Sprintf("%s-svc-account", obj.Name),

		GatewayWgConfigURI:     fmt.Sprintf("%s/gateway/wg-config", r.Env.GatewayAdminSvcAddr),
		GatewayWgConfigHashURI: fmt.Sprintf("%s/gateway/wg-config-hash", r.Env.GatewayAdminSvcAddr),

		GatewayAdminAPIImage: "ghcr.io/kloudlite/operator/networking/cmd/ip-manager:v1.0.7-nightly",
		WebhookServerImage:   "ghcr.io/kloudlite/operator/networking/cmd/webhook:v1.0.7-nightly",

		GatewayWgSecretName:      obj.Spec.WireguardKeysRef.Name,
		GatewayGlobalIP:          obj.Spec.GlobalIP,
		ClusterCIDR:              obj.Spec.ClusterCIDR,
		ServiceCIDR:              obj.Spec.SvcCIDR,
		IPManagerConfigName:      "gateway-ip-manager",
		IPManagerConfigNamespace: r.Env.GatewayAdminNamespace,
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

func (r *Reconciler) setupMutationWebhooks(req *rApi.Request[*networkingv1.Gateway]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(setupMutationWebhooks, req)

	b, err := templates.ParseBytes(r.templateWebhook, templates.WebhookTemplateArgs{
		NamePrefix:         obj.Name,
		Namespace:          r.Env.GatewayAdminNamespace,
		OwnerReferences:    []metav1.OwnerReference{fn.AsOwner(obj, true)},
		WebhookServerImage: "ghcr.io/kloudlite/operator/wireguard/apps/mutation-webhook:v1.0.7-nightly",

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

	// ctx := context.TODO()
	//
	// if err := r.Get(ctx, fn.NN("", r.Env.GatewayAdminNamespace), &corev1.Namespace{}); err != nil {
	// 	if !apiErrors.IsNotFound(err) {
	// 		return err
	// 	}
	//
	// 	if err := r.Create(ctx, &corev1.Namespace{
	// 		ObjectMeta: metav1.ObjectMeta{
	// 			Name: r.Env.GatewayAdminNamespace,
	// 			Annotations: map[string]string{
	// 				constants.DescriptionKey: "Kloudlite Gateway Administration Namespace",
	// 			},
	// 		},
	// 	}); err != nil {
	// 		return err
	// 	}
	// }
	//
	// var gatewayList networkingv1.GatewayList
	// if err := r.List(context.TODO(), &gatewayList); err != nil {
	// 	return err
	// }
	//
	// if len(gatewayList.Items) != 1 {
	// 	return errors.Newf("must be only one gateway, but got %d", len(gatewayList.Items))
	// }
	//
	// gateway := gatewayList.Items[0]
	//
	// wgSecret := &corev1.Secret{}
	// if err := r.Get(ctx, fn.NN(gateway.Spec.WireguardKeysRef.Namespace, gateway.Spec.WireguardKeysRef.Name), wgSecret); err != nil {
	// 	if !apiErrors.IsNotFound(err) {
	// 		return err
	// 	}
	// 	key, err := wgtypes.GenerateKey()
	// 	if err != nil {
	// 		return err
	// 	}
	//
	// 	wgSecret = &corev1.Secret{
	// 		ObjectMeta: metav1.ObjectMeta{
	// 			Name:      gateway.Spec.WireguardKeysRef.Name,
	// 			Namespace: gateway.Spec.WireguardKeysRef.Namespace,
	// 		},
	// 		StringData: map[string]string{
	// 			"private_key": key.String(),
	// 			"public_key":  key.PublicKey().String(),
	// 		},
	// 	}
	//
	// 	if err := r.Create(ctx, wgSecret); err != nil {
	// 		return err
	// 	}
	// }
	//
	// s := strings.SplitN(gateway.Spec.SvcCIDR, "/", 2)
	// if len(s) != 2 {
	// 	return fmt.Errorf("invalid svcCIDR: %s", gateway.Spec.SvcCIDR)
	// }
	// cidrSuffix, err := strconv.Atoi(s[1])
	// if err != nil {
	// 	return err
	// }
	//
	// if r.Env.GatewayAdminHttpPort != 0 {
	// 	go func() {
	// 		if err := HttpServer(HttpServerArgs{
	// 			Port:       r.Env.GatewayAdminHttpPort,
	// 			kcli:       r.Client,
	// 			kclientset: r.yamlClient.Client(),
	//
	// 			IPManagerConfigName:      "gateway-ip-manager",
	// 			IPManagerConfigNamespace: r.Env.GatewayAdminNamespace,
	//
	// 			ClusterCIDR: gateway.Spec.ClusterCIDR,
	//
	// 			PodIPOffset:   int(math.Pow(2, float64(32-cidrSuffix)) + 1),
	// 			PodAllowedIPs: []string{"100.64.0.0/10"},
	//
	// 			PodPeersSecretName:      "gateway-pod-peers",
	// 			PodPeersSecretNamespace: r.Env.GatewayAdminNamespace,
	//
	// 			SvcCIDR: gateway.Spec.SvcCIDR,
	//
	// 			GatewayWgPublicKey:  string(wgSecret.Data["public_key"]),
	// 			GatewayWgEndpoint:   fmt.Sprintf("%s.%s.svc.cluster.local:51820", gateway.Name, r.Env.GatewayAdminNamespace),
	// 			GatewayGlobalIP:     gateway.Spec.GlobalIP,
	// 			GatewayWgPrivateKey: string(wgSecret.Data["private_key"]),
	// 		}); err != nil {
	// 			r.logger.Errorf(err, "Failed to start http server")
	// 			os.Exit(1)
	// 		}
	// 	}()
	// }

	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
