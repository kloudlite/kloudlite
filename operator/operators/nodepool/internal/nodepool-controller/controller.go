package nodepool_controller

import (
	"context"
	"encoding/json"
	"fmt"
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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	templateNodePoolJob []byte
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

	if step := req.EnsureFinalizers(constants.CommonFinalizer); !step.ShouldProceed() {
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

	accessKey, err := func() (string, error) {
		s, err2 := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.IAC.CloudProviderAccessKey.Namespace, obj.Spec.IAC.CloudProviderAccessKey.Name), &corev1.Secret{})
		if err2 != nil {
			return "", err2
		}

		return string(s.Data[obj.Spec.IAC.CloudProviderAccessKey.Key]), nil
	}()
	if err != nil {
		return fail(err)
	}

	secretKey, err := func() (string, error) {
		s, err2 := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.IAC.CloudProviderSecretKey.Namespace, obj.Spec.IAC.CloudProviderSecretKey.Name), &corev1.Secret{})
		if err2 != nil {
			return "", err2
		}

		return string(s.Data[obj.Spec.IAC.CloudProviderSecretKey.Key]), nil
	}()
	if err != nil {
		return fail(err)
	}

	varfileJson, err := r.parseSpecToVarFileJson(ctx, obj, nodesMap, accessKey, secretKey)
	if err != nil {
		return fail(err).Err(nil)
	}

	jobName := obj.Spec.IAC.JobName
	jobNamespace := obj.Spec.IAC.JobNamespace

	job := &batchv1.Job{}
	if err := r.Get(ctx, fn.NN(jobNamespace, jobName), job); err != nil {
		job = nil
	}

	if job == nil {
		b, err := templates.ParseBytes(r.templateNodePoolJob, map[string]any{
			"action": "apply",

			"job-name":      jobName,
			"job-namespace": jobNamespace,
			"annotations": map[string]string{
				annotationNodesChecksum: checksum,
			},
			"labels": map[string]string{
				constants.NodePoolNameKey: obj.Name,
				labelNodePoolApplyJob:     "true",
				labelResourceGeneration:   fmt.Sprintf("%d", obj.Generation),
			},
			// "annotations": obj.Annotations,
			"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},

			"job-node-selector": constants.K8sMasterNodeSelector,

			"service-account-name": "",

			"aws-s3-bucket-name":   obj.Spec.IAC.StateS3BucketName,
			"aws-s3-bucket-region": obj.Spec.IAC.StateS3BucketRegion,
			// "aws-s3-bucket-filepath": fmt.Sprintf("%s/%s/%s/nodepools-%s.tfstate", r.Env.IACStateS3BucketDir, r.Env.KloudliteAccountName, r.Env.KloudliteClusterName, obj.Name),
			"aws-s3-bucket-filepath": obj.Spec.IAC.StateS3BucketFilePath,

			"aws-s3-access-key": accessKey,
			"aws-s3-secret-key": secretKey,

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

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig(), kubectl.YAMLClientOpts{Logger: r.logger})

	var err error
	r.templateNodePoolJob, err = templates.ReadNodepoolJobTemplate()
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
