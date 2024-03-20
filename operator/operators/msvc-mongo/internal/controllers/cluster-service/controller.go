package clusterService

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kloudlite/operator/operators/msvc-mongo/internal/env"
	"github.com/kloudlite/operator/operators/msvc-mongo/internal/templates"
	"github.com/kloudlite/operator/operators/msvc-mongo/internal/types"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	mongodbMsvcv1 "github.com/kloudlite/operator/apis/mongodb.msvc/v1"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/harbor"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Reconciler struct {
	client.Client
	Scheme                     *runtime.Scheme
	Env                        *env.Env
	harborCli                  *harbor.Client
	logger                     logging.Logger
	Name                       string
	yamlClient                 kubectl.YAMLClient
	templateHelmMongoDBCluster []byte
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	AccessCredsGenerated     string = "access-creds-generated"
	MongoDBHelmApplied       string = `mongodb-helm-applied`
	MongoDBHelmReady         string = `mongodb-helm-ready`
	MongoDBHelmDeleted       string = `mongodb-helm-deleted`
	MongoDBStatefulSetsReady string = `mongodb-statefulsets-ready`

	DefaultsPatched string = "defaults-patched"
	KeyMsvcOutput   string = "msvc-output"

	AnnotationCurrentStorageSize string = "kloudlite.io/msvc.storage-size"
)

var ApplyCheckList = []rApi.CheckMeta{
	{Name: DefaultsPatched, Title: "Defaults Patched", Debug: true},
	{Name: AccessCredsGenerated, Title: "Access Credentials Generated"},
	{Name: MongoDBHelmApplied, Title: "MongoDB Helm Applied"},
	{Name: MongoDBHelmReady, Title: "MongoDB Helm Ready"},
	{Name: MongoDBStatefulSetsReady, Title: "MongoDB StatefulSets Ready"},
}

// DefaultsPatched string = "defaults-patched"
var DeleteCheckList = []rApi.CheckMeta{
	{Name: MongoDBHelmDeleted, Title: "MongoDB Helm Deleted"},
}

