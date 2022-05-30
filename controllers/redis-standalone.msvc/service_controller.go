package redisstandalonemsvc

import (
	"context"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"fmt"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	redisStandalone "operators.kloudlite.io/apis/redis-standalone.msvc/v1"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/errors"
	fn "operators.kloudlite.io/lib/functions"
	rApi "operators.kloudlite.io/lib/operator"
	"operators.kloudlite.io/lib/templates"
)

// ServiceReconciler reconciles a Service object
type ServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=redis-standalone.msvc.kloudlite.io,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=redis-standalone.msvc.kloudlite.io,resources=services/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=redis-standalone.msvc.kloudlite.io,resources=services/finalizers,verbs=update

type ServiceReconReq struct {
	stateData map[string]any
	logger    *zap.SugaredLogger
	redisSvc  *redisStandalone.Service
}

const (
	RedisPasswordKey string = "redis-password"
	AclAccountsKey   string = "acl-accounts"
)

type Output struct {
	RootPassword string `json:"ROOT_PASSWORD"`
	Hosts        string `json:"HOSTS"`
	Uri          string `json:"URI"`
}

func (r *ServiceReconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	req := rApi.NewRequest(ctx, r.Client, oReq.NamespacedName, &redisStandalone.Service{})

	if req == nil {
		return ctrl.Result{}, nil
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.Result(), x.Err()
		}
	}

	req.Logger.Info("-------------------- NEW RECONCILATION------------------")

	if x := req.EnsureLabels(); !x.ShouldProceed() {
		return x.Result(), x.Err()
	}

	if x := r.reconcileStatus(req); !x.ShouldProceed() {
		return x.Result(), x.Err()
	}

	if x := r.reconcileOperations(req); !x.ShouldProceed() {
		return x.Result(), x.Err()
	}

	return ctrl.Result{}, nil
}

func (r *ServiceReconciler) finalize(req *rApi.Request[*redisStandalone.Service]) rApi.StepResult {
	return req.Finalize()
}

func (r *ServiceReconciler) reconcileStatus(req *rApi.Request[*redisStandalone.Service]) rApi.StepResult {
	ctx := req.Context()
	redisSvc := req.Object

	isReady := true
	var cs []metav1.Condition

	helmConditions, err := conditions.FromResource(
		ctx, r.Client, constants.HelmRedisGroup, "Helm",
		fn.NamespacedName(redisSvc),
	)
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.FailWithStatusError(err)
		}
		isReady = false
	}
	cs = append(cs, helmConditions...)

	deplConditions, err := conditions.FromResource(
		ctx, r.Client, constants.StatefulsetGroup, "Statefulset",
		fn.NamespacedName(redisSvc),
	)
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.FailWithStatusError(err)
		}
		isReady = false
	}
	cs = append(cs, deplConditions...)

	// STEP: Helm Release Secret
	helmSecret, err := rApi.Get(ctx, r.Client, fn.NN(redisSvc.Namespace, redisSvc.Name), &corev1.Secret{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.FailWithStatusError(err)
		}
		isReady = false
		cs = append(cs, conditions.New("HelmReleaseSecretExists", false, "SecretNotFound", err.Error()))
		helmSecret = nil
	}

	if helmSecret != nil {
		cs = append(cs, conditions.New("HelmReleaseSecretExists", true, "SecretFound"))
	}

	// STEP: ACL configmap
	aclCfg, err := rApi.Get(ctx, r.Client, fn.NN(redisSvc.Namespace, redisSvc.Name), &corev1.ConfigMap{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.FailWithStatusError(err)
		}
		isReady = false
		cs = append(cs, conditions.New("ACLConfigExists", false, "ConfigMapNotFound", err.Error()))
		helmSecret = nil
	}

	if aclCfg != nil {
		cs = append(cs, conditions.New("ACLConfigExists", true, "ConfigMapFound"))
		rApi.SetLocal(req, "AclAccounts", aclCfg.Data)
	}

	// STEP: generated vars
	if _, ok := redisSvc.Status.GeneratedVars.GetString(RedisPasswordKey); !ok {
		cs = append(cs, conditions.New("GeneratedVars", false, "NotGeneratedYet"))
	}

	newConditions, hasUpdated, err := conditions.Patch(redisSvc.Status.Conditions, cs)
	if err != nil {
		return req.FailWithStatusError(errors.NewEf(err, "while patching conditions"))
	}

	if !hasUpdated && isReady == redisSvc.Status.IsReady {
		return req.Next()
	}

	redisSvc.Status.IsReady = isReady
	redisSvc.Status.Conditions = newConditions
	redisSvc.Status.OpsConditions = []metav1.Condition{}
	if err := r.Status().Update(ctx, redisSvc); err != nil {
		return req.FailWithStatusError(err)
	}
	return req.Done()
}

