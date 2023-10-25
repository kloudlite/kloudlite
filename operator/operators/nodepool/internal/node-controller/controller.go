package node_controller

import (
	"context"
	"time"

	"github.com/kloudlite/operator/operators/nodepool/internal/env"

	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
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
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	deleteAfterTimestamp = "kloudlite.io/delete-after-timestamp"
	trackCorev1Node      = "track-corev1-node"

	nodeFinalizer = "finalizers.kloudlite.io/node"
)

// +kubebuilder:rbac:groups=clusters.kloudlite.io,resources=clusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=clusters.kloudlite.io,resources=clusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=clusters.kloudlite.io,resources=clusters/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, request.NamespacedName, &clustersv1.Node{})
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

	if step := r.keepTrackOfCorev1Node(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, nil
}

func (r *Reconciler) keepTrackOfCorev1Node(req *rApi.Request[*clustersv1.Node]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(trackCorev1Node)
	defer req.LogPostCheck(trackCorev1Node)

	failWithErr := func(err error) stepResult.Result {
		return req.CheckFailed(trackCorev1Node, check, err.Error())
	}

	cn := &corev1.Node{}
	if err := r.Get(ctx, fn.NN("", obj.Name), cn); err != nil {
		return failWithErr(err)
	}

	if controllerutil.AddFinalizer(cn, nodeFinalizer) {
		if err := r.Update(ctx, cn); err != nil {
			return failWithErr(err)
		}
	}

	if cn.GetDeletionTimestamp() != nil {
		if err := r.Delete(ctx, obj); err != nil {
			return failWithErr(err)
		}
		return req.Done()
	}

	for i := range cn.Status.Conditions {
		if cn.Status.Conditions[i].Type == "Ready" {
			check.Status = cn.Status.Conditions[i].Status == corev1.ConditionTrue
		}
	}

	if check != obj.Status.Checks[trackCorev1Node] {
		obj.Status.Checks[trackCorev1Node] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

func (r *Reconciler) finalize(req *rApi.Request[*clustersv1.Node]) stepResult.Result {
	ctx, obj := req.Context(), req.Object

	checkName := "finalizing"
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	realNode := &corev1.Node{}
	if err := r.Get(ctx, fn.NN("", obj.Name), realNode); err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.CheckFailed(checkName, check, err.Error())
		}
		// INFO: as real corev1.Node not found
		return req.Finalize()
	}

	hasUpdatedNode := false

	v, ok := realNode.Annotations[deleteAfterTimestamp]
	if !ok {
		if realNode.Annotations == nil {
			realNode.Annotations = make(map[string]string, 1)
		}
		realNode.Annotations[deleteAfterTimestamp] = time.Now().Add(1 * time.Minute).Format(time.RFC3339)
	}

	if !realNode.Spec.Unschedulable {
		realNode.Spec.Unschedulable = true
		hasUpdatedNode = true
	}

	if hasUpdatedNode {
		if err := r.Update(ctx, realNode); err != nil {
			return req.CheckFailed(checkName, check, err.Error())
		}
		return req.Done().RequeueAfter(1 * time.Minute)
	}

	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		req.Logger.Infof("Failed to parse deleteAfterTimestamp: %v", err)
		t = time.Now().Add(-1 * time.Minute)
	}

	if t.Before(time.Now()) {
		controllerutil.RemoveFinalizer(realNode, nodeFinalizer)
		if err := r.Update(ctx, realNode); err != nil {
			return req.CheckFailed(checkName, check, err.Error())
		}

		if err := r.Delete(ctx, realNode); err != nil {
			if !apiErrors.IsNotFound(err) {
				return req.CheckFailed(checkName, check, err.Error())
			}
		}
		if err := r.Delete(ctx, obj); err != nil {
			return req.CheckFailed(checkName, check, err.Error())
		}
		return req.Finalize()
	}

	return req.Done()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

	builder := ctrl.NewControllerManagedBy(mgr).For(&clustersv1.Node{})
	builder.Watches(
		&source.Kind{Type: &corev1.Node{}},
		handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []reconcile.Request {
			if v, ok := obj.GetLabels()[constants.NodeNameKey]; ok {
				return []reconcile.Request{{NamespacedName: fn.NN("", v)}}
			}
			return nil
		}),
	)
	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
