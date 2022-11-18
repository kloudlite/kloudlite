package edgeWorker

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	csiv1 "operators.kloudlite.io/apis/csi/v1"
	extensionsv1 "operators.kloudlite.io/apis/extensions/v1"
	"operators.kloudlite.io/operators/extensions/internal/env"
	"operators.kloudlite.io/pkg/constants"
	"operators.kloudlite.io/pkg/errors"
	fn "operators.kloudlite.io/pkg/functions"
	"operators.kloudlite.io/pkg/harbor"
	"operators.kloudlite.io/pkg/kubectl"
	"operators.kloudlite.io/pkg/logging"
	rApi "operators.kloudlite.io/pkg/operator"
	stepResult "operators.kloudlite.io/pkg/operator/step-result"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	harborCli  *harbor.Client
	logger     logging.Logger
	Name       string
	yamlClient *kubectl.YAMLClient
	Env        *env.Env
}

func getEdgeNamespace(edgeName string) string {
	return `kl-edge-` + edgeName
}

func getProviderNamespace(providerName string) string {
	return `kl-` + providerName
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	ProviderNSReady string = "provider-namespace-ready"
	EdgeNSReady     string = "edge-namespace-ready"
	EdgeRouterReady string = "edge-router-ready"
	CSIDriversReady string = "csi-drivers-ready"
)

const SSLSecretName = "kl-cert-issuer-tls"
const SSLSecretNamespace = "kl-init-cert-manager"

// +kubebuilder:rbac:groups=extensions.kloudlite.io,resources=edgeworkers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=extensions.kloudlite.io,resources=edgeworkers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=extensions.kloudlite.io,resources=edgeworkers/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, request.NamespacedName, &extensionsv1.EdgeWorker{})
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
		req.Logger.Infof("RECONCILATION COMPLETE (isReady=%v)", req.Object.Status.IsReady)
	}()

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	// TODO: initialize all checks here
	if step := req.EnsureChecks(EdgeNSReady); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureProviderNS(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureEdgeNS(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureCSIDrivers(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureEdgeRouters(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) finalize(req *rApi.Request[*extensionsv1.EdgeWorker]) stepResult.Result {
	return req.Finalize()
}

func (r *Reconciler) ensureProviderNS(req *rApi.Request[*extensionsv1.EdgeWorker]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	namespaceName := "kl-" + obj.Spec.Creds.SecretName

	providerNs, err := rApi.Get(ctx, r.Client, fn.NN("", namespaceName), &corev1.Namespace{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.CheckFailed(ProviderNSReady, check, err.Error()).Err(nil)
		}
		req.Logger.Infof("provider namespace (%s) does not exist yet, will be creating it now...", namespaceName)
		providerNs = nil
	}

	if providerNs == nil {
		providerScrt, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.Creds.Namespace, obj.Spec.Creds.SecretName), &corev1.Secret{})
		if err != nil {
			return req.CheckFailed(ProviderNSReady, check, err.Error())
		}

		if err := r.Create(
			ctx, &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:            namespaceName,
					OwnerReferences: []metav1.OwnerReference{fn.AsOwner(providerScrt, true)},
				},
			},
		); err != nil {
			return req.CheckFailed(ProviderNSReady, check, err.Error()).Err(nil)
		}
		return req.Done().RequeueAfter(1 * time.Second)
	}

	check.Status = true
	if check != checks[ProviderNSReady] {
		checks[ProviderNSReady] = check
		return req.UpdateStatus()
	}

	rApi.SetLocal(req, "provider-namespace", namespaceName)
	return req.Next()
}