func (r *ServiceReconciler) reconcileOperations(req *rApi.Request[*redisStandalone.Service]) rApi.StepResult {
	ctx := req.Context()
	redisSvc := req.Object

	if !controllerutil.ContainsFinalizer(redisSvc, constants.CommonFinalizer) {
		controllerutil.AddFinalizer(redisSvc, constants.CommonFinalizer)
		controllerutil.AddFinalizer(redisSvc, constants.ForegroundFinalizer)

		if err := r.Update(ctx, redisSvc); err != nil {
			return req.FailWithStatusError(err)
		}
		return req.Next()
	}

	if !meta.IsStatusConditionTrue(redisSvc.Status.Conditions, "GeneratedVars") {
		if err := redisSvc.Status.GeneratedVars.Set(RedisPasswordKey, fn.CleanerNanoid(40)); err != nil {
			return req.FailWithOpError(err)
		}
		if err := r.Status().Update(ctx, redisSvc); err != nil {
			return req.FailWithOpError(err)
		}
		return req.Done()
	}

	aclAccountsMap, ok := rApi.GetLocal[map[string]string](req, "AclAccounts")
	if !ok {
		return req.FailWithOpError(errors.New("ACL account map not found"))
	}

	redisSvc.Spec.ACLAccounts = aclAccountsMap

	obj, err := templates.ParseObject(templates.RedisStandalone, redisSvc)
	if err != nil {
		return req.FailWithOpError(err)
	}
	err = fn.KubectlApply(req.Context(), r.Client, obj)
	if err != nil {
		return req.FailWithOpError(err)
	}

	redisPasswd, ok := redisSvc.Status.GeneratedVars.GetString(RedisPasswordKey)
	if !ok {
		return req.FailWithOpError(errors.Newf("Bad PreOps, should have had %s key in generatedVars", RedisPasswordKey))
	}
	hostUrl := fmt.Sprintf("%s-headless.%s.svc.cluster.local:6379", redisSvc.Name, redisSvc.Namespace)

	scrt := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("msvc-%s", redisSvc.Name),
			Namespace: redisSvc.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				fn.AsOwner(redisSvc, true),
			},
		},
		StringData: map[string]string{
			"ROOT_PASSWORD": redisPasswd,
			"HOSTS":         hostUrl,
			"URI":           fmt.Sprintf("redis://:%s@%s?allowUsernameInURI=true", redisPasswd, hostUrl),
		},
	}

	if err := fn.KubectlApply(ctx, r.Client, scrt); err != nil {
		return req.FailWithOpError(err)
	}

	cfgName := fmt.Sprintf("msvc-%s-acl-accounts", redisSvc.Name)
	aclCfg := new(corev1.ConfigMap)
	if err := r.Get(ctx, types.NamespacedName{Namespace: redisSvc.Namespace, Name: cfgName}, aclCfg); err != nil {
		if apiErrors.IsNotFound(err) {
			nCfg := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cfgName,
					Namespace: redisSvc.Namespace,
					OwnerReferences: []metav1.OwnerReference{
						fn.AsOwner(redisSvc, true),
					},
					Labels: redisSvc.GetLabels(),
				},
			}
			err := r.Create(ctx, nCfg)
			return req.FailWithStatusError(err)
		}
	}

	return req.Done()
}

func (r *ServiceReconciler) kWatcherMap(o client.Object) []reconcile.Request {
	labels := o.GetLabels()
	if s := labels["app.kubernetes.io/component"]; s != "mongodb" {
		return nil
	}
	if s := labels["app.kubernetes.io/name"]; s != "mongodb" {
		return nil
	}
	resourceName := labels["app.kubernetes.io/instance"]
	nn := types.NamespacedName{Namespace: o.GetNamespace(), Name: resourceName}
	return []reconcile.Request{
		{NamespacedName: nn},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	builder := ctrl.NewControllerManagedBy(mgr).For(&redisStandalone.Service{})

	kinds := []client.Object{
		fn.NewUnstructured(metav1.TypeMeta{APIVersion: constants.MsvcApiVersion, Kind: constants.HelmRedisKind}),
		&appsv1.StatefulSet{},
		&corev1.Pod{},
		&corev1.ConfigMap{},
	}

	for _, kind := range kinds {
		builder.Watches(
			&source.Kind{Type: kind},
			handler.EnqueueRequestsFromMapFunc(
				func(c client.Object) []reconcile.Request {
					s, ok := c.GetLabels()[fmt.Sprintf("%s/ref", redisStandalone.GroupVersion.Group)]
					if !ok {
						return nil
					}
					return []reconcile.Request{
						{NamespacedName: fn.NN(c.GetNamespace(), s)},
					}
				},
			),
		)
	}

	return builder.Complete(r)
}
