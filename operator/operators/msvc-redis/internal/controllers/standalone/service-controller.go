package standalone

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kloudlite/operator/pkg/kubectl"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"

	ct "github.com/kloudlite/operator/apis/common-types"
	redisMsvcv1 "github.com/kloudlite/operator/apis/redis.msvc/v1"
	"github.com/kloudlite/operator/operators/msvc-redis/internal/env"
	"github.com/kloudlite/operator/operators/msvc-redis/internal/types"
	"github.com/kloudlite/operator/pkg/conditions"
	"github.com/kloudlite/operator/pkg/constants"
	"github.com/kloudlite/operator/pkg/errors"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/harbor"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	"github.com/kloudlite/operator/pkg/templates"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type ServiceReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	harborCli  *harbor.Client
	logger     logging.Logger
	Name       string
	Env        *env.Env
	yamlClient *kubectl.YAMLClient
}

func (r *ServiceReconciler) GetName() string {
	return r.Name
}

const (
	HelmReady         string = "helm-ready"
	StsReady          string = "sts-ready"
	AccessCredsReady  string = "access-creds-ready"
	ACLConfigMapReady string = "acl-configmap-ready"
)

const (
	KeyRootPassword     string = "root-password"
	KeyAclConfigMapName string = "acl-configmap-name"
	KeyMsvcOutput       string = "msvc-output"
	DefaultsPatched     string = "defaults-patched"
)

// +kubebuilder:rbac:groups=redis.msvc.kloudlite.io,resources=standaloneservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=redis.msvc.kloudlite.io,resources=standaloneservices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=redis.msvc.kloudlite.io,resources=standaloneservices/finalizers,verbs=update

func (r *ServiceReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &redisMsvcv1.StandaloneService{})
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

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconACLConfigmap(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconAccessCreds(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconHelm(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconSts(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	req.Object.Status.LastReconcileTime = &metav1.Time{Time: time.Now()}

	if err := r.Status().Update(ctx, req.Object); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, nil
}

func (r *ServiceReconciler) finalize(req *rApi.Request[*redisMsvcv1.StandaloneService]) stepResult.Result {
	return req.Finalize()
}

func (r *ServiceReconciler) reconACLConfigmap(req *rApi.Request[*redisMsvcv1.StandaloneService]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(ACLConfigMapReady)
	defer req.LogPostCheck(ACLConfigMapReady)

	aclConfigmapName := "msvc-" + obj.Name + "-acl"

	aclCfgMap, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, aclConfigmapName), &redisMsvcv1.ACLConfigMap{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.CheckFailed(ACLConfigMapReady, check, err.Error())
		}
		req.Logger.Infof("acl configmap (%s) not found, will be creating it", fn.NN(obj.Namespace, obj.Name).String())
	}

	if aclCfgMap == nil {
		if err := r.Create(
			ctx, &redisMsvcv1.ACLConfigMap{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ACLConfigMap",
					APIVersion: "redis.msvc.kloudlite.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:            aclConfigmapName,
					Namespace:       obj.Namespace,
					OwnerReferences: []metav1.OwnerReference{fn.AsOwner(obj, true)},
				},
				Spec: redisMsvcv1.ACLConfigMapSpec{
					MsvcName: obj.Name,
				},
			},
		); err != nil {
			return req.CheckFailed(ACLConfigMapReady, check, err.Error())
		}
	}

	if !aclCfgMap.Status.IsReady {
		if aclCfgMap.Status.Message == nil {
			return req.CheckFailed(ACLConfigMapReady, check, "waiting for acl config map to reconcile").Err(nil)
		}
		b, err := json.Marshal(aclCfgMap.Status.Message)
		if err != nil {
			return req.CheckFailed(ACLConfigMapReady, check, err.Error()).Err(nil)
		}
		return req.CheckFailed(ACLConfigMapReady, check, string(b)).Err(nil)
	}

	check.Status = true
	if check != obj.Status.Checks[ACLConfigMapReady] {
		obj.Status.Checks[ACLConfigMapReady] = check
		req.UpdateStatus()
	}

	rApi.SetLocal(req, KeyAclConfigMapName, aclCfgMap.Name)
	return req.Next()
}

