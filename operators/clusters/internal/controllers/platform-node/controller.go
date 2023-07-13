package platform_node

import (
	"context"
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
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

const (
	tfTemplates string = "./templates/terraform"
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
	NodeReady   string = "node-ready"
	NodeDeleted string = "node-deleted-successfully"
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

	if step := req.EnsureChecks(NodeReady, NodeDeleted); !step.ShouldProceed() {
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

	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	failed := func(e error) stepResult.Result {
		return req.CheckFailed(NodeDeleted, check, e.Error())
	}

	getDeleteAction := func() string {
		if s, ok := obj.Labels[constants.ForceDeleteKey]; ok && s == "true" {
			return "force-delete"
		}
		return "delete"
	}

	getDeleteJobName := func() string {
		return fmt.Sprintf("delete-%s", obj.Name)
	}

	if err := func() error {
		jb, err := rApi.Get(ctx, r.Client, fn.NN(r.Env.JobNamespace, getDeleteJobName()), &batchv1.Job{})

		if err != nil {
			if !apiErrors.IsNotFound(err) {
				return err
			}

			if c, ok := checks[NodeDeleted]; ok && c.Status {
				return nil
			}

			if err := r.createJob(req, getDeleteAction(), getDeleteJobName()); err != nil {
				return err
			}

			return fmt.Errorf("deletion initiated")
		}

		if jb.Status.Succeeded >= 1 {
			err := r.Delete(ctx, jb)
			return err
		}

		return fmt.Errorf("deletion in progress")
	}(); err != nil {
		return failed(err)
	}

	return req.Finalize()
}

func (r *Reconciler) createJob(req *rApi.Request[*clustersv1.Node], action string, name string) error {

	ctx, obj := req.Context(), req.Object

	// fetch the cluster
	cl, err := rApi.Get(ctx, r.Client, fn.NN("", obj.Spec.ClusterName), &clustersv1.Cluster{})
	if err != nil {
		return err
	}

	nodeConfig, err := r.getNodeConfig(cl, obj)
	if err != nil {
		return err
	}

	providerConfig, err := getProviderConfig()
	if err != nil {
		return err
	}

	sProvider, err := getSpecificProvierConfig(ctx, r.Client, cl)
	if err != nil {
		return err
	}

	jobYaml, err := templates.Parse(templates.Clusters.Job,
		map[string]any{
			"name":      name,
			"namespace": r.Env.JobNamespace,
			"ownerRefs": []metav1.OwnerReference{fn.AsOwner(obj)},

			"cloudProvider":  cl.Spec.CloudProvider,
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

func (r *Reconciler) ensureNodeReady(req *rApi.Request[*clustersv1.Node]) stepResult.Result {

	req.LogPreCheck(NodeReady)
	defer req.LogPostCheck(NodeReady)

	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	failed := func(err error) stepResult.Result {
		return req.CheckFailed(NodeReady, check, err.Error())
	}

	// fetch job

	if err := func() error {
		nodeJob, e := rApi.Get(ctx, r.Client, fn.NN(r.Env.JobNamespace, obj.Name), &batchv1.Job{})
		if e == nil {
			if nodeJob.Status.Succeeded >= 1 {
				return r.Delete(ctx, &batchv1.Job{
					ObjectMeta: metav1.ObjectMeta{
						Name:      obj.Name,
						Namespace: r.Env.JobNamespace,
					},
				})
			}

			return fmt.Errorf("creation under progress.")
		}

		if !apiErrors.IsNotFound(e) {
			return e
		}

		if checks[NodeReady].Status {
			if rc, ok := obj.Labels[constants.RecheckClusterKey]; ok && rc == "true" {
				if err := r.createJob(req, getAction(obj), obj.Name); err != nil {
					return err
				}

				obj.Labels[constants.RecheckClusterKey] = "false"
				if err := r.Update(ctx, obj); err != nil {
					return err
				}

				return fmt.Errorf("recheck running")
			}
			return nil
		}

		if err := r.createJob(req, getAction(obj), obj.Name); err != nil {
			return err
		}

		return fmt.Errorf("creation under progress..")
	}(); err != nil {
		return failed(err)
	}

	check.Status = true
	if check != checks[NodeReady] {
		checks[NodeReady] = check
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
				if _, ok := obj.GetLabels()[constants.IsNodeControllerJob]; ok {
					return []reconcile.Request{{NamespacedName: fn.NN("", obj.GetName())}}
				}
				return nil
			}),
	)

	return builder.Complete(r)
}
