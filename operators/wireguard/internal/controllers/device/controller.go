package device

import (
	"context"
	"fmt"
	"slices"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"

	wgv1 "github.com/kloudlite/operator/apis/wireguard/v1"
	devinfo "github.com/kloudlite/operator/apps/coredns/dev-info"
	"github.com/kloudlite/operator/operators/wireguard/internal/env"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	"github.com/kloudlite/operator/pkg/templates"
	networkingv1 "k8s.io/api/networking/v1"
	apiLabels "k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	DEVICE_KEY_PREFIX     = "wg-keys-"
	DEVICE_CONFIG_PREFIX  = "wg-configs-"
	WG_SERVER_NAME_PREFIX = "wg-server-"
	DNS_NAME_PREFIX       = "wg-dns-"
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
	// NSReady        string = "namespace-ready"
	DnsConfigReady string = "dns-config-ready"

	KeysAndSecretReady string = "keys-and-secret-ready"
	ServerSvcReady     string = "server-svc-ready"
	ConfigReady        string = "config-ready"
	ServicesSynced     string = "services-synced"
	ServerReady        string = "server-ready"

	DeviceDeleted string = "device-deleted"
)

var (
	DEV_CHECKLIST = []rApi.CheckMeta{
		{Name: ServerSvcReady, Title: "Ensuring server service is created"},
		{Name: DnsConfigReady, Title: "Ensuring DNS config is ready"},
		{Name: KeysAndSecretReady, Title: "Ensuring keys and secret are ready"},
		{Name: ConfigReady, Title: "Ensuring device config is ready"},
		{Name: ServicesSynced, Title: "Ensuring services are synced"},
		{Name: ServerReady, Title: "Ensuring server is ready"},
	}

	DEV_DESTROY_CHECKLIST = []rApi.CheckMeta{
		{Name: DeviceDeleted, Title: "Cleaning up device resources"},
	}
)

