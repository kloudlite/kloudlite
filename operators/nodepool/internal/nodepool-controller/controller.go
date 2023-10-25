package nodepool_controller

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
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
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
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
	nodepoolNodesHash                  = "nodepool-nodes-hash"
	nodepoolDeleteJob                  = "nodepool-delete-job"
	deleteClusterAutoscalerMarkedNodes = "delete-cluster-autoscaler-marked-nodes"
	trackUpdatesOnNodes                = "track-updates-on-nodes"
	cleanupOrphanNodes                 = "cleanup-orphan-nodes"

	deleteNodeAfterTimestamp    = "kloudlite.io/delete-node-after-timestamp"
	markedForFutureDeletion     = "kloudlite.io/marked-for-future-deletion"
	excludeFromCalculation      = "kloudlite.io/marked-for-future-deletion"
	clusterAutoscalerDeleteNode = "kloudlite.io/cluster-autoscaler-delete-node"

	labelNodePoolApplyJob   = "kloudlite.io/nodepool-apply-job"
	labelNodePoolDeleteJob  = "kloudlite.io/nodepool-delete-job"
	labelResourceGeneration = "kloudlite.io/resource-generation"

	annotationDesiredNodesChecksum = "kloudlite.io/nodepool.nodes-checksum"
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

	if step := r.cleanupOrphanCorev1Nodes(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	// if step := r.cleanupNodesMarkedForDeletion(req); !step.ShouldProceed() {
	// 	return step.ReconcilerResponse()
	// }

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

func listNodesInNodepool(ctx context.Context, cli client.Client, nodepoolName string) ([]clustersv1.Node, error) {
	var nodesList clustersv1.NodeList
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

func (r *Reconciler) cleanupOrphanCorev1Nodes(req *rApi.Request[*clustersv1.NodePool]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(cleanupOrphanNodes)
	defer req.LogPostCheck(cleanupOrphanNodes)

	var nodesList corev1.NodeList
	if err := r.List(ctx, &nodesList, &client.ListOptions{
		LabelSelector: apiLabels.SelectorFromValidatedSet(map[string]string{
			constants.NodePoolNameKey: obj.Name,
		}),
	}); err != nil {
		return req.CheckFailed(cleanupOrphanNodes, check, err.Error())
	}

	for i := range nodesList.Items {
		var managerNode clustersv1.Node
		if err := r.Get(ctx, fn.NN("", nodesList.Items[i].Name), &managerNode); err != nil {
			if !apiErrors.IsNotFound(err) {
				return req.CheckFailed(cleanupOrphanNodes, check, err.Error())
			}

			// INFO: when clustersv1.Node, is not found for a corev1.Node belonging to a nodepool,
			// it means that the node is orphaned, and we need to delete it
			if err := r.Delete(ctx, &nodesList.Items[i]); err != nil {
				return req.CheckFailed(cleanupOrphanNodes, check, err.Error())
			}
		}
	}

	check.Status = true
	if check != obj.Status.Checks[cleanupOrphanNodes] {
		obj.Status.Checks[cleanupOrphanNodes] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

type NodeCalculator struct {
	DesiredNodes       map[string]clustersv1.NodeProps
	MarkedForDeletions map[string]clustersv1.Node
}

func (r *Reconciler) calculateNodes(ctx context.Context, obj *clustersv1.NodePool) (*NodeCalculator, error) {
	nc := &NodeCalculator{
		DesiredNodes:       map[string]clustersv1.NodeProps{},
		MarkedForDeletions: map[string]clustersv1.Node{},
	}

	extractNodeName := func(advertisedName string) string {
		return advertisedName[len(obj.Name)+1:]
	}

	currNodes, err := listNodesInNodepool(ctx, r.Client, obj.Name)
	if err != nil {
		return nil, err
	}

	sort.SliceStable(currNodes, func(i, j int) bool {
		return currNodes[i].CreationTimestamp.Before(&currNodes[j].CreationTimestamp)
	})

	for i := 0; i < len(currNodes); i++ {
		nodeName := extractNodeName(currNodes[i].Name)
		if currNodes[i].GetDeletionTimestamp() == nil {
			nc.DesiredNodes[nodeName] = clustersv1.NodeProps{}
		}
	}

	// INFO: when desirednodes is more than targetcount, we need to delete #diff nodes
	for i := len(nc.DesiredNodes); i > obj.Spec.TargetCount; i -= 1 {
		if err := r.Delete(ctx, &currNodes[i-1]); err != nil {
			return nil, err
		}
	}

	for i := len(nc.DesiredNodes); i < obj.Spec.TargetCount && i < obj.Spec.MaxCount; i++ {
		nodeName := fmt.Sprintf("node-%s", strings.ToLower(fn.CleanerNanoid(8)))
		node := &clustersv1.Node{ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-%s", obj.Name, nodeName)}}
		if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, node, func() error {
			if node.Labels == nil {
				node.Labels = make(map[string]string, 2)
			}
			node.Labels[constants.NodePoolNameKey] = obj.Name
			node.Labels[constants.NodeNameKey] = node.Name
			node.Spec.NodepoolName = obj.Name
			return nil
		}); err != nil {
			return nil, err
		}
		nc.DesiredNodes[nodeName] = clustersv1.NodeProps{}
	}

	return nc, nil
}

func (r *Reconciler) parseSpecToVarFileJson(ctx context.Context, obj *clustersv1.NodePool, nodesMap map[string]clustersv1.NodeProps) (string, error) {
	if obj.Spec.CloudProvider != "aws" || obj.Spec.AWS == nil {
		return "", fmt.Errorf(".spec.aws is nil")
	}

	var poolList clustersv1.NodePoolList
	if err := r.List(ctx, &poolList); err != nil {
		return "", client.IgnoreNotFound(err)
	}

	ec2Nodepools := map[string]any{}
	spotNodepools := map[string]any{}

	switch obj.Spec.AWS.PoolType {
	case "normal":
		{
			ec2Nodepools[obj.Name] = map[string]any{
				"ami":                  obj.Spec.AWS.NormalPool.AMI,
				"ami_ssh_username":     obj.Spec.AWS.NormalPool.AMISSHUsername,
				"availability_zone":    obj.Spec.AWS.NormalPool.AvailabilityZone,
				"nvidia_gpu_enabled":   obj.Spec.AWS.NormalPool.NvidiaGpuEnabled,
				"root_volume_type":     obj.Spec.AWS.NormalPool.RootVolumeType,
				"root_volume_size":     obj.Spec.AWS.NormalPool.RootVolumeSize,
				"iam_instance_profile": obj.Spec.AWS.NormalPool.IAMInstanceProfileRole,
				"instance_type":        obj.Spec.AWS.NormalPool.InstanceType,
				"nodes":                nodesMap,
			}
		}
	case "spot":
		{
			spotNodepools[obj.Name] = map[string]any{
				"ami":                          obj.Spec.AWS.SpotPool.AMI,
				"ami_ssh_username":             obj.Spec.AWS.SpotPool.AMISSHUsername,
				"availability_zone":            obj.Spec.AWS.SpotPool.AvailabilityZone,
				"nvidia_gpu_enabled":           obj.Spec.AWS.SpotPool.NvidiaGpuEnabled,
				"root_volume_type":             obj.Spec.AWS.SpotPool.RootVolumeType,
				"root_volume_size":             obj.Spec.AWS.SpotPool.RootVolumeSize,
				"iam_instance_profile":         obj.Spec.AWS.SpotPool.IAMInstanceProfileRole,
				"spot_fleet_tagging_role_name": obj.Spec.AWS.SpotPool.SpotFleetTaggingRoleName,
				"cpu_node": func() map[string]any {
					if obj.Spec.AWS.SpotPool.CpuNode == nil {
						return nil
					}
					return map[string]any{
						"vcpu": map[string]any{
							"min": obj.Spec.AWS.SpotPool.CpuNode.VCpu.Min,
							"max": obj.Spec.AWS.SpotPool.CpuNode.VCpu.Max,
						},
						"memory_per_vcpu": map[string]any{
							"min": obj.Spec.AWS.SpotPool.CpuNode.MemoryPerVCpu.Min,
							"max": obj.Spec.AWS.SpotPool.CpuNode.MemoryPerVCpu.Max,
						},
					}
				}(),
				"gpu_node": func() map[string]any {
					if obj.Spec.AWS.SpotPool.GpuNode == nil {
						return nil
					}

					return map[string]any{
						"instance_types": obj.Spec.AWS.SpotPool.GpuNode.InstanceTypes,
					}
				}(),
				"nodes": nodesMap,
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

func getJobName(name string) string {
	return fmt.Sprintf("kloudlite-node-pool-job-%s", name)
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

func getSortedKeys(data map[string]clustersv1.NodeProps) []string {
	keys := make([]string, 0, len(data))
	for i := range data {
		keys = append(keys, i)
	}

	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	return keys
}

func (r *Reconciler) startNodepoolApplyJob(req *rApi.Request[*clustersv1.NodePool]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(nodepoolApplyJob)
	defer req.LogPostCheck(nodepoolApplyJob)

	job := &batchv1.Job{}
	jobName := getJobName(obj.Name)
	jobNs := getJobNamespace()

	if err := r.Get(ctx, fn.NN(jobNs, jobName), job); err != nil {
		job = nil
	}

	failedWithErr := func(err error) stepResult.Result {
		return req.CheckFailed(nodepoolApplyJob, check, err.Error())
	}

	nc, err := r.calculateNodes(ctx, obj)
	if err != nil {
		return failedWithErr(err)
	}

	b, err := json.Marshal(nc.DesiredNodes)
	if err != nil {
		return failedWithErr(err)
	}
	checksum := fn.Md5(b)

	if job == nil {
		obj.Annotations[annotationDesiredNodesChecksum] = checksum
		if err := r.Update(ctx, obj); err != nil {
			return failedWithErr(err)
		}

		valuesJson, err := r.parseSpecToVarFileJson(ctx, obj, nc.DesiredNodes)
		if err != nil {
			return req.CheckFailed(nodepoolApplyJob, check, err.Error())
		}

		b, err := templates.ParseBytes(r.templateNodePoolJob, map[string]any{
			"action": "apply",

			"job-name":      jobName,
			"job-namespace": jobNs,
			"labels": map[string]string{
				constants.NodePoolNameKey: obj.Name,
				labelNodePoolApplyJob:     "true",
				labelResourceGeneration:   fmt.Sprintf("%d", obj.Generation),
			},
			"annotations": obj.Annotations,
			"owner-refs":  []metav1.OwnerReference{fn.AsOwner(obj, true)},

			"job-node-selector": constants.K8sMasterNodeSelector,

			"service-account-name": "",

			"aws-s3-bucket-name":     r.Env.IACStateS3BucketName,
			"aws-s3-bucket-region":   r.Env.IACStateS3BucketRegion,
			"aws-s3-bucket-filepath": fmt.Sprintf("%s/%s/%s/nodepool-%s.tfstate", r.Env.IACStateS3BucketDir, r.Env.KloudliteAccountName, r.Env.KloudliteClusterName, obj.Name),

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
	isMyJob = isMyJob && job.Annotations[annotationDesiredNodesChecksum] == checksum

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
		return req.CheckFailed(nodepoolApplyJob, check, "waiting for job to finish execution").Err(nil)
	}

	if err := actOnAutoscalerDeleteNode(ctx, r.Client, job, obj); err != nil {
		return req.CheckFailed(nodepoolApplyJob, check, err.Error())
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

func actOnAutoscalerDeleteNode(ctx context.Context, cli client.Client, job *batchv1.Job, nodepool *clustersv1.NodePool) error {
	npAnnValue, ok := nodepool.Annotations[clusterAutoscalerDeleteNode]
	if !ok {
		return nil
	}

	jobAnnValue, ok := job.Annotations[clusterAutoscalerDeleteNode]
	if !ok {
		return job_manager.DeleteJob(ctx, cli, job.Namespace, job.Name)
	}

	if npAnnValue != jobAnnValue {
		return job_manager.DeleteJob(ctx, cli, job.Namespace, job.Name)
	}

	delete(nodepool.Annotations, clusterAutoscalerDeleteNode)
	nodepool.Spec.TargetCount = nodepool.Spec.TargetCount - 1
	if err := cli.Update(ctx, nodepool); err != nil {
		return err
	}

	if err := cli.Delete(ctx, &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: jobAnnValue}}); err != nil {
		return client.IgnoreNotFound(err)
	}

	return nil
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
	jobName := getJobName(obj.Name)
	jobNs := getJobNamespace()
	if err := r.Get(ctx, fn.NN(jobNs, jobName), job); err != nil {
		job = nil
	}

	nc, err := r.calculateNodes(ctx, obj)
	if err != nil {
		return req.CheckFailed(nodepoolApplyJob, check, err.Error())
	}

	if job == nil {
		valuesJson, err := r.parseSpecToVarFileJson(ctx, obj, nc.DesiredNodes)
		if err != nil {
			return req.CheckFailed(nodepoolApplyJob, check, err.Error())
		}

		b, err := templates.ParseBytes(r.templateNodePoolJob, map[string]any{
			"action": "delete",

			"job-name":      jobName,
			"job-namespace": jobNs,
			"labels": map[string]string{
				constants.NodePoolNameKey: obj.Name,
				labelNodePoolDeleteJob:    "true",
				labelResourceGeneration:   fmt.Sprintf("%d", obj.Generation),
			},
			"annotations": obj.Annotations,
			"owner-refs":  []metav1.OwnerReference{fn.AsOwner(obj, true)},

			"job-node-selector": constants.K8sMasterNodeSelector,

			"service-account-name": "",

			"aws-s3-bucket-name":     r.Env.IACStateS3BucketName,
			"aws-s3-bucket-region":   r.Env.IACStateS3BucketRegion,
			"aws-s3-bucket-filepath": fmt.Sprintf("%s/%s/%s/nodepools-%s.tfstate", r.Env.IACStateS3BucketDir, r.Env.KloudliteAccountName, r.Env.KloudliteClusterName, obj.Name),

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

	watches := []*source.Kind{
		{Type: &corev1.Node{}},
		{Type: &batchv1.Job{}},
	}

	for i := range watches {
		builder.Watches(
			watches[i],
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
	}

	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
