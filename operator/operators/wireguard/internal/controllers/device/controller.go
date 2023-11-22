package device

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
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

/*

note: * denotes completed

ensure device and server keys *
ensure device and server config *

ensure device service is upto date *

ensure deployment created *
*/

const (
	DEVICE_KEY_PREFIX     = "wg-keys-"
	DEVICE_CONFIG_PREFIX  = "wg-configs-"
	WG_SERVER_NAME_PREFIX = "wg-server-"
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
	KeysAndSecretReady string = "keys-ready"
	ServerSvcReady     string = "server-svc-ready"
	DnsReady           string = "dns-ready"
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

	if step := req.EnsureChecks(KeysAndSecretReady, ServerSvcReady, DnsReady, ConfigReady, ServicesSynced, DeviceDeleted); !step.ShouldProceed() {
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

	if step := r.ensureSvcCreated(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureDnsSynced(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureServiceSync(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureDeploy(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureServiceSync(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureConfig(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true

	req.Object.Status.LastReconcileTime = &metav1.Time{Time: time.Now()}
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) rollout(req *rApi.Request[*wgv1.Device]) error {

	ctx, obj := req.Context(), req.Object

	depName := fmt.Sprint(WG_SERVER_NAME_PREFIX, obj.Name)
	_, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, depName), &appsv1.Deployment{})

	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return err
		}
		return nil
	}

	if _, err := fn.Kubectl("-n", obj.Namespace, "rollout", "restart", fmt.Sprintf("deployment/%s", depName)); err != nil {
		return err
	}

	return nil
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

				if err := fn.KubectlApply(ctx, r.Client, &corev1.Secret{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Secret",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: obj.Namespace, Name: name,
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

func (r *Reconciler) ensureSvcCreated(req *rApi.Request[*wgv1.Device]) stepResult.Result {

	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}
	failed := func(err error) stepResult.Result {
		return req.CheckFailed(ServerSvcReady, check, err.Error())
	}

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
	if check != checks[ServerSvcReady] {
		checks[ServerSvcReady] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureDnsSynced(req *rApi.Request[*wgv1.Device]) stepResult.Result {

	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}
	failed := func(err error) stepResult.Result {
		return req.CheckFailed(DnsReady, check, err.Error())
	}

	oldDns := obj.Spec.Dns

	s, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, "kl-coredns"), &corev1.Service{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return failed(err)
		}

		kubeDns, err2 := rApi.Get(ctx, r.Client, fn.NN("kube-system", "kube-dns"), &corev1.Service{})
		if err2 != nil {
			return failed(err2)
		}

		obj.Spec.Dns = &kubeDns.Spec.ClusterIP
	} else {
		obj.Spec.Dns = &s.Spec.ClusterIP
	}

	if err := func() error {

		if *obj.Spec.Dns != *oldDns {

			if err := r.Update(ctx, obj); err != nil {
				return err
			}

			if err := r.upsertConfig(req); err != nil {
				return err
			}

		}

		return nil
	}(); err != nil {
		return failed(err)
	}

	check.Status = true
	if check != checks[DnsReady] {
		checks[DnsReady] = check
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

	serverConf, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, fmt.Sprint(WG_SERVER_NAME_PREFIX, obj.Name)), &corev1.ConfigMap{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return failed(err)
		}

		err := r.upsertConfig(req)
		if err != nil {
			return failed(err)
		}
	} else {
		if serverConf.Annotations == nil {
			serverConf.Annotations = map[string]string{}
		}

		if ip, ok := serverConf.Annotations["kloudlite.io/dns-ip"]; !ok || ip != *obj.Spec.Dns {
			if err := r.upsertConfig(req); err != nil {
				return failed(err)
			}
		}
	}

	_, err = rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, fmt.Sprint(DEVICE_CONFIG_PREFIX, obj.Name)), &corev1.ConfigMap{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return failed(err)
		}

		err := r.upsertConfig(req)
		if err != nil {
			return failed(err)
		}
	}

	check.Status = true
	if check != checks[ConfigReady] {
		checks[ConfigReady] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) upsertConfig(req *rApi.Request[*wgv1.Device]) error {

	ctx, obj := req.Context(), req.Object

	configName := fmt.Sprint(DEVICE_CONFIG_PREFIX, obj.Name)
	secName := fmt.Sprint(DEVICE_KEY_PREFIX, obj.Name)

	wgService, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, fmt.Sprint(WG_SERVER_NAME_PREFIX, obj.Name)), &corev1.Service{})
	if err != nil {
		return err
	}

	if obj.Spec.Dns == nil {
		return fmt.Errorf("dns is not set")
	}

	sec, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, secName), &corev1.Secret{})
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
		ServerEndpoint: fmt.Sprintf("%s.%s.%s:%d", obj.Spec.ClusterName, obj.Spec.AccountName, func() string {
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

	data := Data{
		ServerIp:         string(sIp) + "/32",
		ServerPrivateKey: string(sPriv),
		DNS:              *obj.Spec.Dns,
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

	applyFresh := func() error {
		if err := fn.KubectlApply(ctx, r.Client, &corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Secret",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: configName, Namespace: obj.Namespace,
				Labels:          map[string]string{constants.WGDeviceNameKey: obj.Name},
				OwnerReferences: []metav1.OwnerReference{fn.AsOwner(obj, true)},
				Annotations: map[string]string{
					"kloudlite.io/dns-ip": *obj.Spec.Dns,
				},
			},
			Data: map[string][]byte{"config": out, "server-config": conf, "sysctl": []byte(`
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

	oConf, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, configName), &corev1.Secret{})
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

	if string(oConf.Data["config"]) != string(out) || string(oConf.Data["server-config"]) != string(conf) {
		if err := applyFresh(); err != nil {
			return err
		}
		return fmt.Errorf("device config were updated, updated applied")
	}

	return nil
}

func (r *Reconciler) ensureServiceSync(req *rApi.Request[*wgv1.Device]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	failed := func(err error) stepResult.Result {
		return req.CheckFailed(ServicesSynced, check, err.Error())
	}

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
				Namespace: obj.Namespace, Name: obj.Name,
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
	// check if region updated

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
		if _, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, fmt.Sprint(WG_SERVER_NAME_PREFIX, obj.Name)), &appsv1.Deployment{}); err != nil {
			if !apiErrors.IsNotFound(err) {
				return err
			}

			// created or update wg deployment
			if b, err := templates.Parse(templates.Wireguard.Deploy, map[string]any{
				"name": obj.Name, "isMaster": false, "namespace": obj.Namespace,
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
	if check != checks[ServerReady] {
		checks[ServerReady] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}
	return req.Next()
}

func (r *Reconciler) finalize(req *rApi.Request[*wgv1.Device]) stepResult.Result {
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
		&corev1.Service{},
		&appsv1.Deployment{},
	}

	for i := range watchList {
		builder.Watches(
			&source.Kind{Type: watchList[i]},
			handler.EnqueueRequestsFromMapFunc(
				func(obj client.Object) []reconcile.Request {
					if dev, ok := obj.GetLabels()[constants.WGDeviceNameKey]; ok {
						return []reconcile.Request{{NamespacedName: fn.NN(obj.GetNamespace(), dev)}}
					}

					if _, ok := obj.GetLabels()["kloudlite.io/coredns-svc"]; ok {
						var devices wgv1.DeviceList
						if err := r.List(context.Background(), &devices, &client.ListOptions{
							Namespace: obj.GetNamespace(),
						}); err != nil {
							return nil
						}

						res := make([]reconcile.Request, len(devices.Items))

						for i2, d := range devices.Items {
							res[i2] = reconcile.Request{NamespacedName: fn.NN(d.Namespace, d.Name)}
						}

						return res
					}

					return nil
				}),
		)
	}

	return builder.Complete(r)
}
