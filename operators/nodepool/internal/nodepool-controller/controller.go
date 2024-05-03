package nodepool_controller

import (
	"context"
	"fmt"
	"strings"
	"time"

	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
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

const (
	updateNodeTaintsAndLabels = "update-node-taints-and-labels"
	ensureJobNamespaceRBACs   = "ensure-job-namespace-rbacs"
	syncNodepool              = "sync-nodepool"

	deletingNodepool = "delete-nodepool"
)

var (
	ApplyChecklist = []rApi.CheckMeta{
		{Name: updateNodeTaintsAndLabels, Title: "Update Node Taints and Labels"},
		{Name: ensureJobNamespaceRBACs, Title: "Configure Job Namespace RBACs"},
		{Name: syncNodepool, Title: "Syncing Nodepool"},
	}
	DeleteChecklist = []rApi.CheckMeta{
		{Name: deletingNodepool, Title: "Deleting Nodepool"},
	}
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

	if step := req.EnsureCheckList(ApplyChecklist); !step.ShouldProceed() {
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
	check := rApi.NewRunningCheck(deletingNodepool, req)

	if step := req.EnsureCheckList(DeleteChecklist); !step.ShouldProceed() {
		return step
	}

	nodes, err := nodesBelongingToNodepool(ctx, r.Client, obj.Name)
	if err != nil {
		return check.Failed(err)
	}

	if err := deleteNodes(ctx, r.Client, nodes...); err != nil {
		return check.Failed(err)
	}

	if step := req.CleanupOwnedResources(); !step.ShouldProceed() {
		return step
	}

	if err := deleteFinalizersOnNodes(ctx, r.Client, nodes, nodeFinalizer); err != nil {
		return check.Failed(err)
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
	check.State = rApi.CompletedState
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
	check := rApi.Check{Generation: obj.Generation, State: rApi.RunningState}

	req.LogPreCheck(updateNodeTaintsAndLabels)
	defer req.LogPostCheck(updateNodeTaintsAndLabels)

	fail := func(err error) stepResult.Result {
		return req.CheckFailed(updateNodeTaintsAndLabels, check, err.Error())
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
	check.State = rApi.CompletedState
	if check != obj.Status.Checks[updateNodeTaintsAndLabels] {
		fn.MapSet(&obj.Status.Checks, updateNodeTaintsAndLabels, check)
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

func (r *Reconciler) ensureJobNamespaceRBACs(req *rApi.Request[*clustersv1.NodePool]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation, State: rApi.RunningState}

	req.LogPreCheck(ensureJobNamespaceRBACs)
	defer req.LogPostCheck(ensureJobNamespaceRBACs)

	fail := func(err error) stepResult.Result {
		return req.CheckFailed(ensureJobNamespaceRBACs, check, err.Error())
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
	check.State = rApi.CompletedState
	if check != obj.Status.Checks[ensureJobNamespaceRBACs] {
		fn.MapSet(&obj.Status.Checks, ensureJobNamespaceRBACs, check)
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

func (r *Reconciler) syncNodepool(req *rApi.Request[*clustersv1.NodePool]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(syncNodepool, req)

	nodes, err := nodesBelongingToNodepool(ctx, r.Client, obj.Name)
	if err != nil {
		return check.StillRunning(err)
	}

	if err := addFinalizersOnNodes(ctx, r.Client, nodes, nodeFinalizer); err != nil {
		return check.StillRunning(err)
	}

	markedForDeletion := filterNodesMarkedForDeletion(nodes)

	nodesMap := make(map[string]clustersv1.NodeProps, len(nodes)-len(markedForDeletion))
	for i := range nodes {
		rawName := nodes[i].GetName()[len(obj.GetName())+1:]
		if _, ok := markedForDeletion[nodes[i].GetName()]; !ok {
			nodesMap[rawName] = clustersv1.NodeProps{}
		}
	}

	// checksum := nodesChecksum(nodesMap)

	varfileJson, err := r.parseSpecToVarFileJson(ctx, obj, nodesMap)
	if err != nil {
		return check.Failed(err)
	}

	jobRef := strings.Split(obj.Annotations[annNodepoolJobRef], "/")

	jobNamespace := jobRef[0]
	jobName := jobRef[1]

	b, err := templates.ParseBytes(r.templateNodePoolJob, templates.NodepoolJobVars{
		JobMetadata: metav1.ObjectMeta{
			Name:            jobName,
			Namespace:       jobNamespace,
			Labels:          obj.GetLabels(),
			Annotations:     fn.FilterObservabilityAnnotations(obj.GetAnnotations()),
			OwnerReferences: []metav1.OwnerReference{fn.AsOwner(obj, true)},
		},
		NodeSelector:         constants.K8sMasterNodeSelector,
		JobImage:             r.Env.IACJobImage,
		TFWorkspaceName:      obj.Name,
		TfWorkspaceNamespace: r.Env.TFStateSecretNamespace,
		CloudProvider:        string(obj.Spec.CloudProvider),
		ValuesJSON:           varfileJson,
	})
	if err != nil {
		return check.Failed(err)
	}

	rr, err := r.yamlClient.ApplyYAML(ctx, b)
	if err != nil {
		return check.StillRunning(err)
	}

	req.AddToOwnedResources(rr...)

	job, err := rApi.Get(ctx, r.Client, fn.NN(jobNamespace, jobName), &crdsv1.Job{})
	if err != nil {
		return check.Failed(err)
	}

	if !job.HasCompleted() {
		return check.StillRunning(fmt.Errorf("waiting for job to complete"))
	}

	if job.Status.Phase == crdsv1.JobPhaseFailed {
		return check.Failed(fmt.Errorf("job failed"))
	}

	if markedForDeletion != nil {
		if err := deleteFinalizersOnNodes(ctx, r.Client, fn.MapValues(markedForDeletion), nodeFinalizer); err != nil {
			return check.Failed(err)
		}
		if err := deleteNodes(ctx, r.Client, fn.MapValues(markedForDeletion)...); err != nil {
			return check.Failed(err)
		}
	}

	return check.Completed()
}

func (r *Reconciler) parseSpecToVarFileJson(ctx context.Context, obj *clustersv1.NodePool, nodesMap map[string]clustersv1.NodeProps) (string, error) {
	var poolList clustersv1.NodePoolList
	if err := r.List(ctx, &poolList); err != nil {
		return "", client.IgnoreNotFound(err)
	}

	switch obj.Spec.CloudProvider {
	case ct.CloudProviderAWS:
		return r.AwsJobValuesJson(obj, nodesMap)
		// return r.toAWSVarfileJson(obj, nodesMap)
	case ct.CloudProviderGCP:
		return r.GCPJobValuesJson(obj, nodesMap)
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
