package target_node

import (
	"context"
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

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
	TargetEnv  *env.TargetEnv
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

	ctx, obj := req.Context(), req.Object
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

	getDeletionJobName := func() string {
		return fmt.Sprintf("delete-%s", obj.Name)
	}

	if err := func() error {
		jb, err := rApi.Get(ctx, r.Client, fn.NN(r.Env.JobNamespace, getDeletionJobName()), &batchv1.Job{})
		if err == nil {
			if jb.Status.Succeeded >= 1 {
				return nil
			}

			return fmt.Errorf("deletion in progress")
		}

		if err != nil {
			if !apiErrors.IsNotFound(err) {
				return err
			}
		}

		if err := r.createJob(req, getDeleteAction(), getDeletionJobName()); err != nil {
			return err
		}

		return fmt.Errorf("deletion process initiated")
	}(); err != nil {
		return failed(err)
	}

	return req.Finalize()
}

func (r *Reconciler) createJob(req *rApi.Request[*clustersv1.Node], action string, name string) error {

	ctx, obj := req.Context(), req.Object

	np, err := rApi.Get(ctx, r.Client, fn.NN("", obj.Spec.NodePoolName), &clustersv1.NodePool{})
	if err != nil {
		return err
	}

	nodeConfig, err := r.getNodeConfig(np, obj)
	if err != nil {
		return err
	}

	providerConfig, err := getProviderConfig()
	if err != nil {
		return err
	}

	sProvider, err := r.getSpecificProvierConfig()
	if err != nil {
		return err
	}

	jobYaml, err := templates.Parse(templates.Clusters.Job,
		map[string]any{
			"name":      name,
			"namespace": r.Env.JobNamespace,
			"ownerRefs": []metav1.OwnerReference{fn.AsOwner(obj)},

			"cloudProvider":  r.TargetEnv.CloudProvider,
			"action":         action,
			"nodeConfig":     nodeConfig,
			"providerConfig": providerConfig,

			"AwsProvider":         sProvider,
			"AzureProvider":       sProvider,
			"DoProvider":          sProvider,
			"GCPProvider":         sProvider,
			"agentHelmValues":     "",
			"operatorsHelmValues": "",
			"podAnnotations": map[string]string{
				"kloudlite.io/job_name":      name,
				"kloudlite.io/job_namespace": r.Env.JobNamespace,
				"kloudlite.io/job_type":      action,
			},
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

	if err := func() error {
		var nodes corev1.NodeList
		if err := r.List(ctx, &nodes, &client.ListOptions{
			LabelSelector: labels.SelectorFromValidatedSet(
				map[string]string{
					constants.NodeNameKey: obj.Name,
				},
			),
		}); err != nil {
			if !apiErrors.IsNotFound(err) {
				return err
			}
		}

		if len(nodes.Items) >= 1 {
			ready := false
			for _, n := range nodes.Items {
				for _, nc := range n.Status.Conditions {
					if nc.Type == "Ready" && nc.Status == "True" {
						ready = true
						break
					}
				}

				if ready {
					break
				}
			}

			if ready {
				jb, e := rApi.Get(ctx, r.Client, fn.NN(r.Env.JobNamespace, obj.Name), &batchv1.Job{})
				if e != nil {
					if !apiErrors.IsNotFound(e) {
						return e
					}

					// if nodes are more than 1 with same name means we need to delete not-ready nodes
					if len(nodes.Items) >= 2 {
						for _, n := range nodes.Items {
							notReady := false
							for _, nc := range n.Status.Conditions {
								if nc.Type == "Ready" && nc.Status != "True" {
									notReady = true
									break
								}
							}

							if notReady {
								if err := r.Delete(ctx, &n, &client.DeleteOptions{}); err != nil {
									return err
								}
							}
						}
					}

					return nil
				}

				return r.Delete(ctx, jb, &client.DeleteOptions{})
			}

			return fmt.Errorf("job is success but still waiting for the node to be live")
		}

		if len(nodes.Items) == 0 {
			// check node_job, if not created create
			if _, e := rApi.Get(ctx, r.Client, fn.NN(r.Env.JobNamespace, obj.Name), &batchv1.Job{}); e != nil {
				if !apiErrors.IsNotFound(e) {
					return e
				}

				if err := r.createJob(req, getAction(obj), obj.Name); err != nil {
					return err
				}
				return fmt.Errorf("job created for the node creation")
			}

			return fmt.Errorf("node creation in progress")
		}

		return nil
	}(); err != nil {
		return failed(err)
	}

	// check nodejob
	if err := func() error {
		nodeJob, e := rApi.Get(ctx, r.Client, fn.NN(r.Env.JobNamespace, obj.Name), &batchv1.Job{})
		if e != nil {
			if !apiErrors.IsNotFound(e) {
				return e
			}
			return nil
		}

		if nodeJob.Status.Succeeded >= 1 {
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

		return fmt.Errorf("node creation in progress")
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
		&batchv1.Job{},
		handler.EnqueueRequestsFromMapFunc(
			func(_ context.Context, obj client.Object) []reconcile.Request {
				if _, ok := obj.GetLabels()[constants.IsNodeControllerJob]; ok {
					return []reconcile.Request{{NamespacedName: fn.NN("", obj.GetName())}}
				}
				return nil
			}),
	)

	return builder.Complete(r)
}