// +kubebuilder:rbac:groups=mongodb.msvc.kloudlite.io,resources=clusterServices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mongodb.msvc.kloudlite.io,resources=clusterServices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mongodb.msvc.kloudlite.io,resources=clusterServices/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &mongodbMsvcv1.ClusterService{})
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

	if step := req.EnsureCheckList(ApplyCheckList); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureChecks(DefaultsPatched, AccessCredsGenerated, MongoDBHelmApplied, MongoDBHelmReady, MongoDBStatefulSetsReady); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.generateAccessCredentials(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.applyHelm(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.checkMongoDBStatefulsetsReady(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *Reconciler) finalize(req *rApi.Request[*mongodbMsvcv1.ClusterService]) stepResult.Result {
	checkName := "finalizing"
	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	if step := req.CleanupOwnedResources(); !step.ShouldProceed() {
		return step
	}

	return req.Finalize()
}

func getHelmSecretName(name string) string {
	return fmt.Sprintf("helm-%s-creds", name)
}

func (r *Reconciler) generateAccessCredentials(req *rApi.Request[*mongodbMsvcv1.ClusterService]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation, State: rApi.RunningState}

	checkName := AccessCredsGenerated

	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	fail := func(err error) stepResult.Result {
		return req.CheckFailed(checkName, check, err.Error())
	}

	rootPassword := fn.CleanerNanoid(40)
	replicasetKey := fn.CleanerNanoid(10) // should not be more than 10, as it crashes our mongodb process

	msvcOutput := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: obj.Output.CredentialsRef.Name, Namespace: obj.Namespace}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, msvcOutput, func() error {
		msvcOutput.SetLabels(obj.GetLabels())
		msvcOutput.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})

		if msvcOutput.Data == nil {
			hosts := make([]string, obj.Spec.Replicas)
			for i := 0; i < obj.Spec.Replicas; i++ {
				hosts[i] = fmt.Sprintf("%s-%d.%s-headless.%s.svc.%s:27017", obj.Name, i, obj.Name, obj.Namespace, r.Env.ClusterInternalDNS)
			}

			rootUsername := "root"
			replicaSetName := "rs"

			authSource := "admin"

			output := types.ClusterSvcOutput{
				RootUsername:    rootUsername,
				RootPassword:    rootPassword,
				Hosts:           strings.Join(hosts, ","),
				URI:             fmt.Sprintf("mongodb://%s:%s@%s/%s?authSource=%s&replicaSet=%s", rootUsername, rootPassword, strings.Join(hosts, ","), "admin", authSource, replicaSetName),
				AuthSource:      authSource,
				ReplicasSetName: replicaSetName,
				ReplicaSetKey:   replicasetKey,
			}

			var err error
			msvcOutput.StringData, err = output.ToMap()
			return err
		}
		return nil
	}); err != nil {
		return fail(err)
	}

	req.AddToOwnedResources(rApi.ParseResourceRef(msvcOutput))

	helmSecret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: getHelmSecretName(obj.Name), Namespace: obj.Namespace}}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, helmSecret, func() error {
		helmSecret.SetLabels(obj.GetLabels())
		msvcOutput.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})

		if helmSecret.Data == nil {
			helmSecret.StringData = map[string]string{
				"mongodb-root-password":   rootPassword,
				"mongodb-replica-set-key": replicasetKey,
			}
		}

		return nil
	}); err != nil {
		return fail(err)
	}

	req.AddToOwnedResources(rApi.ParseResourceRef(helmSecret))
	rApi.SetLocal(req, "creds", msvcOutput.Data)

	check.Status = true
	if check != obj.Status.Checks[checkName] {
		fn.MapSet(&obj.Status.Checks, checkName, check)
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

func (r *Reconciler) applyHelm(req *rApi.Request[*mongodbMsvcv1.ClusterService]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation, State: rApi.RunningState}

	checkName := MongoDBHelmApplied

	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	fail := func(err error) stepResult.Result {
		return req.CheckFailed(checkName, check, err.Error())
	}

	creds, ok := rApi.GetLocal[map[string][]byte](req, "creds")
	if !ok {
		return fail(fmt.Errorf("creds not found "))
	}

	b, err := templates.ParseBytes(r.templateHelmMongoDBCluster, map[string]any{
		"name":      obj.Name,
		"namespace": obj.Namespace,

		// "release-name": obj.Name,

		"labels":                      obj.GetLabels(),
		"node-selector":               obj.Spec.NodeSelector,
		"tolerations":                 obj.Spec.Tolerations,
		"topology-spread-constraints": obj.Spec.TopologySpreadConstraints,

		"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},

		"pod-labels":      obj.GetLabels(),
		"pod-annotations": fn.FilterObservabilityAnnotations(obj.GetAnnotations()),

		"storage-class": obj.Spec.Resources.Storage.StorageClass,
		"storage-size":  obj.Spec.Resources.Storage.Size,

		"replica-count":        obj.Spec.Replicas,
		"root-user":            string(creds["USERNAME"]),
		"auth-existing-secret": getHelmSecretName(obj.Name),

		"cpu-min": obj.Spec.Resources.Cpu.Min,
		"cpu-max": obj.Spec.Resources.Cpu.Max,

		"memory-min": obj.Spec.Resources.Memory.Min,
		"memory-max": obj.Spec.Resources.Memory.Max,
	})
	if err != nil {
		return fail(err).Err(nil)
	}

	rr, err := r.yamlClient.ApplyYAML(ctx, b)
	if err != nil {
		return fail(err)
	}

	req.AddToOwnedResources(rr...)

	check.Status = true
	if check != obj.Status.Checks[checkName] {
		fn.MapSet(&obj.Status.Checks, checkName, check)
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

func (r *Reconciler) checkHelmReady(req *rApi.Request[*mongodbMsvcv1.StandaloneService]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation, State: rApi.RunningState}

	checkName := MongoDBHelmReady

	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	fail := func(err error) stepResult.Result {
		return req.CheckFailed(checkName, check, err.Error())
	}

	hc, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Name), &crdsv1.HelmChart{})
	if err != nil {
		return fail(err)
	}

	if !hc.Status.IsReady {
		return fail(fmt.Errorf("waiting for helm installation to complete"))
	}

	check.Status = true
	if check != obj.Status.Checks[checkName] {
		fn.MapSet(&obj.Status.Checks, checkName, check)
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

func (r *Reconciler) checkMongoDBStatefulsetsReady(req *rApi.Request[*mongodbMsvcv1.ClusterService]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation, State: rApi.RunningState}

	checkName := MongoDBStatefulSetsReady

	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	fail := func(err error) stepResult.Result {
		return req.CheckFailed(checkName, check, err.Error())
	}

	var stsList appsv1.StatefulSetList

	if err := r.List(
		ctx, &stsList, &client.ListOptions{
			LabelSelector: labels.SelectorFromValidatedSet(obj.GetEnsuredLabels()),
			Namespace:     obj.Namespace,
		},
	); err != nil {
		return fail(err)
	}

	for i := range stsList.Items {
		item := stsList.Items[i]
		if item.Status.AvailableReplicas != item.Status.Replicas {
			check.Status = false

			var podsList corev1.PodList
			if err := r.List(
				ctx, &podsList, &client.ListOptions{
					LabelSelector: labels.SelectorFromValidatedSet(obj.GetEnsuredLabels()),
				},
			); err != nil {
				return fail(err)
			}

			messages := rApi.GetMessagesFromPods(podsList.Items...)
			if len(messages) > 0 {
				b, err := json.Marshal(messages)
				if err != nil {
					return fail(err)
				}
				return fail(fmt.Errorf("%s", b))
			}
		}
	}

	if len(stsList.Items) == 0 {
		return fail(fmt.Errorf("no statefulset found"))
	}

	check.Status = true
	if check != obj.Status.Checks[checkName] {
		fn.MapSet(&obj.Status.Checks, checkName, check)
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig(), kubectl.YAMLClientOpts{Logger: r.logger})

	var err error
	r.templateHelmMongoDBCluster, err = templates.Read(templates.HelmMongoDBCluster)
	if err != nil {
		return err
	}

	builder := ctrl.NewControllerManagedBy(mgr).For(&mongodbMsvcv1.ClusterService{})
	builder.Owns(&corev1.Secret{})
	builder.Owns(&crdsv1.HelmChart{})
	builder.Watches(
		&appsv1.StatefulSet{}, handler.EnqueueRequestsFromMapFunc(
			func(_ context.Context, obj client.Object) []reconcile.Request {
				v, ok := obj.GetLabels()[constants.MsvcNameKey]
				if !ok {
					return nil
				}
				return []reconcile.Request{{NamespacedName: fn.NN(obj.GetNamespace(), v)}}
			},
		),
	)

	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
