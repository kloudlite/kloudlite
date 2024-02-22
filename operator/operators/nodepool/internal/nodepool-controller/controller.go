package nodepool_controller

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	"github.com/kloudlite/operator/operators/nodepool/internal/env"
	"github.com/kloudlite/operator/operators/nodepool/internal/templates"
	"github.com/kloudlite/operator/pkg/constants"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	ct "github.com/kloudlite/operator/apis/common-types"
	fn "github.com/kloudlite/operator/pkg/functions"
	job_manager "github.com/kloudlite/operator/pkg/job-helper"
	rApi "github.com/kloudlite/operator/pkg/operator"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/util/taints"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Env        *env.Env
	logger     logging.Logger
	Name       string
	yamlClient kubectl.YAMLClient

	templateNodePoolJob   []byte
	templateNamespaceRBAC []byte
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	labelNodePoolApplyJob   string = "kloudlite.io/nodepool-apply-job"
	labelResourceGeneration string = "kloudlite.io/resource-generation"

	labelNodenameWithoutPrefix string = "kloudlite.io/node-name-without-prefix"

	annotationNodesChecksum string = "kloudlite.io/nodes.checksum"

	nodeFinalizer string = "kloudlite.io/nodepool-node-finalizer"
)

const (
	annNodepoolJobRef = "kloudlite.io/nodepool.job-ref"
)

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &clustersv1.NodePool{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	req.PreReconcile()
	defer req.PostReconcile()

	if step := r.patchDefaults(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
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

	if step := req.EnsureFinalizers(constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.updateNodeTaintsAndLabels(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureJobNamespaceRBACs(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.syncNodepool(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *Reconciler) finalize(req *rApi.Request[*clustersv1.NodePool]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	checkName := "finalizing"

	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	fail := func(err error) stepResult.Result {
		return req.CheckFailed(checkName, check, err.Error())
	}

	nodes, err := nodesBelongingToNodepool(ctx, r.Client, obj.Name)
	if err != nil {
		return fail(err)
	}

	if err := deleteNodes(ctx, r.Client, nodes...); err != nil {
		return fail(err)
	}

	if step := r.syncNodepool(req); !step.ShouldProceed() {
		return step
	}

	return req.Finalize()
}

func (r *Reconciler) patchDefaults(req *rApi.Request[*clustersv1.NodePool]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	checkName := "patch-defaults"

	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	fail := func(err error) stepResult.Result {
		return req.CheckFailed(checkName, check, err.Error())
	}

	hasUpdated := false

	if v, ok := obj.Annotations[annNodepoolJobRef]; v != fmt.Sprintf("%s/kloudlite-nodepool-%s", r.Env.JobsNamespace, obj.Name) || !ok {
		hasUpdated = true
		ann := obj.Annotations
		if ann == nil {
			ann = make(map[string]string, 1)
		}
		ann[annNodepoolJobRef] = fmt.Sprintf("%s/kloudlite-nodepool-%s", r.Env.JobsNamespace, obj.Name)
		obj.SetAnnotations(ann)
	}

	if hasUpdated {
		if err := r.Update(ctx, obj); err != nil {
			return fail(err)
		}
		return req.Done()
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

func (r *Reconciler) updateNodeTaintsAndLabels(req *rApi.Request[*clustersv1.NodePool]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	checkName := "node-taints-and-labels"

	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	fail := func(err error) stepResult.Result {
		return req.CheckFailed(checkName, check, err.Error())
	}

	nodes, err := realNodesBelongingToNodepool(ctx, r.Client, obj.Name)
	if err != nil {
		return fail(err)
	}

	for _, node := range nodes {
		for _, taint := range obj.Spec.NodeTaints {
			if !taints.TaintExists(node.Spec.Taints, &taint) {
				node.Spec.Taints = append(node.Spec.Taints, taint)
			}
		}

		if node.Labels == nil {
			node.Labels = map[string]string{}
		}

		for k, v := range obj.Spec.NodeLabels {
			node.Labels[k] = v
		}

		if err := r.Update(ctx, &node); err != nil {
			return fail(err).RequeueAfter(200 * time.Millisecond)
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

func (r *Reconciler) ensureJobNamespaceRBACs(req *rApi.Request[*clustersv1.NodePool]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	checkName := "nodepool-job-namespace"

	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	fail := func(err error) stepResult.Result {
		return req.CheckFailed(checkName, check, err.Error())
	}

	b, err := templates.ParseBytes(r.templateNamespaceRBAC, map[string]any{
		"namespace": r.Env.JobsNamespace,
	})
	if err != nil {
		return fail(err).Err(nil)
	}

	_, err = r.yamlClient.ApplyYAML(ctx, b)
	if err != nil {
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

func (r *Reconciler) syncNodepool(req *rApi.Request[*clustersv1.NodePool]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	checkName := "sync-nodepool"

	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	fail := func(err error) stepResult.Result {
		return req.CheckFailed(checkName, check, err.Error())
	}

	nodes, err := nodesBelongingToNodepool(ctx, r.Client, obj.Name)
	if err != nil {
		return fail(err)
	}

	if err := addFinalizersOnNodes(ctx, r.Client, nodes, nodeFinalizer); err != nil {
		return fail(err)
	}

	markedForDeletion := filterNodesMarkedForDeletion(nodes)

	nodesMap := make(map[string]clustersv1.NodeProps, len(nodes)-len(markedForDeletion))
	for i := range nodes {
		rawName := nodes[i].GetName()[len(obj.GetName())+1:]
		if _, ok := markedForDeletion[nodes[i].GetName()]; !ok {
			nodesMap[rawName] = clustersv1.NodeProps{}
		}
	}

	checksum := nodesChecksum(nodesMap)

	varfileJson, err := r.parseSpecToVarFileJson(ctx, obj, nodesMap)
	if err != nil {
		return fail(err).Err(nil)
	}

	jobRef := strings.Split(obj.Annotations[annNodepoolJobRef], "/")

	jobNamespace := jobRef[0]
	jobName := jobRef[1]

	job := &batchv1.Job{}
	if err := r.Get(ctx, fn.NN(jobNamespace, jobName), job); err != nil {
		job = nil
	}

	if job == nil {
		b, err := templates.ParseBytes(r.templateNodePoolJob, map[string]any{
			"action": "apply",

			"job-name":      jobName,
			"job-namespace": jobNamespace,
			"labels": map[string]string{
				constants.NodePoolNameKey: obj.Name,
				labelNodePoolApplyJob:     "true",
				labelResourceGeneration:   fmt.Sprintf("%d", obj.Generation),
			},
			"annotations": map[string]string{
				annotationNodesChecksum: checksum,
			},
			"pod-annotations": fn.MapMerge(fn.FilterObservabilityAnnotations(obj.GetAnnotations()), map[string]string{
				constants.ObservabilityAccountNameKey: r.Env.AccountName,
				constants.ObservabilityClusterNameKey: r.Env.ClusterName,
			}),

			"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},

			"job-node-selector": constants.K8sMasterNodeSelector,

			"nodepool-name":            obj.Name,
			"tfstate-secret-namespace": r.Env.TFStateSecretNamespace,

			"iac-job-image": r.Env.IACJobImage,

			"values.json": string(varfileJson),
		})
		if err != nil {
			return fail(err).Err(nil)
		}

		rr, err := r.yamlClient.ApplyYAML(ctx, b)
		if err != nil {
			return fail(err).Err(nil)
		}

		req.AddToOwnedResources(rr...)
		req.Logger.Infof("waiting for job to be created")
		return req.Done().RequeueAfter(1 * time.Second)
	}

	isMyJob := job.Labels[labelResourceGeneration] == fmt.Sprintf("%d", obj.Generation) && job.Labels[labelNodePoolApplyJob] == "true" && job.Annotations[annotationNodesChecksum] == checksum

	if !isMyJob {
		if !job_manager.HasJobFinished(ctx, r.Client, job) {
			return fail(fmt.Errorf("waiting for previous jobs to finish execution"))
		}

		if err := job_manager.DeleteJob(ctx, r.Client, job.Namespace, job.Name); err != nil {
			return fail(err)
		}

		return req.Done().RequeueAfter(1 * time.Second)
	}

	if !job_manager.HasJobFinished(ctx, r.Client, job) {
		return fail(fmt.Errorf("waiting for job to finish execution")).Err(nil)
	}

	jobSucceeded := job.Status.Succeeded > 0

	if jobSucceeded {
		if err := deleteFinalizersOnNodes(ctx, r.Client, fn.MapValues(markedForDeletion), nodeFinalizer); err != nil {
			return fail(err)
		}
		if err := deleteNodes(ctx, r.Client, fn.MapValues(markedForDeletion)...); err != nil {
			return fail(err)
		}
	}

	check.Status = jobSucceeded

	if !check.Status {
		check.Message = job_manager.GetTerminationLog(ctx, r.Client, job.Namespace, job.Name)
	}
	if check != obj.Status.Checks[checkName] {
		obj.Status.Checks[checkName] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	if !check.Status {
		return req.Done()
	}

	return req.Next()
}

func (r *Reconciler) toAWSVarfileJson(obj *clustersv1.NodePool, nodesMap map[string]clustersv1.NodeProps) (string, error) {
	if obj.Spec.AWS == nil {
		return "", fmt.Errorf(".spec.aws is nil")
	}

	ec2Nodepools := make(map[string]any, 1)
	spotNodepools := make(map[string]any, 1)

	switch obj.Spec.AWS.PoolType {
	case clustersv1.AWSPoolTypeEC2:
		{
			ec2Nodepools[obj.Name] = map[string]any{
				// "image_id":             obj.Spec.AWS.ImageId,
				// "image_ssh_username":   obj.Spec.AWS.ImageSSHUsername,
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
				// "image_id":                     obj.Spec.AWS.ImageId,
				// "image_ssh_username":           obj.Spec.AWS.ImageSSHUsername,
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

	if r.Env.AWSVpcId == "" || r.Env.AWSVpcPublicSubnets == "" {
		return "", fmt.Errorf("env-var AWS_VPC_ID or AWS_VPC_PUBLIC_SUBNETS is not set, aborting nodepool job")
	}

	var publicsubnets map[string]any
	if err := json.Unmarshal([]byte(r.Env.AWSVpcPublicSubnets), &publicsubnets); err != nil {
		return "", err
	}

	variables := map[string]any{
		// INFO: there will be no aws_access_key, aws_secret_key thing, as we expect this autoscaler to run on AWS instances configured with proper IAM instance profile
		// "aws_access_key":             nil,
		// "aws_secret_key":             nil,
		"aws_region":                 r.Env.CloudProviderRegion,
		"tracker_id":                 fmt.Sprintf("cluster-%s", r.Env.ClusterName),
		"k3s_join_token":             r.Env.K3sJoinToken,
		"k3s_server_public_dns_host": r.Env.K3sServerPublicHost,
		"ec2_nodepools":              ec2Nodepools,
		"spot_nodepools":             spotNodepools,
		"extra_agent_args": []string{
			"--snapshotter",
			"stargz",
		},
		"save_ssh_key_to_path": "",
		"tags": map[string]string{
			"kloudlite-account": r.Env.AccountName,
			"kloudlite-cluster": r.Env.ClusterName,
		},

		"vpc": map[string]any{
			"vpc_id":                r.Env.AWSVpcId,
			"vpc_public_subnet_ids": publicsubnets,
		},

		"kloudlite_release": r.Env.KloudliteRelease,
	}

	b, err := json.Marshal(variables)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// func (r *Reconciler) getAccessAndSecretKey(ctx context.Context, obj *clustersv1.NodePool) (accessKey string, secretKey string, err error) {
// 	s, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.IAC.AccessKey.Namespace, obj.Spec.IAC.AccessKey.Name), &corev1.Secret{})
// 	if err != nil {
// 		return "", "", err
// 	}
//
// 	accessKey = string(s.Data[obj.Spec.IAC.AccessKey.Key])
//
// 	s, err = rApi.Get(ctx, r.Client, fn.NN(obj.Spec.IAC.SecretKey.Namespace, obj.Spec.IAC.SecretKey.Name), &corev1.Secret{})
// 	if err != nil {
// 		return "", "", err
// 	}
//
// 	secretKey = string(s.Data[obj.Spec.IAC.SecretKey.Key])
//
// 	return accessKey, secretKey, nil
// }

func (r *Reconciler) parseSpecToVarFileJson(ctx context.Context, obj *clustersv1.NodePool, nodesMap map[string]clustersv1.NodeProps) (string, error) {
	var poolList clustersv1.NodePoolList
	if err := r.List(ctx, &poolList); err != nil {
		return "", client.IgnoreNotFound(err)
	}

	switch obj.Spec.CloudProvider {
	case ct.CloudProviderAWS:
		return r.toAWSVarfileJson(obj, nodesMap)
	default:
		// accessKey, secretKey, err := r.getAccessAndSecretKey(ctx, obj)
		return "", fmt.Errorf("unsupported cloud provider: %s", obj.Spec.CloudProvider)
	}
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig(), kubectl.YAMLClientOpts{Logger: r.logger})

	var err error
	r.templateNodePoolJob, err = templates.Read(templates.NodepoolJob)
	if err != nil {
		return err
	}

	r.templateNamespaceRBAC, err = templates.Read(templates.NodepoolJobNamespaceRBAC)
	if err != nil {
		return err
	}

	builder := ctrl.NewControllerManagedBy(mgr).For(&clustersv1.NodePool{})

	watches := []client.Object{
		&clustersv1.Node{},
		&batchv1.Job{},
	}

	for _, obj := range watches {
		builder.Watches(
			obj,
			handler.EnqueueRequestsFromMapFunc(
				func(_ context.Context, o client.Object) []reconcile.Request {
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
