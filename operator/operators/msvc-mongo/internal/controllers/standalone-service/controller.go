package standalone_service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	ct "github.com/kloudlite/operator/apis/common-types"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	mongodbMsvcv1 "github.com/kloudlite/operator/apis/mongodb.msvc/v1"
	"github.com/kloudlite/operator/operators/msvc-mongo/internal/env"
	"github.com/kloudlite/operator/operators/msvc-mongo/internal/templates"
	"github.com/kloudlite/operator/operators/msvc-mongo/internal/types"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiLabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	logger     logging.Logger
	Name       string
	Env        *env.Env
	yamlClient kubectl.YAMLClient

	templateHelmMongoDB     []byte
	templateHelmMongoDBAuth []byte
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
	Cleanup         string = "cleanup"
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

// +kubebuilder:rbac:groups=mongodb.msvc.kloudlite.io,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mongodb.msvc.kloudlite.io,resources=services/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mongodb.msvc.kloudlite.io,resources=services/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &mongodbMsvcv1.StandaloneService{})
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

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	// if step := r.patchDefaults(req); !step.ShouldProceed() {
	// 	return step.ReconcilerResponse()
	// }

	if step := r.generateAccessCredentials(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.applyMongoDBStandaloneHelm(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.checkHelmReady(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconSts(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *Reconciler) finalize(req *rApi.Request[*mongodbMsvcv1.StandaloneService]) stepResult.Result {
	check := "finalizing"

	req.LogPreCheck(check)
	defer req.LogPostCheck(check)

	if result := req.CleanupOwnedResources(); !result.ShouldProceed() {
		return result
	}

	return req.Finalize()
}

func getHelmSecretName(name string) string {
	return fmt.Sprintf("helm-%s-creds", name)
}

func (r *Reconciler) generateAccessCredentials(req *rApi.Request[*mongodbMsvcv1.StandaloneService]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation, State: rApi.RunningState}

	checkName := AccessCredsGenerated

	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	fail := func(err error) stepResult.Result {
		return req.CheckFailed(checkName, check, err.Error())
	}

	rootPassword := fn.CleanerNanoid(40)

	msvcOutput := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: obj.Output.CredentialsRef.Name, Namespace: obj.Namespace}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, msvcOutput, func() error {
		msvcOutput.SetLabels(obj.GetLabels())

		msvcOutput.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})

		if msvcOutput.Data == nil {
			// secret does not already exists
			host := fmt.Sprintf("%s-%d.%s.%s.svc.%s:27017", obj.Name, 0, obj.Name, obj.Namespace, r.Env.ClusterInternalDNS)

			rootUsername := "root"

			authSource := "admin"

			output := types.StandaloneSvcOutput{
				RootUsername: rootUsername,
				RootPassword: rootPassword,
				Hosts:        host,
				URI: fmt.Sprintf(
					"mongodb://%s:%s@%s/%s?authSource=%s",
					rootUsername,
					rootPassword,
					host,
					"admin",
					authSource,
				),
				AuthSource: authSource,
			}

			m, err := fn.JsonConvert[map[string]string](output)
			if err != nil {
				return nil
			}

			msvcOutput.StringData = m
		}
		return nil
	}); err != nil {
		return fail(err)
	}

	req.AddToOwnedResources(rApi.ParseResourceRef(msvcOutput))

	helmSecret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: getHelmSecretName(obj.Name), Namespace: obj.Namespace}}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, helmSecret, func() error {
		helmSecret.SetLabels(obj.GetLabels())
		helmSecret.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})

		if helmSecret.Data == nil {
			// secret does not already exists
			helmSecret.StringData = map[string]string{
				"mongodb-root-password": rootPassword,
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

func (r *Reconciler) applyMongoDBStandaloneHelm(req *rApi.Request[*mongodbMsvcv1.StandaloneService]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation, State: rApi.RunningState}

	checkName := MongoDBHelmApplied

	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	fail := func(err error) stepResult.Result {
		return req.CheckFailed(checkName, check, err.Error())
	}

	hc, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Name), &crdsv1.HelmChart{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return fail(err)
		}
		hc = nil
	}

	if hc == nil {
		fn.MapSet(&obj.Annotations, AnnotationCurrentStorageSize, string(obj.Spec.Resources.Storage.Size))
		if err := r.Update(ctx, obj); err != nil {
			return fail(err)
		}

		b, err := templates.ParseBytes(r.templateHelmMongoDB, map[string]any{
			"name":          obj.Name,
			"namespace":     obj.Namespace,
			"labels":        obj.GetLabels(),
			"owner-refs":    []metav1.OwnerReference{fn.AsOwner(obj, true)},
			"node-selector": obj.Spec.NodeSelector,
			"tolerations":   obj.Spec.Tolerations,

			"pod-labels":      obj.GetLabels(),
			"pod-annotations": fn.FilterObservabilityAnnotations(obj.GetAnnotations()),

			"storage-class": obj.Spec.Resources.Storage.StorageClass,
			"storage-size":  obj.Spec.Resources.Storage.Size,

			"requests-cpu": obj.Spec.Resources.Cpu.Min,
			"requests-mem": obj.Spec.Resources.Memory.Min,

			"limits-cpu": obj.Spec.Resources.Cpu.Max,
			"limits-mem": obj.Spec.Resources.Memory.Max,

			"existing-secret": getHelmSecretName(obj.Name),
		})
		if err != nil {
			return fail(err).Err(nil)
		}

		if _, err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
			return fail(err)
		}

		return req.Done().RequeueAfter(500 * time.Millisecond)
	}

	req.AddToOwnedResources(rApi.ParseResourceRef(hc))

	shouldUpdatePVCs := false

	oldSize, ok := obj.GetAnnotations()[AnnotationCurrentStorageSize]
	if ok {
		oldSizeNum, err := ct.StorageSize(oldSize).ToInt()
		if err != nil {
			return fail(err)
		}

		newSizeNum, err := ct.StorageSize(obj.Spec.Resources.Storage.Size).ToInt()
		if err != nil {
			return fail(err)
		}

		if oldSizeNum > newSizeNum {
			return fail(fmt.Errorf("new storage size (%s), must be higher than or equal to old size (%s)", obj.Spec.Resources.Storage.Size, oldSize))
		}
		shouldUpdatePVCs = newSizeNum > oldSizeNum
	}
	if !ok {
		shouldUpdatePVCs = true
	}

	if shouldUpdatePVCs {
		// need to do something
		// 1. Patch the PVC directly
		// 2. Rollout the Statefulsets

		ss, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Name), &appsv1.StatefulSet{})
		if err != nil {
			if !apiErrors.IsNotFound(err) {
				return fail(err)
			}
			ss = nil
		}

		if ss != nil {
			m := types.ExtractPVCLabelsFromStatefulSetLabels(ss.GetLabels())

			var pvclist corev1.PersistentVolumeClaimList
			if err := r.List(ctx, &pvclist, &client.ListOptions{
				LabelSelector: apiLabels.SelectorFromValidatedSet(m),
				Namespace:     obj.Namespace,
			}); err != nil {
				return fail(err)
			}

			for i := range pvclist.Items {
				pvclist.Items[i].Spec.Resources.Requests[corev1.ResourceStorage] = resource.MustParse(string(obj.Spec.Resources.Storage.Size))
				if err := r.Update(ctx, &pvclist.Items[i]); err != nil {
					return fail(err)
				}
			}

			// STEP 2: rollout statefulset
			if err := fn.RolloutRestart(r.Client, fn.StatefulSet, obj.Namespace, ss.GetLabels()); err != nil {
				return fail(err)
			}

			fn.MapSet(&obj.Annotations, AnnotationCurrentStorageSize, string(obj.Spec.Resources.Storage.Size))
			if err := r.Update(ctx, obj); err != nil {
				return fail(err)
			}
		}
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

