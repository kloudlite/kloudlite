package router

import (
	"context"
	"encoding/json"
	"fmt"
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/routers/internal/controllers"
	"github.com/kloudlite/operator/operators/routers/internal/env"
	"github.com/kloudlite/operator/pkg/constants"
	"github.com/kloudlite/operator/pkg/errors"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	"github.com/kloudlite/operator/pkg/templates"
	"golang.org/x/crypto/bcrypt"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strings"
	"time"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	logger     logging.Logger
	Name       string
	Env        *env.Env
	yamlClient *kubectl.YAMLClient
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	IngressReady    string = "ingress-ready"
	BasicAuthReady  string = "basic-auth-ready"
	DefaultsPatched string = "patch-defaults"

	Finalizing string = "finalizing"
)

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, request.NamespacedName, &crdsv1.Router{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !req.ShouldReconcile() {
		return ctrl.Result{}, nil
	}

	req.LogPreReconcile()
	defer req.LogPostReconcile()

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureChecks(IngressReady); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.patchDefaults(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconBasicAuth(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureIngresses(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	req.Object.Status.LastReconcileTime = &metav1.Time{Time: time.Now()}
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) patchDefaults(req *rApi.Request[*crdsv1.Router]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(DefaultsPatched)
	defer req.LogPostCheck(DefaultsPatched)

	hasUpdated := false

	if obj.Spec.BasicAuth.Enabled && obj.Spec.BasicAuth.SecretName == "" {
		hasUpdated = true
		obj.Spec.BasicAuth.SecretName = obj.Name + "-basic-auth"
	}

	if hasUpdated {
		if err := r.Update(ctx, obj); err != nil {
			return req.CheckFailed(DefaultsPatched, check, err.Error()).Err(nil)
		}
		return req.Done().RequeueAfter(1 * time.Second)
	}

	check.Status = true
	if check != checks[DefaultsPatched] {
		checks[DefaultsPatched] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) finalize(req *rApi.Request[*crdsv1.Router]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	if controllerutil.RemoveFinalizer(obj, constants.ForegroundFinalizer) {
		if err := r.Update(ctx, obj); err != nil {
			return req.CheckFailed(Finalizing, check, err.Error()).Err(nil)
		}
		return req.Done().RequeueAfter(1 * time.Second)
	}

	var ingList networkingv1.IngressList
	if err := r.List(ctx, &ingList, &client.ListOptions{
		LabelSelector: labels.SelectorFromValidatedSet(obj.GetEnsuredLabels()),
		Namespace:     obj.Namespace,
	}); err != nil {
		return req.CheckFailed(Finalizing, check, err.Error()).Err(nil)
	}

	for i := range ingList.Items {
		if err := r.Delete(ctx, &ingList.Items[i]); err != nil {
			return req.CheckFailed(Finalizing, check, err.Error()).Err(nil)
		}
	}

	if len(ingList.Items) != 0 {
		return req.CheckFailed(Finalizing, check, "waiting for k8s ingress resources to be deleted")
	}

	if controllerutil.RemoveFinalizer(obj, constants.CommonFinalizer) {
		if err := r.Update(ctx, obj); err != nil {
			return req.CheckFailed(Finalizing, check, err.Error()).Err(nil)
		}
	}

	check.Status = true
	if check != checks[Finalizing] {
		checks[Finalizing] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func isBlueprint(obj client.Object) bool {
	return strings.HasSuffix(obj.GetNamespace(), "-blueprint")
}

func (r *Reconciler) reconBasicAuth(req *rApi.Request[*crdsv1.Router]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(BasicAuthReady)
	defer req.LogPostCheck(BasicAuthReady)

	if obj.Spec.BasicAuth.Enabled {

		if len(obj.Spec.BasicAuth.Username) == 0 {
			return req.CheckFailed(BasicAuthReady, check, fmt.Sprintf(".spec.basicAuth.username must be defined when .spec.basicAuth.enabled is set to true")).Err(nil)
		}

		basicAuthScrt := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.BasicAuth.SecretName, Namespace: obj.Namespace}, Type: "Opaque"}
		if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, basicAuthScrt, func() error {
			basicAuthScrt.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
			if _, ok := basicAuthScrt.Data["password"]; ok {
				return nil
			}

			password := fn.CleanerNanoid(48)
			ePass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				return err
			}
			basicAuthScrt.Data = map[string][]byte{
				"auth":     []byte(fmt.Sprintf("%s:%s", obj.Spec.BasicAuth.Username, ePass)),
				"username": []byte(obj.Spec.BasicAuth.Username),
				"password": []byte(password),
			}
			return nil
		}); err != nil {
			return req.CheckFailed(BasicAuthReady, check, err.Error()).Err(nil)
		}
	}

	check.Status = true
	if check != checks[BasicAuthReady] {
		checks[BasicAuthReady] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) parseAndExtractDomains(req *rApi.Request[*crdsv1.Router], wcDomains []string, nonWcDomains []string) error {
	ctx, obj := req.Context(), req.Object

	issuerName := controllers.GetClusterIssuerName(obj.Spec.Region)
	wcdMap := make(map[string]bool, cap(wcDomains))

	if obj.Spec.Https.Enabled {
		cIssuer, err := rApi.Get(ctx, r.Client, fn.NN("", issuerName), fn.NewUnstructured(constants.ClusterIssuerType))
		if err != nil {
			return err
		}

		var clusterIssuer certmanagerv1.ClusterIssuer
		b, err := json.Marshal(cIssuer.Object)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(b, &clusterIssuer); err != nil {
			return err
		}

		for _, solver := range clusterIssuer.Spec.ACME.Solvers {
			if solver.DNS01 != nil {
				for _, dnsName := range solver.Selector.DNSNames {
					if strings.HasPrefix(dnsName, "*.") {
						wcdMap[dnsName[2:]] = true
					}
					wcdMap[dnsName] = true
					wcDomains = append(wcDomains, dnsName)
				}
			}
		}
	}

	for _, domain := range obj.Spec.Domains {
		if _, ok := wcdMap[domain]; ok {
			continue
		}
		sp := strings.SplitN(domain, ".", 2)
		if len(sp) < 2 {
			continue
		}
		if _, ok := wcdMap[sp[1]]; ok {
			continue
		}
		nonWcDomains = append(nonWcDomains, domain)
	}

	return nil
}

func (r *Reconciler) ensureIngresses(req *rApi.Request[*crdsv1.Router]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(IngressReady)
	defer req.LogPostCheck(IngressReady)

	if len(obj.Spec.Routes) == 0 {
		return req.CheckFailed(IngressReady, check, "no routes specified in ingress resource").Err(nil)
	}

	wcDomains := make([]string, 0, 2)
	nonWcDomains := make([]string, 0, 2)

	if err := r.parseAndExtractDomains(req, wcDomains, nonWcDomains); err != nil {
		return req.CheckFailed(IngressReady, check, err.Error())
	}

	//issuerName := controllers.GetClusterIssuerName(obj.Spec.Region)

	lambdaGroups := map[string][]crdsv1.Route{}
	var appRoutes []crdsv1.Route

	for _, route := range obj.Spec.Routes {
		if route.Lambda != "" {
			if _, ok := lambdaGroups[route.Lambda]; !ok {
				lambdaGroups[route.Lambda] = []crdsv1.Route{}
			}
			lambdaGroups[route.Lambda] = append(lambdaGroups[route.Lambda], route)
		}

		if route.App != "" {
			if isBlueprint(obj) {
				route.App = "env-route-switcher"
				appRoutes = append(appRoutes, route)
				continue
			}
			appRoutes = append(appRoutes, route)
		}

	}

	var kubeYamls [][]byte

	for lName, lRoutes := range lambdaGroups {
		ingName := fmt.Sprintf("r-%s-lambda-%s", obj.Name, lName)

		vals := map[string]any{
			"name":       ingName,
			"namespace":  obj.Namespace,
			"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},
			"labels": map[string]string{
				constants.RouterNameKey: obj.Name,
			},

			"domains":          nonWcDomains,
			"wildcard-domains": wcDomains,

			"router-ref":       obj,
			"routes":           lRoutes,
			"virtual-hostname": fmt.Sprintf("%s.%s", lName, obj.Namespace),

			"ingress-class":      controllers.GetIngressClassName(obj.Spec.Region),
			"cluster-issuer":     controllers.GetClusterIssuerName(obj.Spec.Region),
			"cert-ingress-class": controllers.GetIngressClassName(obj.Spec.Region),
			"is-blueprint":       isBlueprint(obj),
		}

		b, err := templates.Parse(templates.CoreV1.Ingress, vals)
		if err != nil {
			return req.FailWithOpError(
				errors.NewEf(err, "could not parse (template=%s)", templates.Ingress),
			).Err(nil)
		}
		kubeYamls = append(kubeYamls, b)
	}

	if len(appRoutes) > 0 {
		b, err := templates.Parse(
			templates.CoreV1.Ingress, map[string]any{
				"name":             obj.Name,
				"namespace":        obj.Namespace,
				"owner-refs":       []metav1.OwnerReference{fn.AsOwner(obj, true)},
				"domains":          nonWcDomains,
				"wildcard-domains": wcDomains,
				"labels": map[string]any{
					constants.RouterNameKey: obj.Name,
				},
				"router-ref": obj,
				"routes":     appRoutes,

				"ingress-class": func() string {
					if obj.Spec.Region == "" {
						return "ingress-nginx"
					}
					return controllers.GetIngressClassName(obj.Spec.Region)
				}(),
				"is-blueprint":       isBlueprint(obj),
				"cluster-issuer":     controllers.GetClusterIssuerName(obj.Spec.Region),
				"cert-ingress-class": controllers.GetIngressClassName(obj.Spec.Region),
			},
		)
		if err != nil {
			return req.CheckFailed(IngressReady, check, err.Error()).Err(nil)
		}

		kubeYamls = append(kubeYamls, b)
	}

	if err := r.yamlClient.ApplyYAML(ctx, kubeYamls...); err != nil {
		return req.CheckFailed(IngressReady, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != checks[IngressReady] {
		checks[IngressReady] = check
		req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

	builder := ctrl.NewControllerManagedBy(mgr).For(&crdsv1.Router{})
	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.Owns(&networkingv1.Ingress{})
	builder.WithEventFilter(rApi.ReconcileFilter())

	return builder.Complete(r)
}
