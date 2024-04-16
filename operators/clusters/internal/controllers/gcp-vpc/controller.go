package gcp_vpc

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	ct "github.com/kloudlite/operator/apis/common-types"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/clusters/internal/env"
	"github.com/kloudlite/operator/operators/clusters/internal/templates"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
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
)

type GcpVPCReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Env        *env.Env
	logger     logging.Logger
	Name       string
	yamlClient kubectl.YAMLClient

	templateGcpVPCJob []byte
}

func (r *GcpVPCReconciler) GetName() string {
	return r.Name
}

const (
	createVPC     string = "create-vpc"
	deleteVPC     string = "delete-vpc"
	patchDefaults string = "patch-defaults"
)

var ApplyChecklist = []rApi.CheckMeta{
	{Name: patchDefaults, Title: "Patch Defaults", Debug: true},
	{Name: createVPC, Title: "Create VPC"},
}

var DeleteChecklist = []rApi.CheckMeta{
	{Name: deleteVPC, Title: "Delete VPC"},
}

type GCPVpcTFVars struct {
	ProjectID       string `json:"gcp_project_id"`
	Region          string `json:"gcp_region"`
	CredentialsJSON string `json:"gcp_credentials_json"`
	VPCName         string `json:"vpc_name"`
}

// +kubebuilder:rbac:groups=clusters,resources=clusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=clusters,resources=clusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=clusters,resources=clusters/finalizers,verbs=update

func (r *GcpVPCReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, request.NamespacedName, &clustersv1.GcpVPC{})
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

	if step := r.patchDefaults(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
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

	if step := r.applyVPCJob(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *GcpVPCReconciler) patchDefaults(req *rApi.Request[*clustersv1.GcpVPC]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(patchDefaults, req)

	hasUpdate := false
	if obj.Spec.Output.SecretRef.Name == "" {
		hasUpdate = true
		obj.Spec.Output.SecretRef.Name = fmt.Sprintf("gcp-vpc-%s", obj.Name)
	}

	if obj.Spec.Output.SecretRef.Namespace == "" {
		hasUpdate = true
		obj.Spec.Output.SecretRef.Namespace = obj.Namespace
	}

	if obj.Spec.Output.TFWorkspaceName == "" {
		hasUpdate = true
		obj.Spec.Output.TFWorkspaceName = fmt.Sprintf("gcp-vpc-%s", obj.Name)
	}

	if hasUpdate {
		if err := r.Update(ctx, obj); err != nil {
			return check.Failed(err)
		}
	}

	return check.Completed()
}

func (r *GcpVPCReconciler) finalize(req *rApi.Request[*clustersv1.GcpVPC]) stepResult.Result {
	if step := req.EnsureCheckList(DeleteChecklist); !step.ShouldProceed() {
		return step
	}

	if step := req.CleanupOwnedResources(); !step.ShouldProceed() {
		return step
	}

	return req.Finalize()
}

func (r *GcpVPCReconciler) applyVPCJob(req *rApi.Request[*clustersv1.GcpVPC]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(createVPC, req)

	credsSecret := &corev1.Secret{}
	if err := r.Get(ctx, fn.NN(obj.Spec.CredentialsRef.Namespace, obj.Spec.CredentialsRef.Name), credsSecret); err != nil {
		return check.Failed(err)
	}

	gcpCreds, err := fn.ParseFromSecret[clustersv1.GCPCredentials](credsSecret)
	if err != nil {
		return check.Failed(err)
	}

	valuesB, err := json.Marshal(GCPVpcTFVars{
		ProjectID:       obj.Spec.GCPProjectID,
		Region:          obj.Spec.Region,
		CredentialsJSON: base64.StdEncoding.EncodeToString([]byte(gcpCreds.ServiceAccountJSON)),
		VPCName:         obj.Name,
	})
	if err != nil {
		return check.Failed(err)
	}

	b, err := templates.ParseBytes(r.templateGcpVPCJob, templates.GcpVPCJobVars{
		JobMetadata:              metav1.ObjectMeta{Name: obj.Name, Namespace: obj.Namespace},
		JobImage:                 r.Env.IACJobImage,
		JobImagePullPolicy:       "Always",
		TFStateSecretNamespace:   obj.Namespace,
		TFStateSecretName:        obj.Spec.Output.TFWorkspaceName,
		ValuesJSON:               string(valuesB),
		CloudProvider:            ct.CloudProviderGCP,
		VPCOutputSecretName:      obj.Spec.Output.SecretRef.Name,
		VPCOutputSecretNamespace: obj.Spec.Output.SecretRef.Namespace,
	})
	if err != nil {
		return check.Failed(err)
	}

	rr, err := r.yamlClient.ApplyYAML(ctx, b)
	if err != nil {
		return check.StillRunning(err)
	}

	req.AddToOwnedResources(rr...)

	job, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Name), &crdsv1.Job{})
	if err != nil {
		return check.Failed(err)
	}

	if job.Status.Running != nil && *job.Status.Running {
		return check.StillRunning(fmt.Errorf("waiting for job to finish"))
	}

	if job.Status.Failed != nil && *job.Status.Failed {
		return check.Failed(fmt.Errorf("job failed"))
	}

	return check.Completed()
}

func (r *GcpVPCReconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig(), kubectl.YAMLClientOpts{Logger: r.logger})

	var err error
	r.templateGcpVPCJob, err = templates.Read(templates.GcpVPCJob)
	if err != nil {
		return err
	}

	builder := ctrl.NewControllerManagedBy(mgr).For(&clustersv1.GcpVPC{})
	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	return builder.Complete(r)
}
