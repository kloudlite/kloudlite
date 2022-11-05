package edgeRouter

import (
	"context"
	"strings"
	"time"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"github.com/goombaio/namegenerator"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	"operators.kloudlite.io/lib/constants"
	fn "operators.kloudlite.io/lib/functions"
	"operators.kloudlite.io/lib/kubectl"
	"operators.kloudlite.io/lib/logging"
	rApi "operators.kloudlite.io/lib/operator"
	stepResult "operators.kloudlite.io/lib/operator/step-result"
	"operators.kloudlite.io/lib/templates"
	"operators.kloudlite.io/operators/routers/internal/controllers"
	"operators.kloudlite.io/operators/routers/internal/env"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	DefaultsPatched        string = "defaults-patched"
	ClusterIssuerPatched   string = "cluster-issuer-patched"
	ClusterIssuerReady     string = "cluster-issuer-ready"
	IngressControllerReady string = "ingress-controller-ready"
)

const (
	WildcardCertNamespace string = "kl-init-cert-manager"
)

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=edges,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=edges/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=edges/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, request.NamespacedName, &crdsv1.EdgeRouter{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	req.Logger.Infof("NEW RECONCILATION")
	defer func() {
		req.Logger.Infof("RECONCILATION COMPLETED (isReady=%v)", req.Object.Status.IsReady)
	}()

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.RestartIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureChecks(DefaultsPatched, IngressControllerReady, ClusterIssuerPatched); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconDefaults(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconClusterIssuer(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconIngressController(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	// if step := r.patchClusterIssuer(req); !step.ShouldProceed() {
	// 	return step.ReconcilerResponse()
	// }

	req.Object.Status.IsReady = true
	req.Object.Status.LastReconcileTime = metav1.Time{Time: time.Now()}
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) finalize(req *rApi.Request[*crdsv1.EdgeRouter]) stepResult.Result {
	return req.Finalize()
}

func (r *Reconciler) reconDefaults(req *rApi.Request[*crdsv1.EdgeRouter]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	hasPatched := false

	if obj.Spec.ControllerName == "" {
		hasPatched = true
		seed := time.Now().UTC().UnixNano()
		nameGenerator := namegenerator.NewNameGenerator(seed)
		obj.Spec.ControllerName = "ingress-" + nameGenerator.Generate()
	}

	if hasPatched {
		if err := r.Update(ctx, obj); err != nil {
			return req.CheckFailed(DefaultsPatched, check, err.Error())
		}
	}
	return req.Next()
}

func (r *Reconciler) reconIngressController(req *rApi.Request[*crdsv1.EdgeRouter]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	ingressC, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Spec.ControllerName), fn.NewUnstructured(constants.HelmIngressNginx))
	if err != nil {
		req.Logger.Infof("ingress controller (%s) does not exist, will be creating it", fn.NN(obj.Namespace, obj.Spec.ControllerName).String())
	}

	if ingressC == nil || check.Generation > checks[IngressControllerReady].Generation {
		b, err := templates.Parse(
			templates.HelmIngressNginx, map[string]any{
				"obj":             obj,
				"controller-name": obj.Spec.ControllerName,
				"owner-refs":      []metav1.OwnerReference{fn.AsOwner(obj, true)},
				"labels": map[string]string{
					constants.EdgeNameKey: obj.Name,
				},
				"wildcard-cert-name":      controllers.GetClusterIssuerName(obj.Name),
				"wildcard-cert-namespace": WildcardCertNamespace,
				"ingress-class-name":      controllers.GetIngressClassName(obj.Name),
			},
		)
		if err != nil {
			return req.CheckFailed(IngressControllerReady, check, err.Error())
		}

		if err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
			return req.CheckFailed(IngressControllerReady, check, err.Error())
		}

		checks[IngressControllerReady] = check
		return req.UpdateStatus()
	}

	check.Status = true
	if check != checks[IngressControllerReady] {
		checks[IngressControllerReady] = check
		return req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) reconClusterIssuer(req *rApi.Request[*crdsv1.EdgeRouter]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	clusterIssuer, err := rApi.Get(ctx, r.Client, fn.NN("", controllers.GetClusterIssuerName(obj.Name)), &certmanagerv1.ClusterIssuer{})
	if err != nil {
		req.Logger.Infof("cluster issuer does not exist yet, would be creating now...")
		clusterIssuer = nil
	}

	if clusterIssuer == nil || check.Generation > checks[ClusterIssuerReady].Generation {
		b, err := templates.Parse(
			templates.ClusterIssuer, map[string]any{
				"kl-cloudflare-wildcard-domains": strings.Split(r.Env.CloudflareWildcardDomains, ","),
				"kl-cloudflare-email":            r.Env.CloudflareEmail,
				"kl-cloudflare-secret-name":      r.Env.CloudflareSecretName,

				"kl-acme-email": r.Env.AcmeEmail,
				"issuer-name":   controllers.GetClusterIssuerName(obj.Name),
				"ingress-class": controllers.GetIngressClassName(obj.Name),
				"tolerations": []corev1.Toleration{
					{
						Key:      constants.RegionKey,
						Operator: "Equal",
						Value:    obj.Name,
						Effect:   "NoExecute",
					},
				},
				"node-selector": map[string]string{
					constants.RegionKey: obj.Name,
				},
			},
		)

		if err != nil {
			return req.CheckFailed(ClusterIssuerReady, check, err.Error()).Err(nil)
		}

		if err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
			return req.CheckFailed(ClusterIssuerReady, check, err.Error()).Err(nil)
		}
	}

	check.Status = true
	if check != checks[ClusterIssuerReady] {
		checks[ClusterIssuerReady] = check
		return req.UpdateStatus()
	}
	return req.Next()
}