// +kubebuilder:rbac:groups=wireguard.kloudlite.io,resources=devices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=wireguard.kloudlite.io,resources=devices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=wireguard.kloudlite.io,resources=devices/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &wgv1.Device{})
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

	if step := req.EnsureCheckList(DEV_CHECKLIST); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureChecks(DnsConfigReady, KeysAndSecretReady, ConfigReady, ServerSvcReady, ServicesSynced, ServerReady); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureSvcCreated(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureDnsConfig(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureSecretKeys(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureConfig(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureServiceSync(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureDeploy(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *Reconciler) ensureDnsConfig(req *rApi.Request[*wgv1.Device]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation, State: rApi.RunningState}

	failed := func(err error) stepResult.Result {
		return req.CheckFailed(DnsConfigReady, check, err.Error())
	}

	req.LogPreCheck(DnsConfigReady)
	defer req.LogPostCheck(DnsConfigReady)

	kubeDns, err := rApi.Get(ctx, r.Client, fn.NN("kube-system", "kube-dns"), &corev1.Service{})
	if err != nil {
		return failed(err)
	}

	dnsConfigName := fmt.Sprint(DNS_NAME_PREFIX, obj.Name)

	getDnsConfig := func() ([]byte, *string, error) {
		corefile := ""

		for _, cn := range obj.Spec.CNameRecords {
			corefile += fmt.Sprintf("\n\trewrite name %s %s", cn.Host, cn.Target)
		}

		if r.Env.TlsDomainPrefix != "" {
			corefile += fmt.Sprintf("\n\trewrite name %s %s",
				fmt.Sprintf("%s.%s", r.Env.TlsDomainPrefix, r.Env.DnsHostedZone),
				fmt.Sprintf("%s.%s.svc.%s", obj.Name, obj.Namespace, r.Env.ClusterInternalDns),
			)
		}

		corefile += fmt.Sprintf("\n\trewrite name regex (^[a-zA-Z0-9-_]+)[.]local {1}.%s.svc.%s answer auto",
			obj.Namespace,
			r.Env.ClusterInternalDns,
		)

		if obj.Spec.ActiveNamespace != nil {
			corefile += fmt.Sprintf("\n\trewrite name regex ^([a-zA-Z0-9-]+)\\.?[^.]*$ {1}.%s.svc.%s answer auto",
				*obj.Spec.ActiveNamespace,
				r.Env.ClusterInternalDns,
			)

			// device namespace => ingress domains => rewrite rule: rewrite name s1.sample-cluster.kloudlite-dev.tenants.devc.kloudlite.io env-ingress.env-nxtcoder172.svc.cluster.local
			var ingressList networkingv1.IngressList
			if err := r.List(ctx, &ingressList, client.InNamespace(*obj.Spec.ActiveNamespace)); err != nil {
				return nil, nil, err
			}

			domains := make(map[string]struct{})
			for i := range ingressList.Items {
				if ingressList.Items[i].Spec.IngressClassName != nil && *ingressList.Items[i].Spec.IngressClassName == r.Env.DefaultIngressClass {
					continue
				}
				for j := range ingressList.Items[i].Spec.Rules {
					domains[ingressList.Items[i].Spec.Rules[j].Host] = struct{}{}
				}
			}

			for k := range domains {
				corefile += fmt.Sprintf("\n\trewrite name %s %s.%s.svc.%s", k, r.Env.EnvironmentIngressName, *obj.Spec.ActiveNamespace, r.Env.ClusterInternalDns)
			}
		}

		corefile = fmt.Sprintf(
			`
.:53 {
    errors
    health
    ready

%s

    forward . %s
    cache 30
    loop
    reload
    loadbalance
}
`,
			corefile, kubeDns.Spec.ClusterIP)

		b, err := templates.Parse(templates.Wireguard.DnsConfig, map[string]any{
			"namespace":     obj.Namespace,
			"rewrite-rules": corefile,
			"name":          obj.Name,
			"labels":        map[string]string{constants.WGDeviceSeceret: "true", constants.WGDeviceNameKey: obj.Name},
			"ownerRefs":     []metav1.OwnerReference{fn.AsOwner(obj, true)},
		})
		if err != nil {
			return nil, nil, err
		}

		return b, &corefile, nil
	}

	applyDnsConfig := func(configYaml []byte, corefileConfig string) error {
		if _, err := r.yamlClient.ApplyYAML(ctx, configYaml); err != nil {
			return err
		}

		return r.rolloutCoreDNS(req, corefileConfig)
	}

	if err := func() error {
		b, conf, err := getDnsConfig()
		if err != nil {
			return err
		}

		if conf == nil {
			return fmt.Errorf("corefile config is nil, aborting")
		}

		dConf, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, dnsConfigName), &corev1.ConfigMap{})
		if err != nil {
			if !apiErrors.IsNotFound(err) {
				return err
			}

			return applyDnsConfig(b, *conf)
		} else {
			if strings.TrimSpace(string(dConf.Data["Corefile"])) != strings.TrimSpace(string(*conf)) {
				return applyDnsConfig(b, *conf)
			}
		}

		return nil
	}(); err != nil {
		return failed(err)
	}

	check.Status = true
	check.State = rApi.CompletedState
	if check != req.Object.Status.Checks[DnsConfigReady] {
		fn.MapSet(&req.Object.Status.Checks, DnsConfigReady, check)
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}
	return req.Next()
}

func (r *Reconciler) ensureSecretKeys(req *rApi.Request[*wgv1.Device]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation, State: rApi.RunningState}

	failed := func(err error) stepResult.Result {
		return req.CheckFailed(KeysAndSecretReady, check, err.Error())
	}

	req.LogPreCheck(KeysAndSecretReady)
	defer req.LogPostCheck(KeysAndSecretReady)

	name := fmt.Sprint(DEVICE_KEY_PREFIX, obj.Name)

	if err := func() error {
		if _, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, name), &corev1.Secret{}); err != nil {
			if !apiErrors.IsNotFound(err) {
				return err
			}

			// creation new secret
			{
				pub, priv, err := GenerateWgKeys()
				if err != nil {
					return err
				}

				ip := []byte("10.13.0.2")

				sPub, sPriv, err := GenerateWgKeys()
				if err != nil {
					return err
				}

				sIp := []byte("10.13.0.1")

				if _, err := r.yamlClient.Apply(ctx, &corev1.Secret{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Secret",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:            name,
						Namespace:       obj.Namespace,
						Labels:          map[string]string{constants.WGDeviceSeceret: "true", constants.WGDeviceNameKey: obj.Name},
						OwnerReferences: []metav1.OwnerReference{fn.AsOwner(obj, true)},
					},
					Data: map[string][]byte{
						"device-private-key": priv,
						"device-public-key":  pub,
						"device-ip":          ip,

						"server-private-key": sPriv,
						"server-public-key":  sPub,
						"server-ip":          sIp,
					},
				}); err != nil {
					return err
				}

			}
			return fmt.Errorf("secret was not available, new secret of keys and ip created")
		}

		return nil
	}(); err != nil {
		return failed(err)
	}

	check.Status = true
	check.State = rApi.CompletedState
	if check != req.Object.Status.Checks[KeysAndSecretReady] {
		fn.MapSet(&req.Object.Status.Checks, KeysAndSecretReady, check)
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}
	return req.Next()
}

