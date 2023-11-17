package target

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	ct "github.com/kloudlite/operator/apis/common-types"
	redpandav1 "github.com/kloudlite/operator/apis/redpanda.msvc/v1"

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

	templateClusterJob        []byte
	templateRBACForClusterJob []byte
}

func (r *ClusterReconciler) GetName() string {
	return r.Name
}

func getJobName(resourceName string) string {
	return fmt.Sprintf("iac-cluster-job-%s", resourceName)
}

func getBucketFilePath(accountName string, clusterName string) string {
	return fmt.Sprintf("kloudlite/account-%s/cluster-%s.tfstate", accountName, clusterName)
}

const (
	clusterApplyJob   = "clusterApplyJob"
	clusterDestroyJob = "clusterDestroyJob"
	messageQueueTopic = "messageQueueTopic"
	jobRbac           = "job-rbac"
)

const (
	LabelClusterApplyJob    = "kloudlite.io/cluster-apply-job"
	LabelResourceGeneration = "kloudlite.io/resource-generation"
	LabelClusterDestroyJob  = "kloudlite.io/cluster-apply-job"
)

const (
	clusterJobServiceAccount = "cluster-job-sa"
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

	if step := r.ensureJobRBAC(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureMessageQueueTopic(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.startClusterApplyJob(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *ClusterReconciler) patchDefaults(req *rApi.Request[*clustersv1.Cluster]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	checkName := "defaults"

	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	hasUpdated := false

	if obj.Spec.Output == nil {
		hasUpdated = true
		obj.Spec.Output = &clustersv1.ClusterOutput{
			SecretName:            fmt.Sprintf("clusters-%s-credentials", obj.Name),
			KeyKubeconfig:         "kubeconfig",
			KeyK3sServerJoinToken: "k3s_server_token",
			KeyK3sAgentJoinToken:  "k3s_agent_token",
		}
	}

	if hasUpdated {
		if err := r.Update(ctx, obj); err != nil {
			return req.CheckFailed(checkName, check, err.Error())
		}
		return req.Done().RequeueAfter(500 * time.Millisecond)
	}

	check.Status = true
	if check != obj.Status.Checks[checkName] {
		obj.Status.Checks[checkName] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

func (r *ClusterReconciler) finalize(req *rApi.Request[*clustersv1.Cluster]) stepResult.Result {
	_, obj := req.Context(), req.Object

	checkName := "finalizing"
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	if step := r.startClusterDestroyJob(req); !step.ShouldProceed() {
		check.Status = false
		check.Message = "cluster job failed"
		if check != obj.Status.Checks[checkName] {
			if obj.Status.Checks == nil {
				obj.Status.Checks = map[string]rApi.Check{}
			}
			obj.Status.Checks[checkName] = check
			if sr := req.UpdateStatus(); !sr.ShouldProceed() {
				return sr
			}
		}

		return req.Done().Err(nil)
	}

	if obj.Status.Checks[clusterDestroyJob].Status {
		return req.Finalize()
	}

	return req.Done()
}

func (r *ClusterReconciler) ensureMessageQueueTopic(req *rApi.Request[*clustersv1.Cluster]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(messageQueueTopic)
	defer req.LogPostCheck(messageQueueTopic)

	qtopic := &redpandav1.Topic{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.MessageQueueTopicName}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, qtopic, func() error {
		if qtopic.Labels == nil {
			qtopic.Labels = make(map[string]string, 2)
		}
		qtopic.Labels[constants.AccountNameKey] = obj.Spec.AccountName
		qtopic.Labels[constants.ClusterNameKey] = obj.Name

		if qtopic.Annotations == nil {
			qtopic.Annotations = make(map[string]string, 1)
		}
		qtopic.Annotations[constants.DescriptionKey] = "kloudlite cluster incoming message queue topic"
		qtopic.Spec.PartitionCount = 3
		return nil
	}); err != nil {
		return req.CheckFailed(messageQueueTopic, check, err.Error())
	}

	req.AddToOwnedResources(rApi.ParseResourceRef(qtopic))

	check.Status = qtopic.Status.IsReady
	if check != obj.Status.Checks[messageQueueTopic] {
		obj.Status.Checks[messageQueueTopic] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	if !check.Status {
		return req.CheckFailed(messageQueueTopic, check, "waiting for message queue topic to be ready")
	}
	return req.Next()
}

func (r *ClusterReconciler) ensureJobRBAC(req *rApi.Request[*clustersv1.Cluster]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(jobRbac)
	defer req.LogPostCheck(jobRbac)

	b, err := templates.ParseBytes(r.templateRBACForClusterJob, map[string]any{
		"service-account-name": clusterJobServiceAccount,
		"namespace":            obj.Namespace,
	})
	if err != nil {
		return req.CheckFailed(jobRbac, check, err.Error()).Err(nil)
	}

	rr, err := r.yamlClient.ApplyYAML(ctx, b)
	if err != nil {
		return req.CheckFailed(jobRbac, check, err.Error()).Err(nil)
	}
	req.AddToOwnedResources(rr...)

	check.Status = true
	if check != obj.Status.Checks[jobRbac] {
		obj.Status.Checks[jobRbac] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

func (r *ClusterReconciler) parseSpecToVarFileJson(obj *clustersv1.Cluster, providerCreds *corev1.Secret) (string, error) {
	if providerCreds == nil {
		return "", fmt.Errorf("providerCreds is nil")
	}

	clusterTokenScrt := &corev1.Secret{}
	if err := r.Get(context.TODO(), fn.NN(obj.Namespace, obj.Spec.ClusterTokenRef.Name), clusterTokenScrt); err != nil {
		clusterTokenScrt = nil
		return "", err
	}

	switch obj.Spec.CloudProvider {
	case ct.CloudProviderAWS:
		{
			if obj.Spec.AWS == nil {
				return "", fmt.Errorf("when cloudprovider is set to aws, aws config must be provided")
			}

			isAssumeRole := providerCreds.Data[obj.Spec.CredentialKeys.KeyAccessKey] == nil || providerCreds.Data[obj.Spec.CredentialKeys.KeySecretKey] == nil

			valuesBytes, err := json.Marshal(map[string]any{
				"aws_access_key": func() string {
					if !isAssumeRole {
						return string(providerCreds.Data[obj.Spec.CredentialKeys.KeyAccessKey])
					}
					return r.Env.KlAwsAccessKey
				}(),
				"aws_secret_key": func() string {
					if !isAssumeRole {
						return string(providerCreds.Data[obj.Spec.CredentialKeys.KeySecretKey])
					}
					return r.Env.KlAwsSecretKey
				}(),
				"aws_assume_role": func() map[string]any {
					if !isAssumeRole {
						return nil
					}
					return map[string]any{
						"enabled": true,
						// "role_arn":    fmt.Sprintf(r.Env.AWSAssumeTenantRoleFormatString, string(providerCreds.Data[obj.Spec.CredentialKeys.KeyAWSAccountId])),
						"role_arn":    string(providerCreds.Data[obj.Spec.CredentialKeys.KeyAWSAssumeRoleRoleARN]),
						"external_id": string(providerCreds.Data[obj.Spec.CredentialKeys.KeyAWSAssumeRoleExternalID]),
					}
				}(),
				"aws_region": obj.Spec.AWS.Region,

				"tracker_id":                fmt.Sprintf("cluster-%s", obj.Name),
				"enable_nvidia_gpu_support": true,

				"k3s_masters": map[string]any{
					"image_id":           obj.Spec.AWS.K3sMasters.ImageId,
					"image_ssh_username": obj.Spec.AWS.K3sMasters.ImageSSHUsername,
					"instance_type":      obj.Spec.AWS.K3sMasters.InstanceType,
					"nvidia_gpu_enabled": obj.Spec.AWS.K3sMasters.NvidiaGpuEnabled,

					"root_volume_type":          obj.Spec.AWS.K3sMasters.RootVolumeType,
					"root_volume_size":          obj.Spec.AWS.K3sMasters.RootVolumeSize,
					"iam_instance_profile":      obj.Spec.AWS.K3sMasters.IAMInstanceProfileRole,
					"public_dns_host":           obj.Spec.PublicDNSHost,
					"cluster_internal_dns_host": obj.Spec.ClusterInternalDnsHost,
					"cloudflare": map[string]any{
						"enabled":   true,
						"api_token": r.Env.CloudflareApiToken,
						"zone_id":   r.Env.CloudflareZoneId,
						"domain":    obj.Spec.PublicDNSHost,
					},
					"taint_master_nodes": obj.Spec.TaintMasterNodes,
					"backup_to_s3": map[string]any{
						"enabled": obj.Spec.BackupToS3Enabled,
					},
					"nodes": func() map[string]any {
						nodes := make(map[string]any, len(obj.Spec.AWS.K3sMasters.Nodes))
						for k, v := range obj.Spec.AWS.K3sMasters.Nodes {
							nodes[k] = map[string]any{
								"role":              v.Role,
								"availability_zone": v.AvaialbilityZone,
								"last_recreated_at": v.LastRecreatedAt,
							}
						}
						return nodes
					}(),
				},

				"kloudlite_params": map[string]any{
					"release":            obj.Spec.KloudliteRelease,
					"install_crds":       true,
					"install_csi_driver": true,
					"install_operators":  true,
					"install_agent":      true,
					"agent_vars": map[string]any{
						"account_name":             obj.Spec.AccountName,
						"cluster_name":             obj.Name,
						"cluster_token":            string(clusterTokenScrt.Data[obj.Spec.ClusterTokenRef.Key]),
						"message_office_grpc_addr": r.Env.MessageOfficeGRPCAddr,
					},
				},
			})
			if err != nil {
				return "", err
			}
			return string(valuesBytes), nil
		}
	default:
		return "", fmt.Errorf("unknown cloud provider %s", obj.Spec.CloudProvider)
	}
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
				AccountName:    obj.Spec.AccountName,
				BucketRegion:   r.Env.KlS3BucketRegion,
				CredentialsRef: obj.Spec.CredentialsRef,
				CredentialKeys: obj.Spec.CredentialKeys,
			}
			return nil
		}); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("waiting for account-s3-bucket to reconcile")
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

	_ = bucket

	job := &batchv1.Job{}
	if err := r.Get(ctx, fn.NN(obj.Namespace, getJobName(obj.Name)), job); err != nil {
		job = nil
	}

	if job == nil {
		credsSecret := &corev1.Secret{}
		if err := r.Get(ctx, fn.NN(obj.Spec.CredentialsRef.Namespace, obj.Spec.CredentialsRef.Name), credsSecret); err != nil {
			return req.CheckFailed(clusterApplyJob, check, err.Error())
		}

		valuesJson, err := r.parseSpecToVarFileJson(obj, credsSecret)
		if err != nil {
			return req.CheckFailed(clusterApplyJob, check, err.Error()).Err(nil)
		}

		b, err := templates.ParseBytes(r.templateClusterJob, map[string]any{
			"action":        "apply",
			"job-name":      getJobName(obj.Name),
			"job-namespace": obj.Namespace,
			"labels": map[string]string{
				LabelClusterApplyJob:    "true",
				LabelResourceGeneration: fmt.Sprintf("%d", obj.Generation),
			},
			"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},

			"service-account-name": clusterJobServiceAccount,

			"kubeconfig-secret-name":      obj.Spec.Output.SecretName,
			"kubeconfig-secret-namespace": obj.Namespace,
			"kubeconfig-secret-annotations": map[string]string{
				constants.DescriptionKey: fmt.Sprintf("kubeconfig for cluster %s", obj.Name),
			},

			// TODO: move this to our own bucket, as we are creating their masters
			// "aws-s3-bucket-name":     bucket.Name,
			// "aws-s3-bucket-region":   bucket.Spec.BucketRegion,
			// "aws-s3-bucket-filepath": getBucketFilePath(obj.Spec.AccountName, obj.Name),
			//
			// "aws-access-key-id":     string(credsSecret.Data["accessKey"]),
			// "aws-secret-access-key": string(credsSecret.Data["accessSecret"]),

			"aws-s3-bucket-name":     r.Env.KlS3BucketName,
			"aws-s3-bucket-region":   r.Env.KlS3BucketRegion,
			"aws-s3-bucket-filepath": getBucketFilePath(obj.Spec.AccountName, obj.Name),

			"aws-access-key-id":     r.Env.KlAwsAccessKey,
			"aws-secret-access-key": r.Env.KlAwsSecretKey,

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
			return req.CheckFailed(clusterApplyJob, check, "waiting for previous jobs to finish execution")
		}

		if err := job_manager.DeleteJob(ctx, r.Client, job.Namespace, job.Name); err != nil {
			return req.CheckFailed(clusterApplyJob, check, err.Error())
		}

		return req.Done().RequeueAfter(1 * time.Second)
	}

	if !job_manager.HasJobFinished(ctx, r.Client, job) {
		return req.CheckFailed(clusterApplyJob, check, "waiting for job to finish execution").Err(nil)
	}

	check.Message = job_manager.GetTerminationLog(ctx, r.Client, job.Namespace, job.Name)
	check.Status = job.Status.Succeeded > 0
	if check != obj.Status.Checks[clusterApplyJob] {
		obj.Status.Checks[clusterApplyJob] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	if !check.Status {
		req.Logger.Infof("job failed")
		return req.Done()
	}

	return req.Next()
}

func (r *ClusterReconciler) startClusterDestroyJob(req *rApi.Request[*clustersv1.Cluster]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(clusterDestroyJob)
	defer req.LogPostCheck(clusterDestroyJob)

	job := &batchv1.Job{}
	if err := r.Get(ctx, fn.NN(obj.Namespace, getJobName(obj.Name)), job); err != nil {
		job = nil
	}

	if job == nil {
		credsSecret := &corev1.Secret{}
		if err := r.Get(ctx, fn.NN(obj.Spec.CredentialsRef.Namespace, obj.Spec.CredentialsRef.Name), credsSecret); err != nil {
			return req.CheckFailed(clusterDestroyJob, check, err.Error()).Err(nil)
		}

		valuesJson, err := r.parseSpecToVarFileJson(obj, credsSecret)
		if err != nil {
			return req.CheckFailed(clusterDestroyJob, check, err.Error())
		}

		b, err := templates.ParseBytes(r.templateClusterJob, map[string]any{
			"action":        "delete",
			"job-name":      getJobName(obj.Name),
			"job-namespace": obj.Namespace,
			"labels": map[string]string{
				LabelClusterDestroyJob:  "true",
				LabelResourceGeneration: fmt.Sprintf("%d", obj.Generation),
			},
			"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},

			"service-account-name": clusterJobServiceAccount,

			"kubeconfig-secret-name":      fmt.Sprintf("cluster-%s-kubeconfig", obj.Name),
			"kubeconfig-secret-namespace": obj.Namespace,

			"aws-s3-bucket-name":     r.Env.KlS3BucketName,
			"aws-s3-bucket-region":   r.Env.KlS3BucketRegion,
			"aws-s3-bucket-filepath": getBucketFilePath(obj.Spec.AccountName, obj.Name),

			"aws-access-key-id":     r.Env.KlAwsAccessKey,
			"aws-secret-access-key": r.Env.KlAwsSecretKey,

			// "aws-s3-bucket-name":     bucket.Name,
			// "aws-s3-bucket-region":   bucket.Spec.BucketRegion,
			// "aws-s3-bucket-filepath": getBucketFilePath(obj.Spec.AccountName, obj.Name),
			//
			// "aws-access-key-id":     string(credsSecret.Data["accessKey"]),
			// "aws-secret-access-key": string(credsSecret.Data["accessSecret"]),

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
			return req.CheckFailed(clusterDestroyJob, check, "waiting for previous jobs to finish execution")
		}

		if err := job_manager.DeleteJob(ctx, r.Client, job.Namespace, job.Name); err != nil {
			return req.CheckFailed(clusterDestroyJob, check, err.Error())
		}

		return req.Done().RequeueAfter(1 * time.Second)
	}

	if !job_manager.HasJobFinished(ctx, r.Client, job) {
		return req.CheckFailed(clusterDestroyJob, check, "waiting for job to finish execution").Err(nil)
	}

	check.Message = job_manager.GetTerminationLog(ctx, r.Client, job.Namespace, job.Name)
	check.Status = job.Status.Succeeded > 0
	if check != obj.Status.Checks[clusterDestroyJob] {
		obj.Status.Checks[clusterDestroyJob] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	if !check.Status {
		return req.Done()
	}

	return req.Next()
}

func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

	var err error
	r.templateClusterJob, err = templates.Read(templates.ClusterJobTemplate)
	if err != nil {
		return err
	}

	r.templateRBACForClusterJob, err = templates.Read(templates.RBACForClusterJobTemplate)
	if err != nil {
		return err
	}

	builder := ctrl.NewControllerManagedBy(mgr).For(&clustersv1.Cluster{})
	builder.Owns(&batchv1.Job{})
	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
