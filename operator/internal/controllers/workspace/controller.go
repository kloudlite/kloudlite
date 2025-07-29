package workspace

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/codingconcepts/env"
	fn "github.com/kloudlite/kloudlite/operator/toolkit/functions"
	"github.com/kloudlite/kloudlite/operator/toolkit/kubectl"
	"github.com/kloudlite/kloudlite/operator/toolkit/reconciler"
	"github.com/kloudlite/operator/api/v1"
	"github.com/kloudlite/operator/internal/controllers/workspace/internal/templates"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

type envVars struct {
	MaxConcurrentReconciles int `env:"MAX_CONCURRENT_RECONCILES" default:"5"`

	WorkspaceImageInitContainer string `env:"WORKSPACE_IMAGE_INIT_CONTAINER" default:"ghcr.io/kloudlite/iac/workspace:latest"`
	// WorkspaceImageSSH             string `env:"WORKSPACE_IMAGE_SSH" default:"ghcr.io/kloudlite/iac/workspace:latest"`
	WorkspaceImageSSH             string `env:"WORKSPACE_IMAGE_SSH" default:"ghcr.io/kloudlite/kloudlite/operator/workspace-ssh:debug"`
	WorkspaceImageTTYD            string `env:"WORKSPACE_IMAGE_TTYD" default:"ghcr.io/kloudlite/iac/ttyd:latest"`
	WorkspaceImageJupyterNotebook string `env:"WORKSPACE_IMAGE_JUPYTER_NOTEBOOK" default:"ghcr.io/kloudlite/iac/jupyter:latest"`
	// WorkspaceImageCodeServer      string `env:"WORKSPACE_IMAGE_CODE_SERVER" default:"ghcr.io/kloudlite/iac/code-server:latest"`
	WorkspaceImageCodeServer string `env:"WORKSPACE_IMAGE_CODE_SERVER" default:"ghcr.io/kloudlite/kloudlite/operator/workspace-code-server:debug"`
	// WorkspaceImageVscodeServer    string `env:"WORKSPACE_IMAGE_VSCODE_SERVER" default:"ghcr.io/kloudlite/iac/vscode-server:latest"`
	WorkspaceImageVscodeServer string `env:"WORKSPACE_IMAGE_VSCODE_SERVER" default:"ghcr.io/kloudlite/kloudlite/operator/workspace-vscode-server:debug"`
}

// Reconciler reconciles a Workspace object
type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme

	env envVars

	YAMLClient kubectl.YAMLClient

	templateStatefulSet []byte
	templateService     []byte
	templateRouter      []byte
}

func (r *Reconciler) GetName() string {
	return v1.ProjectDomain + "/workspace-controller"
}

const (
	SSHPort             int32 = 22
	TTYDPort            int32 = 54535
	JupyterNotebookPort int32 = 8888
	CodeServerPort      int32 = 8080
)

// +kubebuilder:rbac:groups=kloudlite.io,resources=workspaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kloudlite.io,resources=workspaces/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kloudlite.io,resources=workspaces/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := reconciler.NewRequest(ctx, r.Client, request.NamespacedName, &v1.Workspace{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	req.PreReconcile()
	defer req.PostReconcile()

	return reconciler.ReconcileSteps(req, []reconciler.Step[*v1.Workspace]{
		{
			Name:     "setup statefulset",
			Title:    "Setup Kubernetes StatefulSet",
			OnCreate: r.createStatefulSet,
			OnDelete: r.cleanupStatefulSet,
		},
		{
			Name:     "setup service",
			Title:    "Setup Kubernetes Service",
			OnCreate: r.createService,
			OnDelete: r.cleanupService,
		},
		{
			Name:     "setup router",
			Title:    "Setup Ingress Router",
			OnCreate: r.createRouter,
			OnDelete: r.cleanupRouter,
		},
	})
}

