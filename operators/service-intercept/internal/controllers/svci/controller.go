package svci

import (
	"context"
	"fmt"

	"k8s.io/client-go/tools/record"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/service-intercept/internal/env"
	"github.com/kloudlite/operator/operators/service-intercept/internal/templates"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Logger     logging.Logger
	Name       string
	Env        *env.Env
	YamlClient kubectl.YAMLClient
	recorder   record.EventRecorder

	svcInterceptTemplate []byte
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	CreatedForLabel string = "kloudlite.io/created-for"

	SvcIReconDone           string = "intercept-performed"
	InterceptClosePerformed string = "cleanup"
	SvcInterceptCreated     string = "svc-intercept-created"
)

var DeleteChecklist = []rApi.CheckMeta{
	{Name: SvcIReconDone, Title: "Intercept close performed"},
}

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.Logger), r.Client, request.NamespacedName, &crdsv1.ServiceIntercept{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	req.PreReconcile()
	defer req.PostReconcile()

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureCheckList([]rApi.CheckMeta{
		{Name: SvcIReconDone, Title: func() string {
			return "Intercept performed"
		}(), Hide: false},
	}); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.RestartIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.checkSvcIntercept(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconSvcIntercept(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *Reconciler) finalize(req *rApi.Request[*crdsv1.ServiceIntercept]) stepResult.Result {
	if step := req.EnsureCheckList(DeleteChecklist); !step.ShouldProceed() {
		return step
	}

	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(InterceptClosePerformed, req)

	svc, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Name), &corev1.Service{})
	if err != nil {
		return check.Failed(err)
	}

	var podList corev1.PodList
	if err := r.List(ctx, &podList, &client.ListOptions{
		LabelSelector: labels.SelectorFromValidatedSet(fn.MapMerge(svc.Spec.Selector, map[string]string{
			CreatedForLabel: "intercept",
		})),
		Namespace: obj.Namespace,
	}); err != nil {
		return check.Failed(err)
	}

	for _, p := range podList.Items {
		if err := r.Delete(ctx, &p); err != nil {
			return check.Failed(err)
		}
	}

	if step := req.CleanupOwnedResourcesV2(check); !step.ShouldProceed() {
		return step
	}

	return req.Finalize()
}

func (r *Reconciler) reconSvcIntercept(req *rApi.Request[*crdsv1.ServiceIntercept]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(SvcIReconDone, req)

	svc, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Name), &corev1.Service{})
	if err != nil {
		return check.Failed(err)
	}

	var podList corev1.PodList
	if err := r.List(ctx, &podList, &client.ListOptions{
		LabelSelector: labels.SelectorFromValidatedSet(svc.Spec.Selector),
		Namespace:     obj.Namespace,
	}); err != nil {
		return check.Failed(err)
	}

	for _, p := range podList.Items {
		if cf := p.Labels[CreatedForLabel]; cf == "intercept" {
			continue
		}

		if err := r.Delete(ctx, &p); err != nil {
			return check.Failed(err)
		}
	}

	return check.Completed()
}

func (r *Reconciler) checkSvcIntercept(req *rApi.Request[*crdsv1.ServiceIntercept]) stepResult.Result {
	ctx, obj := req.Context(), req.Object

	check := rApi.NewRunningCheck(SvcInterceptCreated, req)

	podname := obj.Name + "-intercept"
	podns := obj.Namespace
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: podname, Namespace: podns}}

	// if err := r.Delete(ctx, pod); err != nil {
	// 	if !apiErrors.IsNotFound(err) {
	// 		return check.Failed(err)
	// 	}
	// }
	// return check.Completed()

	svciGenerationLabel := "kloudlite.io/svci-generation"

	if err := r.Get(ctx, client.ObjectKeyFromObject(pod), pod); err != nil {
		if apiErrors.IsNotFound(err) {
			portMappings := make(map[uint16]uint16, len(obj.Spec.PortMappings))
			for _, pm := range obj.Spec.PortMappings {
				portMappings[pm.ContainerPort] = pm.ServicePort
			}

			deviceHost := obj.Spec.ToAddr

			if obj.Spec.ToAddr == "" {
				return check.Failed(fmt.Errorf("no address configured on service intercept, failed to intercept")).NoRequeue()
			}

			b, err := templates.ParseBytes(r.svcInterceptTemplate, map[string]any{
				"name":      podname,
				"namespace": podns,
				"labels": fn.MapMerge(fn.MapFilterWithPrefix(obj.Labels, "kloudlite.io/"),
					map[string]string{
						svciGenerationLabel: fmt.Sprintf("%d", obj.Generation),
						CreatedForLabel:     "intercept",
					},
					obj.Status.Selector,
				),
				"owner-references": []metav1.OwnerReference{fn.AsOwner(obj, true)},
				"device-host":      deviceHost,
				"port-mappings":    portMappings,
			})
			if err != nil {
				return check.Failed(err).NoRequeue()
			}

			if _, err := r.YamlClient.ApplyYAML(ctx, b); err != nil {
				return check.Failed(err).NoRequeue()
			}

			return check.StillRunning(fmt.Errorf("waiting for intercept pod to start"))
		}
	}

	if pod.Labels[svciGenerationLabel] != fmt.Sprintf("%d", obj.Generation) {
		if err := r.Delete(ctx, pod); err != nil {
			return check.Failed(err)
		}
		return check.StillRunning(fmt.Errorf("waiting for previous generation pod to be deleted"))
	}

	if pod.Status.Phase != corev1.PodRunning {
		return check.StillRunning(fmt.Errorf("waiting for pod to start running"))
	}

	return check.Completed()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.Logger = logger.WithName(r.Name)
	r.YamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig(), kubectl.YAMLClientOpts{Logger: r.Logger})
	r.recorder = mgr.GetEventRecorderFor(r.GetName())

	var err error
	r.svcInterceptTemplate, err = templates.Read(templates.SvcIntercept)
	if err != nil {
		return err
	}

	builder := ctrl.NewControllerManagedBy(mgr).For(&crdsv1.ServiceIntercept{})

	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
