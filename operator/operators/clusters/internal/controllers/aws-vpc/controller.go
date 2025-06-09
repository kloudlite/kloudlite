package aws_vpc

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	common_types "github.com/kloudlite/operator/apis/common-types"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/clusters/internal/env"
	"github.com/kloudlite/operator/operators/clusters/internal/templates"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

type AwsVPCReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Env        *env.Env
	logger     logging.Logger
	Name       string
	yamlClient kubectl.YAMLClient

	templateAwsVPCJob []byte
}

func (r *AwsVPCReconciler) GetName() string {
	return r.Name
}

const defaultVpcCIDR = "10.0.0.0/16"

func genSubnetCIDR(index int) string {
	return fmt.Sprintf("10.0.%d.0/21", 8*(index))
}

func getPrefixedName(base string) string {
	return fmt.Sprintf("aws-vpc-%s", base)
}

const (
	patchDefaults string = "patch-defaults"
	createVPCJob  string = "create-vpc-job"
	deleteVPCJob  string = "delete-vpc-job"
)

var (
	ApplyChecklist = []rApi.CheckMeta{
		{Name: patchDefaults, Title: "Patch Defaults"},
		{Name: createVPCJob, Title: "Create VPC Lifecycle"},
	}
	DeleteChecklist = []rApi.CheckMeta{{Name: createVPCJob, Title: "Delete VPC Lifecycle"}}
)

// +kubebuilder:rbac:groups=clusters,resources=clusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=clusters,resources=clusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=clusters,resources=clusters/finalizers,verbs=update

func (r *AwsVPCReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &clustersv1.AwsVPC{})
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

	if step := r.applyVPC(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *AwsVPCReconciler) finalize(req *rApi.Request[*clustersv1.AwsVPC]) stepResult.Result {
	if step := req.EnsureCheckList(DeleteChecklist); !step.ShouldProceed() {
		return step
	}

	if step := req.CleanupOwnedResources(); !step.ShouldProceed() {
		return step
	}

	return req.Finalize()
}

func (r *AwsVPCReconciler) patchDefaults(req *rApi.Request[*clustersv1.AwsVPC]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(patchDefaults, req)

	hasUpdated := false

	if obj.Spec.CIDR == "" {
		hasUpdated = true
		obj.Spec.CIDR = defaultVpcCIDR
	}

	if obj.Spec.CIDR == defaultVpcCIDR {
		// INFO: only if VPC CIDR  == defaultVpcCIDR, otherwise the following does not hold any meanings
		if obj.Spec.PublicSubnets == nil {
			hasUpdated = true

			zones, ok := clustersv1.AwsRegionToAZs[clustersv1.AwsRegion(obj.Spec.Region)]
			if !ok {
				return check.Failed(fmt.Errorf("invalid region (%s), no Availability zones defined for it", obj.Spec.Region))
			}

			obj.Spec.PublicSubnets = make([]clustersv1.AwsSubnet, len(zones))
			slices.Sort(zones)

			for i := range zones {
				obj.Spec.PublicSubnets[i] = clustersv1.AwsSubnet{
					AvailabilityZone: string(zones[i]),
					CIDR:             genSubnetCIDR(i),
				}
			}
		}
	}

	if obj.Spec.Output == nil {
		hasUpdated = true
		obj.Spec.Output = &common_types.SecretRef{
			Name:      getPrefixedName(obj.Name),
			Namespace: obj.Namespace,
		}
	}

	if hasUpdated {
		if err := r.Update(ctx, obj); err != nil {
			return check.Failed(err)
		}
		return req.Done().RequeueAfter(500 * time.Millisecond)
	}

	return check.Completed()
}

