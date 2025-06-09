package workspace

import (
	"context"
	"fmt"

	"k8s.io/client-go/tools/record"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/workspace/internal/env"
	"github.com/kloudlite/operator/operators/workspace/internal/templates"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/toolkit/functions"
	"github.com/kloudlite/operator/toolkit/kubectl"
	rApi "github.com/kloudlite/operator/toolkit/reconciler"
	stepResult "github.com/kloudlite/operator/toolkit/reconciler/step-result"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

var PortConfig = templates.PortConfig{
	SSHPort:        22,
	TTYDPort:       56789,
	NotebookPort:   56790,
	CodeServerPort: 56791,
}

const (
	IngressClassName = "nginx"
)

type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Env    *env.Env

	YAMLClient kubectl.YAMLClient
	recorder   record.EventRecorder

	workspaceDeploymentTemplate []byte
	jumpServerTemplate          []byte
	templateWebhook             []byte
}

func (r *Reconciler) GetName() string {
	return "workspace"
}

const (
	CreateDeployment string = "create-deployment"
	// CreateService    string = "create-service"
)

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(ctx, r.Client, request.NamespacedName, &crdsv1.Workspace{})
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

	if step := req.EnsureCheckList([]rApi.CheckMeta{{Name: CreateDeployment}}); !step.ShouldProceed() {
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

	// if step := r.createInterceptableService(req); !step.ShouldProceed() {
	// 	return step.ReconcilerResponse()
	// }

	if step := r.createDeployment(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

// func (r *Reconciler) createInterceptableService(req *rApi.Request[*crdsv1.Workspace]) stepResult.Result {
// 	ctx, obj := req.Context(), req.Object
// 	check := rApi.NewRunningCheck(CreateService, req)

// 	svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: obj.Name, Namespace: obj.Namespace}}
// 	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, svc, func() error {
// 		svc.Spec.Ports = []corev1.ServicePort{
// 			{
// 				Name:       fmt.Sprintf("port-%d", 3000),
// 				Protocol:   "TCP",
// 				Port:       3000,
// 				TargetPort: intstr.FromInt(3000),
// 			},
// 		}
// 		return nil
// 	}); err != nil {
// 		return check.Failed(err)
// 	}

// 	// function-body
// 	return check.Completed()
// }

func (r *Reconciler) createDeployment(req *rApi.Request[*crdsv1.Workspace]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(CreateDeployment, req)

	b, err := templates.ParseBytes(r.workspaceDeploymentTemplate, templates.WorkspaceTemplateArgs{
		Metadata: metav1.ObjectMeta{
			Name:            obj.Name,
			Namespace:       obj.Namespace,
			OwnerReferences: []metav1.OwnerReference{fn.AsOwner(obj, true)},
		},
		KloudliteDomain:    "test.khost.dev",
		WorkMachineName:    obj.Spec.WorkMachine,
		ServiceAccountName: obj.Spec.ServiceAccountName,
		ImageInitContainer: r.Env.WorkspaceImageInitContainer,
		ImageSSH:           r.Env.WorkspaceImageSSH,
		IsOn:               obj.Spec.State == crdsv1.WorkspaceStateOn,
		EnableTTYD:         obj.Spec.EnableTTYD,
		ImageTTYD:          r.Env.WorkspaceImageTTYD,

		EnableJupyterNotebook: obj.Spec.EnableJupyterNotebook,
		ImageJupyterNotebook:  r.Env.WorkspaceImageJupyterNotebook,

		EnableCodeServer: obj.Spec.EnableCodeServer,
		ImageCodeServer:  r.Env.WorkspaceImageCodeServer,

		EnableVSCodeServer: obj.Spec.EnableVSCodeServer,
		ImageVscodeServer:  r.Env.WorkspaceImageVscodeServer,
		PortConfig:         PortConfig,

		ImagePullPolicy:     obj.Spec.ImagePullPolicy,
		KloudliteDeviceFQDN: fmt.Sprintf("%s-headless.%s.svc.cluster.local", obj.Name, obj.Namespace),
	})
	if err != nil {
		return check.Failed(err)
	}

	fmt.Println(string(b))

	rr, err := r.YAMLClient.ApplyYAML(ctx, b)
	if err != nil {
		return check.Failed(err)
	}

	req.AddToOwnedResources(rr...)

	return check.Completed()
}

func (r *Reconciler) finalize(req *rApi.Request[*crdsv1.Workspace]) stepResult.Result {
	if step := req.EnsureCheckList([]rApi.CheckMeta{
		{Name: "uninstall workspace"},
	}); !step.ShouldProceed() {
		return step
	}

	check := rApi.NewRunningCheck("uninstall workspace", req)

	if step := req.CleanupOwnedResources(check); !step.ShouldProceed() {
		return step
	}

	return req.Finalize()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()

	if r.YAMLClient == nil {
		return fmt.Errorf("yaml client must be set")
	}

	r.recorder = mgr.GetEventRecorderFor(r.GetName())

	var err error
	r.workspaceDeploymentTemplate, err = templates.Read(templates.WorkspaceIngressTemplate, templates.WorkspaceSTSTemplate, templates.WorkspaceServiceTemplate)
	if err != nil {
		return err
	}

	builder := ctrl.NewControllerManagedBy(mgr).For(&crdsv1.Workspace{})
	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.Owns(&appsv1.Deployment{})
	builder.Owns(&corev1.Service{})
	builder.Owns(&crdsv1.Router{})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