func (r *Reconciler) ensureEdgeNS(req *rApi.Request[*extensionsv1.EdgeWorker]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	edgeNamespaceName := `kl-edge-` + obj.Name

	edgeNs, err := rApi.Get(ctx, r.Client, fn.NN("", edgeNamespaceName), &corev1.Namespace{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.CheckFailed(EdgeNSReady, check, err.Error())
		}
		req.Logger.Infof("edge namespace (%s) does not exist, yet, would be creating now")
		edgeNs = nil
	}

	if edgeNs == nil {
		if err := r.Create(
			ctx, &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:            edgeNamespaceName,
					Labels:          obj.GetLabels(),
					OwnerReferences: obj.GetOwnerReferences(),
				},
			},
		); err != nil {
			return req.CheckFailed(EdgeNSReady, check, err.Error()).Err(nil)
		}
	}

	check.Status = true
	if check != checks[EdgeNSReady] {
		checks[EdgeNSReady] = check
		return req.UpdateStatus()
	}
	rApi.SetLocal(req, "edge-namespace", edgeNamespaceName)
	return req.Next()
}

func (r *Reconciler) ensureCSIDrivers(req *rApi.Request[*extensionsv1.EdgeWorker]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	csiDriver, err := rApi.Get(ctx, r.Client, fn.NN("", obj.Spec.Creds.SecretName), &csiv1.Driver{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.CheckFailed(CSIDriversReady, check, err.Error()).Err(nil)
		}
		req.Logger.Infof("csi driver (%s) does not exist yet, will be creating it ...")
		csiDriver = nil
	}

	if csiDriver == nil {
		if err := r.Create(
			ctx, &csiv1.Driver{
				ObjectMeta: metav1.ObjectMeta{
					Name:            obj.Spec.Creds.SecretName,
					OwnerReferences: []metav1.OwnerReference{fn.AsOwner(obj, true)},
				},
				Spec: csiv1.DriverSpec{
					Provider:  obj.Spec.Provider,
					SecretRef: obj.Spec.Creds.SecretName,
				},
			},
		); err != nil {
			return req.CheckFailed(CSIDriversReady, check, err.Error()).Err(nil)
		}
		return req.Done().RequeueAfter(2 * time.Second)
	}

	if csiDriver.Status.IsReady != true {
		msg := csiDriver.Status.Message.ToString()
		if len(msg) == 0 {
			msg = "waiting for csi driver controller to setup csi driver"
		}
		return req.CheckFailed(CSIDriversReady, check, msg)
	}

	check.Status = true
	if check != checks[CSIDriversReady] {
		checks[CSIDriversReady] = check
		return req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureEdgeRouters(req *rApi.Request[*extensionsv1.EdgeWorker]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	edgeNamespace, ok := rApi.GetLocal[string](req, "edge-namespace")
	if !ok {
		return req.CheckFailed(EdgeRouterReady, check, errors.NotInLocals("edge-namespace").Error()).Err(nil)
	}

	edgeRouter, err := rApi.Get(ctx, r.Client, fn.NN(edgeNamespace, "ingress-nginx"), &crdsv1.EdgeRouter{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.CheckFailed(EdgeNSReady, check, err.Error()).Err(nil)
		}
		req.Logger.Infof("edge router (%s) does not exist, will be creating now...", obj.Name)
		edgeRouter = nil
	}

	if edgeRouter == nil {
		if err := r.Create(
			ctx, &crdsv1.EdgeRouter{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "ingress-nginx",
					Namespace:       edgeNamespace,
					OwnerReferences: []metav1.OwnerReference{fn.AsOwner(obj, true)},
				},
				Spec: crdsv1.EdgeRouterSpec{
					EdgeName:   obj.Name,
					Region:     obj.Spec.Region,
					AccountRef: obj.Spec.AccountId,
					DefaultSSLCert: crdsv1.SSLCertRef{
						SecretName: SSLSecretName,
						Namespace:  SSLSecretNamespace,
					},
				},
			},
		); err != nil {
			return stepResult.New().Err(err)
		}
	}

	check.Status = true
	if check != checks[EdgeRouterReady] {
		checks[EdgeRouterReady] = check
		return req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

	builder := ctrl.NewControllerManagedBy(mgr).For(&extensionsv1.EdgeWorker{})
	builder.Owns(&csiv1.Driver{})
	return builder.Complete(r)
}
