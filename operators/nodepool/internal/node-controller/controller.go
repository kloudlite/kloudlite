package node_controller

import (
	"context"
	"fmt"
	"time"

	"github.com/kloudlite/operator/operators/nodepool/internal/env"

	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/pkg/util/taints"
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
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	deleteAfterTimestamp = "kloudlite.io/delete-after-timestamp"
	trackCorev1Node      = "track-corev1-node"

	selfFinalizer = "kloudlite.io/node.finalizer"

	finalizingNode = "finalizing-node"
)

// +kubebuilder:rbac:groups=clusters.kloudlite.io,resources=clusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=clusters.kloudlite.io,resources=clusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=clusters.kloudlite.io,resources=clusters/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &clustersv1.Node{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	req.PreReconcile()
	defer req.PostReconcile()

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureCheckList([]rApi.CheckMeta{
		{Name: trackCorev1Node, Title: "Update node status"},
	}); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureChecks(trackCorev1Node); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(selfFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.keepTrackOfCorev1Node(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *Reconciler) keepTrackOfCorev1Node(req *rApi.Request[*clustersv1.Node]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(trackCorev1Node, req)

	cn := &corev1.Node{}
	if err := r.Get(ctx, fn.NN("", obj.Name), cn); err != nil {
		return check.Failed(err)
	}

	if updated := controllerutil.AddFinalizer(cn, selfFinalizer); updated {
		if err := r.Update(ctx, cn); err != nil {
			return check.Failed(err)
		}
	}

	if cn.GetDeletionTimestamp() != nil {
		if err := r.Delete(ctx, obj); err != nil {
			return check.Failed(err)
		}
		return req.Done()
	}

	for i := range cn.Status.Conditions {
		if cn.Status.Conditions[i].Type == "Ready" {
			return check.Completed()
		}
	}

	return check.StillRunning(fmt.Errorf("node is not ready"))
}

func (r *Reconciler) finalize(req *rApi.Request[*clustersv1.Node]) stepResult.Result {
	ctx, obj := req.Context(), req.Object

	if step := req.EnsureCheckList([]rApi.CheckMeta{
		{Name: finalizingNode, Title: "cleaning up resources"},
	}); !step.ShouldProceed() {
		return step
	}

	check := rApi.NewRunningCheck(finalizingNode, req)

	node, err := rApi.Get(ctx, r.Client, fn.NN("", obj.Name), &corev1.Node{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return check.Failed(err)
		}
	}

	deleteAfterTimestamp := "kloudlite.io/node.delete-after"

	if node == nil {
		controllerutil.RemoveFinalizer(obj, selfFinalizer)
		if err := r.Update(ctx, obj); err != nil {
			return check.Failed(err)
		}

		return check.Completed()
	}

	taintKey := "kloudlite.io/node.deleting"

	alreadyHasTaint := false

	for i := range node.Spec.Taints {
		if node.Spec.Taints[i].Key == taintKey {
			alreadyHasTaint = true
			break
		}
	}

	if !alreadyHasTaint {
		node, _, err := taints.AddOrUpdateTaint(node, &corev1.Taint{
			Key:    taintKey,
			Value:  "true",
			Effect: "NoExecute",
		})
		if err != nil {
			return check.Failed(err)
		}

		node.Spec.Unschedulable = true
		fn.MapSet(&node.Annotations, deleteAfterTimestamp, time.Now().Add(1*time.Minute).Format(time.RFC3339))
		if err := r.Update(ctx, node); err != nil {
			return check.Failed(err)
		}
	}

	t, err := time.Parse(time.RFC3339, node.Annotations[deleteAfterTimestamp])
	if err != nil {
		req.Logger.Infof("Failed to parse deleteAfterTimestamp: %v", err)
		t = time.Now().Add(-1 * time.Minute)
	}

	if len(obj.Finalizers) == 1 && obj.Finalizers[0] == selfFinalizer && time.Now().After(t) {
		// it's time to delete the underlying corev1.Node
		controllerutil.RemoveFinalizer(node, selfFinalizer)
		if err := r.Update(ctx, node); err != nil {
			return check.Failed(err)
		}

		if err := r.Delete(ctx, node); err != nil {
			if !apiErrors.IsNotFound(err) {
				return check.Failed(err)
			}
		}

		controllerutil.RemoveFinalizer(obj, selfFinalizer)
		if err := r.Update(ctx, obj); err != nil {
			return check.Failed(err)
		}
		return check.Completed()
	}

	return check.StillRunning(fmt.Errorf("node will be deleted at %s", t.Format(time.RFC3339))).NoRequeue().RequeueAfter(time.Since(t))
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig(), kubectl.YAMLClientOpts{Logger: r.logger})

	builder := ctrl.NewControllerManagedBy(mgr).For(&clustersv1.Node{})

	builder.Watches(
		&corev1.Node{},
		handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
			return []reconcile.Request{{NamespacedName: fn.NN("", obj.GetName())}}
		}),
	)

	// builder.Watches(
	// 	&source.Kind{Type: &corev1.Node{}},
	// 	handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []reconcile.Request {
	// 		if v, ok := obj.GetLabels()[constants.NodeNameKey]; ok {
	// 			return []reconcile.Request{{NamespacedName: fn.NN("", v)}}
	// 		}
	// 		return nil
	// 	}),
	// )

	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
