package nodepool_controller

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kloudlite/operator/operators/nodepool/internal/env"
	"github.com/kloudlite/operator/operators/nodepool/internal/templates"

	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
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
	apiLabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Env        *env.Env
	logger     logging.Logger
	Name       string
	yamlClient kubectl.YAMLClient

	templateNodePoolJob []byte
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	ensureJobNamespace                 = "ensure-job-namespace"
	nodepoolApplyJob                   = "nodepool-apply-job"
	nodepoolDeleteJob                  = "nodepool-delete-job"
	deleteClusterAutoscalerMarkedNodes = "delete-cluster-autoscaler-marked-nodes"

	labelNodePoolApplyJob   = "kloudlite.io/nodepool-apply-job"
	labelNodePoolDeleteJob  = "kloudlite.io/nodepool-delete-job"
	labelResourceGeneration = "kloudlite.io/resource-generation"
)

// +kubebuilder:rbac:groups=clusters.kloudlite.io,resources=clusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=clusters.kloudlite.io,resources=clusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=clusters.kloudlite.io,resources=clusters/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, request.NamespacedName, &clustersv1.NodePool{})
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

	if step := r.ensureNodepoolJobNamespace(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.startNodepoolApplyJob(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, nil
}

func (r *Reconciler) finalize(req *rApi.Request[*clustersv1.NodePool]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks

	checkName := "finalizing"
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	if step := r.startNodepoolDeleteJob(req); !step.ShouldProceed() {
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

	if obj.Status.Checks[nodepoolDeleteJob].Status {
		if err := r.deleteAllNodesOfNodepool(ctx, obj.Name); err != nil {
			return req.CheckFailed(checkName, check, err.Error())
		}

		return req.Finalize()
	}

	return req.Done()
}

func listNodesInNodepool(ctx context.Context, cli client.Client, nodepoolName string) ([]corev1.Node, error) {
	var nodesList corev1.NodeList
	if err := cli.List(ctx, &nodesList, &client.ListOptions{
		LabelSelector: apiLabels.SelectorFromValidatedSet(map[string]string{
			constants.NodePoolNameKey: nodepoolName,
		}),
	}); err != nil {
		return nil, err
	}

	return nodesList.Items, nil
}

func (r *Reconciler) deleteAllNodesOfNodepool(ctx context.Context, nodePoolName string) error {
	var nodesList corev1.NodeList
	if err := r.List(ctx, &nodesList, &client.ListOptions{
		LabelSelector: apiLabels.SelectorFromValidatedSet(map[string]string{
			constants.NodePoolNameKey: nodePoolName,
		}),
	}); err != nil {
		return err
	}

	if len(nodesList.Items) == 0 {
		r.logger.Infof("all nodes belongig to nodepool %s have been deleted", nodePoolName)
		return nil
	}

	for i := range nodesList.Items {
		if err := r.Delete(ctx, &nodesList.Items[i]); err != nil {
			return err
		}
	}

	return fmt.Errorf("waiting for nodes belonging to nodepool %s to be deleted", nodePoolName)
}

func (r *Reconciler) parseSpecToVarFileJson(ctx context.Context, obj *clustersv1.NodePool) (string, error) {
	if obj.Spec.CloudProvider != "aws" || obj.Spec.AWS == nil {
		return "", fmt.Errorf(".spec.aws is nil")
	}

	var poolList clustersv1.NodePoolList
	if err := r.List(ctx, &poolList); err != nil {
		return "", client.IgnoreNotFound(err)
	}

	poolsMap := map[string]clustersv1.NodePool{}
	for i := range poolList.Items {
		if poolList.Items[i].DeletionTimestamp == nil {
			poolsMap[poolList.Items[i].Name] = poolList.Items[i]
		}
	}

	ec2Nodepools := map[string]any{}
	spotNodepools := map[string]any{}

	for k, v := range poolsMap {
		switch v.Spec.AWS.PoolType {
		case "normal":
			{
				ec2Nodepools[k] = map[string]any{
					"ami":                  v.Spec.AWS.NormalPool.AMI,
					"ami_ssh_username":     v.Spec.AWS.NormalPool.AMISSHUsername,
					"availability_zone":    v.Spec.AWS.NormalPool.AvailabilityZone,
					"nvidia_gpu_enabled":   v.Spec.AWS.NormalPool.NvidiaGpuEnabled,
					"root_volume_type":     v.Spec.AWS.NormalPool.RootVolumeType,
					"root_volume_size":     v.Spec.AWS.NormalPool.RootVolumeSize,
					"iam_instance_profile": v.Spec.AWS.NormalPool.IAMInstanceProfileRole,
					"instance_type":        v.Spec.AWS.NormalPool.InstanceType,
					"nodes": func() map[string]any {
						nodes := make(map[string]any, v.Spec.TargetCount)
						for i := 1; i <= v.Spec.TargetCount && i <= v.Spec.MaxCount; i++ {
							nodeName := fmt.Sprintf("node-%d", i)
							nodes[nodeName] = map[string]any{}
							if nv, ok := v.Spec.AWS.NormalPool.Nodes[nodeName]; ok {
								nodes[nodeName] = map[string]any{
									"last_recreated_at": nv.LastRecreatedAt,
								}
							}
						}

						return nodes
					}(),
				}
			}
		case "spot":
			{
				spotNodepools[k] = map[string]any{
					"ami":                          v.Spec.AWS.SpotPool.AMI,
					"ami_ssh_username":             v.Spec.AWS.SpotPool.AMISSHUsername,
					"availability_zone":            v.Spec.AWS.SpotPool.AvailabilityZone,
					"nvidia_gpu_enabled":           v.Spec.AWS.SpotPool.NvidiaGpuEnabled,
					"root_volume_type":             v.Spec.AWS.SpotPool.RootVolumeType,
					"root_volume_size":             v.Spec.AWS.SpotPool.RootVolumeSize,
					"iam_instance_profile":         v.Spec.AWS.SpotPool.IAMInstanceProfileRole,
					"spot_fleet_tagging_role_name": v.Spec.AWS.SpotPool.SpotFleetTaggingRoleName,
					"cpu_node": func() map[string]any {
						if v.Spec.AWS.SpotPool.CpuNode == nil {
							return nil
						}
						return map[string]any{
							"vcpu": map[string]any{
								"min": v.Spec.AWS.SpotPool.CpuNode.VCpu.Min,
								"max": v.Spec.AWS.SpotPool.CpuNode.VCpu.Max,
							},
							"memory_per_vcpu": map[string]any{
								"min": v.Spec.AWS.SpotPool.CpuNode.MemoryPerVCpu.Min,
								"max": v.Spec.AWS.SpotPool.CpuNode.MemoryPerVCpu.Max,
							},
						}
					}(),
					"gpu_node": func() map[string]any {
						if v.Spec.AWS.SpotPool.GpuNode == nil {
							return nil
						}

						return map[string]any{
							"instance_types": v.Spec.AWS.SpotPool.GpuNode.InstanceTypes,
						}
					}(),
					"nodes": func() map[string]any {
						nodes := make(map[string]any, v.Spec.TargetCount)
						for i := 1; i <= v.Spec.TargetCount && i <= v.Spec.MaxCount; i++ {
							nodeName := fmt.Sprintf("node-%d", i)
							nodes[nodeName] = map[string]any{}
							if nv, ok := v.Spec.AWS.SpotPool.Nodes[nodeName]; ok {
								nodes[nodeName] = map[string]any{
									"last_recreated_at": nv.LastRecreatedAt,
								}
							}
						}

						return nodes
					}(),
				}
			}
		}
	}

	variables, err := json.Marshal(map[string]any{
		"aws_access_key":             r.Env.CloudProviderAccessKey,
		"aws_secret_key":             r.Env.CloudProviderSecretKey,
		"aws_region":                 r.Env.CloudProviderRegion,
		"tracker_id":                 "nodepools",
		"k3s_join_token":             r.Env.K3sJoinToken,
		"k3s_server_public_dns_host": r.Env.K3sServerPublicHost,
		"ec2_nodepools":              ec2Nodepools,
		"spot_nodepools":             spotNodepools,
	})
	if err != nil {
		return "", err
	}
	return string(variables), nil
}

func getJobName() string {
	return "kloudlite-node-pool-job"
}

func getJobNamespace() string {
	return "kloudlite-jobs"
}

func (r *Reconciler) ensureNodepoolJobNamespace(req *rApi.Request[*clustersv1.NodePool]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(ensureJobNamespace)
	defer req.LogPostCheck(ensureJobNamespace)

	jobNs := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: getJobNamespace()}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, jobNs, func() error {
		if jobNs.Labels == nil {
			jobNs.Labels = make(map[string]string, 1)
		}
		jobNs.Labels[constants.KloudliteManagedNamespace] = "true"

		if jobNs.Annotations == nil {
			jobNs.Annotations = make(map[string]string, 1)
		}
		jobNs.Annotations[constants.DescriptionKey] = "kloudlite managed namespace for running cluster specific jobs, like nodepool, autoscaling, etc."
		return nil
	}); err != nil {
		return req.CheckFailed(ensureJobNamespace, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != obj.Status.Checks[ensureJobNamespace] {
		obj.Status.Checks[ensureJobNamespace] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

func (r *Reconciler) deleteNodesMarkedForDeletion(req *rApi.Request[*clustersv1.NodePool]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(deleteClusterAutoscalerMarkedNodes)
	defer req.LogPostCheck(deleteClusterAutoscalerMarkedNodes)

	nodes, err := listNodesInNodepool(ctx, r.Client, obj.Name)
	if err != nil {
		return req.CheckFailed(deleteClusterAutoscalerMarkedNodes, check, err.Error())
	}

	for i := range nodes {
	}

	check.Status = true
	if check != obj.Status.Checks[deleteClusterAutoscalerMarkedNodes] {
		obj.Status.Checks[deleteClusterAutoscalerMarkedNodes] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

func (r *Reconciler) startNodepoolApplyJob(req *rApi.Request[*clustersv1.NodePool]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(nodepoolApplyJob)
	defer req.LogPostCheck(nodepoolApplyJob)

	job := &batchv1.Job{}
	jobName := getJobName()
	jobNs := getJobNamespace()
	if err := r.Get(ctx, fn.NN(jobNs, jobName), job); err != nil {
		job = nil
	}

	if job == nil {
		valuesJson, err := r.parseSpecToVarFileJson(ctx, obj)
		if err != nil {
			return req.CheckFailed(nodepoolApplyJob, check, err.Error())
		}

		b, err := templates.ParseBytes(r.templateNodePoolJob, map[string]any{
			"action": "apply",

			"job-name":      jobName,
			"job-namespace": jobNs,
			"labels": map[string]string{
				labelNodePoolApplyJob:   "true",
				labelResourceGeneration: fmt.Sprintf("%d", obj.Generation),
			},
			"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},

			"job-node-selector": constants.K8sMasterNodeSelector,

			"service-account-name": "",

			"aws-s3-bucket-name":     r.Env.IACStateS3BucketName,
			"aws-s3-bucket-region":   r.Env.IACStateS3BucketRegion,
			"aws-s3-bucket-filepath": fmt.Sprintf("%s/%s-%s-nodepools.tfstate", r.Env.IACStateS3BucketDir, r.Env.KloudliteAccountName, r.Env.KloudliteClusterName),

			"aws-access-key-id":     r.Env.CloudProviderAccessKey,
			"aws-secret-access-key": r.Env.CloudProviderSecretKey,

			"values.json": string(valuesJson),
		})
		if err != nil {
			return req.CheckFailed(nodepoolApplyJob, check, err.Error()).Err(nil)
		}

		rr, err := r.yamlClient.ApplyYAML(ctx, b)
		if err != nil {
			return req.CheckFailed(nodepoolApplyJob, check, err.Error()).Err(nil)
		}
		req.AddToOwnedResources(rr...)
		req.Logger.Infof("waiting for job to be created")
		return req.Done().RequeueAfter(1 * time.Second)
	}

	isMyJob := job.Labels[labelResourceGeneration] == fmt.Sprintf("%d", obj.Generation) && job.Labels[labelNodePoolApplyJob] == "true"

	if !isMyJob {
		if !job_manager.HasJobFinished(ctx, r.Client, job) {
			return req.CheckFailed(nodepoolApplyJob, check, "waiting for previous jobs to finish execution")
		}

		if err := job_manager.DeleteJob(ctx, r.Client, job.Namespace, job.Name); err != nil {
			return req.CheckFailed(nodepoolApplyJob, check, err.Error())
		}

		return req.Done().RequeueAfter(1 * time.Second)
	}

	if !job_manager.HasJobFinished(ctx, r.Client, job) {
		// req.Logger.Infof("waiting for job to finish execution")
		return req.CheckFailed(nodepoolApplyJob, check, "waiting for job to finish execution").Err(nil)
	}

	check.Message = job_manager.GetTerminationLog(ctx, r.Client, job.Namespace, job.Name)
	check.Status = job.Status.Succeeded > 0
	if check != obj.Status.Checks[nodepoolApplyJob] {
		obj.Status.Checks[nodepoolApplyJob] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	if !check.Status {
		return req.Done()
	}

	return req.Next()
}

func (r *Reconciler) drainNodesOfNodepool(ctx context.Context, nodepoolName string) error {
	var nodesList corev1.NodeList
	if err := r.List(ctx, &nodesList, &client.ListOptions{
		LabelSelector: apiLabels.SelectorFromValidatedSet(map[string]string{
			constants.NodePoolNameKey: nodepoolName,
		}),
	}); err != nil {
		return err
	}

	for i := range nodesList.Items {
		nodesList.Items[i].Spec.Unschedulable = true
		if err := r.Update(ctx, &nodesList.Items[i]); err != nil {
			return err
		}
	}

	return nil
}

func (r *Reconciler) startNodepoolDeleteJob(req *rApi.Request[*clustersv1.NodePool]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(nodepoolDeleteJob)
	defer req.LogPostCheck(nodepoolDeleteJob)

	// INFO: draining nodes belonging to this nodegroup, prior to deleting them
	if err := r.drainNodesOfNodepool(ctx, obj.Name); err != nil {
		return req.CheckFailed(nodepoolDeleteJob, check, err.Error())
	}

	job := &batchv1.Job{}
	jobName := getJobName()
	jobNs := getJobNamespace()
	if err := r.Get(ctx, fn.NN(jobNs, jobName), job); err != nil {
		job = nil
	}

	if job == nil {
		valuesJson, err := r.parseSpecToVarFileJson(ctx, obj)
		if err != nil {
			return req.CheckFailed(nodepoolApplyJob, check, err.Error())
		}

		b, err := templates.ParseBytes(r.templateNodePoolJob, map[string]any{
			"action": "delete",

			"job-name":      jobName,
			"job-namespace": jobNs,
			"labels": map[string]string{
				labelNodePoolDeleteJob:  "true",
				labelResourceGeneration: fmt.Sprintf("%d", obj.Generation),
			},
			"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},

			"job-node-selector": constants.K8sMasterNodeSelector,

			"service-account-name": "",

			"aws-s3-bucket-name":     r.Env.IACStateS3BucketName,
			"aws-s3-bucket-region":   r.Env.IACStateS3BucketRegion,
			"aws-s3-bucket-filepath": fmt.Sprintf("%s/%s-%s-nodepools.tfstate", r.Env.IACStateS3BucketDir, r.Env.KloudliteAccountName, r.Env.KloudliteClusterName),

			"aws-access-key-id":     r.Env.CloudProviderAccessKey,
			"aws-secret-access-key": r.Env.CloudProviderSecretKey,

			"values.json": string(valuesJson),
		})
		if err != nil {
			return req.CheckFailed(nodepoolDeleteJob, check, err.Error()).Err(nil)
		}

		rr, err := r.yamlClient.ApplyYAML(ctx, b)
		if err != nil {
			return req.CheckFailed(nodepoolDeleteJob, check, err.Error()).Err(nil)
		}
		req.AddToOwnedResources(rr...)
		req.Logger.Infof("waiting for job to be created")
		return req.Done().RequeueAfter(1 * time.Second)
	}

	isMyJob := job.Labels[labelResourceGeneration] == fmt.Sprintf("%d", obj.Generation) && job.Labels[labelNodePoolDeleteJob] == "true"

	if !isMyJob {
		if !job_manager.HasJobFinished(ctx, r.Client, job) {
			return req.CheckFailed(nodepoolDeleteJob, check, "waiting for previous jobs to finish execution")
		}

		if err := job_manager.DeleteJob(ctx, r.Client, job.Namespace, job.Name); err != nil {
			return req.CheckFailed(nodepoolDeleteJob, check, err.Error())
		}

		return req.Done().RequeueAfter(1 * time.Second)
	}

	if !job_manager.HasJobFinished(ctx, r.Client, job) {
		return req.CheckFailed(nodepoolDeleteJob, check, "waiting for job to finish execution").Err(nil)
	}

	check.Message = job_manager.GetTerminationLog(ctx, r.Client, job.Namespace, job.Name)
	check.Status = job.Status.Succeeded > 0
	if check != obj.Status.Checks[nodepoolDeleteJob] {
		obj.Status.Checks[nodepoolDeleteJob] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	if !check.Status {
		return req.Done()
	}

	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

	var err error
	r.templateNodePoolJob, err = templates.ReadNodepoolJobTemplate()
	if err != nil {
		return err
	}

	builder := ctrl.NewControllerManagedBy(mgr).For(&clustersv1.NodePool{})
	builder.Watches(
		&source.Kind{Type: &corev1.Node{}},
		handler.EnqueueRequestsFromMapFunc(
			func(o client.Object) []reconcile.Request {
				npName, ok := o.GetLabels()[constants.NodePoolNameKey]
				if !ok {
					return nil
				}
				return []reconcile.Request{{NamespacedName: fn.NN("", npName)}}
			},
		),
	)

	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
