package device

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	wgctrl_utils "github.com/kloudlite/operator/operators/wireguard/internal/controllers"
	"github.com/kloudlite/operator/operators/wireguard/internal/controllers/server"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	wgv1 "github.com/kloudlite/operator/apis/wireguard/v1"
	"github.com/kloudlite/operator/operators/wireguard/internal/env"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	"github.com/kloudlite/operator/pkg/templates"
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
	KeysAndSecretReady    string = "wg-keys-and-secret-ready"
	DeviceConfigReady     string = "device-config-ready"
	ServicesSynced        string = "services-synced"
	DnsRewriteRulesSynced string = "dns-rewrite-rules-synced"
	DeviceDeleted         string = "device-deleted"
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

	req.LogPreReconcile()
	defer req.LogPostReconcile()

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureChecks(KeysAndSecretReady, DeviceConfigReady, ServicesSynced, DnsRewriteRulesSynced, DeviceDeleted); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
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

	if step := r.reconDnsRewriteRules(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	req.Object.Status.LastReconcileTime = &metav1.Time{Time: time.Now()}

	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) ensureSecretKeys(req *rApi.Request[*wgv1.Device]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}
	failed := func(err error) stepResult.Result {
		return req.CheckFailed(KeysAndSecretReady, check, err.Error())
	}
	name := fmt.Sprintf("wg-device-keys-%s", obj.Name)

	if err := func() error {
		//body

		if _, err := rApi.Get(ctx, r.Client, fn.NN(getNs(obj), name), &corev1.Secret{}); err != nil {
			if !apiErrors.IsNotFound(err) {
				return err
			}

			// creation new secret
			{
				pub, priv, err := wgctrl_utils.GenerateWgKeys()
				if err != nil {
					return err
				}

				ip, err := wgctrl_utils.GetRemoteDeviceIp(int64(obj.Spec.Offset))
				if err != nil {
					return err
				}

				if err := fn.KubectlApply(ctx, r.Client, &corev1.Secret{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Secret",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: getNs(obj), Name: name,
						Labels: map[string]string{constants.WGDeviceSeceret: "true", constants.WGDeviceNameKey: obj.Name},
					},
					Data: map[string][]byte{
						"private-key": priv,
						"public-key":  pub,
						"ip":          ip,
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

func (r *Reconciler) ensureConfig(req *rApi.Request[*wgv1.Device]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}
	failed := func(err error) stepResult.Result {
		return req.CheckFailed(DeviceConfigReady, check, err.Error())
	}

	configName := fmt.Sprintf("wg-device-config-%s", obj.Name)
	secName := fmt.Sprintf("wg-device-keys-%s", obj.Name)

	if err := func() error {
		wgService, err := rApi.Get(ctx, r.Client, fn.NN(getNs(obj), "wireguard-service"), &corev1.Service{})
		if err != nil {
			return err
		}

		server, err := rApi.Get(ctx, r.Client, fn.NN("", obj.Spec.ServerName), &wgv1.Server{})
		if err != nil {
			return err
		}
		if server.Spec.PublicKey == nil {
			return fmt.Errorf("public key not found on server")
		}

		dnsSvc, err := rApi.Get(ctx, r.Client, fn.NN(getNs(obj), "coredns"), &corev1.Service{})
		if err != nil {
			return err
		}

		sec, err := rApi.Get(ctx, r.Client, fn.NN(getNs(obj), secName), &corev1.Secret{})
		if err != nil {
			return err
		}

		_, priv, ip, err := parseDeviceSec(sec)
		if err != nil {
			return err
		}

		out, err := templates.Parse(templates.Wireguard.DeviceConfig, deviceConfig{
			DeviceIp:        string(ip),
			DevicePvtKey:    string(priv),
			ServerPublicKey: *server.Spec.PublicKey,
			ServerEndpoint: fmt.Sprintf("%s.%s.%s:%d", server.Spec.ClusterName, server.Spec.AccountName, func() string {
				if r.Env.DnsHostedZone != "" {
					return r.Env.DnsHostedZone
				}
				return "dns.khost.dev"
			}(), wgService.Spec.Ports[0].NodePort),
			DNS:     dnsSvc.Spec.ClusterIP,
			PodCidr: r.Env.WGPodCidr,
			SvcCidr: r.Env.WGServiceCidr,
		})
		if err != nil {
			return err
		}

		applyFresh := func() error {
			if err := fn.KubectlApply(ctx, r.Client, &corev1.Secret{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Secret",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: configName, Namespace: getNs(obj),
					Labels:          map[string]string{constants.WGDeviceNameKey: obj.Name},
					OwnerReferences: []metav1.OwnerReference{fn.AsOwner(obj, true)},
				},
				Data: map[string][]byte{"config": out},
			}); err != nil {
				return err
			}
			return nil
		}

		oConf, err := rApi.Get(ctx, r.Client, fn.NN(getNs(obj), configName), &corev1.Secret{})
		if err != nil {
			if !apiErrors.IsNotFound(err) {
				return err
			}

			// create config
			if err := applyFresh(); err != nil {
				return err
			}
			return fmt.Errorf("no device config found, created new")
		}
		if string(oConf.Data["config"]) != string(out) {
			if err := applyFresh(); err != nil {
				return err
			}
			return fmt.Errorf("device config were updated, updated applied")
		}
		return nil
	}(); err != nil {
		return failed(err)
	}

	check.Status = true
	if check != checks[DeviceConfigReady] {
		checks[DeviceConfigReady] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureServiceSync(req *rApi.Request[*wgv1.Device]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	failed := func(err error) stepResult.Result {
		return req.CheckFailed(ServicesSynced, check, err.Error())
	}

	config, err := rApi.Get(ctx, r.Client, fn.NN(getNs(obj), "device-proxy-config"), &corev1.ConfigMap{})
	if err != nil {
		return failed(err)
	}

	applyFreshSvc := func() error {
		configData := []server.ConfigService{}

		oConfMap := map[string][]server.ConfigService{}
		if err = json.Unmarshal([]byte(config.Data["config.json"]), &oConfMap); err != nil {
			return err
		}

		if oConfMap["services"] != nil {
			configData = oConfMap["services"]
		}

		sPorts := []corev1.ServicePort{}
		for _, v := range obj.Spec.Ports {
			proxyPort, err := getPort(configData, fmt.Sprint(obj.Name, "-", v.Port))
			if err != nil {
				return err
			}

			sPorts = append(
				sPorts, corev1.ServicePort{
					Name: fmt.Sprint(obj.Name, "-", v.Port),
					Port: v.Port,
					TargetPort: intstr.IntOrString{
						Type:   0,
						IntVal: proxyPort,
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
				Namespace: getNs(obj), Name: obj.Name,
				Labels: map[string]string{constants.WGDeviceNameKey: obj.Name},
			},
			Spec: corev1.ServiceSpec{
				Ports: func() []corev1.ServicePort {
					if len(sPorts) != 0 {
						return sPorts
					}

					return []corev1.ServicePort{{Name: "tmp", Port: 80, TargetPort: intstr.IntOrString{Type: 0, IntVal: 0}}}
				}(),
				Selector: map[string]string{
					"app": "wireguard",
				},
			},
		}); err != nil {
			return err
		}

		return nil
	}

	service, err := rApi.Get(ctx, r.Client, fn.NN(getNs(obj), obj.Name), &corev1.Service{})
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
	// check if region updated

	check.Status = true
	if check != checks[ServicesSynced] {
		checks[ServicesSynced] = check
		return req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) reconDnsRewriteRules(req *rApi.Request[*wgv1.Device]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	failed := func(err error) stepResult.Result {
		return req.CheckFailed(DnsRewriteRulesSynced, check, err.Error())
	}

	dnsConf, err := rApi.Get(ctx, r.Client, fn.NN(getNs(obj), "coredns"), &corev1.ConfigMap{})
	if err != nil {
		return failed(err)
	}

	dnsDevices, ok := dnsConf.Data["devices"]
	if !ok {
		dnsDevices = "[]"
	}

	// dnsSvc, err := rApi.Get(ctx, r.Client, fn.NN(getNs(obj), "coredns"), &corev1.Service{})
	// if err != nil {
	// 	return failed(err)
	// }

	if err := func() error {
		var devices wgv1.DeviceList
		if err := r.List(ctx, &devices, &client.ListOptions{}); err != nil {
			return err
		}

		d := make([]string, 0)
		for _, dev := range devices.Items {
			d = append(d, dev.Name)
		}

		sort.Strings(d)
		var oldDevices []string
		if err := json.Unmarshal([]byte(dnsDevices), &oldDevices); err != nil {
			return err
		}
		sort.Strings(oldDevices)

		ok = func() bool {
			if len(oldDevices) != len(d) {
				return false
			}
			for i := 0; i < len(d); i++ {
				if d[i] != oldDevices[i] {
					return false
				}
			}
			return true
		}()
		if ok {
			return nil
		}

		// update
		{
			kubeDns, err := rApi.Get(ctx, r.Client, fn.NN("kube-system", "kube-dns"), &corev1.Service{})
			if err != nil {
				return err
			}

			rewriteRules := ""
			d := make([]string, 0)
			for _, dev := range devices.Items {
				d = append(d, dev.Name)
				if dev.Name == "" {
					continue
				}

				rewriteRules += fmt.Sprintf(
					"rewrite name %s.%s %s.%s.svc.cluster.local\n        ",
					dev.Name,
					"kl.local",
					dev.Name,
					getNs(obj),
				)
			}

			b, err := templates.Parse(templates.Wireguard.DnsConfig, map[string]any{
				"devices": d, "namespace": getNs(obj), "rewrite-rules": rewriteRules, "dns-ip": kubeDns.Spec.ClusterIP,
			})
			if err != nil {
				return err
			}

			if _, err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
				return err
			}

			if _, err := fn.Kubectl("-n", getNs(obj), "rollout", "restart", "deployment/coredns"); err != nil {
				return err
			}
		}

		return nil
	}(); err != nil {
		return failed(err)
	}

	check.Status = true
	if check != checks[DnsRewriteRulesSynced] {
		checks[DnsRewriteRulesSynced] = check
		return req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) finalize(req *rApi.Request[*wgv1.Device]) stepResult.Result {

	_, obj, _ := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	// TODO: have to write deletion logic
	k := "************** ~~>* deletion of device is not supported yet *<~~ **************"
	fmt.Println(k)
	return req.CheckFailed(DeviceDeleted, check, k)
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
		&corev1.Service{},
	}

	for i := range watchList {
		builder.Watches(
			&source.Kind{Type: watchList[i]},
			handler.EnqueueRequestsFromMapFunc(
				func(obj client.Object) []reconcile.Request {
					if dev, ok := obj.GetLabels()[constants.WGDeviceNameKey]; ok {
						return []reconcile.Request{{NamespacedName: fn.NN("", dev)}}
					}
					return nil
				}),
		)
	}

	return builder.Complete(r)
}
