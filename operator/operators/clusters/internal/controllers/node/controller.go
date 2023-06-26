package node

import (
	"context"
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
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
	"github.com/kloudlite/operator/pkg/templates"
)

// have to fetch these from env
const (
	tfTemplates string = "./templates/terraform"
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
	K8sNodeCreated string = "k8s-node-created"
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

	if step := req.EnsureChecks(K8sNodeCreated); !step.ShouldProceed() {
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

	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	failed := func(e error) stepResult.Result {
		return req.CheckFailed("fail in ensure nodes", check, e.Error())
	}

	np, err := rApi.Get(ctx, r.Client, fn.NN("", obj.Spec.NodePoolName), &clustersv1.NodePool{})
	if err != nil {
		return failed(err)
	}

	nodeConfig, err := r.getNodeConfig(np, obj)
	if err != nil {
		return failed(err)
	}

	providerConfig, err := getProviderConfig()
	if err != nil {
		return failed(err)
	}

	// action := getAction(obj)

	sProvider, err := r.getSpecificProvierConfig()
	if err != nil {
		return failed(err)
	}

	createDeleteNodeJob := func() error {
		jobYaml, err := templates.Parse(templates.Clusters.Job,
			map[string]any{
				"name":      fmt.Sprintf("delete-%s", obj.Name),
				"namespace": r.Env.JobNamespace,
				"ownerRefs": []metav1.OwnerReference{fn.AsOwner(obj)},

				"cloudProvider":  r.Env.CloudProvider,
				"action":         "delete",
				"nodeConfig":     string(nodeConfig),
				"providerConfig": string(providerConfig),

				"AwsProvider":   string(sProvider),
				"AzureProvider": string(sProvider),
				"DoProvider":    string(sProvider),
				"GCPProvider":   string(sProvider),
			},
		)
		if err != nil {
			return err
		}

		if _, err := r.yamlClient.ApplyYAML(ctx, jobYaml); err != nil {
			return err
		}

		return nil
	}

	j, err := rApi.Get(ctx, r.Client, fn.NN(r.Env.JobNamespace, fmt.Sprintf("delete-%s", obj.Name)), &batchv1.Job{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return failed(err)
		}

		if err := createDeleteNodeJob(); err != nil {
			return failed(err)
		}
	}

	if j != nil && j.Status.Succeeded >= 1 {
		return req.Finalize()
	}

	return req.Done()
}

func (r *Reconciler) ensureNodeReady(req *rApi.Request[*clustersv1.Node]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	failed := func(err error) stepResult.Result {
		return req.CheckFailed("ensure-node-ready", check, err.Error())
	}

	req.LogPreCheck(K8sNodeCreated)
	defer req.LogPostCheck(K8sNodeCreated)

	np, err := rApi.Get(ctx, r.Client, fn.NN("", obj.Spec.NodePoolName), &clustersv1.NodePool{})
	if err != nil {
		return failed(err)
	}

	nodeConfig, err := r.getNodeConfig(np, obj)
	if err != nil {
		return failed(err)
	}

	providerConfig, err := getProviderConfig()
	if err != nil {
		return failed(err)
	}

	action := getAction(obj)

	sProvider, err := r.getSpecificProvierConfig()
	if err != nil {
		return failed(err)
	}

	createNode := func() error {
		jobYaml, err := templates.Parse(templates.Clusters.Job,
			map[string]any{
				"name":      obj.Name,
				"namespace": r.Env.JobNamespace,
				"ownerRefs": []metav1.OwnerReference{fn.AsOwner(obj)},

				"cloudProvider":  r.Env.CloudProvider,
				"action":         action,
				"nodeConfig":     nodeConfig,
				"providerConfig": providerConfig,

				"AwsProvider":   sProvider,
				"AzureProvider": sProvider,
				"DoProvider":    sProvider,
				"GCPProvider":   sProvider,
			},
		)
		if err != nil {
			return err
		}

		if _, err := r.yamlClient.ApplyYAML(ctx, jobYaml); err != nil {
			return err
		}

		return nil
	}

	// do your actions here
	if err := func() error {
		if _, err := rApi.Get(ctx, r.Client, fn.NN("", obj.Name), &corev1.Node{}); err != nil {
			if !apiErrors.IsNotFound(err) {
				return err
			}
			// not found do your action

			// check node job if not created create
			if _, e := rApi.Get(ctx, r.Client, fn.NN(r.Env.JobNamespace, obj.Name), &batchv1.Job{}); e != nil {
				if !apiErrors.IsNotFound(e) {
					return e
				}

				if err := createNode(); err != nil {
					return err
				}
			}
		}

		return nil
	}(); err != nil {
		return failed(err)
	}

	// check nodejob

	if err := func() error {
		if nodeJob, e := rApi.Get(ctx, r.Client, fn.NN(r.Env.JobNamespace, obj.Name), &batchv1.Job{}); e != nil {
			if !apiErrors.IsNotFound(e) {
				return e
			}
		} else if nodeJob.Status.Succeeded >= 1 {
			if _, err := rApi.Get(ctx, r.Client, fn.NN("", obj.Name), &corev1.Node{}); err != nil {
				return err
			} else {
				if err := r.Delete(ctx, &batchv1.Job{
					ObjectMeta: metav1.ObjectMeta{
						Name:      obj.Name,
						Namespace: r.Env.JobNamespace,
					},
				}); err != nil {
					return err
				}
			}
		}

		return nil
	}(); err != nil {
		return failed(err)
	}

	// check node attached
	// if not attached then attach then have to attach

	check.Status = true
	if check != checks[K8sNodeCreated] {
		checks[K8sNodeCreated] = check
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

	builder.Watches(
		&source.Kind{Type: &batchv1.Job{}},
		handler.EnqueueRequestsFromMapFunc(
			func(obj client.Object) []reconcile.Request {
				if _, ok := obj.GetLabels()["kloudlite.io/is-nodectrl-job"]; ok {
					return []reconcile.Request{{NamespacedName: fn.NN("", obj.GetName())}}
				}
				return nil
			}))

	return builder.Complete(r)
}