func (r *Reconciler) ensureSvcCreated(req *rApi.Request[*wgv1.Device]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation, State: rApi.RunningState}

	failed := func(err error) stepResult.Result {
		return req.CheckFailed(ServerSvcReady, check, err.Error())
	}

	req.LogPreCheck(ServerSvcReady)
	defer req.LogPostCheck(ServerSvcReady)

	if err := func() error {
		if _, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, fmt.Sprint(WG_SERVER_NAME_PREFIX, obj.Name)), &corev1.Service{}); err != nil {
			if !apiErrors.IsNotFound(err) {
				return err
			}

			// created or update wg deployment
			if b, err := templates.Parse(templates.Wireguard.DeploySvc,
				map[string]any{
					"name":      obj.Name,
					"namespace": obj.Namespace,
					"ownerRefs": []metav1.OwnerReference{fn.AsOwner(obj, true)},
				}); err != nil {
				return err
			} else if _, err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
				return err
			}
		}

		return nil
	}(); err != nil {
		return failed(err)
	}

	check.Status = true
	check.State = rApi.CompletedState
	if check != req.Object.Status.Checks[ServerSvcReady] {
		fn.MapSet(&req.Object.Status.Checks, ServerSvcReady, check)
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureConfig(req *rApi.Request[*wgv1.Device]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{
		Generation: obj.Generation,
		State:      rApi.RunningState,
	}

	failed := func(err error) stepResult.Result {
		return req.CheckFailed(ConfigReady, check, err.Error())
	}

	req.LogPreCheck(ConfigReady)
	defer req.LogPostCheck(ConfigReady)

	devConfig, serverConfig, err := r.generateDeviceConfig(req)
	if err != nil {
		return failed(err)
	}

	conf, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, fmt.Sprint(DEVICE_CONFIG_PREFIX, obj.Name)), &corev1.Secret{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return failed(err)
		}

		if err := r.applyDeviceConfig(req, devConfig, serverConfig); err != nil {
			return failed(err)
		}
	} else {
		if string(conf.Data["config"]) != string(devConfig) || string(conf.Data["server-config"]) != string(serverConfig) {
			if err := r.applyDeviceConfig(req, devConfig, serverConfig); err != nil {
				return failed(err)
			}
		}
	}

	check.Status = true
	check.State = rApi.CompletedState
	if check != req.Object.Status.Checks[ConfigReady] {
		fn.MapSet(&req.Object.Status.Checks, ConfigReady, check)
		return req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) ensureServiceSync(req *rApi.Request[*wgv1.Device]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation, State: rApi.RunningState}

	req.LogPreCheck(ServicesSynced)
	defer req.LogPostCheck(ServicesSynced)

	failed := func(err error) stepResult.Result {
		return req.CheckFailed(ServicesSynced, check, err.Error())
	}

	req.LogPreCheck(ServicesSynced)
	defer req.LogPostCheck(ServicesSynced)

	applyFreshSvc := func() error {
		sPorts := []corev1.ServicePort{}
		for _, v := range obj.Spec.Ports {
			sPorts = append(
				sPorts, corev1.ServicePort{
					Name: fmt.Sprint("port-", v.Port),
					Port: v.Port,
					TargetPort: intstr.IntOrString{
						Type: 0,
						IntVal: func() int32 {
							if v.TargetPort == 0 {
								return v.Port
							}
							return v.TargetPort
						}(),
					},
				},
			)
		}

		if err := fn.KubectlApply(ctx, r.Client, &corev1.Service{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Service",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:            obj.Name,
				Namespace:       obj.Namespace,
				Labels:          map[string]string{constants.WGDeviceNameKey: obj.Name},
				OwnerReferences: []metav1.OwnerReference{fn.AsOwner(obj, true)},
			},
			Spec: corev1.ServiceSpec{
				Ports: func() []corev1.ServicePort {
					ports := []corev1.ServicePort{
						{Name: "kl-coredns", Port: 17171},
						{Name: "kl-coredns-https", Port: 17172}}
					return append(sPorts, ports...)
				}(),
				Selector: map[string]string{
					"kloudlite.io/pod-type": "wireguard-server",
					"kloudlite.io/device":   obj.Name,
				},
			},
		}); err != nil {
			return err
		}

		return nil
	}

	service, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Name), &corev1.Service{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return failed(err)
		}

		// update services
		if err := applyFreshSvc(); err != nil {
			return failed(err)
		}

		return failed(fmt.Errorf("no service was available, created new"))
	}

	if checkPortsDiffer(service.Spec.Ports, obj.Spec.Ports) {
		// update services
		if err := applyFreshSvc(); err != nil {
			return failed(err)
		}
	}

	// external service to active namespace
	if err := func() error {
		if obj.Spec.NoExternalService {
			// external service is not required, skippping this step
			return nil
		}

		var services corev1.ServiceList
		if err := r.List(ctx, &services, &client.ListOptions{
			LabelSelector: apiLabels.SelectorFromValidatedSet(map[string]string{constants.WGDeviceNameKey: obj.Name, "kloudlite.io/wg-svc-type": "external"}),
		}); err != nil {
			if !apiErrors.IsNotFound(err) {
				return err
			}
		}

		// if obj.Spec.AccountName == nil {
		if obj.Spec.ActiveNamespace != nil && services.Items != nil {

			externalSvcExists := false
			for _, svc := range services.Items {
				if svc.Namespace != *obj.Spec.ActiveNamespace {
					if err := r.Delete(ctx, &svc); err != nil {
						return err
					}
				} else {
					externalSvcExists = true
				}
			}

			if !externalSvcExists && *obj.Spec.ActiveNamespace != obj.Namespace {
				if _, err := r.yamlClient.Apply(ctx, &corev1.Service{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Service",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: *obj.Spec.ActiveNamespace,
						Name:      obj.Name,
						Labels:    map[string]string{constants.WGDeviceNameKey: obj.Name, "kloudlite.io/wg-svc-type": "external"},
					},
					Spec: corev1.ServiceSpec{
						ExternalName:    fmt.Sprintf("%s.%s.svc.%s", obj.Name, obj.Namespace, r.Env.ClusterInternalDns),
						SessionAffinity: corev1.ServiceAffinityNone,
						Type:            corev1.ServiceTypeExternalName,
					},
				}); err != nil {
					return err
				}
			}

			return nil
		}

		if obj.Spec.ActiveNamespace != nil && *obj.Spec.ActiveNamespace != obj.Namespace {
			return fn.KubectlApply(ctx, r.Client, &corev1.Service{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Service",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: *obj.Spec.ActiveNamespace, Name: obj.Name,
					Labels:          map[string]string{constants.WGDeviceNameKey: obj.Name, "kloudlite.io/wg-svc-type": "external"},
					OwnerReferences: []metav1.OwnerReference{fn.AsOwner(obj, true)},
				},
				Spec: corev1.ServiceSpec{
					ExternalName:    fmt.Sprintf("%s.%s.svc.%s", obj.Name, obj.Namespace, r.Env.ClusterInternalDns),
					SessionAffinity: corev1.ServiceAffinityNone,
					Type:            corev1.ServiceTypeExternalName,
				},
			})
		}

		return nil
	}(); err != nil {
		return failed(err)
	}

	check.Status = true
	check.State = rApi.CompletedState
	if check != req.Object.Status.Checks[ServicesSynced] {
		fn.MapSet(&req.Object.Status.Checks, ServicesSynced, check)
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

func (r *Reconciler) ensureDeploy(req *rApi.Request[*wgv1.Device]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation, State: rApi.RunningState}

	failed := func(err error) stepResult.Result {
		return req.CheckFailed(ServerReady, check, err.Error())
	}

	req.LogPreCheck(ServerReady)
	defer req.LogPostCheck(ServerReady)

	// check deployment
	if err := func() error {

		deviceInfo := devinfo.DeviceInfo{
			Name:        obj.Name,
			AccountName: r.Env.AccountName,
			ClusterName: r.Env.ClusterName,
		}

		devInfo, err := deviceInfo.ToBase64()
		if err != nil {
			devInfo = fn.Ptr("")
		}

		tlsAvailable := false
		tlsCertSecName := fmt.Sprintf("%s.%s-tls", r.Env.TlsDomainPrefix, r.Env.DnsHostedZone)
		if r.Env.TlsDomainPrefix != "" {
			if _, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, tlsCertSecName), &corev1.Secret{}); err == nil {
				tlsAvailable = true
			}
		}

		dep, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, fmt.Sprint(WG_SERVER_NAME_PREFIX, obj.Name)), &appsv1.Deployment{})
		if err != nil {
			if !apiErrors.IsNotFound(err) {
				return err
			}

			// created or update wg deployment
			b, err := templates.Parse(templates.Wireguard.Deploy, map[string]any{
				"name":          obj.Name,
				"namespace":     obj.Namespace,
				"ownerRefs":     []metav1.OwnerReference{fn.AsOwner(obj, true)},
				"tolerations":   []corev1.Toleration{{Operator: "Exists"}},
				"node-selector": obj.Spec.NodeSelector,
				"tls-cert-sec-name": func() string {
					if tlsAvailable {
						return tlsCertSecName
					}
					return ""
				}(),
				"dev-info": devInfo,
			})
			if err != nil {
				return err
			}

			if obj.Spec.Disabled {
				return nil
			}

			if _, err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
				return err
			}
		}

		if dep != nil && obj.Spec.Disabled {
			return r.Delete(ctx, dep)
		}

		return nil
	}(); err != nil {
		return failed(err)
	}

	check.Status = true
	if check != req.Object.Status.Checks[ServerReady] {
		fn.MapSet(&req.Object.Status.Checks, ServerReady, check)
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}
	return req.Next()
}

