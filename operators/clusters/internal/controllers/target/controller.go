package target

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	ct "github.com/kloudlite/operator/apis/common-types"

	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	"github.com/kloudlite/operator/operators/clusters/internal/env"
	"github.com/kloudlite/operator/operators/clusters/internal/templates"
	"github.com/kloudlite/operator/pkg/constants"
	"github.com/kloudlite/operator/pkg/errors"
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
	LabelClusterApplyJob    = "kloudlite.io/cluster-apply-job"
	LabelResourceGeneration = "kloudlite.io/resource-generation"
	LabelClusterDestroyJob  = "kloudlite.io/cluster-apply-job"
)

const (
	clusterJobServiceAccount = "cluster-job-sa"
)

const (
	ClusterPrerequisitesReady string = "cluster-prerequisites-ready"

	ClusterJobRBACReady             string = `cluster-job-rbac-ready`
	ClusterCreateJobAppliedAndReady string = `cluster-create-job-applied-and-ready`
	ClusterDeleteJobApplied         string = `cluster-delete-job-applied`

	DefaultsPatched string = "defaults-patched"
	KeyMsvcOutput   string = "msvc-output"

	AnnotationCurrentStorageSize string = "kloudlite.io/msvc.storage-size"
)

var ApplyCheckList = []rApi.CheckMeta{
	{Name: DefaultsPatched, Title: "Defaults Patched", Debug: true},
	{Name: ClusterPrerequisitesReady, Title: "Cluster Pre-Requisites Ready"},
	{Name: ClusterJobRBACReady, Title: "Cluster Job RBAC Ready", Debug: true},
	{Name: ClusterCreateJobAppliedAndReady, Title: "Cluster Create Job Applied"},
}

// DefaultsPatched string = "defaults-patched"
var DeleteCheckList = []rApi.CheckMeta{}

// +kubebuilder:rbac:groups=clusters.kloudlite.io,resources=clusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=clusters.kloudlite.io,resources=clusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=clusters.kloudlite.io,resources=clusters/finalizers,verbs=update

func (r *ClusterReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &clustersv1.Cluster{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

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

	if step := req.EnsureCheckList(ApplyCheckList); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureChecks(DefaultsPatched, ClusterPrerequisitesReady, ClusterJobRBACReady, ClusterCreateJobAppliedAndReady); !step.ShouldProceed() {
		return step.ReconcilerResponse()
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

	check := rApi.NewRunningCheck(DefaultsPatched, req)

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
			return check.Failed(err)
		}
		return req.Done().RequeueAfter(500 * time.Millisecond)
	}

	return check.Completed()
}

func (r *ClusterReconciler) finalize(req *rApi.Request[*clustersv1.Cluster]) stepResult.Result {
	_, obj := req.Context(), req.Object

	if !slices.Equal(obj.Status.CheckList, DeleteCheckList) {
		req.Object.Status.CheckList = DeleteCheckList
		if step := req.UpdateStatus(); !step.ShouldProceed() {
			return step
		}
	}

	check := rApi.NewRunningCheck("finalizing", req)

	if step := r.startClusterDestroyJob(req); !step.ShouldProceed() {
		return check.StillRunning(fmt.Errorf("waiting for cluster destroy job check to be completed"))
	}

	return req.Finalize()
}

func (r *ClusterReconciler) ensureJobRBAC(req *rApi.Request[*clustersv1.Cluster]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(ClusterJobRBACReady, req)

	b, err := templates.ParseBytes(r.templateRBACForClusterJob, map[string]any{
		"service-account-name": clusterJobServiceAccount,
		"namespace":            obj.Namespace,
	})
	if err != nil {
		return check.Failed(err).Err(nil)
	}

	rr, err := r.yamlClient.ApplyYAML(ctx, b)
	if err != nil {
		return check.Failed(err)
	}
	req.AddToOwnedResources(rr...)

	return check.Completed()
}

