package target

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	"github.com/kloudlite/operator/operators/clusters/internal/env"
	"github.com/kloudlite/operator/operators/clusters/internal/iac"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	job_manager "github.com/kloudlite/operator/pkg/job-helper"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	"github.com/kloudlite/operator/pkg/templates"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type ClusterReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Env        *env.Env
	logger     logging.Logger
	Name       string
	yamlClient kubectl.YAMLClient

	templateClusterApplyJob   []byte
	templateClusterDestroyJob []byte
}

func (r *ClusterReconciler) GetName() string {
	return r.Name
}

func getJobName(resourceName string) string {
	return fmt.Sprintf("iac-job-%s", resourceName)
}

func getBucketFilePath(accountName string, clusterName string) string {
	return fmt.Sprintf("kloudlite/account-%s/cluster-%s.tfstate", accountName, clusterName)
}

const (
	clusterApplyJob   = "clusterApplyJob"
	clusterDestroyJob = "clusterDestroyJob"
)

const (
	LabelClusterApplyJob    = "kloudlite.io/cluster-apply-job"
	LabelResourceGeneration = "kloudlite.io/resource-generation"
	LabelClusterDestroyJob  = "kloudlite.io/cluster-apply-job"
)

// +kubebuilder:rbac:groups=clusters.kloudlite.io,resources=clusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=clusters.kloudlite.io,resources=clusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=clusters.kloudlite.io,resources=clusters/finalizers,verbs=update

func (r *ClusterReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, request.NamespacedName, &clustersv1.Cluster{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	req.LogPreReconcile()
	defer req.LogPostReconcile()

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

	if step := r.startClusterApplyJob(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, nil
}

func (r *ClusterReconciler) finalize(req *rApi.Request[*clustersv1.Cluster]) stepResult.Result {
	_, obj, checks := req.Context(), req.Object, req.Object.Status.Checks

	checkName := "finalizing"
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	if step := r.startClusterDestroyJob(req); !step.ShouldProceed() {
		check.Status = false
		check.Message = "waiting for cluster destroy job to finish execution"
		if check != checks[checkName] {
			checks[checkName] = check
			if sr := req.UpdateStatus(); !sr.ShouldProceed() {
				return sr
			}
		}

		return step
	}

	if obj.Status.Checks[clusterDestroyJob].Status {
		return req.Finalize()
	}

	return req.Done()
}

func (r *ClusterReconciler) parseSpecToVarFileJson(obj *clustersv1.Cluster, accessKeyId string, secretAccessKey string) (string, error) {
	if obj.Spec.AWS == nil {
		return "", fmt.Errorf(".spec.aws is nil")
	}

	valuesBytes, err := json.Marshal(map[string]any{
		"aws_access_key": accessKeyId,
		"aws_secret_key": secretAccessKey,
		"aws_region":     obj.Spec.AWS.Region,
		"aws_ami":        obj.Spec.AWS.AMI,

		"aws_iam_instance_profile_role": obj.Spec.AWS.IAMInstanceProfileRole,
		"cloudflare_api_token":          r.Env.CloudflareApiToken,
		// "cloudflare_domain":             "dev3.kloudlite.io",
		"cloudflare_domain":  fmt.Sprintf("cluster-%s.account-%s.cnames.kloudlite.io", obj.Name, obj.Spec.AccountName),
		"cloudflare_zone_id": r.Env.CloudflareZoneId,

		"ec2_nodes_config": func() map[string]any {
			m := make(map[string]any, len(obj.Spec.AWS.EC2NodesConfig))
			for k, v := range obj.Spec.AWS.EC2NodesConfig {
				m[k] = map[string]any{
					"az":               v.AvailabilityZone,
					"instance_type":    v.InstanceType,
					"role":             v.Role,
					"root_volume_size": v.RootVolumeSize,
				}
			}

			return m
		}(),

		"spot_settings": map[string]any{
			"enabled": obj.Spec.AWS.SpotSettings != nil,
			"spot_fleet_tagging_role_name": func() string {
				if obj.Spec.AWS.SpotSettings != nil {
					return obj.Spec.AWS.SpotSettings.SpotFleetTaggingRoleName
				}
				return ""
			}(),
		},

		"spot_nodes_config": func() map[string]any {
			m := make(map[string]any, len(obj.Spec.AWS.SpotNodesConfig))

			for k, sn := range obj.Spec.AWS.SpotNodesConfig {
				m[k] = map[string]any{
					"vcpu": map[string]any{
						"min": sn.VCpu.Min,
						"max": sn.VCpu.Max,
					},
					"memory_per_vcpu": map[string]any{
						"min": sn.MemPerVCpu.Min,
						"max": sn.MemPerVCpu.Max,
					},
					"root_volume_size": sn.RootVolumeSize,
					"allow_public_ip":  false,
				}
			}

			return m
		}(),

		"disable_ssh": obj.Spec.DisableSSH,
	})

	if err != nil {
		return "", err
	}
	return string(valuesBytes), nil
}

func (r *ClusterReconciler) findAccountS3Bucket(ctx context.Context, obj *clustersv1.Cluster) (*clustersv1.AccountS3Bucket, error) {
	var bucketList clustersv1.AccountS3BucketList
	if err := r.List(ctx, &bucketList, client.InNamespace(obj.Namespace)); err != nil {
		return nil, err
	}
	if len(bucketList.Items) == 0 {
		// TODO: create account-s3-bucket
		s3Bucket := &clustersv1.AccountS3Bucket{ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("kl-%s", obj.Spec.AccountId), Namespace: obj.Namespace}}
		if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, s3Bucket, func() error {
			s3Bucket.Spec = clustersv1.AccountS3BucketSpec{
				AccountName:  obj.Spec.AccountName,
				BucketRegion: r.Env.KlS3BucketRegion,
			}
			return nil
		}); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("waiting for account-s3-bucket to reconcile")
		// return nil, fmt.Errorf("no account-s3-bucket found in namespace %s", namespace)
	}

	if len(bucketList.Items) != 1 {
		return nil, fmt.Errorf("multiple account-s3-bucket found in namespace %s", obj.Namespace)
	}

	if !bucketList.Items[0].Status.IsReady {
		return nil, fmt.Errorf("bucket %s (in region: %s), is not ready yet", bucketList.Items[0].Name, bucketList.Items[0].Spec.BucketRegion)
	}

	return &bucketList.Items[0], nil
}

