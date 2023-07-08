package router

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"golang.org/x/crypto/bcrypt"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/routers/internal/controllers"
	"github.com/kloudlite/operator/operators/routers/internal/env"
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

func getIngressClassName(obj *crdsv1.Router) string {
	if obj.Spec.IngressClass != "" {
		return obj.Spec.IngressClass
	}
	return controllers.GetIngressClassName(obj.Spec.Region)
}

func getClusterIssuer(obj *crdsv1.Router) string {
	if obj.Spec.Https.ClusterIssuer != "" {
		return obj.Spec.Https.ClusterIssuer
	}
	return controllers.GetClusterIssuerName(obj.Spec.Region)
}

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &crdsv1.Router{})
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
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, nil
}

func (r *Reconciler) patchDefaults(req *rApi.Request[*crdsv1.Router]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(DefaultsPatched)
	defer req.LogPostCheck(DefaultsPatched)

	hasUpdated := false

	if obj.Spec.BasicAuth != nil && obj.Spec.BasicAuth.Enabled && obj.Spec.BasicAuth.SecretName == "" {
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
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
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
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}
	return req.Next()
}

func (r *Reconciler) isInProjectNamespace(ctx context.Context, obj client.Object) bool {
	n, err := rApi.Get(context.TODO(), r.Client, fn.NN("", obj.GetNamespace()), &corev1.Namespace{})
	if err != nil {
		return false
	}

	if _, ok := n.Labels[constants.EnvNameKey]; ok {
		return false
	}

	if _, ok := n.Labels[constants.ProjectNameKey]; ok {
		return true
	}
	return false
}