func (r *Reconciler) finalize(req *rApi.Request[*wgv1.Device]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation, State: rApi.RunningState}

	if !slices.Equal(obj.Status.CheckList, DEV_DESTROY_CHECKLIST) {
		req.Object.Status.CheckList = DEV_DESTROY_CHECKLIST
		if step := req.UpdateStatus(); !step.ShouldProceed() {
			return step
		}
	}

	failed := func(err error) stepResult.Result {
		check.State = rApi.ErroredState
		return req.CheckFailed(DeviceDeleted, check, err.Error())
	}

	var services corev1.ServiceList
	if err := r.List(ctx, &services, &client.ListOptions{
		LabelSelector: apiLabels.SelectorFromValidatedSet(map[string]string{constants.WGDeviceNameKey: obj.Name, "kloudlite.io/wg-svc-type": "external"}),
	}); err != nil {
		if !apiErrors.IsNotFound(err) {
			return failed(err)
		}
	} else {
		for _, svc := range services.Items {
			if err := r.Delete(ctx, &svc); err != nil {
				return failed(err)
			}
		}
	}

	return req.Finalize()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig(), kubectl.YAMLClientOpts{Logger: r.logger})

	builder := ctrl.NewControllerManagedBy(mgr).For(&wgv1.Device{})
	builder.WithEventFilter(rApi.ReconcileFilter())

	watchList := []client.Object{
		&corev1.Secret{},
		&corev1.ConfigMap{},
		&corev1.Service{},
		&appsv1.Deployment{},
	}

	for _, object := range watchList {
		builder.Watches(
			object,
			handler.EnqueueRequestsFromMapFunc(
				func(_ context.Context, obj client.Object) []reconcile.Request {
					if dev, ok := obj.GetLabels()[constants.WGDeviceNameKey]; ok {
						return []reconcile.Request{{NamespacedName: fn.NN(obj.GetNamespace(), dev)}}
					}

					return nil
				}),
		)
	}
	builder.Watches(
		&networkingv1.Ingress{},
		handler.EnqueueRequestsFromMapFunc(
			func(ctx context.Context, obj client.Object) []reconcile.Request {
				var devices wgv1.DeviceList
				if err := r.List(ctx, &devices, &client.ListOptions{
					LabelSelector: apiLabels.SelectorFromValidatedSet(map[string]string{
						"kloudlite.io/active.namespace": obj.GetNamespace(),
					}),
				}); err != nil {
					return nil
				}

				rr := make([]reconcile.Request, 0, len(devices.Items))
				for i := range devices.Items {
					rr = append(rr, reconcile.Request{NamespacedName: fn.NN(devices.Items[i].Namespace, devices.Items[i].Name)})
				}

				return rr
			}),
	)

	return builder.Complete(r)
}
