package redisstandalonemsvc

import (
	"context"
	"fmt"
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
	"operators.kloudlite.io/lib/redis"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// ACLAccountReconciler reconciles a ACLAccount object
type ACLAccountReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

type Key string

const (
	GeneratedVars = "GeneratedVars"
)

const (
	UserPassword = "UserPassword"
	RedisMsvcKey = "RedisMsvc"
)

// +kubebuilder:rbac:groups=redis-standalone.msvc.kloudlite.io,resources=aclaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=redis-standalone.msvc.kloudlite.io,resources=aclaccounts/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=redis-standalone.msvc.kloudlite.io,resources=aclaccounts/finalizers,verbs=update

func (r *ACLAccountReconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	req, _ := rApi.NewRequest(ctx, r.Client, oReq.NamespacedName, &redisStandalone.ACLAccount{})

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

func (r *ACLAccountReconciler) finalize(req *rApi.Request[*redisStandalone.ACLAccount]) rApi.StepResult {
	// TODO: ACL finalizer not deleting entry from ACL configmap
	ctx := req.Context()
	obj := req.Object

	// remove ACL Entry for user
	aclCfg, err := rApi.Get(
		ctx, r.Client, fn.NN(obj.GetNamespace(), ACLConfigMapName.Format(obj.Name)),
		&corev1.ConfigMap{},
	)
	if err != nil {
		if apiErrors.IsNotFound(err) {
			return req.Finalize()
		}
		return req.FailWithOpError(err)
	}

	delete(aclCfg.Data, obj.Name)

	if err := fn.KubectlApply(ctx, r.Client, fn.ParseConfigMap(aclCfg)); err != nil {
		return req.FailWithOpError(err)
	}

	return req.Finalize()
}

func (r *ACLAccountReconciler) reconcileStatus(req *rApi.Request[*redisStandalone.ACLAccount]) rApi.StepResult {
	ctx := req.Context()
	aclObj := req.Object

	var cs []metav1.Condition
	isReady := true

	// ASSERT: whether redis service is ready
	redisMsvc, err := rApi.Get(
		ctx,
		r.Client,
		fn.NN(aclObj.Namespace, aclObj.Spec.ManagedSvcName),
		&redisStandalone.Service{},
	)
	if err != nil {
		return req.FailWithStatusError(errors.NewEf(err, "could not get msvc"))
	}

	if !redisMsvc.Status.IsReady {
		return req.FailWithStatusError(errors.Newf("msvc is not ready"))
	}

	rApi.SetLocal(req, RedisMsvcKey, redisMsvc)

	// ASSERT: whether managed service output is available or not
	msvcOutput, err := rApi.Get(
		ctx,
		r.Client,
		fn.NN(redisMsvc.Namespace, fmt.Sprintf("msvc-%s", redisMsvc.Name)),
		&corev1.Secret{},
	)

	if err != nil {
		return req.FailWithStatusError(errors.NewEf(err, "msvc output is not available"))
	}

	if msvcOutput != nil {
		cs = append(cs, conditions.New("MsvcOutputExists", true, "Found"))
		rApi.SetLocal(req, "MsvcOutput", msvcOutput)
	}

	// CHECK: whether redis user exists
	redisCli, err := redis.NewClient(string(msvcOutput.Data["HOSTS"]), "", string(msvcOutput.Data["ROOT_PASSWORD"]))
	if err != nil {
		return req.FailWithStatusError(errors.NewEf(err, "could not create redis client"))
	}
	rApi.SetLocal(req, "RedisClient", redisCli)

	exists, err := redisCli.UserExists(ctx, aclObj.Name)
	if err != nil || !exists {
		isReady = false
		cs = append(cs, conditions.New("ACLUserExists", false, "NotFound", err.Error()))
	}
	if exists {
		cs = append(cs, conditions.New("ACLUserExists", true, "Found"))
	}

	// CHECK whether ACL config map is found or not
	aclCfg, err := rApi.Get(
		ctx, r.Client, fn.NN(redisMsvc.Namespace, fmt.Sprintf("msvc-%s-acl-accounts", redisMsvc.Name)),
		&corev1.ConfigMap{},
	)
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.FailWithStatusError(err)
		}
		isReady = false
		aclCfg = nil
		cs = append(cs, conditions.New("ACLConfigExists", false, "NotFound", err.Error()))
	}

	if aclCfg != nil {
		cs = append(cs, conditions.New("ACLConfigExists", true, "Found"))
		rApi.SetLocal(req, "AclConfigmap", aclCfg)
	}

	// CHECK: generated vars
	_, ok := aclObj.Status.GeneratedVars.GetString("UserPassword")
	if ok {
		cs = append(cs, conditions.New("GeneratedVars", true, "Generated"))
	} else {
		cs = append(cs, conditions.New("GeneratedVars", false, "NotGeneratedYet"))
	}

	// CHECK: output exists
	aclOutput, err := rApi.Get(
		ctx, r.Client, fn.NN(aclObj.Namespace, fmt.Sprintf("mres-%s", aclObj.Name)),
		&corev1.Secret{},
	)

	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.FailWithStatusError(err)
		}
		isReady = false
		aclOutput = nil
		cs = append(cs, conditions.New("OutputExists", false, "NotFound", err.Error()))
	}

	if aclOutput != nil {
		cs = append(cs, conditions.New("OutputExists", true, "Found"))
	}

	// STEP: save conditions and move ahead
	newConditions, hasUpdated, err := conditions.Patch(aclObj.Status.Conditions, cs)
	if err != nil {
		return req.FailWithStatusError(err)
	}

	if !hasUpdated && isReady == aclObj.Status.IsReady {
		return req.Next()
	}

	aclObj.Status.IsReady = isReady
	aclObj.Status.Conditions = newConditions
	aclObj.Status.OpsConditions = []metav1.Condition{}
	if err := r.Status().Update(ctx, aclObj); err != nil {
		return req.FailWithStatusError(err)
	}
	return req.Done()
}

