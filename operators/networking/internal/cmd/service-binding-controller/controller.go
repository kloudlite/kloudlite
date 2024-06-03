package main

import (
	"context"
	"fmt"
	"net/http"

	networkingv1 "github.com/kloudlite/operator/apis/networking/v1"
	"github.com/kloudlite/operator/operators/networking/internal/cmd/service-binding-controller/env"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Env        *env.Env
	logger     logging.Logger
	Name       string
	yamlClient kubectl.YAMLClient
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	bindService string = "bind-service"
)

func getJobSvcAccountName() string {
	return "job-svc-account"
}

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=lifecycles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=lifecycles/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=lifecycles/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &networkingv1.ServiceBinding{})
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

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	// if step := req.EnsureFinalizers(constants.CommonFinalizer); !step.ShouldProceed() {
	// 	return step.ReconcilerResponse()
	// }

	if step := req.EnsureCheckList([]rApi.CheckMeta{
		{Name: bindService, Title: "Bind Service to Global Network IP"},
	}); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.patchServiceBinding(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *Reconciler) finalize(req *rApi.Request[*networkingv1.ServiceBinding]) stepResult.Result {
	rApi.NewRunningCheck("finalizing", req)
	return req.Finalize()
}

func (r *Reconciler) patchServiceBinding(req *rApi.Request[*networkingv1.ServiceBinding]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(bindService, req)

	svc, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.ServiceRef.Namespace, obj.Spec.ServiceRef.Name), &corev1.Service{})
	if err != nil {
		return check.Failed(err)
	}

	if svc.Spec.Type != corev1.ServiceTypeClusterIP {
		return check.Completed()
	}

	obj.Spec.ServiceIP = &svc.Spec.ClusterIP
	obj.Spec.Ports = svc.Spec.Ports

	if err := r.Update(ctx, obj); err != nil {
		return check.Failed(err)
	}

	r2, err := http.NewRequestWithContext(ctx, http.MethodPut, fmt.Sprintf("%s/service/%s", r.Env.GatewayAdminApiAddr, obj.Name), nil)
	if err != nil {
		return check.Failed(err)
	}

	if _, err := http.DefaultClient.Do(r2); err != nil {
		return check.Failed(err)
	}

	return check.Completed()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig(), kubectl.YAMLClientOpts{Logger: r.logger})

	builder := ctrl.NewControllerManagedBy(mgr).For(&networkingv1.ServiceBinding{})
	builder.Watches(&corev1.Service{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
		if v, ok := obj.GetLabels()["kloudlite.io/servicebinding.enabled"]; v == "true" && ok {
			return []reconcile.Request{{NamespacedName: fn.NN(obj.GetNamespace(), obj.GetName())}}
		}
		return nil
	}))

	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
