package globalvpn

import (
	"bytes"
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"

	ctrl "sigs.k8s.io/controller-runtime"

	wgv1 "github.com/kloudlite/operator/apis/wireguard/v1"
	appCommon "github.com/kloudlite/operator/operators/wireguard/apps/multi-cluster/apps/common"
	"github.com/kloudlite/operator/operators/wireguard/apps/multi-cluster/apps/server"
	"github.com/kloudlite/operator/operators/wireguard/apps/multi-cluster/mpkg/wg"
	"github.com/kloudlite/operator/operators/wireguard/internal/controllers/global-vpn/templates"
	"github.com/kloudlite/operator/operators/wireguard/internal/env"
	"github.com/kloudlite/operator/pkg/constants"
	"github.com/kloudlite/operator/pkg/errors"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	ResourceNamespace = "kl-global-vpn"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	logger     logging.Logger
	Name       string
	yamlClient kubectl.YAMLClient
	Env        *env.Env
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	NSReady                   string = "namespace-ready"
	GWReady                   string = "gateway-ready"
	AgtReady                  string = "agent-ready"
	SpecReady                 string = "spec-ready"
	TrackNodePort             string = "track-node-port"
	UpdateCustomCoreDNSConfig string = "update-custom-coredns-config"

	// ConnectDeleted string = "connect-deleted"
)

var CONN_CHECKLIST = []rApi.CheckMeta{
	{Name: NSReady, Title: "making sure namespace is ready"},
	{Name: SpecReady, Title: "making sure spec data is ready"},
	{Name: GWReady, Title: "making sure gateway is ready"},
	{Name: AgtReady, Title: "making sure agent is ready"},
	{Name: TrackNodePort, Title: "making sure agent is ready", Debug: true},
	{Name: UpdateCustomCoreDNSConfig, Title: "updates coredns config to acknowledge gateway"},
}

// CONN_DESTROY_CHECKLIST = []rApi.CheckMeta{
// 	{Name: ConnectDeleted, Title: "Cleaning up resources"},
// }

// +kubebuilder:rbac:groups=wireguard.kloudlite.io,resources=connections,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=wireguard.kloudlite.io,resources=connections/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=wireguard.kloudlite.io,resources=connections/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &wgv1.GlobalVPN{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}

		return ctrl.Result{}, nil
	}

	req.PreReconcile()
	defer req.PostReconcile()

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureCheckList(CONN_CHECKLIST); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	// TODO: add checks here
	if step := req.EnsureChecks(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureNs(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.patchDefaults(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconGateway(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconAgent(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.trackNodePort(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.updateCoreDNSConfig(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *Reconciler) ensureNs(req *rApi.Request[*wgv1.GlobalVPN]) stepResult.Result {
	ctx, _ := req.Context(), req.Object
	check := rApi.NewRunningCheck(NSReady, req)

	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ResourceNamespace}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, ns, func() error {
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	return check.Completed()
}

func (r *Reconciler) patchDefaults(req *rApi.Request[*wgv1.GlobalVPN]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(SpecReady, req)

	updated := false
	if obj.Spec.GatewayResources == nil {
		updated = true
		obj.Spec.GatewayResources = &corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("50m"),
				corev1.ResourceMemory: resource.MustParse("64Mi"),
			},
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("25m"),
				corev1.ResourceMemory: resource.MustParse("32Mi"),
			},
		}
	}

	if obj.Spec.AgentsResources == nil {
		updated = true
		obj.Spec.AgentsResources = &corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("20m"),
				corev1.ResourceMemory: resource.MustParse("24Mi"),
			},
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("10m"),
				corev1.ResourceMemory: resource.MustParse("12Mi"),
			},
		}
	}

	if obj.Spec.WgInterface == nil {
		updated = true
		obj.Spec.WgInterface = fn.New("kl0")
	}

	wgCreds, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.WgRef.Namespace, obj.Spec.WgRef.Name), &corev1.Secret{})
	if err != nil {
		return check.Failed(err)
	}

	wc, err := fn.ParseFromSecret[wgv1.WgParams](wgCreds)
	if err != nil {
		return check.Failed(err)
	}

	if wc.IP == "" {
		return check.Failed(fmt.Errorf("wg gateway IP must be provided"))
	}

	if wc.DNSServer == nil {
		s, err := rApi.Get(ctx, r.Client, fn.NN("kube-system", "kube-dns"), &corev1.Service{})
		if err != nil {
			return check.Failed(err)
		}
		wc.DNSServer = &s.Spec.ClusterIP
	}

	if wc.WgPrivateKey == "" {
		publ, priv, err := wg.GenerateWgKeys()
		if err != nil {
			return check.Failed(err)
		}

		wc.WgPublicKey = string(publ)
		wc.WgPrivateKey = string(priv)

		m, err := fn.JsonConvert[map[string]string](wc)
		if err != nil {
			return check.Failed(err)
		}

		wgCreds.StringData = m
		if err := r.Update(ctx, wgCreds); err != nil {
			return check.StillRunning(err)
		}
		return check.StillRunning(fmt.Errorf("waiting for controller to patch wg private/public keys"))
	}

	if updated {
		if err := r.Update(ctx, obj); err != nil {
			return check.Failed(err)
		}

		return check.StillRunning(fmt.Errorf("waiting for spec data to be updated"))
	}

	rApi.SetLocal(req, "wg-params", *wc)
	return check.Completed()
}

