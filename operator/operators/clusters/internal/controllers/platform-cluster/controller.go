package platform_cluster

import (
	"context"
	"fmt"
	"time"

	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	"github.com/kloudlite/operator/operators/clusters/internal/controllers/platform-node"
	"github.com/kloudlite/operator/operators/clusters/internal/env"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
)

type Reconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	logger      logging.Logger
	Name        string
	yamlClient  *kubectl.YAMLClient
	Env         *env.Env
	PlatformEnv *env.PlatformEnv
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	ClusterDeleted string = "cluster-deleted"
	ClusterReady   string = "cluster-ready"
)

// +kubebuilder:rbac:groups=clusters.kloudlite.io,resources=clusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=clusters.kloudlite.io,resources=clusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=clusters.kloudlite.io,resources=clusters/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &clustersv1.Cluster{})
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

	if step := req.EnsureChecks(ClusterReady, ClusterDeleted); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureNodesCreated(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	req.Object.Status.LastReconcileTime = &metav1.Time{Time: time.Now()}

	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) finalize(req *rApi.Request[*clustersv1.Cluster]) stepResult.Result {

	_, obj, _ := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	// TODO: have to write deletion logic
	k := "************** ~~>* deletion of cluster is not supported yet *<~~ **************"
	fmt.Println(k)
	return req.CheckFailed(ClusterDeleted, check, k)
}

func (r *Reconciler) ensureNodesCreated(req *rApi.Request[*clustersv1.Cluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(ClusterReady)
	defer req.LogPostCheck(ClusterReady)

	failed := func(e error) stepResult.Result {
		return req.CheckFailed(ClusterReady, check, e.Error())
	}

	masterName := func(nodeSuffix string) string {
		return fmt.Sprintf("kl-master-%s-%s", obj.Name, nodeSuffix)
	}

	// logic to create cluster
	createNodeWithAction := func(nodeSuffix string) error {
		if err := r.Create(ctx, &clustersv1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: masterName(nodeSuffix),
			},
			Spec: clustersv1.NodeSpec{
				ClusterName: obj.Name,
				NodeType: func() clustersv1.NodeType {
					if nodeSuffix == "01" {
						return "cluster"
					}
					return "master"
				}(),
			},
		}); err != nil {
			return err
		}

		return nil
	}

	// master name -> kl-clustername-master01
	// secondary master -> kl-clustername-master02,03...

	cluster, err := rApi.Get(ctx, r.Client, fn.NN("", masterName("01")), &clustersv1.Node{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return failed(err)
		}

		if err := createNodeWithAction("01"); err != nil {
			return failed(err)
		}

		return failed(fmt.Errorf("node %s set to create", masterName("01")))
	}

	if c, ok := cluster.Status.Checks[platform_node.NodeReady]; true {
		if !ok {
			return failed(fmt.Errorf("can't fetch status for the cluster, please wait"))
		}

		if !c.Status {
			return failed(fmt.Errorf("cluster not ready yet reason [ %q ], please wait", c.Message))
		}
	}

	// cluster present check either secondary master required [ha enabled]

	if obj.Spec.AvailablityMode == "HA" {
		// check for second master
		_, err := rApi.Get(ctx, r.Client, fn.NN("", masterName("02")), &clustersv1.Node{})
		if err != nil {
			if !apiErrors.IsNotFound(err) {
				return failed(err)
			}

			if err := createNodeWithAction("02"); err != nil {
				return failed(err)
			}

			return failed(fmt.Errorf("node %s set to create", masterName("02")))
		}

		// check for third master
		_, err = rApi.Get(ctx, r.Client, fn.NN("", masterName("03")), &clustersv1.Node{})
		if err != nil {
			if !apiErrors.IsNotFound(err) {
				return failed(err)
			}

			if err := createNodeWithAction("03"); err != nil {
				return failed(err)
			}

			return failed(fmt.Errorf("node %s set to create", masterName("03")))
		}
	}

	// fetch only without GetDeletionTimestamp

	check.Status = true
	if check != checks[ClusterReady] {
		checks[ClusterReady] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

	builder := ctrl.NewControllerManagedBy(mgr).For(&clustersv1.Cluster{})
	builder.WithEventFilter(rApi.ReconcileFilter())

	builder.Watches(
		&source.Kind{Type: &clustersv1.Node{}},
		handler.EnqueueRequestsFromMapFunc(
			func(obj client.Object) []reconcile.Request {
				if cl, ok := obj.GetLabels()[constants.ClusterNameKey]; ok {
					return []reconcile.Request{{NamespacedName: fn.NN("", cl)}}
				}
				return nil
			}),
	)

	return builder.Complete(r)
}
