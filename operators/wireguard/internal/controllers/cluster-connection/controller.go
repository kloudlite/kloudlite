package cluster_connection

import (
	"context"
	"fmt"
	"slices"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"

	ctrl "sigs.k8s.io/controller-runtime"

	wgv1 "github.com/kloudlite/operator/apis/wireguard/v1"
	appCommon "github.com/kloudlite/operator/operators/wireguard/apps/multi-cluster/apps/common"
	"github.com/kloudlite/operator/operators/wireguard/apps/multi-cluster/apps/server"
	"github.com/kloudlite/operator/operators/wireguard/apps/multi-cluster/mpkg/wg"
	"github.com/kloudlite/operator/operators/wireguard/internal/controllers/cluster-connection/templates"
	"github.com/kloudlite/operator/operators/wireguard/internal/env"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

/*
steps to be implemented:
[x] ensure namespace is ready
[x] ensure spec datas
[x] ensure gateway created and up to date
[x] ensure agent created and up to date
[x] handle delete

TODO: yet to decide on the following:
[ ] service discovery ( service discovery is not decided yet )
*/

const (
	ResourceNamespace = "kl-cluster-connection"
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
	NSReady   string = "namespace-ready"
	GWReady   string = "gateway-ready"
	AgtReady  string = "agent-ready"
	SpecReady string = "spec-ready"

	// ConnectDeleted string = "connect-deleted"
)

var (
	CONN_CHECKLIST = []rApi.CheckMeta{
		{Name: NSReady, Title: "making sure namespace is ready"},
		{Name: SpecReady, Title: "making sure spec data is ready"},
		{Name: GWReady, Title: "making sure gateway is ready"},
		{Name: AgtReady, Title: "making sure agent is ready"},
	}

	// CONN_DESTROY_CHECKLIST = []rApi.CheckMeta{
	// 	{Name: ConnectDeleted, Title: "Cleaning up resources"},
	// }
)

// +kubebuilder:rbac:groups=wireguard.kloudlite.io,resources=connections,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=wireguard.kloudlite.io,resources=connections/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=wireguard.kloudlite.io,resources=connections/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &wgv1.ClusterConnection{})
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

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *Reconciler) ensureNs(req *rApi.Request[*wgv1.ClusterConnection]) stepResult.Result {
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

func (r *Reconciler) patchDefaults(req *rApi.Request[*wgv1.ClusterConnection]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(SpecReady, req)

	updated := false

	if obj.Spec.Id == 0 || obj.Spec.Id > 499 {
		return check.Failed(fmt.Errorf("id should be between 1 and 499"))
	}

	if obj.Spec.DnsServer == nil {
		s, err := rApi.Get(ctx, r.Client, fn.NN("kube-system", "kube-dns"), &corev1.Service{})
		if err != nil {
			return check.Failed(err)
		}

		obj.Spec.DnsServer = fn.New(s.Spec.ClusterIP)
		updated = true
	}

	ip, err := wg.GetRemoteDeviceIp(int64(obj.Spec.Id), r.Env.WgIpBase)
	if err != nil {
		return check.Failed(err)
	}

	if obj.Spec.IpAddress == nil || *obj.Spec.IpAddress != string(ip) {
		obj.Spec.IpAddress = fn.New(string(ip))
		updated = true
	}

	secName := fmt.Sprintf("%s-gateway-configs", obj.Name)

	createSec := func() (*string, *string, error) {
		publ, priv, err := wg.GenerateWgKeys()
		if err != nil {
			return nil, nil, err
		}

		se := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: secName, Namespace: ResourceNamespace}}
		if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, se, func() error {
			se.Data = map[string][]byte{
				"private-key": priv,
			}
			return nil
		}); err != nil {
			return nil, nil, err
		}

		return fn.New(string(publ)), fn.New(string(priv)), nil
	}

	if err := func() error {
		sec, err := rApi.Get(ctx, r.Client, fn.NN(ResourceNamespace, secName), &corev1.Secret{})
		if err != nil {
			if apiErrors.IsNotFound(err) {
				return nil
			}

			pub, _, err := createSec()
			if err != nil {
				return err
			}

			obj.Spec.PublicKey = pub
			updated = true

			return nil
		}

		pvKey, ok := sec.Data["private-key"]

		if !ok {

			pub, _, err := createSec()
			if err != nil {
				return err
			}

			obj.Spec.PublicKey = pub
			updated = true

			return nil
		}

		b, err := wg.GeneratePublicKey(string(pvKey))
		if obj.Spec.PublicKey == nil || string(b) != *obj.Spec.PublicKey {
			obj.Spec.PublicKey = fn.New(string(b))
			updated = true
			return nil
		}

		return nil
	}(); err != nil {
		return check.Failed(err)
	}

	if obj.Spec.PublicKey == nil {
		s, err := rApi.Get(ctx, r.Client, fn.NN(ResourceNamespace, secName), &corev1.Secret{})
		if err != nil {
			return check.Failed(err)
		}

		pvKey, ok := s.Data["private-key"]
		if !ok {
			return check.Failed(fmt.Errorf("private key not found"))
		}

		b, err := wg.GeneratePublicKey(string(pvKey))
		if err != nil {
			return check.Failed(err)
		}

		obj.Spec.PublicKey = fn.New(string(b))

		updated = true
	}

	if obj.Spec.Interface == nil {
		obj.Spec.Interface = fn.New(fmt.Sprintf("kl-%d", obj.Spec.Id))
		updated = true
	}

	if obj.Spec.Nodeport == nil {
		if s, err := rApi.Get(ctx, r.Client, fn.NN(ResourceNamespace, fmt.Sprintf("%s-gateway-external", obj.Name)), &corev1.Service{}); err == nil {
			obj.Spec.Nodeport = fn.New(int(s.Spec.Ports[0].NodePort))
			updated = true
		}
	}

	if obj.Spec.GatewayResources == nil {
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
		updated = true
	}

	if obj.Spec.AgentsResources == nil {
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
		updated = true
	}

	if updated {
		if err := r.Client.Update(ctx, obj); err != nil {
			return check.Failed(err)
		}

		return check.StillRunning(fmt.Errorf("waiting for spec data to be updated"))
	}

	return check.Completed()
}