func (r *Reconciler) createStatefulSet(check *reconciler.Check[*v1.Workspace], obj *v1.Workspace) reconciler.StepResult {
	ws, err := reconciler.Get(check.Context(), r.Client, fn.NN("", obj.Spec.WorkMachine), &v1.Workmachine{})
	if err != nil {
		return check.Failed(err)
	}

	b, err := templates.ParseBytes(r.templateStatefulSet, templates.StatefulSetTemplateArgs{
		Metadata: metav1.ObjectMeta{
			Name:            obj.Name,
			Namespace:       obj.Namespace,
			OwnerReferences: []metav1.OwnerReference{fn.AsOwner(obj, true)},
		},
		Selector: map[string]string{
			v1.WorkspaceNameKey: obj.Name,
		},
		PodLabels:                map[string]string{},
		WorkMachineTolerationKey: v1.WorkspaceNameKey,
		WorkMachineName:          obj.Spec.WorkMachine,
		ServiceAccountName:       obj.Spec.ServiceAccountName,
		ImageInitContainer:       r.env.WorkspaceImageInitContainer,
		ImageSSH:                 r.env.WorkspaceImageSSH,
		Paused:                   obj.Spec.Paused,
		EnableTTYD:               obj.Spec.EnableTTYD,
		ImageTTYD:                r.env.WorkspaceImageTTYD,

		EnableJupyterNotebook: obj.Spec.EnableJupyterNotebook,
		ImageJupyterNotebook:  r.env.WorkspaceImageJupyterNotebook,

		EnableCodeServer: obj.Spec.EnableCodeServer,
		ImageCodeServer:  r.env.WorkspaceImageCodeServer,

		EnableVSCodeServer: obj.Spec.EnableVSCodeServer,
		ImageVSCodeServer:  r.env.WorkspaceImageVscodeServer,

		ImagePullPolicy:     obj.Spec.ImagePullPolicy,
		KloudliteDeviceFQDN: fmt.Sprintf("%s-headless.%s.svc.cluster.local", obj.Name, obj.Namespace),
		SSHSecretName:       ws.Spec.SSH.Secret.Name,
	})
	if err != nil {
		return check.Failed(err)
	}

	objects, err := r.YAMLClient.ApplyYAML(check.Context(), b)
	if err != nil {
		return check.Failed(err)
	}

	if len(objects) != 1 {
		return check.Failed(fmt.Errorf("expected 1 object from YAMLClient.ApplyYAML, but got %d", len(objects)))
	}

	ss, err := fn.FromUnstructured(objects[0], &appsv1.StatefulSet{})
	if err != nil {
		return check.Failed(err)
	}

	if ss.Status.Replicas != ss.Status.ReadyReplicas {
		return check.Failed(fmt.Errorf("waiting for replica to be ready"))
	}

	return check.Passed()
}

func (r *Reconciler) cleanupStatefulSet(check *reconciler.Check[*v1.Workspace], obj *v1.Workspace) reconciler.StepResult {
	if err := fn.DeleteAndWait(check.Context(), r.Client, &appsv1.StatefulSet{ObjectMeta: metav1.ObjectMeta{Name: obj.Name, Namespace: obj.Namespace}}); err != nil {
		return check.Failed(err)
	}
	return check.Passed()
}

func (r *Reconciler) createService(check *reconciler.Check[*v1.Workspace], obj *v1.Workspace) reconciler.StepResult {
	b, err := templates.ParseBytes(r.templateService, templates.ServiceTemplateArgs{
		Metadata: metav1.ObjectMeta{
			Name:      obj.Name,
			Namespace: obj.Namespace,
			Labels: map[string]string{
				v1.WorkspaceNameKey: obj.Name,
			},
			OwnerReferences: []metav1.OwnerReference{fn.AsOwner(obj, true)},
		},
		Selector: map[string]string{
			v1.WorkspaceNameKey: obj.Name,
		},
		EnableJupyterNotebook: obj.Spec.EnableJupyterNotebook,
		EnableCodeServer:      obj.Spec.EnableCodeServer,
		EnableTTYD:            obj.Spec.EnableTTYD,

		SSHPort:        SSHPort,
		TTYDPort:       TTYDPort,
		NotebookPort:   JupyterNotebookPort,
		CodeServerPort: CodeServerPort,
	})
	if err != nil {
		return check.Failed(err)
	}

	if _, err := r.YAMLClient.ApplyYAML(check.Context(), b); err != nil {
		return check.Failed(err)
	}

	return check.Passed()
}