func (r *Reconciler) reconGateway(req *rApi.Request[*wgv1.GlobalVPN]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(GWReady, req)

	wc, ok := rApi.GetLocal[wgv1.WgParams](req, "wg-params")
	if !ok {
		return check.Failed(errors.NotInLocals("wg-params"))
	}

	peers := make([]appCommon.Peer, 0, len(obj.Spec.Peers))
	for _, peer := range obj.Spec.Peers {
		if peer.PublicKey == wc.WgPublicKey {
			continue
		}

		ipCidr := fmt.Sprintf("%s/32", peer.IP)

		ai := peer.AllowedIPs
		if !slices.Contains(ai, ipCidr) {
			ai = append(ai, ipCidr)
		}

		peers = append(peers, appCommon.Peer{
			PublicKey:  peer.PublicKey,
			Endpoint:   peer.Endpoint,
			AllowedIPs: ai,
		})
	}

	// ipCidr := fmt.Sprintf("%s/32", wc.IP)

	ipMap := map[string]string{}

	if wc.VirtualCidr != "" {
		var namespaces corev1.NamespaceList
		if err := r.List(
			ctx, &namespaces, &client.ListOptions{
				LabelSelector: labels.SelectorFromValidatedSet(map[string]string{constants.GVPNExposeNamespaceKey: "true"}),
			},
		); err != nil {
			return check.Failed(err)
		}

		ipIndex := 0
		for i := range namespaces.Items {
			var svcList corev1.ServiceList
			if err := r.List(ctx, &svcList, &client.ListOptions{Namespace: namespaces.Items[i].Name}); err != nil {
				r.logger.Error(err)
				continue
			}

			for j := range svcList.Items {
				ip, err := wg.GenIPAddr(ipIndex, wc.VirtualCidr)
				if err != nil {
					r.logger.Error(err)
				}

				ipMap[ip] = svcList.Items[j].Spec.ClusterIP
				ipIndex++
			}
		}
	}

	if wc.DNSServer == nil {
		return check.Failed(fmt.Errorf("dns server must be provided"))
	}

	sec := server.Config{
		PrivateKey:      string(wc.WgPrivateKey),
		IpAddress:       fmt.Sprintf("%s/32", wc.IP),
		Peers:           peers,
		IpForwardingMap: ipMap,
		DnsServer:       *wc.DNSServer,
	}

	secBytes, err := sec.ToYaml()
	if err != nil {
		return check.Failed(err)
	}

	gw, err := templates.ParseTemplate(templates.Gateway, map[string]interface{}{
		"name":      fmt.Sprintf("%s-gateway", obj.Name),
		"namespace": ResourceNamespace,
		"image": func() string {
			if r.Env.WgGatewayImage == "" {
				return constants.DefaultWgGatewayImage
			}
			return r.Env.WgGatewayImage
		}(),
		"resources":    *obj.Spec.GatewayResources,
		"serverConfig": string(secBytes),
		"ownerRefs":    []metav1.OwnerReference{fn.AsOwner(obj, true)},
		"interface":    obj.Spec.WgInterface,
		// "coredns-svc-ip": wc.DNSServer,
		// "nodeport":     wc.NodePort,
	})
	if err != nil {
		return check.Failed(err).NoRequeue()
	}

	if _, err = r.yamlClient.ApplyYAML(ctx, gw); err != nil {
		return check.Failed(err)
	}

	s, err := rApi.Get(ctx, r.Client, fn.NN(ResourceNamespace, fmt.Sprintf("%s-gateway-configs", obj.Name)), &corev1.Secret{})

	if err == nil && !slices.Equal(bytes.TrimSpace(secBytes), bytes.TrimSpace(s.Data["server-config"])) {
		if err := fn.RolloutRestart(r.Client, fn.Deployment, ResourceNamespace, map[string]string{
			constants.WGConnectionNameKey:         fmt.Sprintf("%s-gateway", obj.Name),
			"kloudlite.io/wg-global-vpn.resource": "gateway",
		}); err != nil {
			return check.Failed(err)
		}
	}
	return check.Completed()
}

func (r *Reconciler) reconAgent(req *rApi.Request[*wgv1.GlobalVPN]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(AgtReady, req)

	wc, ok := rApi.GetLocal[wgv1.WgParams](req, "wg-params")
	if !ok {
		return check.Failed(errors.NotInLocals("wg-params"))
	}

	agent, err := templates.ParseTemplate(templates.Agent, map[string]interface{}{
		"gatewayName":  fmt.Sprintf("%s-gateway", obj.Name),
		"name":         fmt.Sprintf("%s-agent", obj.Name),
		"namespace":    ResourceNamespace,
		"corednsSvcIp": wc.DNSServer,
		"resources":    *obj.Spec.AgentsResources,
		"image": func() string {
			if r.Env.WgAgentImage == "" {
				return constants.DefaultWgAgentImage
			}
			return r.Env.WgAgentImage
		}(),
		"interface": obj.Spec.WgInterface,
		"ownerRefs": []metav1.OwnerReference{fn.AsOwner(obj, true)},
	})
	if err != nil {
		return check.Failed(err).Err(nil)
	}

	if _, err = r.yamlClient.ApplyYAML(ctx, agent); err != nil {
		return check.Failed(err).Err(nil)
	}

	return check.Completed()
}