func genTFValues(obj *clustersv1.AwsVPC, credsSecret *corev1.Secret, ev *env.Env) ([]byte, error) {
	switch obj.Spec.Credentials.AuthMechanism {
	case clustersv1.AwsAuthMechanismSecretKeys:
		{
			awscreds, err := fn.ParseFromSecret[clustersv1.AwsAuthSecretKeys](credsSecret)
			if err != nil {
				return nil, err
			}

			return json.Marshal(map[string]any{
				"aws_access_key": awscreds.AccessKey,
				"aws_secret_key": awscreds.SecretKey,
				"aws_region":     obj.Spec.Region,
				// "aws_assume_role": map[string]any{"enabled": false},
				"vpc_name": getPrefixedName(obj.Name),
				"vpc_cidr": obj.Spec.CIDR,
				"public_subnets": func() []map[string]any {
					results := make([]map[string]any, len(obj.Spec.PublicSubnets))
					for i := range obj.Spec.PublicSubnets {
						results[i] = map[string]any{
							"availability_zone": obj.Spec.PublicSubnets[i].AvailabilityZone,
							"cidr":              obj.Spec.PublicSubnets[i].CIDR,
						}
					}
					return results
				}(),
				"tags": obj.Labels,
			})
		}
	case clustersv1.AwsAuthMechanismAssumeRole:
		{

			awscreds, err := fn.ParseFromSecret[clustersv1.AwsAssumeRoleParams](credsSecret)
			if err != nil {
				return nil, err
			}

			return json.Marshal(map[string]any{
				"aws_access_key": ev.KlAwsAccessKey,
				"aws_secret_key": ev.KlAwsSecretKey,
				"aws_region":     obj.Spec.Region,
				"aws_assume_role": map[string]any{
					"enabled":     true,
					"role_arn":    string(awscreds.RoleARN),
					"external_id": string(awscreds.ExternalID),
				},

				"vpc_name": getPrefixedName(obj.Name),
				"vpc_cidr": obj.Spec.CIDR,
				"public_subnets": func() []map[string]any {
					results := make([]map[string]any, len(obj.Spec.PublicSubnets))
					for i := range obj.Spec.PublicSubnets {
						results[i] = map[string]any{
							"availability_zone": obj.Spec.PublicSubnets[i].AvailabilityZone,
							"cidr":              obj.Spec.PublicSubnets[i].CIDR,
						}
					}
					return results
				}(),
				"tags": obj.Labels,
			})
		}
	default:
		{
			return nil, fmt.Errorf("unknown auth mechanism %s", obj.Spec.Credentials.AuthMechanism)
		}
	}
}

func (r *AwsVPCReconciler) applyVPC(req *rApi.Request[*clustersv1.AwsVPC]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(createVPCJob, req)

	vpcOutput, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.Output.Namespace, obj.Spec.Output.Name), &corev1.Secret{})
	if err != nil {
		if apiErrors.IsNotFound(err) {
			return check.StillRunning(err)
		}
		vpcOutput = nil
	}

	if vpcOutput != nil {
		return check.Completed()
	}

	credsSecret := &corev1.Secret{}
	if err := r.Get(ctx, fn.NN(obj.Spec.Credentials.SecretRef.Namespace, obj.Spec.Credentials.SecretRef.Name), credsSecret); err != nil {
		return check.Failed(err).Err(nil)
	}

	valuesBytes, err := genTFValues(obj, credsSecret, r.Env)
	if err != nil {
		return check.Failed(err).Err(nil)
	}

	b, err := templates.ParseBytes(r.templateAwsVPCJob, templates.AwsVPCJobVars{
		JobMetadata: metav1.ObjectMeta{
			Name:            getPrefixedName(obj.Name),
			Namespace:       obj.Namespace,
			OwnerReferences: []metav1.OwnerReference{fn.AsOwner(obj, true)},
			Labels:          obj.GetLabels(),
			Annotations:     fn.FilterObservabilityAnnotations(obj.GetAnnotations()),
		},
		NodeSelector: r.Env.IACJobNodeSelector,
		Tolerations:  r.Env.IACJobTolerations,
		JobImage:     r.Env.IACJobImage,

		TFWorkspaceName:            fmt.Sprintf("aws-vpc-%s", obj.Name),
		TFWorkspaceSecretNamespace: obj.Namespace,
		ValuesJSON:                 string(valuesBytes),
		// AWS: templates.AWSClusterJobParams{
		// 	AccessKeyID:     r.Env.KlAwsAccessKey,
		// 	AccessKeySecret: r.Env.KlAwsSecretKey,
		// },
		VPCOutputSecretName:      obj.Spec.Output.Name,
		VPCOutputSecretNamespace: obj.Spec.Output.Namespace,
	})
	if err != nil {
		return check.Failed(err).NoRequeue()
	}

	rr, err := r.yamlClient.ApplyYAML(ctx, b)
	if err != nil {
		return check.Failed(err)
	}

	req.AddToOwnedResources(rr...)

	job, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, getPrefixedName(obj.Name)), &crdsv1.Lifecycle{})
	if err != nil {
		return check.Failed(err)
	}

	if job.HasCompleted() {
		if job.Status.Phase == crdsv1.JobPhaseFailed {
			return check.Failed(fmt.Errorf("job failed"))
		}
		return check.Completed()
	}

	return check.StillRunning(fmt.Errorf("waiting for job to complete"))
}

func (r *AwsVPCReconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig(), kubectl.YAMLClientOpts{Logger: r.logger})

	b, err := templates.Read(templates.AwsVPCJob)
	if err != nil {
		return err
	}
	r.templateAwsVPCJob = b

	builder := ctrl.NewControllerManagedBy(mgr).For(&clustersv1.AwsVPC{})
	builder.Owns(&batchv1.Job{})
	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
