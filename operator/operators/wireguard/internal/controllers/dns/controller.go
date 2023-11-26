package dns

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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

ensure configmap present and uptodate for dns rewrite rules
ensure coredns deployment is present
*/

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
	ConfigReady     string = "config-ready-and-synced"
	DeploymentReady string = "deployment-&-svc-ready"
)

// +kubebuilder:rbac:groups=wireguard.kloudlite.io,resources=dns,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=wireguard.kloudlite.io,resources=dns/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=wireguard.kloudlite.io,resources=dns/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &wgv1.Dns{})
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

	if step := req.EnsureChecks(ConfigReady, DeploymentReady); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureDnsConfigReady(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureDnsSvcAndReady(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true

	req.Object.Status.LastReconcileTime = &metav1.Time{Time: time.Now()}
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) ensureDnsSvcAndReady(req *rApi.Request[*wgv1.Dns]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}
	failed := func(err error) stepResult.Result {
		return req.CheckFailed(DeploymentReady, check, err.Error())
	}

	// check Deployment
	if err := func() error {
		if _, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, "coredns"), &appsv1.Deployment{}); err != nil {
			if !apiErrors.IsNotFound(err) {
				return err
			}

			configExists := true
			if _, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, "coredns"), &corev1.ConfigMap{}); err != nil {
				if !apiErrors.IsNotFound(err) {
					return err
				}
				configExists = false
			}

			if b, err := templates.Parse(templates.Wireguard.CoreDns,
				map[string]any{
					"name":         obj.Name,
					"configExists": configExists,
					"namespace":    obj.Namespace,
					"ownerRefs":    []metav1.OwnerReference{fn.AsOwner(obj, true)},

					"tolerations": []corev1.Toleration{
						{Operator: "Exists"},
					},
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

	// check Svc
	if err := func() error {
		dns, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, "kl-coredns"), &corev1.Service{})
		if err != nil {
			if !apiErrors.IsNotFound(err) {
				return err
			}

			if b, err := templates.Parse(templates.Wireguard.CoreDnsSvc,
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

		if dns != nil {
			obj.Spec.DNS = &dns.Spec.ClusterIP
		}

		return nil
	}(); err != nil {
		return failed(err)
	}

	check.Status = true
	if check != checks[DeploymentReady] {
		checks[DeploymentReady] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}
	return req.Next()
}

func (r *Reconciler) ensureDnsConfigReady(req *rApi.Request[*wgv1.Dns]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	failed := func(err error) stepResult.Result {
		return req.CheckFailed(ConfigReady, check, err.Error())
	}

	upsertRewriteRules := func(devices *wgv1.DeviceList) error {
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
				"rewrite name %s.%s %s.%s.svc.%s\n        ",
				dev.Name,
				"kl.local",
				dev.Name,
				obj.Namespace,
				r.Env.ClusterInternalDns,
			)
		}

		b, err := templates.Parse(templates.Wireguard.DnsConfig, map[string]any{
			"devices": d, "namespace": obj.Namespace, "rewrite-rules": rewriteRules, "dns-ip": kubeDns.Spec.ClusterIP,
		})
		if err != nil {
			return err
		}

		if _, err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
			return err
		}

		if _, err := fn.Kubectl("-n", obj.Namespace, "rollout", "restart", "deployment/coredns"); err != nil {
			return err
		}

		obj.Spec.MainDns = &kubeDns.Spec.ClusterIP
		req.UpdateStatus()

		return nil
	}

	if err := func() error {
		dConf, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, "coredns"), &corev1.ConfigMap{})
		if err != nil {
			if !apiErrors.IsNotFound(err) {
				return err
			}

			if err := upsertRewriteRules(&wgv1.DeviceList{}); err != nil {
				return err
			}
		}

		var devices wgv1.DeviceList
		if err := r.List(ctx, &devices, &client.ListOptions{
			Namespace: obj.Namespace,
		}); err != nil {
			return err
		}

		d := make([]string, len(devices.Items))
		for i, dev := range devices.Items {
			d[i] = dev.Name
		}

		sort.Strings(d)

		var d2 []string
		if err := json.Unmarshal([]byte(dConf.Data["devices"]), &d2); err != nil {
			return upsertRewriteRules(&devices)
		}

		sort.Strings(d2)
		if fmt.Sprint(d) != fmt.Sprint(d2) {
			return upsertRewriteRules(&devices)
		}

		return nil
	}(); err != nil {
		return failed(err)
	}

	check.Status = true
	if check != checks[ConfigReady] {
		checks[ConfigReady] = check
		return req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) finalize(req *rApi.Request[*wgv1.Dns]) stepResult.Result {
	return req.Finalize()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

	builder := ctrl.NewControllerManagedBy(mgr).For(&wgv1.Dns{})
	builder.WithEventFilter(rApi.ReconcileFilter())

	watchList := []client.Object{
		&corev1.Service{},
	}

	for i := range watchList {
		builder.Watches(
			&source.Kind{Type: watchList[i]},
			handler.EnqueueRequestsFromMapFunc(
				func(obj client.Object) []reconcile.Request {
					if dnsName, ok := obj.GetLabels()["kloudlite.io/coredns-svc"]; ok && dnsName != "" {
						return []reconcile.Request{{NamespacedName: fn.NN(obj.GetNamespace(), dnsName)}}
					}
					return nil
				}),
		)
	}

	return builder.Complete(r)
}
