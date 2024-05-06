package vm_controller

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	ct "github.com/kloudlite/operator/apis/common-types"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/virtual-machine/internal/env"
	"github.com/kloudlite/operator/operators/virtual-machine/internal/templates"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Env        *env.Env
	logger     logging.Logger
	Name       string
	yamlClient kubectl.YAMLClient

	templateVMJob []byte
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	DefaultsPatched string = "defaults-patched"
	CreateVM        string = "create-vm"
	TrackVMStatus   string = "track-vm-status"
	DeleteVM        string = "delete-vm"
)

var ApplyChecklist = []rApi.CheckMeta{
	{Name: DefaultsPatched, Title: "Defaults Patched", Debug: true},
	{Name: CreateVM, Title: "Create Virtual Machine"},
	{Name: TrackVMStatus, Title: "Track VM Status"},
}

var DeleteChecklist = []rApi.CheckMeta{}

// +kubebuilder:rbac:groups=clusters,resources=virtualmachines,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=clusters,resources=virtualmachines/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=clusters,resources=virtualmachines/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &clustersv1.VirtualMachine{})
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

	if step := req.EnsureCheckList(ApplyChecklist); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.patchDefaults(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.applyVMJob(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.trackJobStatus(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *Reconciler) patchDefaults(req *rApi.Request[*clustersv1.VirtualMachine]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(DefaultsPatched, req)

	hasUpdate := false

	switch obj.Spec.CloudProvider {
	case ct.CloudProviderGCP:
		{
			if obj.Spec.GCP.VPC == nil {
				hasUpdate = true
				obj.Spec.GCP.VPC = &clustersv1.GcpVPCParams{Name: "default"}
			}

			if obj.Spec.GCP.AvailabilityZone == "" {
				hasUpdate = true
				obj.Spec.GCP.AvailabilityZone = fmt.Sprintf("%s-a", obj.Spec.GCP.Region)
			}
		}
	}

	if obj.Spec.ControllerParams.TFWorkspaceName == "" {
		hasUpdate = true
		obj.Spec.ControllerParams.TFWorkspaceName = obj.Name
	}

	if obj.Spec.ControllerParams.TFWorkspaceNamespace == "" {
		hasUpdate = true
		obj.Spec.ControllerParams.TFWorkspaceNamespace = obj.Namespace
	}

	if obj.Spec.ControllerParams.JobRef.Name == "" {
		hasUpdate = true
		obj.Spec.ControllerParams.JobRef.Name = fmt.Sprintf("vm-job-%s", obj.Name)
	}

	if obj.Spec.ControllerParams.JobRef.Namespace == "" {
		hasUpdate = true
		obj.Spec.ControllerParams.JobRef.Namespace = obj.Namespace
	}

	if hasUpdate {
		if err := r.Update(ctx, obj); err != nil {
			return check.Failed(err)
		}
		return check.StillRunning(fmt.Errorf("waiting for resource to be patched"))
	}

	return check.Completed()
}

func (r *Reconciler) finalize(req *rApi.Request[*clustersv1.VirtualMachine]) stepResult.Result {
	if step := req.EnsureCheckList(DeleteChecklist); !step.ShouldProceed() {
		return step
	}

	if step := req.CleanupOwnedResources(); !step.ShouldProceed() {
		return step
	}

	return req.Finalize()
}

func (r *Reconciler) parseSpecToValuesJSON(req *rApi.Request[*clustersv1.VirtualMachine]) ([]byte, error) {
	ctx, obj := req.Context(), req.Object

	switch obj.Spec.CloudProvider {
	case ct.CloudProviderAWS:
		{
			return nil, fmt.Errorf("aws support incoming")
		}

	case ct.CloudProviderGCP:
		{
			g, err := obj.Spec.GCP.RetrieveCreds(ctx, r.Client)
			if err != nil {
				return nil, err
			}

			return json.Marshal(templates.TFVarsGcpVM{
				AllowIncomingHttpTraffic: obj.Spec.GCP.AllowIncomingHttpTraffic,
				AllowSsh:                 obj.Spec.GCP.AllowSSH,
				AvailabilityZone:         obj.Spec.GCP.AvailabilityZone,
				BootvolumeSize:           float64(obj.Spec.GCP.BootVolumeSize),
				BootvolumeType:           "pd-ssd",
				GcpCredentialsJson:       base64.StdEncoding.EncodeToString([]byte(strings.TrimSpace(g.ServiceAccountJSON))),
				GcpProjectID:             obj.Spec.GCP.GCPProjectID,
				GcpRegion:                obj.Spec.GCP.Region,
				Labels: map[string]string{
					"kloudlite-account": obj.Spec.KloudliteAccount,
				},
				MachineState:   obj.Spec.MachineState,
				MachineType:    obj.Spec.GCP.MachineType,
				NamePrefix:     "vm",
				Network:        obj.Spec.GCP.VPC.Name,
				ProvisionMode:  string(obj.Spec.GCP.PoolType),
				ServiceAccount: obj.Spec.GCP.ServiceAccount,
				StartupScript:  obj.Spec.GCP.StartupScript,
				VmName:         obj.Name,
			})
		}
	default:
		{
			return nil, fmt.Errorf("unsupported cloudprovider: %s", obj.Spec.CloudProvider)
		}
	}
}

func (r *Reconciler) applyVMJob(req *rApi.Request[*clustersv1.VirtualMachine]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(CreateVM, req)

	valuesJSON, err := r.parseSpecToValuesJSON(req)
	if err != nil {
		return check.Failed(err)
	}

	b, err := templates.ParseBytes(r.templateVMJob, templates.VMJobVars{
		JobMetadata: metav1.ObjectMeta{
			Name:              fmt.Sprintf("vm-job-%s", obj.Name),
			Namespace:         obj.Namespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
			Labels:            obj.GetLabels(),
			Annotations:       fn.FilterObservabilityAnnotations(obj.GetAnnotations()),
			OwnerReferences:   []metav1.OwnerReference{fn.AsOwner(obj, true)},
		},
		NodeSelector:  r.Env.IACJobNodeSelector,
		Tolerations:   r.Env.IACJobTolerations,
		CloudProvider: string(obj.Spec.CloudProvider),

		JobImage:             r.Env.IACJobImage,
		JobImagePullPolicy:   "IfNotPresent",
		TFWorkspaceName:      obj.Spec.ControllerParams.TFWorkspaceName,
		TFWorkspaceNamespace: obj.Spec.ControllerParams.TFWorkspaceNamespace,
		ValuesJSON:           string(valuesJSON),
	})
	if err != nil {
		return check.Failed(err)
	}

	rr, err := r.yamlClient.ApplyYAML(ctx, b)
	if err != nil {
		return check.Failed(err)
	}

	req.AddToOwnedResources(rr...)

	return check.Completed()
}

func (r *Reconciler) trackJobStatus(req *rApi.Request[*clustersv1.VirtualMachine]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(TrackVMStatus, req)

	job, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.ControllerParams.JobRef.Namespace, obj.Spec.ControllerParams.JobRef.Name), &crdsv1.Lifecycle{})
	if err != nil {
		return check.Failed(err)
	}

	if !job.HasCompleted() {
		return check.StillRunning(fmt.Errorf("waiting for job to complete"))
	}

	if job.Status.Phase == crdsv1.JobPhaseFailed {
		return check.Failed(fmt.Errorf("job failed"))
	}

	return check.Completed()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig(), kubectl.YAMLClientOpts{Logger: r.logger})

	var err error
	r.templateVMJob, err = templates.Read(templates.VMJobTemplate)
	if err != nil {
		return err
	}

	builder := ctrl.NewControllerManagedBy(mgr).For(&clustersv1.VirtualMachine{})
	builder.Owns(&crdsv1.Lifecycle{})
	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