func (r *Reconciler) reconSts(req *rApi.Request[*mongodbMsvcv1.StandaloneService]) stepResult.Result {
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
			LabelSelector: apiLabels.SelectorFromValidatedSet(
				map[string]string{constants.MsvcNameKey: obj.Name},
			),
			Namespace: obj.Namespace,
		},
	); err != nil {
		return fail(err)
	}

	if len(stsList.Items) == 0 {
		return fail(fmt.Errorf("no statefulset pods found, waiting for helm controller to reconcile the resource"))
	}

	for i := range stsList.Items {
		item := stsList.Items[i]
		if item.Status.AvailableReplicas != item.Status.Replicas {
			check.Status = false

			var podsList corev1.PodList
			if err := r.List(
				ctx, &podsList, &client.ListOptions{
					LabelSelector: apiLabels.SelectorFromValidatedSet(
						map[string]string{
							constants.MsvcNameKey: obj.Name,
						},
					),
				},
			); err != nil {
				return fail(err)
			}

			messages := rApi.GetMessagesFromPods(podsList.Items...)
			if len(messages) > 0 {
				b, err := json.Marshal(messages)
				if err != nil {
					return fail(err).Err(nil)
				}
				return fail(fmt.Errorf("%s", b)).Err(nil)
			}

			return fail(fmt.Errorf("waiting for statefulset pods to start"))
		}
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
	r.templateHelmMongoDB, err = templates.Read(templates.HelmMongoDBStandalone)
	if err != nil {
		return err
	}

	r.templateHelmMongoDBAuth, err = templates.Read(templates.HelmMongoDBStandaloneAuth)
	if err != nil {
		return err
	}

	builder := ctrl.NewControllerManagedBy(mgr).For(&mongodbMsvcv1.StandaloneService{})
	builder.Owns(&corev1.Secret{})
	builder.Owns(&crdsv1.HelmChart{})

	watchList := []client.Object{
		&appsv1.StatefulSet{},
	}

	for _, obj := range watchList {
		builder.Watches(
			obj,
			handler.EnqueueRequestsFromMapFunc(
				func(_ context.Context, obj client.Object) []reconcile.Request {
					value, ok := obj.GetLabels()[constants.MsvcNameKey]
					if !ok {
						return nil
					}
					return []reconcile.Request{
						{NamespacedName: fn.NN(obj.GetNamespace(), value)},
					}
				},
			),
		)
	}

	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
