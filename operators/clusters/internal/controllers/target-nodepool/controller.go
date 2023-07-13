package target_nodepool

import (
	"context"
	"fmt"
	"time"

	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiLabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	"github.com/kloudlite/operator/operators/clusters/internal/env"
	"github.com/kloudlite/operator/pkg/constants"
	"github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	logger     logging.Logger
	Name       string
	yamlClient *kubectl.YAMLClient
	Env        *env.Env
	TargetEnv  *env.TargetEnv
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	K8sNodePoolCreated string = "k8s-nodepool-created"
	NodePoolDeletion   string = "nodepool-deletion"
)

// +kubebuilder:rbac:groups=clusters.kloudlite.io,resources=nodepools,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=clusters.kloudlite.io,resources=nodepools/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=clusters.kloudlite.io,resources=nodepools/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &clustersv1.NodePool{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	req.LogPreReconcile()
	defer req.LogPostReconcile()

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureChecks(K8sNodePoolCreated); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureNodesAsPerReq(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	req.Object.Status.LastReconcileTime = &metav1.Time{Time: time.Now()}

	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) finalize(req *rApi.Request[*clustersv1.NodePool]) stepResult.Result {

	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	failed := func(e error) stepResult.Result {
		return req.CheckFailed(NodePoolDeletion, check, e.Error())
	}

	//  fetch nodees of clusterv1
	var nodes clustersv1.NodeList
	if err := r.List(ctx, &nodes, &client.ListOptions{
		LabelSelector: apiLabels.SelectorFromValidatedSet(
			apiLabels.Set{constants.NodePoolKey: obj.Name},
		),
	}); err != nil {
		if !apiErrors.IsNotFound(err) {
			return failed(err)
		}
		// no nodes present finalize it
		return req.Finalize()
	}

	// if no nodes present finalize it
	if len(nodes.Items) == 0 {
		return req.Finalize()
	}

	//  have to delete one by one
	for _, n := range nodes.Items {
		if err := r.Delete(ctx, &n); err != nil {
			return failed(err)
		}
	}

	return failed(fmt.Errorf("nodes are set to delete and waiting to be deleted"))
}

func (r *Reconciler) ensureNodesAsPerReq(req *rApi.Request[*clustersv1.NodePool]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(K8sNodePoolCreated)
	defer req.LogPostCheck(K8sNodePoolCreated)

	failed := func(e error) stepResult.Result {
		return req.CheckFailed("fail in ensure nodes", check, e.Error())
	}

	// fetch all nodes and check either it is same as target or not, if not do the needful
	var nodes clustersv1.NodeList
	if err := r.List(ctx, &nodes, &client.ListOptions{
		LabelSelector: apiLabels.SelectorFromValidatedSet(
			apiLabels.Set{constants.NodePoolKey: obj.Name},
		),
	}); err != nil {
		return failed(err)
	}

	length := len(nodes.Items)
	rLength := 0

	// fetch only without GetDeletionTimestamp
	for _, n := range nodes.Items {
		if n.GetDeletionTimestamp() == nil {
			rLength += 1
		}
	}

	if length < obj.Spec.TargetCount {
		for i := length; i < obj.Spec.TargetCount; i++ {
			if err := r.Create(ctx, &clustersv1.Node{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "kl-worker-",
				},
				Spec: clustersv1.NodeSpec{
					NodePoolName: obj.Name,
					NodeType:     "worker",
				},
			}); err != nil {
				return failed(err)
			}
		}
	} else if (rLength > obj.Spec.TargetCount) && (length > 0) {
		// needs to delete
		n := ""
		for _, n2 := range nodes.Items {
			if n2.GetDeletionTimestamp() == nil {
				n = n2.Name
				break
			}
		}

		if n == "" {
			return failed(fmt.Errorf("no nodes found without deletion timestamp to delete"))
		}

		if err := r.Delete(
			ctx, &clustersv1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: n,
				},
			},
		); err != nil {
			return failed(err)
		}
	}

	check.Status = true
	if check != checks[K8sNodePoolCreated] {
		checks[K8sNodePoolCreated] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

	builder := ctrl.NewControllerManagedBy(mgr).For(&clustersv1.NodePool{})
	builder.WithEventFilter(rApi.ReconcileFilter())

	builder.Watches(
		&source.Kind{Type: &clustersv1.Node{}},
		handler.EnqueueRequestsFromMapFunc(
			func(obj client.Object) []reconcile.Request {
				if np, ok := obj.GetLabels()[constants.NodePoolKey]; ok {
					return []reconcile.Request{{NamespacedName: functions.NN("", np)}}
				}
				return nil
			}),
	)

	return builder.Complete(r)
}
