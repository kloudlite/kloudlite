package redisstandalonemsvc

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	redisStandalone "operators.kloudlite.io/apis/redis-standalone.msvc/v1"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/errors"
	fn "operators.kloudlite.io/lib/functions"
	rApi "operators.kloudlite.io/lib/operator"
	"operators.kloudlite.io/lib/templates"
	t "operators.kloudlite.io/lib/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// ServiceReconciler reconciles a Service object
type ServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
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

const (
	ACLConfigMapName t.Fstring = "msvc-%s-acl-accounts"
	RedisStsName     t.Fstring = "%s-master"
)

// +kubebuilder:rbac:groups=redis-standalone.msvc.kloudlite.io,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=redis-standalone.msvc.kloudlite.io,resources=services/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=redis-standalone.msvc.kloudlite.io,resources=services/finalizers,verbs=update

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

	// CHECK: fetch conditions from helm resource
	helmConditions, err := conditions.FromResource(
		ctx, r.Client, constants.HelmRedisGroup, "Helm",
		fn.NN(redisSvc.Namespace, redisSvc.Name),
	)
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.FailWithStatusError(err)
		}
		cs = append(cs, conditions.New("HelmResourceExists", false, "NotFound", err.Error()))
		isReady = false
		helmConditions = nil
	}

	if helmConditions != nil {
		cs = append(cs, conditions.New("HelmResourceExists", true, "Found"))
		cs = append(cs, helmConditions...)
	}

	// CHECK: fetch conditions from redis statefulset
	stsConditions, err := conditions.FromResource(
		ctx, r.Client, constants.StatefulsetGroup, "Statefulset",
		fn.NN(redisSvc.Namespace, RedisStsName.Format(redisSvc.Name)),
		// fn.NN(redisSvc.Namespace, fmt.Sprintf(RedisStsName, redisSvc.Name)),
	)
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.FailWithStatusError(err)
		}
		cs = append(cs, conditions.New("StatefulsetExists", false, "NotFound", err.Error()))
		isReady = false
		stsConditions = nil
	}

	if stsConditions != nil {
		cs = append(cs, conditions.New("StatefulsetExists", true, "Found"))
		cs = append(cs, stsConditions...)
	}

	// CHECK: whether Helm Release Secret Exists
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

	// CHECK: wether ACL configmap exists
	aclCfg, err := rApi.Get(
		ctx, r.Client, fn.NN(redisSvc.Namespace, ACLConfigMapName.Format(redisSvc.Name)),
		&corev1.ConfigMap{},
	)
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.FailWithStatusError(err)
		}
		isReady = false
		cs = append(cs, conditions.New("ACLConfigExists", false, "ConfigMapNotFound", err.Error()))
		aclCfg = nil
	}

	if aclCfg != nil {
		cs = append(cs, conditions.New("ACLConfigExists", true, "ConfigMapFound"))
		rApi.SetLocal(req, "AclAccounts", aclCfg.Data)
	}

	// CHECK: whether generated vars ?
	_, ok := redisSvc.Status.GeneratedVars.GetString(RedisPasswordKey)

	if ok {
		cs = append(cs, conditions.New("GeneratedVars", true, "Generated"))
	} else {
		cs = append(cs, conditions.New("GeneratedVars", false, "NotGeneratedYet"))
	}

	// CHECK: whether service output exists
	svcOutput, err := rApi.Get(
		ctx,
		r.Client,
		fn.NN(redisSvc.Namespace, fmt.Sprintf("msvc-%s", redisSvc.Name)), &corev1.Secret{},
	)

	if err != nil {
		isReady = false
		svcOutput = nil
		cs = append(cs, conditions.New("OutputExists", false, "NotFound", err.Error()))
	}

	if svcOutput != nil {
		cs = append(cs, conditions.New("OutputExists", true, "Found"))
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
	if ok && aclAccountsMap != nil {
		redisSvc.Spec.ACLAccounts = aclAccountsMap
	}

	obj, err := templates.ParseObject(templates.RedisStandalone, redisSvc)
	if err != nil {
		return req.FailWithOpError(err)
	}

	if err := fn.KubectlApply(ctx, r.Client, obj); err != nil {
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

	if err := fn.KubectlApply(ctx, r.Client, fn.ParseSecret(scrt)); err != nil {
		return req.FailWithOpError(err)
	}

	cfgName := fmt.Sprintf("msvc-%s-acl-accounts", redisSvc.Name)
	aclCfg := new(corev1.ConfigMap)
	if err := r.Get(ctx, fn.NN(redisSvc.Namespace, cfgName), aclCfg); err != nil {
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
			if err := r.Create(ctx, nCfg); err != nil {
				return req.FailWithOpError(err)
			}
		}
	}

	return req.Done()
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&redisStandalone.Service{}).
		Owns(fn.NewUnstructured(metav1.TypeMeta{Kind: constants.HelmRedisKind, APIVersion: constants.MsvcApiVersion})).
		Owns(&redisStandalone.ACLAccount{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Pod{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