func (r *ClusterReconciler) ensureCloudproviderStuffs(req *rApi.Request[*clustersv1.Cluster]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(ClusterPrerequisitesReady, req)

	switch obj.Spec.CloudProvider {
	case ct.CloudProviderAWS:
		{
			if obj.Spec.AWS == nil {
				return check.Failed(fmt.Errorf(".spec.aws must be set when cloudprovider is aws")).Err(nil)
			}

			if obj.Spec.AWS.VPC == nil {
				namespace := obj.Namespace
				name := fmt.Sprintf("vpc-%s", obj.Spec.AWS.Region)
				awsvpc, err := rApi.Get(ctx, r.Client, fn.NN(namespace, name), &clustersv1.AwsVPC{})
				if err != nil {
					if !apiErrors.IsNotFound(err) {
						return check.Failed(err)
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
							Credentials: obj.Spec.AWS.Credentials,
							Region:      obj.Spec.AWS.Region,
						},
					}
					if err := r.Create(ctx, awsvpc); err != nil {
						return check.Failed(err)
					}
				}

				if !awsvpc.Status.IsReady {
					return check.StillRunning(fmt.Errorf("aws vpc (%s) is not ready", name))
				}

				secret, err := rApi.Get(ctx, r.Client, fn.NN(awsvpc.Spec.Output.Namespace, awsvpc.Spec.Output.Name), &corev1.Secret{})
				if err != nil {
					return check.Failed(err)
				}

				var m []map[string]string

				if err := json.Unmarshal(secret.Data["vpc_public_subnets"], &m); err != nil {
					return check.Failed(err)
				}

				vpcPublicSubnets := make([]clustersv1.AwsSubnetWithID, 0, len(m))
				for _, v := range m {
					vpcPublicSubnets = append(vpcPublicSubnets, clustersv1.AwsSubnetWithID{
						AvailabilityZone: v["availability_zone"],
						ID:               v["id"],
					})
				}

				obj.Spec.AWS.VPC = &clustersv1.AwsVPCParams{
					ID:            string(bytes.Trim(bytes.TrimSpace(secret.Data["vpc_id"]), "\n")),
					PublicSubnets: vpcPublicSubnets,
				}

				if err := r.Update(ctx, obj); err != nil {
					return check.Failed(err)
				}

				return req.Done().RequeueAfter(500 * time.Millisecond)
			}
		}
	default:
		{
			return check.Failed(fmt.Errorf("unsupported cloudprovider %s", obj.Spec.CloudProvider)).Err(nil)
		}
	}

	return check.Completed()
}