// func (r *Reconciler) patchClusterIssuer(req *rApi.Request[*crdsv1.EdgeRouter]) stepResult.Result {
// 	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
// 	check := rApi.Check{Generation: obj.Generation}
//
// 	clusterIssuer, err := rApi.Get(ctx, r.Client, fn.NN("", r.Env.ClusterCertIssuer), &certmanagerv1.ClusterIssuer{})
// 	if err != nil {
// 		return req.CheckFailed(ClusterIssuerPatched, check, err.Error()).Err(nil)
// 	}
//
// 	ingressClassName := fmt.Sprintf("ingress-nginx-%s", obj.Name)
//
// 	solverExists := false
//
// 	for _, solver := range clusterIssuer.Spec.ACME.Solvers {
// 		if solver.HTTP01 != nil && solver.HTTP01.Ingress != nil && solver.HTTP01.Ingress.Class != nil {
// 			if *solver.HTTP01.Ingress.Class == ingressClassName {
// 				solverExists = true
// 			}
// 		}
// 	}
//
// 	if !solverExists {
// 		clusterIssuer.Spec.ACME.Solvers = append(
// 			clusterIssuer.Spec.ACME.Solvers, acmev1.ACMEChallengeSolver{
// 				HTTP01: &acmev1.ACMEChallengeSolverHTTP01{
// 					Ingress: &acmev1.ACMEChallengeSolverHTTP01Ingress{
// 						Class: &ingressClassName,
// 						PodTemplate: &acmev1.ACMEChallengeSolverHTTP01IngressPodTemplate{
// 							ACMEChallengeSolverHTTP01IngressPodObjectMeta: acmev1.ACMEChallengeSolverHTTP01IngressPodObjectMeta{
// 								Labels: map[string]string{
// 									"kloudlite.io/ingress-class": ingressClassName,
// 								},
// 							},
// 							Spec: acmev1.ACMEChallengeSolverHTTP01IngressPodSpec{},
// 						},
// 					},
// 				},
// 			},
// 		)
//
// 		if err := r.Update(ctx, clusterIssuer); err != nil {
// 			return req.CheckFailed(ClusterIssuerPatched, check, err.Error())
// 		}
// 	}
//
// 	check.Status = true
// 	if check != checks[ClusterIssuerPatched] {
// 		checks[ClusterIssuerPatched] = check
// 		return req.UpdateStatus()
// 	}
// 	return req.Next()
// }

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

	builder := ctrl.NewControllerManagedBy(mgr).For(&crdsv1.EdgeRouter{})
	builder.Owns(fn.NewUnstructured(constants.HelmIngressNginx))
	builder.Owns(&appsv1.DaemonSet{})
	return builder.Complete(r)
}
