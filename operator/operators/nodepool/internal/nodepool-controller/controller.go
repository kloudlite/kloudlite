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
	ct "github.com/kloudlite/operator/apis/common-types"
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
	defaultsPatched                    = "defaults-patched"

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
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &clustersv1.NodePool{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	req.PreReconcile()
	defer req.PostReconcile()

	if v, ok := req.Object.Annotations[constants.AnnotationReconcileScheduledAfter]; ok {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return ctrl.Result{}, err
		}
		if time.Now().Before(t) {
			req.Logger.Infof("reconcile has been scheduled after %s, will reque after that", t)
			return ctrl.Result{RequeueAfter: time.Until(t)}, nil
		}
		delete(req.Object.Annotations, constants.AnnotationReconcileScheduledAfter)
		if err := r.Update(ctx, req.Object); err != nil {
			return ctrl.Result{}, err
		}
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

	if step := r.patchDefaults(req); !step.ShouldProceed() {
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

	if step := r.startNodepoolApplyJob(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *Reconciler) patchDefaults(req *rApi.Request[*clustersv1.NodePool]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(defaultsPatched)
	defer req.LogPostCheck(defaultsPatched)

	hasUpdated := false

	if obj.Spec.IAC.JobName == "" {
		hasUpdated = true
		obj.Spec.IAC.JobName = fmt.Sprintf("kloudlite-nodepool-job-%s", obj.Name)
	}

	if obj.Spec.IAC.JobNamespace == "" {
		hasUpdated = true
		obj.Spec.IAC.JobNamespace = "kloudlite-jobs"
	}

	if hasUpdated {
		if err := r.Update(ctx, obj); err != nil {
			return req.CheckFailed(defaultsPatched, check, err.Error())
		}
		return req.Done()
	}

	check.Status = true
	if check != obj.Status.Checks[defaultsPatched] {
		obj.Status.Checks[defaultsPatched] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

func (r *Reconciler) finalize(req *rApi.Request[*clustersv1.NodePool]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks

	checkName := "finalizing"
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	if step := r.startNodepoolDeleteJob(req); !step.ShouldProceed() {
		check.Status = false
		check.Message = "waiting for nodepool delete job to finish execution"
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

		dt := currNodes[i].GetDeletionTimestamp().IsZero()
		if dt {
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

func toAWSVarfileJson(obj *clustersv1.NodePool, ev *env.Env, nodesMap map[string]clustersv1.NodeProps, accessKey, secretKey string) (string, error) {
	if obj.Spec.AWS == nil {
		return "", fmt.Errorf(".spec.aws is nil")
	}

	ec2Nodepools := make(map[string]any, 1)
	spotNodepools := make(map[string]any, 1)

	switch obj.Spec.AWS.PoolType {
	case clustersv1.AWSPoolTypeEC2:
		{
			ec2Nodepools[obj.Name] = map[string]any{
				"image_id":             obj.Spec.AWS.ImageId,
				"image_ssh_username":   obj.Spec.AWS.ImageSSHUsername,
				"availability_zone":    obj.Spec.AWS.AvailabilityZone,
				"nvidia_gpu_enabled":   obj.Spec.AWS.NvidiaGpuEnabled,
				"root_volume_type":     obj.Spec.AWS.RootVolumeType,
				"root_volume_size":     obj.Spec.AWS.RootVolumeSize,
				"iam_instance_profile": obj.Spec.AWS.IAMInstanceProfileRole,
				"instance_type":        obj.Spec.AWS.EC2Pool.InstanceType,
				"nodes":                nodesMap,
			}
		}
	case clustersv1.AWSPoolTypeSpot:
		{
			if obj.Spec.AWS.SpotPool == nil {
				return "", fmt.Errorf(".spec.aws.spotPool is nil")
			}

			spotNodepools[obj.Name] = map[string]any{
				"image_id":                     obj.Spec.AWS.ImageId,
				"image_ssh_username":           obj.Spec.AWS.ImageSSHUsername,
				"availability_zone":            obj.Spec.AWS.AvailabilityZone,
				"nvidia_gpu_enabled":           obj.Spec.AWS.NvidiaGpuEnabled,
				"root_volume_type":             obj.Spec.AWS.RootVolumeType,
				"root_volume_size":             obj.Spec.AWS.RootVolumeSize,
				"iam_instance_profile":         obj.Spec.AWS.IAMInstanceProfileRole,
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

	variables := map[string]any{
		// "aws_access_key":             nil,
		// "aws_secret_key":             nil,
		"aws_region":                 ev.CloudProviderRegion,
		"tracker_id":                 fmt.Sprintf("nodepool-%s", obj.Name),
		"k3s_join_token":             ev.K3sJoinToken,
		"k3s_server_public_dns_host": ev.K3sServerPublicHost,
		"ec2_nodepools":              ec2Nodepools,
		"spot_nodepools":             spotNodepools,
		"extra_agent_args":           []string{},
		"save_ssh_key_to_path":       "",
	}

	if accessKey != "" {
		variables["aws_access_key"] = accessKey
	}

	if secretKey != "" {
		variables["aws_secret_key"] = secretKey
	}

	b, err := json.Marshal(variables)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (r *Reconciler) parseSpecToVarFileJson(ctx context.Context, obj *clustersv1.NodePool, nodesMap map[string]clustersv1.NodeProps, accessKey, secretKey string) (string, error) {
	var poolList clustersv1.NodePoolList
	if err := r.List(ctx, &poolList); err != nil {
		return "", client.IgnoreNotFound(err)
	}

	switch obj.Spec.CloudProvider {
	case ct.CloudProviderAWS:
		return toAWSVarfileJson(obj, r.Env, nodesMap, accessKey, secretKey)
	default:
		return "", fmt.Errorf("unsupported cloud provider: %s", obj.Spec.CloudProvider)
	}
}

func (r *Reconciler) ensureNodepoolJobNamespace(req *rApi.Request[*clustersv1.NodePool]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(ensureJobNamespace)
	defer req.LogPostCheck(ensureJobNamespace)

	jobNs := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.IAC.JobNamespace}}
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
	jobName := obj.Spec.IAC.JobName
	jobNs := obj.Spec.IAC.JobNamespace

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

		accessKey, err := func() (string, error) {
			s, err2 := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.IAC.CloudProviderAccessKey.Namespace, obj.Spec.IAC.CloudProviderAccessKey.Name), &corev1.Secret{})
			if err2 != nil {
				return "", err2
			}

			return string(s.Data[obj.Spec.IAC.CloudProviderAccessKey.Key]), nil
		}()
		if err != nil {
			return req.CheckFailed(nodepoolDeleteJob, check, err.Error())
		}

		secretKey, err := func() (string, error) {
			s, err2 := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.IAC.CloudProviderSecretKey.Namespace, obj.Spec.IAC.CloudProviderSecretKey.Name), &corev1.Secret{})
			if err2 != nil {
				return "", err2
			}

			return string(s.Data[obj.Spec.IAC.CloudProviderSecretKey.Key]), nil
		}()
		if err != nil {
			return req.CheckFailed(nodepoolDeleteJob, check, err.Error())
		}

		valuesJson, err := r.parseSpecToVarFileJson(ctx, obj, nc.DesiredNodes, accessKey, secretKey)
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

			"aws-s3-bucket-name":   obj.Spec.IAC.StateS3BucketName,
			"aws-s3-bucket-region": obj.Spec.IAC.StateS3BucketRegion,
			// "aws-s3-bucket-filepath": fmt.Sprintf("%s/%s/%s/nodepools-%s.tfstate", r.Env.IACStateS3BucketDir, r.Env.KloudliteAccountName, r.Env.KloudliteClusterName, obj.Name),
			"aws-s3-bucket-filepath": obj.Spec.IAC.StateS3BucketFilePath,

			"aws-s3-access-key": accessKey,
			"aws-s3-secret-key": secretKey,

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
	jobName := obj.Spec.IAC.JobName
	jobNs := obj.Spec.IAC.JobNamespace
	if err := r.Get(ctx, fn.NN(jobNs, jobName), job); err != nil {
		job = nil
	}

	nc, err := r.calculateNodes(ctx, obj)
	if err != nil {
		return req.CheckFailed(nodepoolDeleteJob, check, err.Error())
	}

	if job == nil {
		accessKey, err := func() (string, error) {
			s, err2 := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.IAC.CloudProviderAccessKey.Namespace, obj.Spec.IAC.CloudProviderAccessKey.Name), &corev1.Secret{})
			if err2 != nil {
				return "", err2
			}

			return string(s.Data[obj.Spec.IAC.CloudProviderAccessKey.Key]), nil
		}()
		if err != nil {
			return req.CheckFailed(nodepoolDeleteJob, check, err.Error())
		}

		secretKey, err := func() (string, error) {
			s, err2 := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.IAC.CloudProviderSecretKey.Namespace, obj.Spec.IAC.CloudProviderSecretKey.Name), &corev1.Secret{})
			if err2 != nil {
				return "", err2
			}

			return string(s.Data[obj.Spec.IAC.CloudProviderSecretKey.Key]), nil
		}()
		if err != nil {
			return req.CheckFailed(nodepoolDeleteJob, check, err.Error())
		}

		valuesJson, err := r.parseSpecToVarFileJson(ctx, obj, nc.DesiredNodes, accessKey, secretKey)
		if err != nil {
			return req.CheckFailed(nodepoolDeleteJob, check, err.Error())
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

			"aws-s3-bucket-name":   obj.Spec.IAC.StateS3BucketName,
			"aws-s3-bucket-region": obj.Spec.IAC.StateS3BucketRegion,
			// "aws-s3-bucket-filepath": fmt.Sprintf("%s/%s/%s/nodepools-%s.tfstate", r.Env.IACStateS3BucketDir, r.Env.KloudliteAccountName, r.Env.KloudliteClusterName, obj.Name),
			"aws-s3-bucket-filepath": obj.Spec.IAC.StateS3BucketFilePath,

			"aws-s3-access-key": accessKey,
			"aws-s3-secret-key": secretKey,

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
			return req.CheckFailed(nodepoolDeleteJob, check, "waiting for previous jobs to finish execution").Err(nil)
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

	watches := []client.Object{
		&corev1.Node{},
		&batchv1.Job{},
	}

	for _, obj := range watches {
		builder.Watches(
			obj,
			handler.EnqueueRequestsFromMapFunc(
				func(ctx context.Context, o client.Object) []reconcile.Request {
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