func (r *Reconciler) finalize(req *rApi.Request[*wgv1.GlobalVPN]) stepResult.Result {
	// INFO: currently all resources will consist owner reference, so will be deleted automatically

	return req.Finalize()
}

func (r *Reconciler) trackNodePort(req *rApi.Request[*wgv1.GlobalVPN]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(TrackNodePort, req)

	svc, err := rApi.Get(ctx, r.Client, fn.NN(ResourceNamespace, fmt.Sprintf("%s-gateway-external", obj.Name)), &corev1.Service{})
	if err != nil {
		return check.Failed(err)
	}

	var nodeport *int32

	for i := range svc.Spec.Ports {
		if svc.Spec.Type == corev1.ServiceTypeNodePort {
			nodeport = &svc.Spec.Ports[i].NodePort
		}
	}

	if nodeport == nil {
		return check.Failed(fmt.Errorf("failed to find a nodeport for our gateway service"))
	}

	wgCreds, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.WgRef.Namespace, obj.Spec.WgRef.Name), &corev1.Secret{})
	if err != nil {
		return check.Failed(err)
	}

	wc, err := fn.ParseFromSecret[wgv1.WgParams](wgCreds)
	if err != nil {
		return check.Failed(err)
	}

	wc.NodePort = fn.New(fmt.Sprintf("%d", *nodeport))

	m, err := fn.JsonConvert[map[string]string](wc)
	if err != nil {
		return check.Failed(err)
	}

	wgCreds.StringData = m
	if err := r.Update(ctx, wgCreds); err != nil {
		return check.StillRunning(err)
	}

	return check.Completed()
}

func (r *Reconciler) updateCoreDNSConfig(req *rApi.Request[*wgv1.GlobalVPN]) stepResult.Result {
	ctx := req.Context()
	check := rApi.NewRunningCheck(UpdateCustomCoreDNSConfig, req)

	wc, ok := rApi.GetLocal[wgv1.WgParams](req, "wg-params")
	if !ok {
		return check.Failed(errors.NotInLocals("wg-params"))
	}

	configName := "coredns-custom"
	configmap := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: configName, Namespace: "kube-system"}}

	hasCorednsConfigUpdated := false

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, configmap, func() error {
		key := "kl-globalvpn.server"
		current := configmap.Data[key]
		updated, err := r.getCorednsConfig(req, current, *wc.DNSServer)
		if err != nil {
			return err
		}
		hasCorednsConfigUpdated = strings.TrimSpace(current) == updated
		if configmap.Data == nil {
			configmap.Data = make(map[string]string, 1)
		}
		if hasCorednsConfigUpdated {
			configmap.Data[key] = updated
			fn.MapSet(&configmap.Annotations, "kloudlite.io/last-updated-at", time.Now().Format(time.RFC3339))
		}
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	if hasCorednsConfigUpdated {
		if err := fn.RolloutRestart(r.Client, fn.Deployment, "kube-system", map[string]string{"k8s-app": "kube-dns"}); err != nil {
			return check.StillRunning(err)
		}
	}

	return check.Completed()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig(), kubectl.YAMLClientOpts{Logger: r.logger})

	builder := ctrl.NewControllerManagedBy(mgr).For(&wgv1.GlobalVPN{})
	builder.WithEventFilter(rApi.ReconcileFilter())

	watchList := []client.Object{
		&corev1.Secret{},
		&corev1.Service{},
		&appsv1.Deployment{},
		&appsv1.DaemonSet{},
	}

	for _, object := range watchList {
		builder.Watches(
			object,
			handler.EnqueueRequestsFromMapFunc(
				func(_ context.Context, obj client.Object) []reconcile.Request {
					if conn, ok := obj.GetLabels()[constants.WGConnectionNameKey]; ok {
						return []reconcile.Request{{NamespacedName: fn.NN("", conn)}}
					}

					return nil
				}),
		)
	}

	builder.Watches(&corev1.Namespace{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, o client.Object) []reconcile.Request {
		if v, ok := o.GetLabels()[constants.GVPNExposeNamespaceKey]; ok && v == "true" {

			var gvpnList wgv1.GlobalVPNList
			if err := r.List(ctx, &gvpnList, &client.ListOptions{}); err != nil {
				return nil
			}

			if len(gvpnList.Items) != 1 {
				return nil
			}

			return []reconcile.Request{{NamespacedName: fn.NN(gvpnList.Items[0].GetNamespace(), gvpnList.Items[0].GetName())}}
		}
		return nil
	}))

	return builder.Complete(r)
}
