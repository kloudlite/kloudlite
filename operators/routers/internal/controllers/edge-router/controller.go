package edgeRouter

import (
	"context"
	acmev1 "github.com/cert-manager/cert-manager/pkg/apis/acme/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
	Finalizing             string = "finalizing"
)

const (
	WildcardCertName      string = "kl-cert-issuer-tls"
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

	req.LogPreReconcile()
	defer req.LogPostReconcile()

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
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

	if step := r.ensureClusterIssuer(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureIngressController(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	req.Object.Status.LastReconcileTime = &metav1.Time{Time: time.Now()}
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) finalize(req *rApi.Request[*crdsv1.EdgeRouter]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	// STEP 1: ensure all ingress nginx controllers are deleted
	nginxes := unstructured.UnstructuredList{
		Object: map[string]any{
			"apiVersion": constants.HelmIngressNginx.APIVersion,
			"kind":       constants.HelmIngressNginx.Kind,
		},
	}
	if err := r.List(ctx, &nginxes, &client.ListOptions{
		Namespace:     obj.Namespace,
		LabelSelector: labels.SelectorFromValidatedSet(obj.GetEnsuredLabels()),
	}); err != nil {
		return req.CheckFailed(Finalizing, check, err.Error()).Err(nil)
	}

	for i := range nginxes.Items {
		if nginxes.Items[i].GetDeletionTimestamp() == nil {
			if err := r.Delete(ctx, &nginxes.Items[i]); err != nil {
				if !apiErrors.IsNotFound(err) {
					return req.CheckFailed(Finalizing, check, err.Error()).Err(nil)
				}
			}
		}
	}

	if len(nginxes.Items) != 0 {
		return req.CheckFailed(Finalizing, check, "waiting for nginx ingress controllers to be deleted")
	}

	// STEP 2: ensure all cluster issuers are deleted
	issuerName := controllers.GetClusterIssuerName(obj.Spec.EdgeName)
	if err := r.Delete(ctx, &certmanagerv1.ClusterIssuer{ObjectMeta: metav1.ObjectMeta{Name: issuerName}}); err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.CheckFailed(Finalizing, check, "waiting for cluster issuer to be deleted")
		}
	}

	// STEP 3: clear all finalizers
	controllerutil.RemoveFinalizer(obj, constants.CommonFinalizer)
	if err := r.Update(ctx, obj); err != nil {
		return req.CheckFailed(Finalizing, check, err.Error())
	}

	return req.Done()
}

func (r *Reconciler) ensureClusterIssuer(req *rApi.Request[*crdsv1.EdgeRouter]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(ClusterIssuerReady)
	defer req.LogPostCheck(ClusterIssuerReady)

	issuerName := controllers.GetClusterIssuerName(obj.Spec.EdgeName)

	// STEP 1: copy dns solvers from default cluster issuer
	defaultIssuer, err := rApi.Get(ctx, r.Client, fn.NN("", r.Env.DefaultClusterIssuerName), &certmanagerv1.ClusterIssuer{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.CheckFailed(ClusterIssuerReady, check, err.Error())
		}
		req.Logger.Infof("default cluster issuer (%s) not found, skipping reading them", r.Env.DefaultClusterIssuerName)
	}

	var acmeDnsSolvers []acmev1.ACMEChallengeSolver

	if defaultIssuer != nil && defaultIssuer.Spec.ACME != nil {
		for _, s := range defaultIssuer.Spec.ACME.Solvers {
			if s.DNS01 != nil {
				acmeDnsSolvers = append(acmeDnsSolvers, s)
			}
		}
	}

	// STEP 2: create new cluster issuer for this edge
	b, err := templates.Parse(
		templates.ClusterIssuer, map[string]any{
			"owner-refs":       []metav1.OwnerReference{fn.AsOwner(obj, true)},
			"namespace":        obj.Namespace,
			"acme-dns-solvers": acmeDnsSolvers,

			"kl-acme-email": r.Env.AcmeEmail,
			"issuer-name":   issuerName,
			"ingress-class": controllers.GetIngressClassName(obj.Spec.EdgeName),
			"tolerations": []corev1.Toleration{
				{
					Key:      constants.RegionKey,
					Operator: "Equal",
					Value:    obj.Spec.EdgeName,
					Effect:   "NoExecute",
				},
			},
			"node-selector": map[string]string{
				constants.RegionKey: obj.Spec.EdgeName,
			},
		},
	)

	if err != nil {
		return req.CheckFailed(ClusterIssuerReady, check, err.Error()).Err(nil)
	}

	if err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
		return req.CheckFailed(ClusterIssuerReady, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != checks[ClusterIssuerReady] {
		checks[ClusterIssuerReady] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureIngressController(req *rApi.Request[*crdsv1.EdgeRouter]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(IngressControllerReady)
	defer req.LogPostCheck(IngressControllerReady)

	b, err := templates.Parse(
		templates.HelmIngressNginx, map[string]any{
			"name":                    obj.Name,
			"namespace":               obj.Namespace,
			"region":                  obj.Spec.EdgeName,
			"owner-refs":              []metav1.OwnerReference{fn.AsOwner(obj, true)},
			"labels":                  obj.Labels,
			"wildcard-cert-name":      WildcardCertName,
			"wildcard-cert-namespace": WildcardCertNamespace,
			"ingress-class-name":      controllers.GetIngressClassName(obj.Spec.EdgeName),
		},
	)
	if err != nil {
		return req.CheckFailed(IngressControllerReady, check, err.Error()).Err(nil)
	}

	if err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
		return req.CheckFailed(IngressControllerReady, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != checks[IngressControllerReady] {
		checks[IngressControllerReady] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

	builder := ctrl.NewControllerManagedBy(mgr).For(&crdsv1.EdgeRouter{})
	builder.WithOptions(controller.Options{
		MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles,
	})
	builder.Owns(fn.NewUnstructured(constants.HelmIngressNginx))
	builder.Owns(&appsv1.DaemonSet{})
	builder.Owns(fn.NewUnstructured(constants.ClusterIssuerType))
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