func (r *Reconciler) reconBasicAuth(req *rApi.Request[*crdsv1.Router]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(BasicAuthReady)
	defer req.LogPostCheck(BasicAuthReady)

	if obj.Spec.BasicAuth != nil && obj.Spec.BasicAuth.Enabled {
		if len(obj.Spec.BasicAuth.Username) == 0 {
			return req.CheckFailed(BasicAuthReady, check, ".spec.basicAuth.username must be defined when .spec.basicAuth.enabled is set to true").Err(nil)
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
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}
	return req.Next()
}

// func (r *Reconciler) parseAndExtractDomains(req *rApi.Request[*crdsv1.Router], wcDomains []string, nonWcDomains []string) error {
func (r *Reconciler) parseAndExtractDomains(req *rApi.Request[*crdsv1.Router]) ([]string, []string, error) {
	ctx, obj := req.Context(), req.Object

	var wcDomains, nonWcDomains []string

	issuerName := getClusterIssuer(obj)
	wcdMap := make(map[string]bool, cap(wcDomains))

	if obj.Spec.Https.Enabled {
		cIssuer, err := rApi.Get(ctx, r.Client, fn.NN("", issuerName), fn.NewUnstructured(constants.ClusterIssuerType))
		if err != nil {
			if !apiErrors.IsNotFound(err) {
				return nil, nil, err
			}
		}

		if cIssuer != nil {
			var clusterIssuer certmanagerv1.ClusterIssuer
			b, err := json.Marshal(cIssuer.Object)
			if err != nil {
				return nil, nil, err
			}
			if err := json.Unmarshal(b, &clusterIssuer); err != nil {
				return nil, nil, err
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

	return wcDomains, nonWcDomains, nil
}

func (r *Reconciler) ensureIngresses(req *rApi.Request[*crdsv1.Router]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(IngressReady)
	defer req.LogPostCheck(IngressReady)

	if len(obj.Spec.Routes) == 0 {
		return req.CheckFailed(IngressReady, check, "no routes specified in ingress resource").Err(nil)
	}

	wcDomains, nonWcDomains, err := r.parseAndExtractDomains(req)
	if err != nil {
		return req.CheckFailed(IngressReady, check, err.Error())
	}

	annotations := make(map[string]string, 5)

	annotations["nginx.ingress.kubernetes.io/preserve-trailing-slash"] = "true"

	if obj.Spec.MaxBodySizeInMB != nil {
		annotations["nginx.ingress.kubernetes.io/proxy-body-size"] = fmt.Sprintf("%vm", *obj.Spec.MaxBodySizeInMB)
	}

	if obj.Spec.Https != nil && obj.Spec.Https.Enabled {
		annotations["cert-manager.io/cluster-issuer"] = getClusterIssuer(obj)
		annotations["nginx.kubernetes.io/ssl-redirect"] = "true"
		annotations["nginx.ingress.kubernetes.io/force-ssl-redirect"] = fmt.Sprintf("%v", obj.Spec.Https.ForceRedirect)
	}

	if obj.Spec.RateLimit != nil && obj.Spec.RateLimit.Enabled {
		if obj.Spec.RateLimit.Rps > 0 {
			annotations["nginx.ingress.kubernetes.io/limit-rps"] = fmt.Sprintf("%v", obj.Spec.RateLimit.Rps)
		}
		if obj.Spec.RateLimit.Rpm > 0 {
			annotations["nginx.ingress.kubernetes.io/limit-rpm"] = fmt.Sprintf("%v", obj.Spec.RateLimit.Rpm)
		}
		if obj.Spec.RateLimit.Connections > 0 {
			annotations["nginx.ingress.kubernetes.io/limit-connections"] = fmt.Sprintf("%v", obj.Spec.RateLimit.Connections)
		}
	}

	if obj.Spec.Cors != nil && obj.Spec.Cors.Enabled {
		annotations["nginx.ingress.kubernetes.io/enable-cors"] = "true"
		annotations["nginx.ingress.kubernetes.io/cors-allow-origin"] = strings.Join(obj.Spec.Cors.Origins, ",")
		annotations["nginx.ingress.kubernetes.io/cors-allow-credentials"] = fmt.Sprintf("%v", obj.Spec.Cors.AllowCredentials)
	}

	if obj.Spec.BackendProtocol != nil {
		annotations["nginx.ingress.kubernetes.io/backend-protocol"] = *obj.Spec.BackendProtocol
	}

	if obj.Spec.BasicAuth != nil && obj.Spec.BasicAuth.Enabled {
		annotations["nginx.ingress.kubernetes.io/auth-type"] = "basic"
		annotations["nginx.ingress.kubernetes.io/auth-secret"] = obj.Spec.BasicAuth.SecretName
		annotations["nginx.ingress.kubernetes.io/auth-realm"] = "route is protected by basic auth"
	}

	annotations["nginx.ingress.kubernetes.io/rewrite-target"] = "/$1"

	// issuerName := controllers.GetClusterIssuerName(obj.Spec.Region)

	lambdaGroups := map[string][]crdsv1.Route{}
	var appRoutes []crdsv1.Route

	for _, route := range obj.Spec.Routes {
		if route.Lambda != "" {
			if _, ok := lambdaGroups[route.Lambda]; !ok {
				lambdaGroups[route.Lambda] = []crdsv1.Route{}
			}
			lambdaGroups[route.Lambda] = append(lambdaGroups[route.Lambda], route)
			annotations["nginx.ingress.kubernetes.io/upstream-vhost"] = fmt.Sprintf("%s.%s", route.Lambda, obj.Namespace)
		}

		if route.App != "" {
			if r.isInProjectNamespace(ctx, obj) {
				route.App = r.Env.KloudliteEnvRouteSwitcher
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
			"name":        ingName,
			"namespace":   obj.Namespace,
			"owner-refs":  []metav1.OwnerReference{fn.AsOwner(obj, true)},
			"labels":      obj.GetLabels(),
			"annotations": annotations,

			"domains":          nonWcDomains,
			"wildcard-domains": wcDomains,

			"router-ref":       obj,
			"routes":           lRoutes,
			"virtual-hostname": fmt.Sprintf("%s.%s", lName, obj.Namespace),

			"is-in-project-namespace": r.isInProjectNamespace(ctx, obj),
			"ingress-class":           getIngressClassName(obj),
			"cluster-issuer":          getClusterIssuer(obj),
		}

		b, err := templates.Parse(templates.CoreV1.Ingress, vals)
		if err != nil {
			return req.CheckFailed(IngressReady, check, "cloud not parse ingress template").Err(nil)
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

				"labels":      obj.GetLabels(),
				"annotations": annotations,

				"router-ref": obj,
				"routes":     appRoutes,

				"is-in-project-namespace": r.isInProjectNamespace(ctx, obj),
				"ingress-class":           getIngressClassName(obj),
				"cluster-issuer":          getClusterIssuer(obj),
			},
		)
		if err != nil {
			return req.CheckFailed(IngressReady, check, err.Error()).Err(nil)
		}

		kubeYamls = append(kubeYamls, b)
	}

	rr, err := r.yamlClient.ApplyYAML(ctx, kubeYamls...)
	if err != nil {
		return req.CheckFailed(IngressReady, check, err.Error()).Err(nil)
	}

	req.AddToOwnedResources(rr...)

	for i := range rr {
		ing := &networkingv1.Ingress{ObjectMeta: metav1.ObjectMeta{Name: rr[i].Name, Namespace: rr[i].Namespace}}
		if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, ing, func() error {
			matches := true

			for k := range ing.GetAnnotations() {
				if _, ok := annotations[k]; !ok && !strings.HasPrefix(k, "kloudlite.io") {
					matches = false
					break
				}
			}

			if !matches {
				ing.SetAnnotations(annotations)
			}
			return nil
		}); err != nil {
			return req.CheckFailed(IngressReady, check, err.Error()).Err(nil)
		}
	}

	check.Status = true
	if check != checks[IngressReady] {
		checks[IngressReady] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

	builder := ctrl.NewControllerManagedBy(mgr).For(&crdsv1.Router{})
	builder.Owns(&networkingv1.Ingress{})
	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
