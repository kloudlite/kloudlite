package account_s3_bucket

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
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
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Env    *env.Env
	logger logging.Logger
	Name   string

	yamlClient kubectl.YAMLClient

	templateS3BucketJob []byte
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	BucketCreateJob  = "bucket-create-job"
	BucketDestroyJob = "bucket-destroy-job"
)

const (
	LabelBucketCreateJob    = "kloudlite.io/bucket-create-job"
	LabelBucketDestroyJob   = "kloudlite.io/bucket-destroy-job"
	LabelResourceGeneration = "kloudlite.io/resource-generation"
)

func getJobName(resourceName string) string {
	return fmt.Sprintf("%s-s3", resourceName)
}

// +kubebuilder:rbac:groups=clusters.kloudlite.io,resources=clusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=clusters.kloudlite.io,resources=clusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=clusters.kloudlite.io,resources=clusters/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, request.NamespacedName, &clustersv1.AccountS3Bucket{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	req.PreReconcile()
	defer req.PostReconcile()

	// ensure there can be only one instance of this resource in a namespace
	var bucketsList clustersv1.AccountS3BucketList
	if err := r.List(ctx, &bucketsList, client.InNamespace(req.Object.Namespace)); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if len(bucketsList.Items) > 1 {
		return ctrl.Result{}, fmt.Errorf("there can be only one instance of AccountS3Bucket in a namespace")
	}

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

	if step := r.StartBucketCreateJob(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *Reconciler) finalize(req *rApi.Request[*clustersv1.AccountS3Bucket]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	checkName := "finalizing"

	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	if step := r.StartBucketDestroyJob(req); !step.ShouldProceed() {
		return step
	}

	if !obj.Status.Checks[BucketDestroyJob].Status {
		return req.Finalize()
	}

	check.Message = job_manager.GetTerminationLog(ctx, r.Client, obj.Namespace, getJobName(obj.Name))
	check.Status = false

	req.Logger.Infof("finalizing failed for account-s3-bucket %s", obj.Name)

	if check != obj.Status.Checks[checkName] {
		obj.Status.Checks[checkName] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Done()
}

func (r *Reconciler) StartBucketCreateJob(req *rApi.Request[*clustersv1.AccountS3Bucket]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(BucketCreateJob)
	defer req.LogPostCheck(BucketCreateJob)

	job := &batchv1.Job{}
	if err := r.Get(ctx, fn.NN(obj.Namespace, getJobName(obj.Name)), job); err != nil {
		job = nil
	}

	if job == nil {
		credsSecret := &corev1.Secret{}
		if err := r.Get(ctx, fn.NN(obj.Spec.CredentialsRef.Namespace, obj.Spec.CredentialsRef.Name), credsSecret); err != nil {
			return req.CheckFailed(BucketCreateJob, check, err.Error()).Err(nil)
		}

		valuesBytes, err := json.Marshal(map[string]any{
			"aws_access_key": r.Env.KlAwsAccessKey,
			"aws_secret_key": r.Env.KlAwsSecretKey,
			"aws_assume_role": map[string]any{
				"enabled":     true,
				"role_arn":    string(credsSecret.Data[obj.Spec.CredentialKeys.KeyAWSAssumeRoleRoleARN]),
				"external_id": string(credsSecret.Data[obj.Spec.CredentialKeys.KeyAWSAssumeRoleExternalID]),
			},
			"aws_region":  r.Env.KlS3BucketRegion,
			"bucket_name": obj.Name,
			"tracker_id":  obj.Name,
		})
		if err != nil {
			return req.CheckFailed(BucketCreateJob, check, err.Error()).Err(nil)
		}

		b, err := templates.ParseBytes(r.templateS3BucketJob, map[string]any{
			"action": "apply",

			"job-name":      getJobName(obj.Name),
			"job-namespace": obj.Namespace,

			"labels": map[string]string{
				LabelBucketCreateJob:    "true",
				LabelResourceGeneration: fmt.Sprintf("%d", obj.Generation),
			},
			"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},

			"aws-s3-bucket-name":     r.Env.KlS3BucketName,
			"aws-s3-bucket-region":   r.Env.KlS3BucketRegion,
			"aws-s3-bucket-filepath": fmt.Sprintf("kloudlite/accounts/%s/%s", obj.Spec.AccountName, obj.Name),

			"aws-access-key-id":     r.Env.KlAwsAccessKey,
			"aws-secret-access-key": r.Env.KlAwsSecretKey,

			"values.json": string(valuesBytes),

			"job-image": r.Env.IACJobImage,
		})
		if err != nil {
			return req.CheckFailed(BucketCreateJob, check, err.Error()).Err(nil)
		}

		rr, err := r.yamlClient.ApplyYAML(ctx, b)
		if err != nil {
			return req.CheckFailed(BucketCreateJob, check, err.Error())
		}

		req.AddToOwnedResources(rr...)
		return req.CheckFailed(BucketCreateJob, check, "waiting for job to be created")
	}

	isMyJob := job.Labels[LabelResourceGeneration] == fmt.Sprintf("%d", obj.Generation) && job.Labels[LabelBucketCreateJob] == "true"

	if !isMyJob {
		if !job_manager.HasJobFinished(ctx, r.Client, job) {
			return req.CheckFailed(BucketCreateJob, check, "waiting for previous jobs to finish execution")
		}

		if err := job_manager.DeleteJob(ctx, r.Client, job.Namespace, job.Name); err != nil {
			return req.CheckFailed(BucketCreateJob, check, err.Error())
		}
	}

	if !job_manager.HasJobFinished(ctx, r.Client, job) {
		return req.CheckFailed(BucketCreateJob, check, "waiting for job to finish execution")
	}

	tlog := job_manager.GetTerminationLog(ctx, r.Client, job.Namespace, job.Name)
	check.Message = tlog
	if tlog == "" {
		check.Message = "bucket creation job failed"
	}

	check.Status = job.Status.Succeeded > 0
	if check != obj.Status.Checks[BucketCreateJob] {
		obj.Status.Checks[BucketCreateJob] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	if !check.Status {
		req.Logger.Infof("bucket creation failed for account-s3-bucket %s", obj.Name)
		return req.Done()
	}

	return req.Next()
}

func (r *Reconciler) StartBucketDestroyJob(req *rApi.Request[*clustersv1.AccountS3Bucket]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(BucketDestroyJob)
	defer req.LogPostCheck(BucketDestroyJob)

	job := &batchv1.Job{}
	if err := r.Get(ctx, fn.NN(obj.Namespace, getJobName(obj.Name)), job); err != nil {
		job = nil
	}

	if job == nil {
		credsSecret := &corev1.Secret{}
		if err := r.Get(ctx, fn.NN(obj.Spec.CredentialsRef.Namespace, obj.Spec.CredentialsRef.Name), credsSecret); err != nil {
			return req.CheckFailed(BucketCreateJob, check, err.Error())
		}

		valuesBytes, err := json.Marshal(map[string]any{
			"aws_access_key": r.Env.KlAwsAccessKey,
			"aws_secret_key": r.Env.KlAwsSecretKey,
			"aws_assume_role": map[string]any{
				"enabled":     true,
				"role_arn":    string(credsSecret.Data[obj.Spec.CredentialKeys.KeyAWSAssumeRoleRoleARN]),
				"external_id": string(credsSecret.Data[obj.Spec.CredentialKeys.KeyAWSAssumeRoleExternalID]),
			},
			"aws_region":  r.Env.KlS3BucketRegion,
			"bucket_name": obj.Name,
			"tracker_id":  obj.Name,
		})
		if err != nil {
			return req.CheckFailed(BucketDestroyJob, check, err.Error())
		}

		b, err := templates.ParseBytes(r.templateS3BucketJob, map[string]any{
			"action":        "delete",
			"job-name":      getJobName(obj.Name),
			"job-namespace": obj.Namespace,
			"labels": map[string]string{
				LabelBucketDestroyJob:   "true",
				LabelResourceGeneration: fmt.Sprintf("%d", obj.Generation),
			},
			"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},

			"aws-s3-bucket-name":     r.Env.KlS3BucketName,
			"aws-s3-bucket-region":   r.Env.KlS3BucketRegion,
			"aws-s3-bucket-filepath": fmt.Sprintf("kloudlite/accounts/%s/%s", obj.Spec.AccountName, obj.Name),

			"aws-access-key-id":     r.Env.KlAwsAccessKey,
			"aws-secret-access-key": r.Env.KlAwsSecretKey,

			"values.json": string(valuesBytes),

			"job-image": r.Env.IACJobImage,
		})
		if err != nil {
			return req.CheckFailed(BucketDestroyJob, check, err.Error()).Err(nil)
		}

		rr, err := r.yamlClient.ApplyYAML(ctx, b)
		if err != nil {
			return req.CheckFailed(BucketDestroyJob, check, err.Error())
		}
		req.AddToOwnedResources(rr...)
		return req.Done().RequeueAfter(1 * time.Second)
	}

	isMyJob := job.Labels[LabelResourceGeneration] == fmt.Sprintf("%d", obj.Generation) && job.Labels[LabelBucketDestroyJob] == "true"
	if !isMyJob {
		// wait for completion
		if !job_manager.HasJobFinished(ctx, r.Client, job) {
			return req.CheckFailed(BucketDestroyJob, check, "waiting for previous jobs to finish execution")
		}
		if err := r.Delete(ctx, job); err != nil {
			if !apiErrors.IsNotFound(err) {
				return req.CheckFailed(BucketDestroyJob, check, err.Error())
			}
		}
	}

	if !job_manager.HasJobFinished(ctx, r.Client, job) {
		return req.CheckFailed(BucketDestroyJob, check, "waiting for job to finish execution")
	}

	check.Message = job_manager.GetTerminationLog(ctx, r.Client, job.Namespace, job.Name)
	check.Status = job.Status.Succeeded > 0
	if check != obj.Status.Checks[BucketDestroyJob] {
		obj.Status.Checks[BucketDestroyJob] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig(), kubectl.YAMLClientOpts{Logger: r.logger})

	var err error
	r.templateS3BucketJob, err = templates.Read(templates.S3BucketJobTemplate)
	if err != nil {
		return err
	}

	builder := ctrl.NewControllerManagedBy(mgr).For(&clustersv1.AccountS3Bucket{})
	builder.Owns(&batchv1.Job{})
	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