func (r *ServiceReconciler) reconAccessCreds(req *rApi.Request[*redisMsvcv1.StandaloneService]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(AccessCredsReady)
	defer req.LogPostCheck(AccessCredsReady)

	secretName := "msvc-" + obj.Name
	scrt, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, secretName), &corev1.Secret{})
	if err != nil {
		req.Logger.Infof("secret %s does not exist yet, would be creating it ...", fn.NN(obj.Namespace, secretName).String())
	}

	if scrt == nil {
		rootPassword := fn.CleanerNanoid(40)
		host := fmt.Sprintf("%s-headless.%s.svc.cluster.local:6379", obj.Name, obj.Namespace)
		b, err := templates.Parse(
			templates.Secret, map[string]any{
				"name":       secretName,
				"namespace":  obj.Namespace,
				"labels":     obj.GetLabels(),
				"owner-refs": obj.GetOwnerReferences(),
				"string-data": types.MsvcOutput{
					RootPassword: rootPassword,
					Hosts:        host,
					Uri:          fmt.Sprintf("redis://:%s@%s?allowUsernameInURI=true", rootPassword, host),
				},
			},
		)

		if err != nil {
			return req.CheckFailed(AccessCredsReady, check, err.Error())
		}

		if _, err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
			return req.CheckFailed(AccessCredsReady, check, err.Error())
		}
	}

	if !fn.IsOwner(obj, fn.AsOwner(scrt)) {
		obj.SetOwnerReferences(append(obj.GetOwnerReferences(), fn.AsOwner(scrt)))
		if err := r.Update(ctx, obj); err != nil {
			return req.FailWithOpError(err)
		}
		return req.Done().RequeueAfter(100 * time.Millisecond)
	}

	check.Status = true
	if check != obj.Status.Checks[AccessCredsReady] {
		obj.Status.Checks[AccessCredsReady] = check
		req.UpdateStatus()
	}

	msvcOutput, err := fn.ParseFromSecret[types.MsvcOutput](scrt)
	if err != nil {
		return req.CheckFailed(AccessCredsReady, check, err.Error()).Err(nil)
	}

	rApi.SetLocal(req, KeyMsvcOutput, *msvcOutput)
	return req.Next()
}

func (r *ServiceReconciler) reconHelm(req *rApi.Request[*redisMsvcv1.StandaloneService]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(HelmReady)
	defer req.LogPostCheck(HelmReady)

	helmRes, err := rApi.Get(
		ctx, r.Client, fn.NN(obj.Namespace, obj.Name), fn.NewUnstructured(constants.HelmRedisType),
	)
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.CheckFailed(HelmReady, check, err.Error())
		}
		req.Logger.Infof("helm resource (%s) not found, will be creating it", fn.NN(obj.Namespace, obj.Name).String())
	}

	msvcOutput, ok := rApi.GetLocal[types.MsvcOutput](req, KeyMsvcOutput)
	if !ok {
		return req.CheckFailed(HelmReady, check, errors.NotInLocals(KeyRootPassword).Error())
	}
	aclConfigmapName, ok := rApi.GetLocal[string](req, KeyAclConfigMapName)
	if !ok {
		return req.CheckFailed(HelmReady, check, errors.NotInLocals(KeyAclConfigMapName).Error())
	}

	sc := func() string {
		if obj.Spec.Resources.Storage != nil && obj.Spec.Resources.Storage.StorageClass != "" {
			return obj.Spec.Resources.Storage.StorageClass
		}
		return fmt.Sprintf("%s-%s", obj.Spec.Region, ct.Ext4)
	}()

	b, err := templates.Parse(
		templates.RedisStandalone, map[string]any{
			"object":             obj,
			"freeze":             obj.GetLabels()[constants.LabelKeys.Freeze] == "true",
			"storage-class":      sc,
			"owner-refs":         obj.GetOwnerReferences(),
			"acl-configmap-name": aclConfigmapName,
			"root-password":      msvcOutput.RootPassword,
		},
	)

	if err != nil {
		return req.CheckFailed(HelmReady, check, err.Error()).Err(nil)
	}

	if _, err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
		return req.CheckFailed(HelmReady, check, err.Error()).Err(nil)
	}

	cds, err := conditions.FromObject(helmRes)
	if err != nil {
		return req.CheckFailed(HelmReady, check, err.Error())
	}

	deployedC := meta.FindStatusCondition(cds, "Deployed")
	if deployedC == nil {
		return req.Done()
	}

	if deployedC.Status == metav1.ConditionFalse {
		return req.CheckFailed(HelmReady, check, deployedC.Message)
	}

	if deployedC.Status == metav1.ConditionTrue {
		check.Status = true
	}

	if check != obj.Status.Checks[HelmReady] {
		obj.Status.Checks[HelmReady] = check
		req.UpdateStatus()
	}

	return req.Next()
}

func getStsName(objName string) string {
	return objName + "-master"
}

func (r *ServiceReconciler) reconSts(req *rApi.Request[*redisMsvcv1.StandaloneService]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(StsReady)
	defer req.LogPostCheck(StsReady)

	sts, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, getStsName(obj.Name)), &appsv1.StatefulSet{})
	if err != nil {
		return req.CheckFailed(StsReady, check, err.Error())
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
			return req.FailWithOpError(err)
		}

		messages := rApi.GetMessagesFromPods(podsList.Items...)
		if len(messages) > 0 {
			b, err := json.Marshal(messages)
			if err != nil {
				return req.CheckFailed(StsReady, check, err.Error())
			}
			return req.CheckFailed(StsReady, check, string(b))
		}
		return req.CheckFailed(StsReady, check, "waiting for pods to start ...")
	}

	check.Status = true
	if check != obj.Status.Checks[StsReady] {
		obj.Status.Checks[StsReady] = check
		req.UpdateStatus()
	}

	return req.Next()
}

func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

	builder := ctrl.NewControllerManagedBy(mgr).For(&redisMsvcv1.StandaloneService{})
	builder.Owns(&corev1.Secret{})
	builder.Owns(&redisMsvcv1.ACLConfigMap{})
	builder.Owns(fn.NewUnstructured(constants.HelmRedisType))

	builder.Watches(
		&source.Kind{Type: &appsv1.StatefulSet{}}, handler.EnqueueRequestsFromMapFunc(
			func(obj client.Object) []reconcile.Request {
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
