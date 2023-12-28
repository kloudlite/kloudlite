package device

import (
	"context"
	"fmt"
	"sort"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"

	wgv1 "github.com/kloudlite/operator/apis/wireguard/v1"
	"github.com/kloudlite/operator/operators/wireguard/internal/env"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	"github.com/kloudlite/operator/pkg/templates"
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
	NSReady        string = "namespace-ready"
	DnsConfigReady string = "dns-config-ready"

	KeysAndSecretReady string = "keys-and-secret-ready"
	ServerSvcReady     string = "server-svc-ready"
	ConfigReady        string = "config-ready"
	ServicesSynced     string = "services-synced"
	ServerReady        string = "server-ready"

	DeviceDeleted string = "device-deleted"
)

// +kubebuilder:rbac:groups=wireguard.kloudlite.io,resources=devices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=wireguard.kloudlite.io,resources=devices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=wireguard.kloudlite.io,resources=devices/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &wgv1.Device{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if r.Env.DeviceInfoNamespace == "" {
		r.Env.DeviceInfoNamespace = "device-info"
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	req.LogPreReconcile()
	defer req.LogPostReconcile()

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureChecks(NSReady, DnsConfigReady, KeysAndSecretReady, ConfigReady, ServerSvcReady, ServicesSynced, ServerReady, DeviceDeleted); !step.ShouldProceed() {
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
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, nil
}