func (r *ClusterReconciler) parseSpecToVarFileJson(ctx context.Context, obj *clustersv1.Cluster) (string, error) {
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

			credsSecret := &corev1.Secret{}
			if err := r.Get(ctx, fn.NN(obj.Spec.AWS.Credentials.SecretRef.Namespace, obj.Spec.AWS.Credentials.SecretRef.Name), credsSecret); err != nil {
				return "", errors.NewEf(err, "failed to get aws credentials")
			}

			if obj.Spec.AWS.VPC == nil {
				return "", fmt.Errorf(".spec.aws.vpc must be provided")
			}

			azToSubnetId := make(map[string]string, len(obj.Spec.AWS.VPC.PublicSubnets))
			for _, v := range obj.Spec.AWS.VPC.PublicSubnets {
				azToSubnetId[v.AvailabilityZone] = v.ID
			}

			values := map[string]any{
				// "aws_access_key": func() string {
				// 	if !isAssumeRole {
				// 		return string(providerCreds.Data[obj.Spec.CredentialKeys.KeyAccessKey])
				// 	}
				// 	return r.Env.KlAwsAccessKey
				// }(),
				// "aws_secret_key": func() string {
				// 	if !isAssumeRole {
				// 		return string(providerCreds.Data[obj.Spec.CredentialKeys.KeySecretKey])
				// 	}
				// 	return r.Env.KlAwsSecretKey
				// }(),
				// "aws_assume_role": func() map[string]any {
				// 	if !isAssumeRole {
				// 		return nil
				// 	}
				// 	return map[string]any{
				// 		"enabled":     true,
				// 		"role_arn":    string(providerCreds.Data[obj.Spec.CredentialKeys.KeyAWSAssumeRoleRoleARN]),
				// 		"external_id": string(providerCreds.Data[obj.Spec.CredentialKeys.KeyAWSAssumeRoleExternalID]),
				// 	}
				// }(),

				"aws_region":                obj.Spec.AWS.Region,
				"tracker_id":                fmt.Sprintf("cluster-%s", obj.Name),
				"enable_nvidia_gpu_support": obj.Spec.AWS.K3sMasters.NvidiaGpuEnabled,

				"vpc_id": obj.Spec.AWS.VPC.ID,

				"k3s_masters": map[string]any{
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
								zones, ok := clustersv1.AwsRegionToAZs[clustersv1.AwsRegion(obj.Spec.AWS.Region)]
								if !ok {
									continue
								}
								az = string(zones[0])
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
			}

			switch obj.Spec.AWS.Credentials.AuthMechanism {
			case clustersv1.AwsAuthMechanismSecretKeys:
				{
					awscreds, err := fn.ParseFromSecret[clustersv1.AwsAuthSecretKeys](credsSecret)
					if err != nil {
						return "", err
					}

					values["aws_access_key"] = awscreds.AccessKey
					values["aws_secret_key"] = awscreds.SecretKey
					values["aws_assume_role"] = map[string]any{
						"enabled": false,
					}
				}
			case clustersv1.AwsAuthMechanismAssumeRole:
				{
					awscreds, err := fn.ParseFromSecret[clustersv1.AwsAssumeRoleParams](credsSecret)
					if err != nil {
						return "", err
					}

					values["aws_access_key"] = r.Env.KlAwsAccessKey
					values["aws_secret_key"] = r.Env.KlAwsSecretKey
					values["aws_assume_role"] = map[string]any{
						"enabled":     true,
						"role_arn":    awscreds.RoleARN,
						"external_id": awscreds.ExternalID,
					}
				}
			}

			valuesBytes, err := json.Marshal(values)
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
	check := rApi.NewRunningCheck(ClusterCreateJobAppliedAndReady, req)

	job := &batchv1.Job{}
	if err := r.Get(ctx, fn.NN(obj.Spec.Output.JobNamespace, obj.Spec.Output.JobName), job); err != nil {
		job = nil
	}

	if job == nil {
		valuesJson, err := r.parseSpecToVarFileJson(ctx, obj)
		if err != nil {
			return check.Failed(err).Err(nil)
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
			return check.Failed(err).Err(nil)
		}

		rr, err := r.yamlClient.ApplyYAML(ctx, b)
		if err != nil {
			return check.Failed(err)
		}
		req.AddToOwnedResources(rr...)
		req.Logger.Infof("waiting for job to be created")
		return req.Done().RequeueAfter(1 * time.Second)
	}

	isMyJob := job.Labels[LabelResourceGeneration] == fmt.Sprintf("%d", obj.Generation) && job.Labels[LabelClusterApplyJob] == "true"

	if !isMyJob {
		if !job_manager.HasJobFinished(ctx, r.Client, job) {
			return check.Failed(fmt.Errorf("waiting for previous jobs to finish execution")).Err(nil)
		}

		if err := job_manager.DeleteJob(ctx, r.Client, job.Namespace, job.Name); err != nil {
			return check.Failed(err)
		}

		return req.Done().RequeueAfter(1 * time.Second)
	}

	if !job_manager.HasJobFinished(ctx, r.Client, job) {
		return check.Failed(fmt.Errorf("waiting for job to finish execution"))
	}

	if job.Status.Succeeded == 0 {
		return check.Failed(fmt.Errorf("cluster creation job did not succeed"))
	}

	return check.Completed()
}

func (r *ClusterReconciler) startClusterDestroyJob(req *rApi.Request[*clustersv1.Cluster]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(ClusterDeleteJobApplied, req)

	job := &batchv1.Job{}
	if err := r.Get(ctx, fn.NN(obj.Spec.Output.JobNamespace, obj.Spec.Output.JobName), job); err != nil {
		job = nil
	}

	if job == nil {
		valuesJson, err := r.parseSpecToVarFileJson(ctx, obj)
		if err != nil {
			return check.Failed(err).Err(nil)
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
			return check.Failed(err).Err(nil)
		}

		rr, err := r.yamlClient.ApplyYAML(ctx, b)
		if err != nil {
			return check.Failed(err)
		}
		req.AddToOwnedResources(rr...)
		req.Logger.Infof("waiting for job to be created")
		return req.Done().RequeueAfter(1 * time.Second)
	}

	isMyJob := job.Labels[LabelResourceGeneration] == fmt.Sprintf("%d", obj.Generation) && job.Labels[LabelClusterDestroyJob] == "true"

	if !isMyJob {
		if !job_manager.HasJobFinished(ctx, r.Client, job) {
			return check.StillRunning(fmt.Errorf("waiting for previous jobs to finish execution"))
		}

		if err := job_manager.DeleteJob(ctx, r.Client, job.Namespace, job.Name); err != nil {
			return check.Failed(err)
		}

		return req.Done().RequeueAfter(1 * time.Second)
	}

	if !job_manager.HasJobFinished(ctx, r.Client, job) {
		return check.StillRunning(fmt.Errorf("waiting for job to finish execution"))
	}

	if job.Status.Succeeded < 1 {
		// means job failed
		return check.Failed(fmt.Errorf("job failed, checkout logs for more details")).Err(nil)
	}

	// check.Message = job_manager.GetTerminationLog(ctx, r.Client, job.Namespace, job.Name)

	return check.Completed()
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