func (r *Reconciler) cleanupService(check *reconciler.Check[*v1.Workspace], obj *v1.Workspace) reconciler.StepResult {
	svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: obj.Name, Namespace: obj.Namespace}}
	if err := fn.DeleteAndWait(check.Context(), r.Client, svc); err != nil {
		return check.Failed(err)
	}

	return check.Passed()
}

func (r *Reconciler) createRouter(check *reconciler.Check[*v1.Workspace], obj *v1.Workspace) reconciler.StepResult {
	b, err := templates.ParseBytes(r.templateRouter, templates.RouterTemplateArgs{
		Metadata: metav1.ObjectMeta{
			Name:      obj.Name,
			Namespace: obj.Namespace,
			Labels: map[string]string{
				v1.WorkspaceNameKey: obj.Name,
			},
			OwnerReferences: []metav1.OwnerReference{fn.AsOwner(obj, true)},
		},
		WorkMachineName:       obj.Spec.WorkMachine,
		KloudliteDomain:       "kloudlite.local",
		EnableJupyterNotebook: obj.Spec.EnableJupyterNotebook,
		EnableCodeServer:      obj.Spec.EnableCodeServer,
		EnableTTYD:            obj.Spec.EnableTTYD,
		TTYDPort:              TTYDPort,
		NotebookPort:          JupyterNotebookPort,
		CodeServerPort:        CodeServerPort,

		ServiceName: obj.Name,
		ServicePath: "/",
	})
	if err != nil {
		return check.Failed(err)
	}

	objects, err := r.YAMLClient.ApplyYAML(check.Context(), b)
	if err != nil {
		return check.Failed(err)
	}

	if len(objects) != 1 {
		return check.Failed(fmt.Errorf("expected 1 object from YAMLClient.ApplyYAML, but got %d", len(objects)))
	}

	router, err := fn.FromUnstructured(objects[0], &v1.Router{})
	if err != nil {
		return check.Failed(err)
	}

	if !router.Status.IsReady {
		return check.Failed(err)
	}

	return check.Passed()
}

func (r *Reconciler) cleanupRouter(check *reconciler.Check[*v1.Workspace], obj *v1.Workspace) reconciler.StepResult {
	svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: obj.Name, Namespace: obj.Namespace}}
	if err := fn.DeleteAndWait(check.Context(), r.Client, svc); err != nil {
		return check.Failed(err)
	}

	return check.Passed()
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()

	if err := env.Set(&r.env); err != nil {
		return fmt.Errorf("failed to set env: %w", err)
	}

	if r.YAMLClient == nil {
		return fmt.Errorf("reconciler.YAMLClient must be set")
	}

	var err error
	r.templateStatefulSet, err = templates.Read(templates.StatefulSetTemplate)
	if err != nil {
		return fmt.Errorf("failed to read statefulset template file: %w", err)
	}

	r.templateService, err = templates.Read(templates.ServiceTemplate)
	if err != nil {
		return fmt.Errorf("failed to read service template file: %w", err)
	}

	r.templateRouter, err = templates.Read(templates.RouterTemplate)
	if err != nil {
		return fmt.Errorf("failed to read service template file: %w", err)
	}

	builder := ctrl.NewControllerManagedBy(mgr).For(&v1.Workspace{}).Named(r.GetName())
	builder.Owns(&appsv1.StatefulSet{})

	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.env.MaxConcurrentReconciles})
	builder.WithEventFilter(reconciler.ReconcileFilter(mgr.GetEventRecorderFor(r.GetName())))
	return builder.Complete(r)
}
