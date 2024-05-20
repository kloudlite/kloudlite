package external_app

import (
	"context"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/app-n-lambda/internal/env"
	"github.com/kloudlite/operator/pkg/constants"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	controllerutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type ExternalAppReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Env        *env.Env
	logger     logging.Logger
	Name       string
	yamlClient kubectl.YAMLClient
}

func (r *ExternalAppReconciler) GetName() string {
	return r.Name
}

const (
	createExternalNameService = "createExternalNameService"
)

var ApplyChecklist = []rApi.CheckMeta{
	{Name: createExternalNameService, Title: "Creates External Name Service"},
}

// +kubebuilder:rbac:groups=crdsv1,resources=external_apps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crdsv1,resources=external_apps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crdsv1,resources=external_apps/finalizers,verbs=update

func (r *ExternalAppReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, request.NamespacedName, &crdsv1.ExternalApp{})
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

	if step := req.EnsureFinalizers(constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.createExternalService(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *ExternalAppReconciler) createExternalService(req *rApi.Request[*crdsv1.ExternalApp]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(createExternalNameService, req)

	svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: obj.Name, Namespace: obj.Namespace}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, svc, func() error {
		svc.Spec.Type = corev1.ServiceTypeExternalName
		switch obj.Spec.RecordType {
		case crdsv1.ExternalAppRecordTypeCNAME:
			{
				svc.Spec.ExternalName = obj.Spec.Record
			}
		case crdsv1.ExternalAppRecordTypeIPAddr:
			{
				svc.Spec.ExternalIPs = []string{obj.Spec.Record}
			}
		}
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	return check.Completed()
}

func (r *ExternalAppReconciler) finalize(req *rApi.Request[*crdsv1.ExternalApp]) stepResult.Result {
	return req.Finalize()
}

func (r *ExternalAppReconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig(), kubectl.YAMLClientOpts{Logger: r.logger})

	builder := ctrl.NewControllerManagedBy(mgr).For(&crdsv1.ExternalApp{})
	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	return builder.Complete(r)
}
