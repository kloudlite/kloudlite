package standalone

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	redisMsvcv1 "github.com/kloudlite/operator/apis/redis.msvc/v1"
	"github.com/kloudlite/operator/operators/msvc-redis/internal/controllers/standalone/templates"
	"github.com/kloudlite/operator/operators/msvc-redis/internal/env"
	"github.com/kloudlite/operator/operators/msvc-redis/internal/types"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
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
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ServiceReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	logger     logging.Logger
	Name       string
	Env        *env.Env
	yamlClient kubectl.YAMLClient

	templateHelmRedisStandalone []byte
}

func (r *ServiceReconciler) GetName() string {
	return r.Name
}

const (
	HelmReady              string = "helm-ready"
	StsReady               string = "sts-ready"
	AccessCredsGenerated   string = "access-creds-generated"
	RedisHelmApplied       string = `redis-helm-applied`
	RedisHelmReady         string = `redis-helm-ready`
	RedisHelmDeleted       string = `redis-helm-deleted`
	RedisStatefulSetsReady string = `redis-statefulsets-ready`
)

const (
	KeyMsvcOutput   string = "msvc-output"
	DefaultsPatched string = "defaults-patched"
)

var ApplyCheckList = []rApi.CheckMeta{
	{Name: DefaultsPatched, Title: "Defaults Patched", Debug: true},
	{Name: AccessCredsGenerated, Title: "Access Credentials Generated"},
	{Name: RedisHelmApplied, Title: "Redis Helm Applied"},
	{Name: RedisHelmReady, Title: "Redis Helm Ready"},
	{Name: RedisStatefulSetsReady, Title: "Redis StatefulSets Ready"},
}

var DeleteCheckList = []rApi.CheckMeta{
	{Name: RedisHelmDeleted, Title: "Redis Helm Deleted"},
}

// +kubebuilder:rbac:groups=redis.msvc.kloudlite.io,resources=standaloneservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=redis.msvc.kloudlite.io,resources=standaloneservices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=redis.msvc.kloudlite.io,resources=standaloneservices/finalizers,verbs=update

