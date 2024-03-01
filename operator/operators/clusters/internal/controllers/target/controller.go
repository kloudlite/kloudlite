package target

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	common_types "github.com/kloudlite/operator/apis/common-types"
	ct "github.com/kloudlite/operator/apis/common-types"

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
	apiLabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
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

	NotifyOnClusterUpdate func(ctx context.Context, obj *clustersv1.Cluster) error
}

func (r *ClusterReconciler) GetName() string {
	return r.Name
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
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &clustersv1.Cluster{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// if req.Object.Namespace != "kl-account-dev-team" || req.Object.Name != "testing-nxtcoder17" {
	// 	return ctrl.Result{}, nil
	// }

	req.PreReconcile()
	defer req.PostReconcile()

	notifyAndExit := func(step stepResult.Result) (ctrl.Result, error) {
		if err := r.NotifyOnClusterUpdate(ctx, req.Object); err != nil {
			return ctrl.Result{}, err
		}
		return step.ReconcilerResponse()
	}

	if step := r.patchDefaults(req); !step.ShouldProceed() {
		return notifyAndExit(step)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return notifyAndExit(x)
		}
		return ctrl.Result{}, nil
	}

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return notifyAndExit(step)
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return notifyAndExit(step)
	}

	if step := req.EnsureFinalizers(constants.CommonFinalizer); !step.ShouldProceed() {
		return notifyAndExit(step)
	}

	if step := r.ensureJobRBAC(req); !step.ShouldProceed() {
		return notifyAndExit(step)
	}

	if step := r.ensureCloudproviderStuffs(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.startClusterApplyJob(req); !step.ShouldProceed() {
		return notifyAndExit(step)
	}

	req.Object.Status.IsReady = true
	if err := r.NotifyOnClusterUpdate(ctx, req.Object); err != nil {
		return ctrl.Result{}, err
	}
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
			JobName:      fmt.Sprintf("iac-cluster-job-%s", obj.Name),
			JobNamespace: obj.Namespace,

			SecretName: fmt.Sprintf("clusters-%s-credentials", obj.Name),

			KeyKubeconfig:          "kubeconfig",
			KeyK3sServerJoinToken:  "k3s_server_token",
			KeyK3sAgentJoinToken:   "k3s_agent_token",
			KeyAWSVPCId:            "aws_vpc_id",
			KeyAWSVPCPublicSubnets: "aws_vpc_public_subnets",
		}
	}

	if obj.Spec.Output.JobName == "" {
		hasUpdated = true
		obj.Spec.Output.JobName = fmt.Sprintf("iac-cluster-job-%s", obj.Name)
	}

	if obj.Spec.Output.JobNamespace == "" {
		hasUpdated = true
		obj.Spec.Output.JobNamespace = obj.Namespace
	}

	ann := obj.GetAnnotations()
	annKey := "kloudlite.io/cluster.job-ref"
	if _, ok := ann[annKey]; !ok {
		hasUpdated = true
		fn.MapSet(&ann, annKey, fmt.Sprintf("%s/%s", obj.Spec.Output.JobNamespace, obj.Spec.Output.JobName))
		obj.SetAnnotations(ann)
	}

	if hasUpdated {
		if err := r.Update(ctx, obj); err != nil {
			return req.CheckFailed(checkName, check, err.Error())
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

func (r *ClusterReconciler) finalize(req *rApi.Request[*clustersv1.Cluster]) stepResult.Result {
	_, obj := req.Context(), req.Object

	checkName := "finalizing"
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	if step := r.startClusterDestroyJob(req); !step.ShouldProceed() {
		check.Status = false
		check.Message = "waiting for cluster destroy job check to be completed"
		if check != obj.Status.Checks[checkName] {
			fn.MapSet(&obj.Status.Checks, checkName, check)
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

func (r *ClusterReconciler) ensureCloudproviderStuffs(req *rApi.Request[*clustersv1.Cluster]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	checkName := "kloudlite-vpc"

	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	fail := func(err error) stepResult.Result {
		return req.CheckFailed(checkName, check, err.Error())
	}

	switch obj.Spec.CloudProvider {
	case common_types.CloudProviderAWS:
		{
			if obj.Spec.AWS.VPC == nil {
				namespace := obj.Namespace
				name := fmt.Sprintf("vpc-%s", obj.Spec.AWS.Region)
				awsvpc, err := rApi.Get(ctx, r.Client, fn.NN(namespace, name), &clustersv1.AwsVPC{})
				if err != nil {
					if !apiErrors.IsNotFound(err) {
						return fail(err)
					}
					// create vpc
					awsvpc = &clustersv1.AwsVPC{
						TypeMeta: metav1.TypeMeta{
							Kind:       "AwsVPC",
							APIVersion: "clusters.kloudlite.io/v1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      name,
							Namespace: namespace,
						},
						Spec: clustersv1.AwsVPCSpec{
							CredentialsRef: obj.Spec.CredentialsRef,
							CredentialKeys: *obj.Spec.CredentialKeys,
							Region:         obj.Spec.AWS.Region,
						},
					}
					if err := r.Create(ctx, awsvpc); err != nil {
						return fail(err)
					}
				}

				if !awsvpc.Status.IsReady {
					return fail(fmt.Errorf("aws vpc (%s) is not ready", name))
				}

				secret, err := rApi.Get(ctx, r.Client, fn.NN(awsvpc.Spec.Output.Namespace, awsvpc.Spec.Output.Name), &corev1.Secret{})
				if err != nil {
					return fail(err)
				}

				var m []map[string]string

				if err := json.Unmarshal(secret.Data["vpc_public_subnets"], &m); err != nil {
					return fail(err)
				}

				vpcPublicSubnets := make([]clustersv1.AwsSubnetWithID, 0, len(m))
				for _, v := range m {
					vpcPublicSubnets = append(vpcPublicSubnets, clustersv1.AwsSubnetWithID{
						AvailabilityZone: clustersv1.AwsAZ(v["availability_zone"]),
						ID:               v["id"],
					})
				}

				obj.Spec.AWS.VPC = &clustersv1.AwsVPCParams{
					ID:            string(bytes.Trim(bytes.TrimSpace(secret.Data["vpc_id"]), "\n")),
					PublicSubnets: vpcPublicSubnets,
				}

				if err := r.Update(ctx, obj); err != nil {
					return fail(err).RequeueAfter(500 * time.Millisecond)
				}

				return req.Done().RequeueAfter(500 * time.Millisecond)
			}
		}
	default:
		{
			return fail(fmt.Errorf("unsupported cloudprovider %s", obj.Spec.CloudProvider))
		}
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
			azToSubnetId := make(map[clustersv1.AwsAZ]string, len(obj.Spec.AWS.VPC.PublicSubnets))
			for _, v := range obj.Spec.AWS.VPC.PublicSubnets {
				azToSubnetId[v.AvailabilityZone] = v.ID
			}

			valuesBytes, err := json.Marshal(map[string]any{
				"tracker_id": fmt.Sprintf("cluster-%s", obj.Name),
				"aws_region": obj.Spec.AWS.Region,
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

				"enable_nvidia_gpu_support": obj.Spec.AWS.K3sMasters.NvidiaGpuEnabled,

				"vpc_id": obj.Spec.AWS.VPC.ID,

				"k3s_masters": map[string]any{
					// "image_id":           obj.Spec.AWS.K3sMasters.ImageId,
					// "image_ssh_username": obj.Spec.AWS.K3sMasters.ImageSSHUsername,
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
						"enabled": false,
					},
					"nodes": func() map[string]any {
						nodes := make(map[string]any, len(obj.Spec.AWS.K3sMasters.Nodes))
						for k, v := range obj.Spec.AWS.K3sMasters.Nodes {
							az := v.AvailabilityZone
							if az == "" {
								zones, ok := clustersv1.AwsRegionToAZs[obj.Spec.AWS.Region]
								if !ok {
									continue
								}
								az = zones[0]
							}

							nodes[k] = map[string]any{
								"role":              v.Role,
								"availability_zone": az,
								"vpc_subnet_id":     azToSubnetId[az],
								"last_recreated_at": v.LastRecreatedAt,
								"kloudlite_release": v.KloudliteRelease,
							}
						}
						return nodes
					}(),
				},

				"kloudlite_params": map[string]any{
					"release":            obj.Spec.KloudliteRelease,
					"install_crds":       true,
					"install_csi_driver": true,
					"install_operators":  false,
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

func (r *ClusterReconciler) startClusterApplyJob(req *rApi.Request[*clustersv1.Cluster]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	checkName := "checkName"

	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	fail := func(err error) stepResult.Result {
		return req.CheckFailed(checkName, check, err.Error())
	}

	job := &batchv1.Job{}
	if err := r.Get(ctx, fn.NN(obj.Spec.Output.JobNamespace, obj.Spec.Output.JobName), job); err != nil {
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
			"job-name":      obj.Spec.Output.JobName,
			"job-namespace": obj.Namespace,

			"labels": map[string]string{
				LabelClusterApplyJob:    "true",
				LabelResourceGeneration: fmt.Sprintf("%d", obj.Generation),
			},
			"pod-annotations": fn.MapMerge(fn.FilterObservabilityAnnotations(obj.GetAnnotations()), map[string]string{
				constants.ObservabilityAccountNameKey: obj.Spec.AccountName,
				constants.ObservabilityClusterNameKey: obj.Name,
			}),

			"job-node-selector": r.Env.IACJobNodeSelector,
			"job-tolerations":   r.Env.IACJobTolerations,

			"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},

			"service-account-name": clusterJobServiceAccount,

			"kubeconfig-secret-name":      obj.Spec.Output.SecretName,
			"kubeconfig-secret-namespace": obj.Namespace,
			"kubeconfig-secret-annotations": map[string]string{
				constants.DescriptionKey: fmt.Sprintf("kubeconfig for cluster %s", obj.Name),
			},

			"cluster-name":              obj.Name,
			"tf-state-secret-namespace": obj.Namespace,

			"aws-access-key-id":     r.Env.KlAwsAccessKey,
			"aws-secret-access-key": r.Env.KlAwsSecretKey,

			"values.json": string(valuesJson),
			"job-image":   r.Env.IACJobImage,
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
			return req.CheckFailed(clusterApplyJob, check, "waiting for previous jobs to finish execution").Err(nil)
		}

		if err := job_manager.DeleteJob(ctx, r.Client, job.Namespace, job.Name); err != nil {
			return req.CheckFailed(clusterApplyJob, check, err.Error())
		}

		return req.Done().RequeueAfter(1 * time.Second)
	}

	if !job_manager.HasJobFinished(ctx, r.Client, job) {
		return req.CheckFailed(clusterApplyJob, check, "waiting for job to finish execution").Err(nil)
	}

	if job.Status.Succeeded == 0 {
		return fail(fmt.Errorf("cluster creation job did not succeed"))
	}

	check.Status = true
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

	job := &batchv1.Job{}
	if err := r.Get(ctx, fn.NN(obj.Spec.Output.JobNamespace, obj.Spec.Output.JobName), job); err != nil {
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
			"job-name":      obj.Spec.Output.JobName,
			"job-namespace": obj.Spec.Output.JobNamespace,
			"labels": map[string]string{
				LabelClusterDestroyJob:  "true",
				LabelResourceGeneration: fmt.Sprintf("%d", obj.Generation),
			},
			"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},

			"job-node-selector": r.Env.IACJobNodeSelector,
			"job-tolerations":   r.Env.IACJobTolerations,

			"service-account-name": clusterJobServiceAccount,

			"kubeconfig-secret-name":      obj.Spec.Output.SecretName,
			"kubeconfig-secret-namespace": obj.Namespace,

			"cluster-name":              obj.Name,
			"tf-state-secret-namespace": obj.Namespace,

			"aws-access-key-id":     r.Env.KlAwsAccessKey,
			"aws-secret-access-key": r.Env.KlAwsSecretKey,
			"values.json":           string(valuesJson),

			"job-image": r.Env.IACJobImage,
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
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig(), kubectl.YAMLClientOpts{Logger: r.logger})

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
	builder.Watches(&clustersv1.AwsVPC{}, handler.EnqueueRequestsFromMapFunc(
		func(ctx context.Context, obj client.Object) []reconcile.Request {
			if v, ok := obj.GetLabels()[constants.RegionKey]; ok {
				var clist clustersv1.ClusterList
				if err := r.List(ctx, &clist, &client.ListOptions{
					LabelSelector: apiLabels.SelectorFromValidatedSet(map[string]string{
						constants.RegionKey: v,
					}),
					Namespace: obj.GetNamespace(),
				}); err != nil {
					return nil
				}

				rr := make([]reconcile.Request, 0, len(clist.Items))
				for i := range clist.Items {
					rr = append(rr, reconcile.Request{
						NamespacedName: fn.NN(clist.Items[i].GetNamespace(), clist.Items[i].GetName()),
					})
				}

				return rr
			}
			return nil
		},
	))

	builder.Watches(&clustersv1.AccountS3Bucket{}, handler.EnqueueRequestsFromMapFunc(
		func(ctx context.Context, obj client.Object) []reconcile.Request {
			if v, ok := obj.GetLabels()[constants.AccountNameKey]; ok {
				var clusterlist clustersv1.ClusterList
				if err := r.List(ctx, &clusterlist, &client.ListOptions{
					LabelSelector: apiLabels.SelectorFromValidatedSet(map[string]string{
						constants.AccountNameKey: v,
					}),
				}); err != nil {
					return nil
				}

				rreq := make([]reconcile.Request, 0, len(clusterlist.Items))
				for i := range clusterlist.Items {
					rreq = append(rreq, reconcile.Request{NamespacedName: fn.NN(clusterlist.Items[i].GetNamespace(), clusterlist.Items[i].GetName())})
				}
				return rreq
			}
			return nil
		}))

	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
