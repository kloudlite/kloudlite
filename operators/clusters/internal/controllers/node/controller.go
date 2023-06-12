package node

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	"github.com/kloudlite/operator/operators/clusters/internal/env"
	"github.com/kloudlite/operator/pkg/constants"
	"github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
)

// have to fetch these from env
const (
	tfTemplates  string = ""
	accountName  string = "sample-account"
	accessKey    string = "accessKey"
	accessSecret string = "accessSecret"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	logger     logging.Logger
	Name       string
	yamlClient *kubectl.YAMLClient
	Env        *env.Env
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	K8sSecretCreated string = "k8s-secret-created"
)

// +kubebuilder:rbac:groups=clusters.kloudlite.io,resources=nodes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=clusters.kloudlite.io,resources=nodes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=clusters.kloudlite.io,resources=nodes/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &clustersv1.Node{})
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

	if step := req.RestartIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureChecks(K8sSecretCreated); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureNodeReady(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	req.Object.Status.LastReconcileTime = &metav1.Time{Time: time.Now()}
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) finalize(req *rApi.Request[*clustersv1.Node]) stepResult.Result {
	// return req.Finalize()
	// finalize only if node deleted
	return nil
}

func (r *Reconciler) ensureNodeReady(req *rApi.Request[*clustersv1.Node]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(K8sSecretCreated)
	defer req.LogPostCheck(K8sSecretCreated)

	// do your actions here
	if err := func() error {
		mNode, err := rApi.Get(ctx, r.Client, functions.NN("", obj.Name), &corev1.Node{})
		if err != nil {
			if !apiErrors.IsNotFound(err) {
				return err
			}
			// not found do your action
		}

		fmt.Println(mNode)
		return nil
	}(); err != nil {
		return req.CheckFailed("failed", check, err.Error())
	}

	// check node attached
	// if not attached then attach then have to attach

	// checking if node attach
	if err := func() error {
		// mNode := &corev1.Node{}
		// if err := r.Get(ctx, functions.NN("", obj.Name), mNode); err != nil {
		// 	if !apiErrors.IsNotFound(err) {
		// 		return err
		// 	}
		// 	// not found do your action
		// }

		// mNode, err := rApi.Get(ctx, r.Client, functions.NN("", obj.Name), &corev1.Node{})
		// if err != nil {
		// 	if !apiErrors.IsNotFound(err) {
		// 		return err
		// 	}
		// 	// not found do your action
		// }

		ymlFile := []byte("")

		/*
			needed
			env:
			 access sec, key, provider[aws, gcp]
			 accountName

			 tfTemplates
			 cr node labels: -> core node labels
			 cr node spec taints -> core node tains
			 sshPath -> env
		*/

		if _, err := r.yamlClient.ApplyYAML(ctx, ymlFile); err != nil {
			return err
		}

		return nil
	}(); err != nil {
		fmt.Println(err)
		panic(err)
	}

	check.Status = true
	if check != checks[K8sSecretCreated] {
		checks[K8sSecretCreated] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

	builder := ctrl.NewControllerManagedBy(mgr).For(&clustersv1.Node{})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
