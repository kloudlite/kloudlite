package target_cluster

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
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
	ClusterDeleted string = "cluster-deleted"
	IpsUpToDate    string = "ips-up-to-date"
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

	if step := req.EnsureChecks(IpsUpToDate, ClusterDeleted); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureIpsUpdated(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	req.Object.Status.LastReconcileTime = &metav1.Time{Time: time.Now()}

	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) finalize(req *rApi.Request[*clustersv1.Cluster]) stepResult.Result {

	ctx, obj, _ := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	failed := func(e error) stepResult.Result {
		return req.CheckFailed(ClusterDeleted, check, e.Error())
	}

	if err := func() error {
		var nodePools clustersv1.NodePoolList
		if err := r.List(ctx, &nodePools, &client.ListOptions{}); err != nil {
			return err
		}

		if len(nodePools.Items) == 0 {
			return nil
		}

		count := 0
		for _, np := range nodePools.Items {
			if np.GetDeletionTimestamp() == nil {
				count += 1
				if err := r.Delete(ctx, &np, &client.DeleteOptions{}); err != nil {
					return err
				}
			}
		}

		if count >= 0 {
			return fmt.Errorf("%d nodepools initiated for deletion", count)
		}

		return fmt.Errorf("nodepools are still under deletion")
	}(); err != nil {
		return failed(err)
	}

	return req.Finalize()
}

func (r *Reconciler) ensureIpsUpdated(req *rApi.Request[*clustersv1.Cluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(IpsUpToDate)
	defer req.LogPostCheck(IpsUpToDate)

	failed := func(e error) stepResult.Result {
		return req.CheckFailed(IpsUpToDate, check, e.Error())
	}

	var nodes corev1.NodeList

	if err := r.List(ctx, &nodes, &client.ListOptions{}); err != nil {
		return failed(err)
	}

	var ips []string

	for _, n := range nodes.Items {
		if ip, ok := n.Labels[constants.PublicIpKey]; ok {
			ips = append(ips, ip)
		} else {
			fmt.Printf("ip not found for the node %s", n.Name)
		}
	}

	isEqual := func(a, b []string) bool {
		sort.Strings(a)
		sort.Strings(b)
		if len(a) != len(b) {
			return false
		}

		return reflect.DeepEqual(a, b)
	}

	if !isEqual(ips, obj.Spec.NodeIps) {
		obj.Spec.NodeIps = ips
		if err := r.Update(ctx, obj, &client.UpdateOptions{}); err != nil {
			return failed(err)
		}
		return req.Next()
	}

	check.Status = true
	if check != checks[IpsUpToDate] {
		checks[IpsUpToDate] = check
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
		&source.Kind{Type: &corev1.Node{}},
		handler.EnqueueRequestsFromMapFunc(
			func(_ client.Object) []reconcile.Request {
				var clist clustersv1.ClusterList
				if err := r.List(context.TODO(), &clist, &client.ListOptions{}); err != nil {
					fmt.Println(err.Error())
					return nil
				}

				if len(clist.Items) == 0 {
					fmt.Println("expected at least one cluster resource to be there")
					return nil
				}

				return []reconcile.Request{{NamespacedName: fn.NN("", clist.Items[0].Name)}}
			}),
	)

	return builder.Complete(r)
}