func (r *ServiceReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &redisMsvcv1.StandaloneService{})
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

	if step := req.EnsureCheckList(ApplyCheckList); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureChecks(DefaultsPatched, AccessCredsGenerated, RedisHelmApplied, RedisHelmReady, RedisStatefulSetsReady); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.patchDefaults(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.generateAccessCreds(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.applyHelm(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.checkHelmReady(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.checkStsReady(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *ServiceReconciler) finalize(req *rApi.Request[*redisMsvcv1.StandaloneService]) stepResult.Result {
	checkName := "finalizing"

	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	obj := req.Object
	if !slices.Equal(obj.Status.CheckList, DeleteCheckList) {
		obj.Status.CheckList = nil
	}

	if step := req.EnsureCheckList(DeleteCheckList); !step.ShouldProceed() {
		return step
	}

	if step := req.CleanupOwnedResources(); !step.ShouldProceed() {
		return step
	}

	return req.Finalize()
}

func (r *ServiceReconciler) patchDefaults(req *rApi.Request[*redisMsvcv1.StandaloneService]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation, State: rApi.RunningState}

	checkName := DefaultsPatched

	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	fail := func(err error) stepResult.Result {
		return req.CheckFailed(checkName, check, err.Error())
	}

	hasPatched := false

	// function-body
	if obj.Spec.Output.Credentials.Name == "" {
		hasPatched = true
		obj.Spec.Output.Credentials.Name = fmt.Sprintf("msvc-%s-creds", obj.Name)
	}

	if obj.Spec.Output.Credentials.Namespace == "" {
		hasPatched = true
		obj.Spec.Output.Credentials.Namespace = obj.Namespace
	}

	if hasPatched {
		if err := r.Update(ctx, obj); err != nil {
			return fail(err)
		}

		return req.Done().RequeueAfter(500 * time.Millisecond)
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

func (r *ServiceReconciler) generateAccessCreds(req *rApi.Request[*redisMsvcv1.StandaloneService]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation, State: rApi.RunningState}

	checkName := AccessCredsGenerated

	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	fail := func(err error) stepResult.Result {
		return req.CheckFailed(checkName, check, err.Error())
	}

	accessCreds := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.Output.Credentials.Name, Namespace: obj.Spec.Output.Credentials.Namespace}}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, accessCreds, func() error {
		obj.SetLabels(obj.GetLabels())
		obj.SetOwnerReferences(obj.GetOwnerReferences())

		if accessCreds.Data != nil {
			// means secret already exists, it is not getting created
			return nil
		}

		rootPassword := fn.CleanerNanoid(40)
		host := fmt.Sprintf("%s-headless.%s.svc.%s", obj.Name, obj.Namespace, r.Env.ClusterInternalDNS)
		port := 6379

		var m map[string]string

		out := types.MsvcOutput{
			Host:         host,
			Port:         fmt.Sprintf("%d", port),
			Addr:         fmt.Sprintf("%s:%d", host, port),
			Uri:          fmt.Sprintf("redis://:%s@%s?allowUsernameInURI=true", rootPassword, host),
			RootPassword: rootPassword,
		}

		m, err := out.ToMap()
		if err != nil {
			return err
		}

		accessCreds.StringData = m

		return nil
	}); err != nil {
		return fail(err)
	}

	msvcOutput, err := fn.ParseFromSecret[types.MsvcOutput](accessCreds)
	if err != nil {
		return fail(err).Err(nil)
	}
	rApi.SetLocal(req, KeyMsvcOutput, *msvcOutput)

	check.Status = true
	if check != obj.Status.Checks[checkName] {
		fn.MapSet(&obj.Status.Checks, checkName, check)
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

func (r *ServiceReconciler) applyHelm(req *rApi.Request[*redisMsvcv1.StandaloneService]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation, State: rApi.RunningState}

	checkName := RedisHelmApplied

	fail := func(err error) stepResult.Result {
		return req.CheckFailed(checkName, check, err.Error())
	}

	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	msvcOutput, ok := rApi.GetLocal[types.MsvcOutput](req, KeyMsvcOutput)
	if !ok {
		return fail(rApi.ErrNotInReqLocals(KeyMsvcOutput))
	}

	b, err := templates.ParseBytes(r.templateHelmRedisStandalone, map[string]any{
		"name":      obj.Name,
		"namespace": obj.Namespace,

		"labels":     map[string]string{constants.MsvcNameKey: obj.Name},
		"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},

		"pod-labels":      obj.GetLabels(),
		"pod-annotations": fn.FilterObservabilityAnnotations(obj.GetAnnotations()),

		"node-selector": obj.Spec.NodeSelector,
		"tolerations":   obj.Spec.Tolerations,

		"storage-size":  obj.Spec.Resources.Storage.Size,
		"storage-class": obj.Spec.Resources.Storage.StorageClass,

		"requests-cpu": obj.Spec.Resources.Cpu.Min,
		"requests-mem": obj.Spec.Resources.Memory.Min,

		"limits-cpu": obj.Spec.Resources.Cpu.Min,
		"limits-mem": obj.Spec.Resources.Memory.Max,

		"root-password": msvcOutput.RootPassword,
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

func (r *ServiceReconciler) checkHelmReady(req *rApi.Request[*redisMsvcv1.StandaloneService]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation, State: rApi.RunningState}

	checkName := RedisHelmReady

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

func (r *ServiceReconciler) checkStsReady(req *rApi.Request[*redisMsvcv1.StandaloneService]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation, State: rApi.RunningState}

	checkName := RedisStatefulSetsReady

	fail := func(err error) stepResult.Result {
		return req.CheckFailed(checkName, check, err.Error())
	}

	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	sts, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Name+"-master"), &appsv1.StatefulSet{})
	if err != nil {
		return fail(err)
	}

	if sts.Status.AvailableReplicas != sts.Status.Replicas {
		check.Status = false

		var podsList corev1.PodList
		if err := r.List(
			ctx, &podsList, &client.ListOptions{
				LabelSelector: labels.SelectorFromValidatedSet(
					map[string]string{constants.MsvcNameKey: obj.Name},
				),
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
			return fail(fmt.Errorf(string(b)))
		}
		return fail(fmt.Errorf("waiting for statefulset pods to kick in"))
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

func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig(), kubectl.YAMLClientOpts{Logger: r.logger})

	b, err := templates.Read(templates.HelmStandaloneRedisTemplate)
	if err != nil {
		return err
	}
	r.templateHelmRedisStandalone = b

	builder := ctrl.NewControllerManagedBy(mgr).For(&redisMsvcv1.StandaloneService{})
	builder.Owns(&corev1.Secret{})
	builder.Owns(&redisMsvcv1.ACLConfigMap{})
	builder.Owns(&crdsv1.HelmChart{})

	builder.Watches(
		&appsv1.StatefulSet{}, handler.EnqueueRequestsFromMapFunc(
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

	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
