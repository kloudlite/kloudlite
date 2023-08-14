package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiLabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	wgv1 "github.com/kloudlite/operator/apis/wireguard/v1"
	wgctrl_utils "github.com/kloudlite/operator/operators/wireguard/internal/controllers"
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
	yamlClient *kubectl.YAMLClient
	Env        *env.Env
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	NSReady              string = "namespace-ready"
	WGKeysAndSecretReady string = "wg-keys-and-secret-ready"
	DeviceProxyReady     string = "device-proxy-ready"
	WGDeployReady        string = "wg-deployment-ready"
	CorednsDeployReady   string = "coredns-deployment-ready"
	ServerDeleted        string = "server-deleted"
)

// +kubebuilder:rbac:groups=wireguard.kloudlite.io,resources=servers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=wireguard.kloudlite.io,resources=servers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=wireguard.kloudlite.io,resources=servers/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &wgv1.Server{})
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

	if step := req.EnsureChecks(NSReady, WGKeysAndSecretReady, DeviceProxyReady, WGDeployReady, CorednsDeployReady); !step.ShouldProceed() {
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

	if step := r.ensureKeysAndSecret(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensurDevProxy(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureDeploy(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureCoreDNS(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	req.Object.Status.LastReconcileTime = &metav1.Time{Time: time.Now()}

	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) ensureNs(req *rApi.Request[*wgv1.Server]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	failed := func(err error) stepResult.Result {
		return req.CheckFailed(NSReady, check, err.Error())
	}

	if _, err := rApi.Get(ctx, r.Client, fn.NN("", getNs(obj)), &corev1.Namespace{}); err != nil {
		if !apiErrors.IsNotFound(err) {
			return failed(err)
		}

		if err := r.Create(ctx, &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: getNs(obj),
				Labels: map[string]string{
					constants.ClusterNameKey:  obj.Spec.ClusterName,
					constants.AccountNameKey:  obj.Spec.AccountName,
					constants.WGServerNameKey: obj.Name,
				},
				OwnerReferences: []metav1.OwnerReference{fn.AsOwner(obj, true)},
			},
		}); err != nil {
			return failed(err)
		}
	}

	check.Status = true
	if check != checks[NSReady] {
		checks[NSReady] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureKeysAndSecret(req *rApi.Request[*wgv1.Server]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}
	failed := func(err error) stepResult.Result {
		return req.CheckFailed(WGKeysAndSecretReady, check, err.Error())
	}

	var sec *corev1.Secret
	var err error

	// check keys and if not present, generate
	if err := func() error {
		sec, err = rApi.Get(ctx, r.Client, fn.NN(getNs(obj), "wg-server-keys"), &corev1.Secret{})
		if err != nil {
			if !apiErrors.IsNotFound(err) {
				return err
			}
			// have to create
			pub, priv, err := wgctrl_utils.GenerateWgKeys()
			if err != nil {
				return err
			}

			if err = r.Create(ctx,
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "wg-server-keys",
						Namespace: getNs(obj),
						Labels: map[string]string{
							constants.ClusterNameKey:  obj.Spec.ClusterName,
							constants.AccountNameKey:  obj.Spec.AccountName,
							constants.WGServerNameKey: obj.Name,
						},
						OwnerReferences: []metav1.OwnerReference{
							fn.AsOwner(obj, true),
						},
					},
					Data: map[string][]byte{
						"private-key": priv,
						"public-key":  pub,
					},
				},
			); err != nil {
				return err
			}

			return fmt.Errorf("wg secret created")
		}

		return nil
	}(); err != nil {
		return failed(err)
	}

	if err := func() error {
		pub, priv, err := parseWgSec(sec)
		if err != nil {
			return err
		}

		if obj.Spec.PublicKey == nil || string(pub) != *obj.Spec.PublicKey {
			pubS := string(pub)
			obj.Spec.PublicKey = &pubS
			if err := r.Update(ctx, obj); err != nil {
				return err
			}
		}

		var devSecrets corev1.SecretList
		if err := r.List(
			ctx, &devSecrets,
			&client.ListOptions{
				LabelSelector: apiLabels.SelectorFromValidatedSet(
					map[string]string{
						constants.WGDeviceSeceret: "true",
					},
				),
				Namespace: getNs(obj),
			},
		); err != nil {
			return err
		}

		var data data
		data.ServerIp = "10.13.13.1/32"
		data.ServerPrivateKey = string(priv)

		sort.Slice(
			devSecrets.Items, func(i, j int) bool {
				return devSecrets.Items[i].Name < devSecrets.Items[j].Name
			},
		)

		for _, dev := range devSecrets.Items {
			devIp, devPub, err := parseDevSec(dev)
			if err != nil {
				continue
			}

			data.Peers = append(data.Peers, Peer{
				PublicKey:  string(devPub),
				AllowedIps: string(devIp),
			})
		}

		conf, err := templates.Parse(templates.Wireguard.Config, data)
		if err != nil {
			return err
		}

		oConf, err := rApi.Get(ctx, r.Client, fn.NN(getNs(obj), "wg-server-config"), &corev1.Secret{})
		if err != nil {
			if !apiErrors.IsNotFound(err) {
				return err
			}

			if err := r.Create(ctx, &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "wg-server-config",
					Namespace: getNs(obj),
					Labels: map[string]string{
						constants.ClusterNameKey:  obj.Spec.ClusterName,
						constants.AccountNameKey:  obj.Spec.AccountName,
						constants.WGServerNameKey: obj.Name,
					},
				},
				Data: map[string][]byte{
					"data": conf,
				},
			}); err != nil {
				return err
			}
		}

		if oConf != nil && (string(oConf.Data["data"]) != string(conf)) {

			if err := r.Update(ctx, &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "wg-server-config",
					Namespace: getNs(obj),
					Labels: map[string]string{
						constants.ClusterNameKey:  obj.Spec.ClusterName,
						constants.AccountNameKey:  obj.Spec.AccountName,
						constants.WGServerNameKey: obj.Name,
					},
				},
				Data: map[string][]byte{
					"data": conf,
				},
			}); err != nil {
				return err
			}

			// TODO: (testing is strictly needed) ( please remove this comment if test is done)
			if _, err := http.Post(
				fmt.Sprintf("https://wg-api-service.%s.svc.cluster.local:2998/port", getNs(obj)), "application/json", bytes.NewBuffer(conf),
			); err != nil {
				return err
			}
		}

		return nil
	}(); err != nil {
		return failed(err)
	}

	check.Status = true
	if check != checks[WGKeysAndSecretReady] {
		checks[WGKeysAndSecretReady] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensurDevProxy(req *rApi.Request[*wgv1.Server]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}
	failed := func(err error) stepResult.Result {
		return req.CheckFailed(DeviceProxyReady, check, err.Error())
	}
	name := "device-proxy-config"

	if err := func() error {
		oConf, err := rApi.Get(ctx, r.Client, fn.NN(getNs(obj), name), &corev1.ConfigMap{})
		if err != nil {
			if !apiErrors.IsNotFound(err) {
				return err
			}

			if err := r.Create(ctx, &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: getNs(obj),
					Labels: map[string]string{
						constants.ClusterNameKey:  obj.Spec.ClusterName,
						constants.AccountNameKey:  obj.Spec.AccountName,
						constants.WGServerNameKey: obj.Name,
					},

					OwnerReferences: []metav1.OwnerReference{fn.AsOwner(obj)},
				},
				Data: map[string]string{
					"config.json": `{"services":[]}`,
				},
			}); err != nil {
				return err
			}
			return fmt.Errorf("no device-proxy-config was created, so created now")
		}

		configs := []ConfigService{}
		oConfigs := []ConfigService{}
		configData := map[string]*ConfigService{}

		if oConf != nil {
			oCMap := map[string][]ConfigService{}
			if err := json.Unmarshal([]byte(oConf.Data["config.json"]), &oCMap); err != nil {
				return err
			}

			if oCMap["services"] != nil {
				oConfigs = oCMap["services"]
			}
		}

		for _, cs := range oConfigs {
			k := cs
			configData[cs.Id] = &k
		}

		var devices wgv1.DeviceList
		if err := r.List(ctx, &devices, &client.ListOptions{
			LabelSelector: apiLabels.SelectorFromValidatedSet(
				map[string]string{
					constants.WGServerNameKey: obj.Name,
				},
			),
		}); err != nil {
			if apiErrors.IsNotFound(err) {
				return err
			}
		}

		for _, d := range devices.Items {
			for _, p := range d.Spec.Ports {
				tempPort := getTempPort(configData, fmt.Sprint(d.Name, "-", p.Port), configData)

				dIp, err := wgctrl_utils.GetRemoteDeviceIp(int64(d.Spec.Offset))
				if err != nil {
					return err
				}
				configs = append(configs, ConfigService{
					Id:   fmt.Sprint(d.Name, "-", p.Port),
					Name: string(dIp),
					ServicePort: func() int32 {
						if p.TargetPort != 0 {
							return p.TargetPort
						}
						return p.Port
					}(),
					ProxyPort: tempPort,
				})
			}
		}

		sort.Slice(
			configs, func(i, j int) bool {
				return configs[i].Name < configs[j].Name
			},
		)
		c, err := json.Marshal(
			map[string][]ConfigService{
				"services": configs,
			},
		)
		if err != nil {
			return err
		}

		if len(configs) == 0 {
			return nil
		}

		equal, err := JSONStringsEqual(oConf.Data["config.json"], string(c))
		if err != nil {
			return err
		}

		if !equal {
			sort.Slice(
				configs, func(i, j int) bool {
					return configs[i].Name < configs[j].Name
				},
			)

			configJson, err := json.Marshal(map[string][]ConfigService{
				"services": configs,
			})
			if err != nil {
				return err
			}

			if err := r.Update(ctx, &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: getNs(obj),
					Labels: map[string]string{
						constants.ClusterNameKey: obj.Spec.ClusterName,
						constants.AccountNameKey: obj.Spec.AccountName,
					},
					OwnerReferences: []metav1.OwnerReference{
						fn.AsOwner(obj),
					},
				},
				Data: map[string]string{
					"config.json": string(configJson),
				},
			}); err != nil {
				return err
			}

			if _, err := http.Post(
				fmt.Sprintf("https://wg-api-service.%s.svc.cluster.local:2999/port", getNs(obj)), "application/json", bytes.NewBuffer(configJson),
			); err != nil {
				return err
			}
		}

		return nil
	}(); err != nil {
		return failed(err)
	}

	check.Status = true
	if check != checks[DeviceProxyReady] {
		checks[DeviceProxyReady] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureDeploy(req *rApi.Request[*wgv1.Server]) stepResult.Result {

	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}
	failed := func(err error) stepResult.Result {
		return req.CheckFailed(WGDeployReady, check, err.Error())
	}

	var devices wgv1.DeviceList
	if err := r.List(ctx, &devices); err != nil {
		if !apiErrors.IsNotFound(err) {
			return failed(err)
		}
	}

	// check deployment
	if err := func() error {
		if dep, err := rApi.Get(ctx, r.Client, fn.NN(getNs(obj), "wg-deployment"), &appsv1.Deployment{}); err != nil {
			if !apiErrors.IsNotFound(err) {
				return err
			}

			if len(devices.Items) == 0 {
				return fmt.Errorf("no devices created yet, so no wireguard deployment needed")
			}

			// created or update wg deployment
			if b, err := templates.Parse(templates.Wireguard.Deploy, map[string]any{
				"name": obj.Name, "isMaster": false,
			}); err != nil {
				return err
			} else if _, err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
				return err
			}
		} else if len(devices.Items) == 0 {
			return r.Delete(ctx, dep)
		}

		return nil
	}(); err != nil {
		return failed(err)
	}

	check.Status = true
	if check != checks[WGDeployReady] {
		checks[WGDeployReady] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureCoreDNS(req *rApi.Request[*wgv1.Server]) stepResult.Result {

	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}
	failed := func(err error) stepResult.Result {
		return req.CheckFailed(CorednsDeployReady, check, err.Error())
	}

	if err := func() error {
		if _, err := rApi.Get(ctx, r.Client, fn.NN(getNs(obj), "coredns"), &appsv1.Deployment{}); err != nil {
			if !apiErrors.IsNotFound(err) {
				return err
			}

			configExists := true
			if _, err := rApi.Get(ctx, r.Client, fn.NN(getNs(obj), "coredns"), &corev1.ConfigMap{}); err != nil {
				if !apiErrors.IsNotFound(err) {
					return err
				}
				configExists = false
			}

			if b, err := templates.Parse(templates.Wireguard.CoreDns,
				map[string]any{"name": obj.Name, "configExists": configExists}); err != nil {
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
	if check != checks[CorednsDeployReady] {
		checks[CorednsDeployReady] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) finalize(req *rApi.Request[*wgv1.Server]) stepResult.Result {

	_, obj, _ := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	// TODO: have to write deletion logic
	k := "************** ~~>* deletion of server is not supported yet *<~~ **************"
	fmt.Println(k)
	return req.CheckFailed(ServerDeleted, check, k)
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

	builder := ctrl.NewControllerManagedBy(mgr).For(&wgv1.Server{})
	builder.WithEventFilter(rApi.ReconcileFilter())

	watchList := []client.Object{
		&wgv1.Device{},
		&corev1.Secret{},
		&corev1.ConfigMap{},
		&corev1.Namespace{},
		&appsv1.Deployment{},
	}

	for i := range watchList {
		builder.Watches(
			watchList[i],
			handler.EnqueueRequestsFromMapFunc(
				func(_ctx context.Context, obj client.Object) []reconcile.Request {
					if serverName, ok := obj.GetLabels()[constants.WGServerNameKey]; ok {
						return []reconcile.Request{{NamespacedName: fn.NN("", serverName)}}
					}
					return nil
				}),
		)
	}

	return builder.Complete(r)
}