func (r *ClusterReconciler) startClusterApplyJob(req *rApi.Request[*clustersv1.Cluster]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(clusterApplyJob)
	defer req.LogPostCheck(clusterApplyJob)

	bucket, err := r.findAccountS3Bucket(ctx, obj)
	if err != nil {
		return req.CheckFailed(clusterApplyJob, check, err.Error())
	}

	job := &batchv1.Job{}
	if err := r.Get(ctx, fn.NN(obj.Namespace, getJobName(obj.Name)), job); err != nil {
		job = nil
	}

	if job == nil {
		credsSecret := &corev1.Secret{}
		if err := r.Get(ctx, fn.NN(obj.Spec.CredentialsRef.Namespace, obj.Spec.CredentialsRef.Name), credsSecret); err != nil {
			return req.CheckFailed(clusterApplyJob, check, err.Error())
		}

		valuesJson, err := r.parseSpecToVarFileJson(obj, string(credsSecret.Data["accessKey"]), string(credsSecret.Data["accessSecret"]))
		if err != nil {
			return req.CheckFailed(clusterApplyJob, check, err.Error())
		}

		b, err := templates.ParseBytes(r.templateClusterApplyJob, map[string]any{
			"job-name":      getJobName(obj.Name),
			"job-namespace": obj.Namespace,
			"labels": map[string]string{
				LabelClusterApplyJob:    "true",
				LabelResourceGeneration: fmt.Sprintf("%d", obj.Generation),
			},
			"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},

			"aws-s3-bucket-name":     bucket.Name,
			"aws-s3-bucket-region":   bucket.Spec.BucketRegion,
			"aws-s3-bucket-filepath": getBucketFilePath(obj.Spec.AccountName, obj.Name),

			"aws-access-key-id":     string(credsSecret.Data["accessKey"]),
			"aws-secret-access-key": string(credsSecret.Data["accessSecret"]),

			"values.json": string(valuesJson),
		})
		if err != nil {
			return req.CheckFailed(clusterApplyJob, check, err.Error()).Err(nil)
		}

		rr, err := r.yamlClient.ApplyYAML(ctx, b)
		if err != nil {
			return req.CheckFailed(clusterApplyJob, check, err.Error()).Err(nil)
		}
		req.AddToOwnedResources(rr...)
		req.Logger.Infof("waiting for job to be created")
		return req.Done().RequeueAfter(1 * time.Second)
	}

	isMyJob := job.Labels[LabelResourceGeneration] == fmt.Sprintf("%d", obj.Generation) && job.Labels[LabelClusterApplyJob] == "true"

	if !isMyJob {
		if !job_manager.HasJobFinished(ctx, r.Client, job) {
			return req.CheckFailed(clusterApplyJob, check, fmt.Sprintf("waiting for previous jobs to finish execution"))
		}

		if err := job_manager.DeleteJob(ctx, r.Client, job.Namespace, job.Name); err != nil {
			return req.CheckFailed(clusterApplyJob, check, err.Error())
		}

		return req.Done().RequeueAfter(1 * time.Second)
	}

	if !job_manager.HasJobFinished(ctx, r.Client, job) {
		return req.CheckFailed(clusterApplyJob, check, "waiting for job to finish execution")
	}

	check.Message = job_manager.GetTerminationLog(ctx, r.Client, job.Namespace, job.Name)
	check.Status = job.Status.Succeeded > 0
	if check != obj.Status.Checks[clusterApplyJob] {
		obj.Status.Checks[clusterApplyJob] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

func (r *ClusterReconciler) startClusterDestroyJob(req *rApi.Request[*clustersv1.Cluster]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(clusterDestroyJob)
	defer req.LogPostCheck(clusterDestroyJob)

	bucket, err := r.findAccountS3Bucket(ctx, obj)
	if err != nil {
		return req.CheckFailed(clusterDestroyJob, check, err.Error())
	}

	job := &batchv1.Job{}
	if err := r.Get(ctx, fn.NN(obj.Namespace, getJobName(obj.Name)), job); err != nil {
		job = nil
	}

	if job == nil {
		credsSecret := &corev1.Secret{}
		if err := r.Get(ctx, fn.NN(obj.Spec.CredentialsRef.Namespace, obj.Spec.CredentialsRef.Name), credsSecret); err != nil {
			return req.CheckFailed(clusterApplyJob, check, err.Error()).Err(nil)
		}

		valuesJson, err := r.parseSpecToVarFileJson(obj, string(credsSecret.Data["accessKey"]), string(credsSecret.Data["accessSecret"]))
		if err != nil {
			return req.CheckFailed(clusterApplyJob, check, err.Error())
		}

		b, err := templates.ParseBytes(r.templateClusterDestroyJob, map[string]any{
			"job-name":      getJobName(obj.Name),
			"job-namespace": obj.Namespace,
			"labels": map[string]string{
				LabelClusterDestroyJob:  "true",
				LabelResourceGeneration: fmt.Sprintf("%d", obj.Generation),
			},
			"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},

			"aws-s3-bucket-name":     bucket.Name,
			"aws-s3-bucket-region":   bucket.Spec.BucketRegion,
			"aws-s3-bucket-filepath": getBucketFilePath(obj.Spec.AccountName, obj.Name),

			"aws-access-key-id":     string(credsSecret.Data["accessKey"]),
			"aws-secret-access-key": string(credsSecret.Data["accessSecret"]),

			"values.json": string(valuesJson),
		})
		if err != nil {
			return req.CheckFailed(clusterDestroyJob, check, err.Error()).Err(nil)
		}

		rr, err := r.yamlClient.ApplyYAML(ctx, b)
		if err != nil {
			return req.CheckFailed(clusterDestroyJob, check, err.Error()).Err(nil)
		}
		req.AddToOwnedResources(rr...)
		req.Logger.Infof("waiting for job to be created")
		return req.Done().RequeueAfter(1 * time.Second)
	}

	isMyJob := job.Labels[LabelResourceGeneration] == fmt.Sprintf("%d", obj.Generation) && job.Labels[LabelClusterDestroyJob] == "true"

	if !isMyJob {
		if !job_manager.HasJobFinished(ctx, r.Client, job) {
			return req.CheckFailed(clusterDestroyJob, check, fmt.Sprintf("waiting for previous jobs to finish execution"))
		}

		if err := job_manager.DeleteJob(ctx, r.Client, job.Namespace, job.Name); err != nil {
			return req.CheckFailed(clusterDestroyJob, check, err.Error())
		}

		return req.Done().RequeueAfter(1 * time.Second)
	}

	if !job_manager.HasJobFinished(ctx, r.Client, job) {
		return req.CheckFailed(clusterDestroyJob, check, "waiting for job to finish execution")
	}

	check.Message = job_manager.GetTerminationLog(ctx, r.Client, job.Namespace, job.Name)
	check.Status = job.Status.Succeeded > 0
	if check != obj.Status.Checks[clusterDestroyJob] {
		obj.Status.Checks[clusterDestroyJob] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

	var err error
	r.templateClusterApplyJob, err = iac.TemplatesDir.ReadFile("templates/cluster-plan-and-apply-job.yml.tpl")
	if err != nil {
		return err
	}

	r.templateClusterDestroyJob, err = iac.TemplatesDir.ReadFile("templates/cluster-destroy-job.yml.tpl")
	if err != nil {
		return err
	}

	builder := ctrl.NewControllerManagedBy(mgr).For(&clustersv1.Cluster{})
	builder.Owns(&batchv1.Job{})
	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	return builder.Complete(r)
}
