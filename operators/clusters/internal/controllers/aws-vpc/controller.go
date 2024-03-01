package aws_vpc

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	common_types "github.com/kloudlite/operator/apis/common-types"
	"github.com/kloudlite/operator/operators/clusters/internal/env"
	"github.com/kloudlite/operator/operators/clusters/internal/templates"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	job_manager "github.com/kloudlite/operator/pkg/job-helper"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
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

	templateAwsVPCJob     []byte
	templateAwsVPCJobRBAC []byte
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

const vpcJobServiceAccount = "aws-vpc-job-sa"

const (
	labelResourceGeneration string = "kloudlite.io/controller.resource-generation"
	labelVPCCreate          string = "kloudlite.io/aws-vpc.create"
	labelVPCDelete          string = "kloudlite.io/aws-vpc.delete"
)

// var (
// 	creationCheckList = []string{"patch-defaults", "ensures-job-rbac", "create-vpc-job"}
// 	deletionCheckList = []string{"patch-defaults", "ensures-job-rbac", "delete-vpc-job"}
// )

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

	if step := r.patchDefaults(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureJobRBAC(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.createVPC(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *AwsVPCReconciler) finalize(req *rApi.Request[*clustersv1.AwsVPC]) stepResult.Result {
	checkName := "finalizing"

	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	if step := r.cleanupVPC(req); !step.ShouldProceed() {
		return step
	}

	return req.Finalize()
}

func (r *AwsVPCReconciler) patchDefaults(req *rApi.Request[*clustersv1.AwsVPC]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	checkName := "patch-defaults"

	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	fail := func(err error) stepResult.Result {
		return req.CheckFailed(checkName, check, err.Error())
	}

	hasUpdated := false

	if obj.Spec.CIDR == "" {
		hasUpdated = true
		obj.Spec.CIDR = defaultVpcCIDR
	}

	if obj.Spec.CIDR == defaultVpcCIDR {
		// INFO: only if VPC CIDR  == defaultVpcCIDR, otherwise the following does not hold any meanings
		if obj.Spec.PublicSubnets == nil {
			hasUpdated = true

			zones, ok := clustersv1.AwsRegionToAZs[obj.Spec.Region]
			if !ok {
				return fail(fmt.Errorf("invalid region (%s), no Availability zones defined for it", obj.Spec.Region)).Err(nil)
			}

			obj.Spec.PublicSubnets = make([]clustersv1.AwsSubnet, len(zones))

			slices.Sort(zones)

			for i := range zones {
				obj.Spec.PublicSubnets[i] = clustersv1.AwsSubnet{
					AvailabilityZone: zones[i],
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
			return fail(err)
		}
		return req.Done().RequeueAfter(500 * time.Millisecond)
	}

	check.Status = true
	if check != obj.Status.Checks[checkName] {
		fn.MapSet(&obj.Status.Checks, checkName, check)
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

func (r *AwsVPCReconciler) ensureJobRBAC(req *rApi.Request[*clustersv1.AwsVPC]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	checkName := "job-rbac"

	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	fail := func(err error) stepResult.Result {
		return req.CheckFailed(checkName, check, err.Error())
	}

	b, err := templates.ParseBytes(r.templateAwsVPCJobRBAC, map[string]any{
		"service-account-name": vpcJobServiceAccount,
		"namespace":            obj.Namespace,
	})
	if err != nil {
		return fail(err).Err(nil)
	}

	if _, err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
		return fail(err)
	}

	check.Status = true
	if check != obj.Status.Checks[checkName] {
		fn.MapSet(&obj.Status.Checks, checkName, check)
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

func genTFValues(obj *clustersv1.AwsVPC, credsSecret *corev1.Secret, ev *env.Env) ([]byte, error) {
	return json.Marshal(map[string]any{
		"aws_access_key": ev.KlAwsAccessKey,
		"aws_secret_key": ev.KlAwsSecretKey,
		"aws_region":     obj.Spec.Region,
		"aws_assume_role": map[string]any{
			"enabled":     true,
			"role_arn":    string(credsSecret.Data[obj.Spec.CredentialKeys.KeyAWSAssumeRoleRoleARN]),
			"external_id": string(credsSecret.Data[obj.Spec.CredentialKeys.KeyAWSAssumeRoleExternalID]),
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

func (r *AwsVPCReconciler) createVPC(req *rApi.Request[*clustersv1.AwsVPC]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	checkName := "checkName"

	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	fail := func(err error) stepResult.Result {
		return req.CheckFailed(checkName, check, err.Error())
	}

	job := &batchv1.Job{}
	if err := r.Get(ctx, fn.NN(obj.Namespace, getPrefixedName(obj.Name)), job); err != nil {
		job = nil
	}

	if job == nil {
		credsSecret := &corev1.Secret{}
		if err := r.Get(ctx, fn.NN(obj.Spec.CredentialsRef.Namespace, obj.Spec.CredentialsRef.Name), credsSecret); err != nil {
			return fail(err)
		}

		valuesBytes, err := genTFValues(obj, credsSecret, r.Env)
		if err != nil {
			return fail(err).Err(nil)
		}

		b, err := templates.ParseBytes(r.templateAwsVPCJob, map[string]any{
			"action": "apply",

			"job-name":          getPrefixedName(obj.Name),
			"job-namespace":     obj.Namespace,
			"job-image":         r.Env.IACJobImage,
			"job-tolerations":   r.Env.IACJobTolerations,
			"job-node-selector": r.Env.IACJobNodeSelector,

			"labels": map[string]string{
				labelVPCCreate:          "true",
				labelResourceGeneration: fmt.Sprintf("%d", obj.Generation),
			},

			"pod-annotations": fn.FilterObservabilityAnnotations(obj.GetAnnotations()),

			"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},

			"service-account-name": vpcJobServiceAccount,

			"tf-state-secret-name":      fmt.Sprintf("aws-vpc-%s", obj.Name),
			"tf-state-secret-namespace": obj.GetNamespace(),

			"aws-access-key-id":     r.Env.KlAwsAccessKey,
			"aws-secret-access-key": r.Env.KlAwsSecretKey,

			"values.json": string(valuesBytes),

			"vpc-output-secret-name":      obj.Spec.Output.Name,
			"vpc-output-secret-namespace": obj.Spec.Output.Namespace,
		})
		if err != nil {
			return fail(err).Err(nil)
		}

		rr, err := r.yamlClient.ApplyYAML(ctx, b)
		if err != nil {
			return fail(err)
		}

		req.AddToOwnedResources(rr...)
		return fail(fmt.Errorf("waiting for job to be created"))
	}

	isMyJob := job.Labels[labelResourceGeneration] == fmt.Sprintf("%d", obj.Generation) && job.Labels[labelVPCCreate] == "true"

	if !isMyJob {
		if !job_manager.HasJobFinished(ctx, r.Client, job) {
			return fail(fmt.Errorf("waiting for previous jobs to finish execution"))
		}

		if err := job_manager.DeleteJob(ctx, r.Client, job.Namespace, job.Name); err != nil {
			return fail(err)
		}
	}

	if !job_manager.HasJobFinished(ctx, r.Client, job) {
		return fail(fmt.Errorf("waiting for job to finish execution"))
	}

	// tlog := job_manager.GetTerminationLog(ctx, r.Client, job.Namespace, job.Name)
	// check.Message = tlog
	// if tlog == "" {
	// 	check.Message = "bucket creation job failed"
	// }

	if job.Status.Failed >= 1 {
		// means job failed
		return fail(fmt.Errorf("job failed, checkout logs for more details")).Err(nil)
	}

	check.Status = true
	if check != obj.Status.Checks[checkName] {
		fn.MapSet(&obj.Status.Checks, checkName, check)
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

func (r *AwsVPCReconciler) cleanupVPC(req *rApi.Request[*clustersv1.AwsVPC]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	checkName := "cleanup-vpc"

	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	fail := func(err error) stepResult.Result {
		return req.CheckFailed(checkName, check, err.Error())
	}

	job := &batchv1.Job{}
	if err := r.Get(ctx, fn.NN(obj.Namespace, getPrefixedName(obj.Name)), job); err != nil {
		job = nil
	}

	if job == nil {
		credsSecret := &corev1.Secret{}
		if err := r.Get(ctx, fn.NN(obj.Spec.CredentialsRef.Namespace, obj.Spec.CredentialsRef.Name), credsSecret); err != nil {
			return fail(err)
		}

		valuesBytes, err := genTFValues(obj, credsSecret, r.Env)
		if err != nil {
			return fail(err).Err(nil)
		}

		b, err := templates.ParseBytes(r.templateAwsVPCJob, map[string]any{
			"action": "delete",

			"job-name":          getPrefixedName(obj.Name),
			"job-namespace":     obj.Namespace,
			"job-image":         r.Env.IACJobImage,
			"job-tolerations":   r.Env.IACJobTolerations,
			"job-node-selector": r.Env.IACJobNodeSelector,

			"labels": map[string]string{
				labelVPCDelete:          "true",
				labelResourceGeneration: fmt.Sprintf("%d", obj.Generation),
			},

			"pod-annotations": fn.FilterObservabilityAnnotations(obj.GetAnnotations()),

			"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},

			"service-account-name": vpcJobServiceAccount,

			"tf-state-secret-name":      fmt.Sprintf("aws-vpc-%s", obj.Name),
			"tf-state-secret-namespace": obj.GetNamespace(),

			"aws-access-key-id":     r.Env.KlAwsAccessKey,
			"aws-secret-access-key": r.Env.KlAwsSecretKey,

			"values.json": string(valuesBytes),

			"vpc-output-secret-name":      obj.Spec.Output.Name,
			"vpc-output-secret-namespace": obj.Spec.Output.Namespace,
		})
		if err != nil {
			return fail(err).Err(nil)
		}

		rr, err := r.yamlClient.ApplyYAML(ctx, b)
		if err != nil {
			return fail(err)
		}

		req.AddToOwnedResources(rr...)
		return fail(fmt.Errorf("waiting for job to be created"))
	}

	isMyJob := job.Labels[labelResourceGeneration] == fmt.Sprintf("%d", obj.Generation) && job.Labels[labelVPCDelete] == "true"

	if !isMyJob {
		if !job_manager.HasJobFinished(ctx, r.Client, job) {
			return fail(fmt.Errorf("waiting for previous jobs to finish execution"))
		}

		if err := job_manager.DeleteJob(ctx, r.Client, job.Namespace, job.Name); err != nil {
			return fail(err)
		}

		return req.Done()
	}

	if !job_manager.HasJobFinished(ctx, r.Client, job) {
		return fail(fmt.Errorf("waiting for job to finish execution"))
	}

	// tlog := job_manager.GetTerminationLog(ctx, r.Client, job.Namespace, job.Name)
	// check.Message = tlog
	// if tlog == "" {
	// 	check.Message = "bucket creation job failed"
	// }

	if job.Status.Succeeded < 1 {
		// means job failed
		return fail(fmt.Errorf("job failed, checkout logs for more details")).Err(nil)
	}

	check.Status = true
	if check != obj.Status.Checks[checkName] {
		fn.MapSet(&obj.Status.Checks, checkName, check)
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
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

	b, err = templates.Read(templates.RBACForClusterJobTemplate)
	if err != nil {
		return err
	}
	r.templateAwsVPCJobRBAC = b

	builder := ctrl.NewControllerManagedBy(mgr).For(&clustersv1.AwsVPC{})
	builder.Owns(&batchv1.Job{})
	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