func (r *Reconciler) reconGateway(req *rApi.Request[*wgv1.ClusterConnection]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(GWReady, req)

	var peers []appCommon.Peer
	for i, peer := range obj.Spec.Peers {
		if peer.Id == 0 || peer.Id > 499 {
			return check.Failed(fmt.Errorf("peer [%d]: id should be between 1 and 499", i)).Err(nil)
		}

		if peer.Id == obj.Spec.Id {
			continue
		}

		ip, err := wg.GetRemoteDeviceIp(int64(peer.Id), r.Env.WgIpBase)
		if err != nil {
			return check.Failed(err)
		}
		ipCidr := fmt.Sprintf("%s/32", ip)

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

	secName := fmt.Sprintf("%s-gateway-configs", obj.Name)
	xSec, err := rApi.Get(ctx, r.Client, fn.NN(ResourceNamespace, secName), &corev1.Secret{})
	if err != nil {
		return check.Failed(err)
	}

	pvKey, ok := xSec.Data["private-key"]
	if !ok {
		return check.Failed(fmt.Errorf("private key not found"))
	}

	sec := server.Config{
		PrivateKey: string(pvKey),
		IpAddress:  fmt.Sprintf("%s/32", *obj.Spec.IpAddress),
		Peers:      peers,
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
		"interface":    *obj.Spec.Interface,
		"nodeport": func() int32 {
			if obj.Spec.Nodeport == nil {
				return 0
			}

			return int32(*obj.Spec.Nodeport)
		}(),
	})
	if err != nil {
		return check.Failed(err).Err(nil)
	}

	if _, err = r.yamlClient.ApplyYAML(ctx, gw); err != nil {
		return check.Failed(err).Err(nil)
	}

	s, err := rApi.Get(ctx, r.Client, fn.NN(ResourceNamespace, fmt.Sprintf("%s-gateway-configs", obj.Name)), &corev1.Secret{})
	if err == nil && !slices.Equal(secBytes, s.Data["server-config"]) {
		if err := fn.RolloutRestart(r.Client, fn.Deployment, ResourceNamespace, map[string]string{
			constants.WGConnectionNameKey:                 fmt.Sprintf("%s-gateway", obj.Name),
			"kloudlite.io/wg-cluster-connection.resource": "gateway",
		}); err != nil {
			return check.Failed(err)
		}
	}
	return check.Completed()
}

func (r *Reconciler) reconAgent(req *rApi.Request[*wgv1.ClusterConnection]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(AgtReady, req)

	agent, err := templates.ParseTemplate(templates.Agent, map[string]interface{}{
		"gatewayName":  fmt.Sprintf("%s-gateway", obj.Name),
		"name":         fmt.Sprintf("%s-agent", obj.Name),
		"namespace":    ResourceNamespace,
		"corednsSvcIp": *obj.Spec.DnsServer,
		"resources":    *obj.Spec.AgentsResources,
		"image": func() string {
			if r.Env.WgAgentImage == "" {
				return constants.DefaultWgAgentImage
			}
			return r.Env.WgAgentImage
		}(),
		"interface": *obj.Spec.Interface,
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

func (r *Reconciler) finalize(req *rApi.Request[*wgv1.ClusterConnection]) stepResult.Result {
	// INFO: currently all resources will consist owner reference, so will be deleted automatically

	return req.Finalize()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig(), kubectl.YAMLClientOpts{Logger: r.logger})

	builder := ctrl.NewControllerManagedBy(mgr).For(&wgv1.ClusterConnection{})
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

	return builder.Complete(r)
}
