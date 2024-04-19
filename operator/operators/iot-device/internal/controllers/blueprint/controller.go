package blueprint

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/yaml"

	crds1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/iot-device/internal/env"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	apiLabels "k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
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
	AppsReady           = "apps-ready"
	BlueprintFinalizing = "finalizing-blueprint"
)

var (
	blueprintApplyCheckList = []rApi.CheckMeta{
		{Name: AppsReady, Title: "App Reconcilation"},
	}

	blueprintDestroyChecklist = []rApi.CheckMeta{
		{Name: BlueprintFinalizing, Title: "Finalizing Blueprint"},
	}
)

// +kubebuilder:rbac:groups=wireguard.kloudlite.io,resources=devices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=wireguard.kloudlite.io,resources=devices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=wireguard.kloudlite.io,resources=devices/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &crds1.Blueprint{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}

		return ctrl.Result{}, nil
	}

	req.PreReconcile()
	defer req.PostReconcile()

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureCheckList(blueprintApplyCheckList); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconApps(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *Reconciler) reconApps(req *rApi.Request[*crds1.Blueprint]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(AppsReady, req)

	lk := "kloudlite.io/blueprint.name"

	for i, a := range obj.Spec.Apps {
		if a.GetLabels()[lk] == "" {
			obj.Spec.Apps[i].Labels[lk] = obj.Name
		}

		obj.Spec.Apps[i].OwnerReferences = append(a.OwnerReferences, fn.AsOwner(obj, true))
	}

	b, err := yaml.Marshal(obj.Spec.Apps)
	if err != nil {
		return check.Failed(err)
	}

	refs, err := r.yamlClient.ApplyYAML(ctx, b)
	if err != nil {
		return check.Failed(err)
	}

	req.AddToOwnedResources(refs...)

	var apps crds1.AppList
	if err := r.List(ctx, &apps, &client.ListOptions{
		LabelSelector: apiLabels.SelectorFromValidatedSet(map[string]string{lk: obj.Name}),
	}); err != nil {
		return check.Failed(err)
	}

	var as []crds1.AppStatus
	for _, a := range apps.Items {
		as = append(as, crds1.AppStatus{Name: a.Name, Status: a.Status})
	}

	obj.Status.Apps = as

	if err := r.Status().Update(ctx, obj); err != nil {
		return check.Failed(err)
	}

	return check.Completed()
}

func (r *Reconciler) finalize(req *rApi.Request[*crds1.Blueprint]) stepResult.Result {
	if step := req.EnsureCheckList(blueprintDestroyChecklist); !step.ShouldProceed() {
		return step
	}

	if step := req.CleanupOwnedResources(); !step.ShouldProceed() {
		return step
	}

	return req.Finalize()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig(), kubectl.YAMLClientOpts{Logger: r.logger})

	builder := ctrl.NewControllerManagedBy(mgr).For(&crds1.Blueprint{})
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
				func(_ context.Context, obj client.Object) []reconcile.Request {
					if dev, ok := obj.GetLabels()[constants.WGDeviceNameKey]; ok {
						return []reconcile.Request{{NamespacedName: fn.NN(obj.GetNamespace(), dev)}}
					}

					return nil
				}),
		)
	}

	return builder.Complete(r)
}
