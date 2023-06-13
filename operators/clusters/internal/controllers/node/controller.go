package node

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	"github.com/kloudlite/operator/operators/clusters/internal/env"
	"github.com/kloudlite/operator/pkg/constants"
	"github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	"github.com/kloudlite/operator/pkg/templates"
)

// have to fetch these from env
const (
	tfTemplates string = ""
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

	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	failed := func(e error) stepResult.Result {
		return req.CheckFailed("fail in ensure nodes", check, e.Error())
	}

	np, err := rApi.Get(ctx, r.Client, functions.NN("", obj.Spec.NodePoolName), &clustersv1.NodePool{})
	if err != nil {
		return failed(err)
	}

	getNodeConfig := func() ([]byte, error) {
		switch r.Env.CloudProvider {
		case "aws":
			var awsNode AWSNode
			if err := json.Unmarshal([]byte(np.Spec.NodeConfig), &awsNode); err != nil {
				return nil, err
			}

			awsbyte, err := yaml.Marshal(awsNode)
			if err != nil {
				return nil, err
			}
			return awsbyte, nil

		case "do", "azure", "gcp":
			panic("unimplemented")
		default:
			return nil, fmt.Errorf("this type of cloud provider not supported for now")
		}
	}

	getProviderConfig := func() ([]byte, error) {
		pd := CommonProviderData{
			TfTemplates: tfTemplates,
			Labels:      map[string]string{},
			Taints:      []string{},
			SSHPath:     "",
		}
		return yaml.Marshal(pd)
	}

	nodeConfig, err := getNodeConfig()
	if err != nil {
		return failed(err)
	}

	providerConfig, err := getProviderConfig()
	if err != nil {
		return failed(err)
	}

	getAction := func() string {
		switch obj.Spec.NodeType {
		case "worker", "master", "cluster":
			return "delete"
		default:
			return "unknown"
		}
	}

	action := getAction()

	getSpecificProvierConfig := func() ([]byte, error) {
		switch r.Env.CloudProvider {
		case "aws":
			return json.Marshal(AwsProviderConfig{
				AccessKey:    r.Env.AccessKey,
				AccessSecret: r.Env.AccessSecret,
				AccountName:  r.Env.AccountName,
			})
		default:
			return nil, fmt.Errorf("cloud provider %s not supported for now", r.Env.CloudProvider)
		}
	}

	sProvider, err := getSpecificProvierConfig()
	if err != nil {
		return failed(err)
	}

	createDeleteNodeJob := func() error {
		jobYaml, err := templates.Parse(templates.Clusters.Job,
			map[string]any{
				"name":      fmt.Sprintf("delete-%s", obj.Name),
				"namespace": r.Env.JobNamespace,
				"ownerRefs": functions.AsOwner(obj),

				"cloudProvider":  r.Env.CloudProvider,
				"action":         action,
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

	j, err := rApi.Get(ctx, r.Client, functions.NN(r.Env.JobNamespace, fmt.Sprintf("delete-%s", obj.Name)), &batchv1.Job{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return failed(err)
		}

		if err := createDeleteNodeJob(); err != nil {
			return failed(err)
		}
	}

	if j.Status.Succeeded == 1 {
		return req.Finalize()
	}

	return nil
}

func (r *Reconciler) ensureNodeReady(req *rApi.Request[*clustersv1.Node]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	failed := func(err error) stepResult.Result {
		return req.CheckFailed("ensure-node-ready", check, err.Error())
	}

	req.LogPreCheck(K8sSecretCreated)
	defer req.LogPostCheck(K8sSecretCreated)

	np, err := rApi.Get(ctx, r.Client, functions.NN("", obj.Spec.NodePoolName), &clustersv1.NodePool{})
	if err != nil {
		return failed(err)
	}

	getNodeConfig := func() ([]byte, error) {
		switch r.Env.CloudProvider {
		case "aws":
			var awsNode AWSNode
			if err := json.Unmarshal([]byte(np.Spec.NodeConfig), &awsNode); err != nil {
				return nil, err
			}

			awsbyte, err := yaml.Marshal(awsNode)
			if err != nil {
				return nil, err
			}
			return awsbyte, nil

		case "do", "azure", "gcp":
			panic("unimplemented")
		default:
			return nil, fmt.Errorf("this type of cloud provider not supported for now")
		}
	}

	getProviderConfig := func() ([]byte, error) {
		pd := CommonProviderData{
			TfTemplates: tfTemplates,
			Labels:      map[string]string{},
			Taints:      []string{},
			SSHPath:     "",
		}
		return yaml.Marshal(pd)
	}

	nodeConfig, err := getNodeConfig()
	if err != nil {
		return failed(err)
	}

	providerConfig, err := getProviderConfig()
	if err != nil {
		return failed(err)
	}

	getAction := func() string {
		switch obj.Spec.NodeType {
		case "worker":
			return "add-worker"
		case "master":
			return "add-master"
		case "cluster":
			return "create-cluster"
		default:
			return "unknown"
		}
	}

	action := getAction()

	getSpecificProvierConfig := func() ([]byte, error) {
		switch r.Env.CloudProvider {
		case "aws":
			return json.Marshal(AwsProviderConfig{
				AccessKey:    r.Env.AccessKey,
				AccessSecret: r.Env.AccessSecret,
				AccountName:  r.Env.AccountName,
			})
		default:
			return nil, fmt.Errorf("cloud provider %s not supported for now", r.Env.CloudProvider)
		}
	}

	sProvider, err := getSpecificProvierConfig()
	if err != nil {
		return failed(err)
	}

	createNode := func() error {
		jobYaml, err := templates.Parse(templates.Clusters.Job,
			map[string]any{
				"name":      obj.Name,
				"namespace": r.Env.JobNamespace,
				"ownerRefs": functions.AsOwner(obj),

				"cloudProvider":  r.Env.CloudProvider,
				"action":         action,
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

	// do your actions here
	if err := func() error {
		_, err := rApi.Get(ctx, r.Client, functions.NN("", obj.Name), &corev1.Node{})
		if err != nil {
			if !apiErrors.IsNotFound(err) {
				return err
			}
			// not found do your action

			// check node job if not created create
			if _, e := rApi.Get(ctx, r.Client, functions.NN(r.Env.JobNamespace, obj.Name), &batchv1.Job{}); e != nil {
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
		return req.CheckFailed("failed", check, err.Error())
	}

	// check node attached
	// if not attached then attach then have to attach

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