func (r *Reconciler) rollout(req *rApi.Request[*wgv1.Device]) error {
	ctx, obj := req.Context(), req.Object

	depName := fmt.Sprint(WG_SERVER_NAME_PREFIX, obj.Name)
	_, err := rApi.Get(ctx, r.Client, fn.NN(r.Env.DeviceInfoNamespace, depName), &appsv1.Deployment{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return err
		}
		return nil
	}

	if _, err := fn.Kubectl("-n", r.Env.DeviceInfoNamespace, "rollout", "restart", fmt.Sprintf("deployment/%s", depName)); err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) ensureNs(req *rApi.Request[*wgv1.Device]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	failed := func(err error) stepResult.Result {
		return req.CheckFailed(NSReady, check, err.Error())
	}

	_, err := rApi.Get(ctx, r.Client, fn.NN("", r.Env.DeviceInfoNamespace), &corev1.Namespace{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return failed(err)
		}

		if _, err := r.yamlClient.Apply(ctx, &corev1.Namespace{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Namespace",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: r.Env.DeviceInfoNamespace,
			},
		}); err != nil {
			return failed(err)
		}
	}

	check.Status = true
	if check != checks[NSReady] {
		checks[DnsConfigReady] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureDnsConfig(req *rApi.Request[*wgv1.Device]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	failed := func(err error) stepResult.Result {
		return req.CheckFailed(DnsConfigReady, check, err.Error())
	}

	kubeDns, err := rApi.Get(ctx, r.Client, fn.NN("kube-system", "kube-dns"), &corev1.Service{})
	if err != nil {
		return failed(err)
	}

	dnsConfigName := fmt.Sprint(DNS_NAME_PREFIX, obj.Name)

	getDnsConfig := func(devices []wgv1.Device) ([]byte, *string, error) {
		sort.Slice(devices, func(i, j int) bool {
			return devices[i].Name < devices[j].Name
		})

		rewriteRules := ""
		for _, dev := range devices {
			if dev.Name == "" {
				continue
			}

			rewriteRules += fmt.Sprintf(
				"\n    rewrite name %s.%s %s.%s.svc.%s",
				dev.Name,
				"kl.local",
				dev.Name,
				r.Env.DeviceInfoNamespace,
				func() string {
					if r.Env.ClusterInternalDns == "" {
						return "cluster.local"
					}
					return r.Env.ClusterInternalDns
				}(),
			)
		}

		if obj.Spec.DeviceNamespace != nil {
			rewriteRules += fmt.Sprintf("\n\trewrite name regex ^([a-zA-Z0-9-]+)\\.?[^.]*$ {1}.%s.svc.%s answer auto",
				*obj.Spec.DeviceNamespace,
				func() string {
					if r.Env.ClusterInternalDns == "" {
						return "cluster.local"
					}
					return r.Env.ClusterInternalDns
				}(),
			)
		}

		rewriteRules = fmt.Sprintf(
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
			rewriteRules, kubeDns.Spec.ClusterIP)

		b, err := templates.Parse(templates.Wireguard.DnsConfig, map[string]any{
			"namespace":     r.Env.DeviceInfoNamespace,
			"rewrite-rules": rewriteRules,
			"name":          obj.Name,
			"labels":        map[string]string{constants.WGDeviceSeceret: "true", constants.WGDeviceNameKey: obj.Name},
			"ownerRefs":     []metav1.OwnerReference{fn.AsOwner(obj, true)},
		})

		if err != nil {
			return nil, nil, err
		}

		return b, &rewriteRules, nil
	}

	applyDnsCoinfig := func(configYaml []byte) error {

		if _, err := r.yamlClient.ApplyYAML(ctx, configYaml); err != nil {
			return err
		}

		return r.rollout(req)
	}

	if err := func() error {

		var devices wgv1.DeviceList
		if err := r.List(ctx, &devices, &client.ListOptions{
			Namespace: r.Env.DeviceInfoNamespace,
		}); err != nil {
			return err
		}

		b, conf, err := getDnsConfig(devices.Items)
		if err != nil {
			return err
		}

		dConf, err := rApi.Get(ctx, r.Client, fn.NN(r.Env.DeviceInfoNamespace, dnsConfigName), &corev1.ConfigMap{})
		if err != nil {
			if !apiErrors.IsNotFound(err) {
				return err
			}

			return applyDnsCoinfig(b)
		} else {

			if strings.TrimSpace(string(dConf.Data["Corefile"])) != strings.TrimSpace(string(*conf)) {
				return applyDnsCoinfig(b)
			}
		}

		return nil
	}(); err != nil {
		return failed(err)
	}

	check.Status = true
	if check != checks[DnsConfigReady] {
		checks[DnsConfigReady] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureSecretKeys(req *rApi.Request[*wgv1.Device]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}
	failed := func(err error) stepResult.Result {
		return req.CheckFailed(KeysAndSecretReady, check, err.Error())
	}

	name := fmt.Sprint(DEVICE_KEY_PREFIX, obj.Name)

	if err := func() error {
		// body

		if _, err := rApi.Get(ctx, r.Client, fn.NN(r.Env.DeviceInfoNamespace, name), &corev1.Secret{}); err != nil {
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
						Namespace: r.Env.DeviceInfoNamespace, Name: name,
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
	if check != checks[KeysAndSecretReady] {
		checks[KeysAndSecretReady] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) generateDeviceConfig(req *rApi.Request[*wgv1.Device]) (devConfig []byte, serverConfig []byte, errr error) {
	ctx, obj := req.Context(), req.Object

	secName := fmt.Sprint(DEVICE_KEY_PREFIX, obj.Name)

	if err := func() error {

		wgService, err := rApi.Get(ctx, r.Client, fn.NN(r.Env.DeviceInfoNamespace, fmt.Sprint(WG_SERVER_NAME_PREFIX, obj.Name)), &corev1.Service{})
		if err != nil {
			return err
		}

		sec, err := rApi.Get(ctx, r.Client, fn.NN(r.Env.DeviceInfoNamespace, secName), &corev1.Secret{})
		if err != nil {
			return err
		}

		pub, priv, ip, err := parseDeviceSec(sec)
		if err != nil {
			return err
		}

		serverPublicKey, sPriv, sIp, err := parseServerSec(sec)
		if err != nil {
			return err
		}

		out, err := templates.Parse(templates.Wireguard.DeviceConfig, deviceConfig{
			DeviceIp:        string(ip),
			DevicePvtKey:    string(priv),
			ServerPublicKey: string(serverPublicKey),
			ServerEndpoint: fmt.Sprintf("%s:%d", func() string {
				if r.Env.DnsHostedZone != "" {
					return r.Env.DnsHostedZone
				}
				return "dev.kloudlite.io"
			}(), wgService.Spec.Ports[0].NodePort),
			DNS:     "10.13.0.3",
			PodCidr: r.Env.WGPodCidr,
			SvcCidr: r.Env.WGServiceCidr,
		})
		if err != nil {
			return err
		}

		// setting devConfig
		devConfig = out

		data := Data{
			ServerIp:         string(sIp) + "/32",
			ServerPrivateKey: string(sPriv),
			DNS:              "127.0.0.1",
			Peers: []Peer{
				{
					PublicKey:  string(pub),
					AllowedIps: string(ip),
				},
			},
		}

		conf, err := templates.Parse(templates.Wireguard.Config, data)
		if err != nil {
			return err
		}

		serverConfig = conf
		return nil
	}(); err != nil {
		return nil, nil, err
	}

	return devConfig, serverConfig, nil
}

func (r *Reconciler) applyDeviceConfig(req *rApi.Request[*wgv1.Device], deviceConfig []byte, serverConfig []byte) error {

	ctx, obj := req.Context(), req.Object
	configName := fmt.Sprint(DEVICE_CONFIG_PREFIX, obj.Name)

	if err := fn.KubectlApply(ctx, r.Client, &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: configName, Namespace: r.Env.DeviceInfoNamespace,
			Labels:          map[string]string{constants.WGDeviceNameKey: obj.Name},
			OwnerReferences: []metav1.OwnerReference{fn.AsOwner(obj, true)},
		},
		Data: map[string][]byte{"config": deviceConfig, "server-config": serverConfig, "sysctl": []byte(`
net.ipv4.ip_forward=1
				`)},
	}); err != nil {
		return err
	}

	if err := r.rollout(req); err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) ensureSvcCreated(req *rApi.Request[*wgv1.Device]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}
	failed := func(err error) stepResult.Result {
		return req.CheckFailed(ServerSvcReady, check, err.Error())
	}

	if err := func() error {
		if _, err := rApi.Get(ctx, r.Client, fn.NN(r.Env.DeviceInfoNamespace, fmt.Sprint(WG_SERVER_NAME_PREFIX, obj.Name)), &corev1.Service{}); err != nil {
			if !apiErrors.IsNotFound(err) {
				return err
			}

			// created or update wg deployment
			if b, err := templates.Parse(templates.Wireguard.DeploySvc,
				map[string]any{
					"name":      obj.Name,
					"namespace": r.Env.DeviceInfoNamespace,
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
	if check != checks[ServerSvcReady] {
		checks[ServerSvcReady] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureConfig(req *rApi.Request[*wgv1.Device]) stepResult.Result {

	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}
	failed := func(err error) stepResult.Result {
		return req.CheckFailed(ConfigReady, check, err.Error())
	}

	devConfig, serverConfig, err := r.generateDeviceConfig(req)
	if err != nil {
		return failed(err)
	}

	conf, err := rApi.Get(ctx, r.Client, fn.NN(r.Env.DeviceInfoNamespace, fmt.Sprint(DEVICE_CONFIG_PREFIX, obj.Name)), &corev1.Secret{})
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
	if check != checks[ConfigReady] {
		checks[ConfigReady] = check
		return req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) ensureServiceSync(req *rApi.Request[*wgv1.Device]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

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
				Namespace: r.Env.DeviceInfoNamespace, Name: obj.Name,
				Labels:          map[string]string{constants.WGDeviceNameKey: obj.Name},
				OwnerReferences: []metav1.OwnerReference{fn.AsOwner(obj, true)},
			},
			Spec: corev1.ServiceSpec{
				Ports: func() []corev1.ServicePort {
					if len(sPorts) != 0 {
						return sPorts
					}

					return []corev1.ServicePort{{Name: "port-default", Port: 80, TargetPort: intstr.IntOrString{Type: 0, IntVal: 0}}}
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

	service, err := rApi.Get(ctx, r.Client, fn.NN(r.Env.DeviceInfoNamespace, obj.Name), &corev1.Service{})
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

	if err := func() error {
		var services corev1.ServiceList
		if err := r.List(ctx, &services, &client.ListOptions{
			LabelSelector: labels.SelectorFromValidatedSet(map[string]string{constants.WGDeviceNameKey: obj.Name, "kloudlite.io/wg-svc-type": "external"}),
		}); err != nil {
			if !apiErrors.IsNotFound(err) {
				return err
			}
		}

		// if obj.Spec.AccountName == nil {
		if obj.Spec.DeviceNamespace != nil && services.Items != nil {

			externalSvcExists := false
			for _, svc := range services.Items {
				if svc.Namespace != *obj.Spec.DeviceNamespace {
					if err := r.Delete(ctx, &svc); err != nil {
						return err
					}
				} else {
					externalSvcExists = true
				}
			}

			if !externalSvcExists && *obj.Spec.DeviceNamespace != r.Env.DeviceInfoNamespace {
				if _, err := r.yamlClient.Apply(ctx, &corev1.Service{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Service",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: *obj.Spec.DeviceNamespace,
						Name:      obj.Name,
						Labels:    map[string]string{constants.WGDeviceNameKey: obj.Name, "kloudlite.io/wg-svc-type": "external"},
					},
					Spec: corev1.ServiceSpec{
						ExternalName: fmt.Sprintf("%s.%s.svc.%s", obj.Name, r.Env.DeviceInfoNamespace, func() string {
							if r.Env.ClusterInternalDns == "" {
								return "cluster.local"
							}
							return r.Env.ClusterInternalDns
						}()),
						SessionAffinity: corev1.ServiceAffinityNone,
						Type:            corev1.ServiceTypeExternalName,
					},
				}); err != nil {
					return err
				}
			}

			return nil
		}

		if obj.Spec.DeviceNamespace != nil && *obj.Spec.DeviceNamespace != r.Env.DeviceInfoNamespace {
			return fn.KubectlApply(ctx, r.Client, &corev1.Service{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Service",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: *obj.Spec.DeviceNamespace, Name: obj.Name,
					Labels:          map[string]string{constants.WGDeviceNameKey: obj.Name, "kloudlite.io/wg-svc-type": "external"},
					OwnerReferences: []metav1.OwnerReference{fn.AsOwner(obj, true)},
				},
				Spec: corev1.ServiceSpec{
					ExternalName: fmt.Sprintf("%s.%s.svc.%s", obj.Name, r.Env.DeviceInfoNamespace, func() string {
						if r.Env.ClusterInternalDns == "" {
							return "cluster.local"
						}
						return r.Env.ClusterInternalDns
					}()),
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
	if check != checks[ServicesSynced] {
		checks[ServicesSynced] = check
		return req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) ensureDeploy(req *rApi.Request[*wgv1.Device]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}
	failed := func(err error) stepResult.Result {
		return req.CheckFailed(ServerReady, check, err.Error())
	}

	// check deployment
	if err := func() error {
		if _, err := rApi.Get(ctx, r.Client, fn.NN(r.Env.DeviceInfoNamespace, fmt.Sprint(WG_SERVER_NAME_PREFIX, obj.Name)), &appsv1.Deployment{}); err != nil {
			if !apiErrors.IsNotFound(err) {
				return err
			}

			// created or update wg deployment
			if b, err := templates.Parse(templates.Wireguard.Deploy, map[string]any{
				"name":          obj.Name,
				"namespace":     r.Env.DeviceInfoNamespace,
				"ownerRefs":     []metav1.OwnerReference{fn.AsOwner(obj, true)},
				"tolerations":   []corev1.Toleration{{Operator: "Exists"}},
				"node-selector": obj.Spec.NodeSelector,
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
	if check != checks[ServerReady] {
		checks[ServerReady] = check
		return req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) finalize(req *rApi.Request[*wgv1.Device]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	failed := func(err error) stepResult.Result {
		return req.CheckFailed(DeviceDeleted, check, err.Error())
	}

	var services corev1.ServiceList
	if err := r.List(ctx, &services, &client.ListOptions{
		LabelSelector: labels.SelectorFromValidatedSet(map[string]string{constants.WGDeviceNameKey: obj.Name, "kloudlite.io/wg-svc-type": "external"}),
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
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

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
				func(ctx context.Context, obj client.Object) []reconcile.Request {
					if dev, ok := obj.GetLabels()[constants.WGDeviceNameKey]; ok {
						return []reconcile.Request{{NamespacedName: fn.NN(obj.GetNamespace(), dev)}}
					}

					return nil
				}),
		)
	}

	return builder.Complete(r)
}