func (r *ACLAccountReconciler) reconcileOperations(req *rApi.Request[*redisStandalone.ACLAccount]) rApi.StepResult {
	ctx := req.Context()
	aclAccObj := req.Object

	if !controllerutil.ContainsFinalizer(aclAccObj, constants.CommonFinalizer) {
		controllerutil.AddFinalizer(aclAccObj, constants.CommonFinalizer)
		controllerutil.AddFinalizer(aclAccObj, constants.ForegroundFinalizer)

		if err := r.Update(ctx, aclAccObj); err != nil {
			return req.FailWithStatusError(err)
		}
		return req.Next()
	}

	if meta.IsStatusConditionFalse(aclAccObj.Status.Conditions, GeneratedVars) {
		if err := aclAccObj.Status.GeneratedVars.Set(UserPassword, fn.CleanerNanoid(40)); err != nil {
			return req.FailWithOpError(err)
		}
		if err := r.Status().Update(ctx, aclAccObj); err != nil {
			return req.FailWithOpError(err)
		}
		return req.Done()
	}

	userPassword, ok := aclAccObj.Status.GeneratedVars.GetString(UserPassword)
	if !ok {
		return req.FailWithOpError(errors.Newf("%s not found in .Status.GeneratedVars", UserPassword))
	}

	// ACLUser Upsert
	prefix, ok := aclAccObj.Spec.Inputs.GetString("prefix")
	if !ok {
		return req.FailWithOpError(errors.Newf("prefix not found in .Spec.Inputs"))
	}

	redisCli, ok := rApi.GetLocal[*redis.Client](req, "RedisClient")
	if !ok {
		return req.FailWithOpError(errors.Newf("RedisClient key not found in req locals"))
	}

	if err := redisCli.UpsertUser(ctx, prefix, aclAccObj.Name, userPassword); err != nil {
		return req.FailWithOpError(errors.NewEf(err, "failed to upsert (user=%s)", aclAccObj.Name))
	}

	// ACLConfigMap entry
	aclCfg, ok := rApi.GetLocal[*corev1.ConfigMap](req, "AclConfigmap")
	if !ok {
		return req.FailWithOpError(errors.Newf("AclConfigmap not found in req locals"))
	}

	if aclCfg.Data == nil {
		aclCfg.Data = map[string]string{}
	}
	aclCfg.Data[aclAccObj.Name] = fmt.Sprintf(
		"user %s on ~%s:* +@all -@dangerous +info resetpass >%s",
		aclAccObj.Name,
		prefix,
		userPassword,
	)

	if err := fn.KubectlApply(ctx, r.Client, fn.ParseConfigMap(aclCfg)); err != nil {
		return req.FailWithOpError(err)
	}

	// OUTPUT
	msvcOutput, ok := rApi.GetLocal[*corev1.Secret](req, "MsvcOutput")
	if !ok {
		return req.FailWithOpError(errors.Newf("MsvcOutput does not exist in req locals"))
	}

	host := string(msvcOutput.Data["HOSTS"])

	scrt := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("mres-%s", aclAccObj.Name),
			Namespace: aclAccObj.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				fn.AsOwner(aclAccObj, true),
			},
		},
		StringData: map[string]string{
			"HOSTS":    host,
			"PASSWORD": userPassword,
			"URI":      fmt.Sprintf("redis://%s:%s@%s?allowUsernameInURI=true", aclAccObj.Name, userPassword, host),
		},
	}

	if err := fn.KubectlApply(ctx, r.Client, fn.ParseSecret(scrt)); err != nil {
		return req.FailWithOpError(err)
	}

	return req.Done()
}

// SetupWithManager sets up the controller with the Manager.
func (r *ACLAccountReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&redisStandalone.ACLAccount{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
