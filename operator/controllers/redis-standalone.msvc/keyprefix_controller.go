package redisstandalonemsvc

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"operators.kloudlite.io/lib/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	redisStandalone "operators.kloudlite.io/apis/redis-standalone.msvc/v1"
	"operators.kloudlite.io/controllers/crds"
	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/finalizers"
	fn "operators.kloudlite.io/lib/functions"
	reconcileResult "operators.kloudlite.io/lib/reconcile-result"
)

// KeyPrefixReconciler reconciles a KeyPrefix object
type KeyPrefixReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

type KeyPrefixReconReq struct {
	ctrl.Request
	stateData map[stateKey]string
	logger    *zap.SugaredLogger
	keyPrefix *redisStandalone.KeyPrefix
}

type stateKey string

const (
	RootPasswordKey stateKey = "root-password"
	HostsKey        stateKey = "hosts"
)

// +kubebuilder:rbac:groups=redis-standalone.msvc.kloudlite.io,resources=keyprefixes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=redis-standalone.msvc.kloudlite.io,resources=keyprefixes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=redis-standalone.msvc.kloudlite.io,resources=keyprefixes/finalizers,verbs=update

func (r *KeyPrefixReconciler) Reconcile(ctx context.Context, orgReq ctrl.Request) (ctrl.Result, error) {
	req := &KeyPrefixReconReq{
		Request:   orgReq,
		logger:    crds.GetLogger(orgReq.NamespacedName),
		keyPrefix: new(redisStandalone.KeyPrefix),
		stateData: map[stateKey]string{},
	}

	if err := r.Get(ctx, orgReq.NamespacedName, req.keyPrefix); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !req.keyPrefix.HasLabels() {
		req.keyPrefix.EnsureLabels()
		return ctrl.Result{}, r.Update(ctx, req.keyPrefix)
	}

	if req.keyPrefix.GetDeletionTimestamp() != nil {
		return r.finalize(ctx, req)
	}

	ctrlReq, err := r.reconcileStatus(ctx, req)
	if err != nil {
		return r.failWithErr(ctx, req, err)
	}
	if ctrlReq != nil {
		return *ctrlReq, nil
	}

	return r.reconcileOperations(ctx, req)
}

func (r *KeyPrefixReconciler) finalize(ctx context.Context, req *KeyPrefixReconReq) (ctrl.Result, error) {
	req.logger.Infof("finalizing: %+v", req.keyPrefix.NameRef())
	if err := r.Delete(
		ctx, &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": constants.MsvcApiVersion,
				"kind":       constants.HelmRedisKind,
				"metadata": map[string]interface{}{
					"name":      req.keyPrefix.Name,
					"namespace": req.keyPrefix.Namespace,
				},
			},
		},
	); err != nil {
		req.logger.Infof("could not delete helm resource: %+v", err)
		if !apiErrors.IsNotFound(err) {
			return reconcileResult.FailedE(err)
		}
	}
	controllerutil.RemoveFinalizer(req.keyPrefix, finalizers.MsvcCommonService.String())
	if err := r.Update(ctx, req.keyPrefix); err != nil {
		return reconcileResult.FailedE(err)
	}
	return reconcileResult.OK()
}

func (r *KeyPrefixReconciler) failWithErr(ctx context.Context, req *KeyPrefixReconReq, err error) (ctrl.Result, error) {
	fn.Conditions2.MarkNotReady(&req.keyPrefix.Status.OpsConditions, err, "ReconFailedWithErr")
	if err2 := r.Status().Update(ctx, req.keyPrefix); err2 != nil {
		if apiErrors.IsConflict(err2) {
			return reconcileResult.OK()
		}
		return ctrl.Result{}, err2
	}
	//return reconcileResult.FailedE(err)
	return ctrl.Result{}, nil

}

func (r *KeyPrefixReconciler) reconcileStatus(ctx context.Context, req *KeyPrefixReconReq) (*ctrl.Result, error) {
	var conditions []metav1.Condition

	redisMsvcSecret := new(corev1.Secret)
	nn := types.NamespacedName{
		Namespace: req.keyPrefix.Namespace, Name: fmt.Sprintf("msvc-%s", req.keyPrefix.Spec.ManagedSvcName),
	}

	if err := r.Get(ctx, nn, redisMsvcSecret); err != nil {
		if !apiErrors.IsNotFound(err) {
			return nil, err
		}
		req.logger.Infof("secret for managed resource (%s) is not available yet, aborting ...", nn)
		fn.Conditions2.Build(
			&conditions, "Msvc", metav1.Condition{
				Type:    "OutputExists",
				Status:  metav1.ConditionFalse,
				Reason:  "SecretNotFound",
				Message: err.Error(),
			},
		)
		redisMsvcSecret = nil
	}

	if redisMsvcSecret != nil {
		req.stateData[RootPasswordKey] = string(redisMsvcSecret.Data["ROOT_PASSWORD"])
		req.stateData[HostsKey] = string(redisMsvcSecret.Data["HOSTS"])
	}

	// output cfgMap exists
	cfgMap := new(corev1.ConfigMap)
	if err := r.Get(
		ctx,
		types.NamespacedName{
			Namespace: req.keyPrefix.Namespace, Name: fmt.Sprintf("msvc-%s-acl-accounts", req.keyPrefix.Spec.ManagedSvcName),
		},
		cfgMap,
	); err != nil {
		if !apiErrors.IsNotFound(err) {
			return nil, err
		}
		fn.Conditions2.Build(
			&conditions,
			"Output", metav1.Condition{
				Type:    "NotFound",
				Status:  "True",
				Reason:  "ConfigmapNotFound",
				Message: fmt.Sprintf("keyPrefix output does not exist yet"),
			},
		)
		cfgMap = nil
	}

	if cfgMap != nil {
		if _, ok := cfgMap.Data[req.keyPrefix.Name]; ok {
			fn.Conditions2.Build(
				&conditions,
				"ACLAccounts", metav1.Condition{
					Type:    "EntryExists",
					Status:  metav1.ConditionTrue,
					Reason:  "Exists",
					Message: "Entry in configmap exists",
				},
			)
		}
	}

	if fn.Conditions2.Equal(conditions, req.keyPrefix.Status.Conditions) {
		req.logger.Infof("Status is already in sync, so moving forward with ops")
		return nil, nil
	}

	req.logger.Infof("status is different, so updating status ...")
	req.logger.Debugf("conditions: %+v", conditions)
	req.logger.Debugf("req.keyPrefix.Status.Conditions: %+v", req.keyPrefix.Status.Conditions)
	req.keyPrefix.Status.Conditions = conditions
	if err := r.Status().Update(ctx, req.keyPrefix); err != nil {
		return nil, err
	}

	return reconcileResult.OKP()
}

func newRedisClient(ctx context.Context, hosts, authPassword string) (*redis.Client, error) {
	rCli := redis.NewClient(
		&redis.Options{
			Addr:     hosts,
			Password: authPassword,
		},
	)
	if rCli == nil {
		return nil, errors.Newf("could not build redis client")
	}

	if err := rCli.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return rCli, nil
	//if err := rCli.Ping(ctx).Err(); err != nil {
	//	fn.Conditions2.Build(
	//		&conditions,
	//		"Redis", metav1.Condition{
	//			Type:    "PingFailed",
	//			Status:  "True",
	//			Reason:  "BadHostAddr",
	//			Message: fmt.Sprintf("could not reach to redis hosts %s", req.GetStateData("HOSTS")),
	//		},
	//	)
	//}
}

func createACLAcc(
	ctx context.Context, redisCli *redis.Client, username, prefix,
	password string,
) error {
	if err := redisCli.Do(
		ctx,
		"ACL", "SETUSER", username, "on",
		fmt.Sprintf("~%s:*", prefix),
		"+@all", "-@dangerous", "+info", "resetpass", fmt.Sprintf(">%s", password),
	).Err(); err != nil {
		return err
	}
	return nil
	//return redisCli.Do(ctx, "CONFIG", "REWRITE").Err()
}

func (r *KeyPrefixReconciler) preOps(ctx context.Context, req *KeyPrefixReconReq) error {
	//authP, b := rawJson.Get[string](req.keyPrefix.Status.GeneratedVars, "auth-password")
	//if _, ok := gVars["auth-password"]; !ok {
	//	if gVars == nil {
	//		gVars = map[string]any{}
	//	}
	//	gVars["auth-password"] = fn.CleanerNanoid(40)
	//	// FIXME
	//	//if err := req.keyPrefix.Status.GeneratedVars.Patch(gVars); err != nil {
	//	//	return err
	//	//}
	//	return r.Status().Update(ctx, req.keyPrefix)
	//}
	return nil
}

func (r *KeyPrefixReconciler) reconcileOperations(ctx context.Context, req *KeyPrefixReconReq) (ctrl.Result, error) {
	rootPassword := req.stateData[RootPasswordKey]
	hosts := req.stateData[HostsKey]

	if err := r.preOps(ctx, req); err != nil {
		return ctrl.Result{}, err
	}

	redisCli, err := newRedisClient(ctx, hosts, rootPassword)
	if redisCli == nil || err != nil {
		return reconcileResult.FailedE(err)
	}

	authPasswd, err := fn.JsonGet[string](req.keyPrefix.Status.GeneratedVars, "auth-password")
	if err != nil {
		return r.failWithErr(ctx, req, err)
	}

	if err := createACLAcc(
		ctx, redisCli, req.keyPrefix.Name, req.keyPrefix.Name,
		authPasswd,
	); err != nil {
		return r.failWithErr(ctx, req, err)
	}

	cfgMap := new(corev1.ConfigMap)
	if err := r.Get(
		ctx,
		types.NamespacedName{
			Namespace: req.keyPrefix.Namespace, Name: fmt.Sprintf("msvc-%s-acl-accounts", req.keyPrefix.Spec.ManagedSvcName),
		},
		cfgMap,
	); err != nil {
		return reconcileResult.Failed()
	}

	if cfgMap.Data == nil {
		cfgMap.Data = map[string]string{}
	}
	cfgMap.Data[req.keyPrefix.Name] = fmt.Sprintf(
		"USER %s on ~%s:* +@all -@dangerous +info resetpass >%s",
		req.keyPrefix.Name,
		req.keyPrefix.Name,
		authPasswd,
	)
	if err := fn.KubectlApply(ctx, r.Client, cfgMap); err != nil {
		return ctrl.Result{}, err
	}
	if err := r.reconcileOutput(ctx, req); err != nil {
		return r.failWithErr(ctx, req, err)
	}
	return reconcileResult.OK()
}

func (r *KeyPrefixReconciler) reconcileOutput(ctx context.Context, req *KeyPrefixReconReq) error {
	password, ok := req.keyPrefix.Status.GeneratedVars.Get("auth-password")
	if !ok {
		return errors.New("auth-password not found")
	}
	scrt := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("mres-%s", req.keyPrefix.Name),
			Namespace: req.keyPrefix.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				fn.AsOwner(req.keyPrefix, true),
			},
			Labels: req.keyPrefix.GetLabels(),
		},
		StringData: map[string]string{
			"USERNAME":   req.keyPrefix.Name,
			"PASSWORD":   password.(string),
			"KEY_PREFIX": req.keyPrefix.Name,
			"HOSTS":      req.stateData[HostsKey],
			"URI": fmt.Sprintf(
				"redis://%s:%s@%s?allowUsernameInUri=true", req.keyPrefix.Name, password, req.stateData[HostsKey],
			),
		},
	}
	return fn.KubectlApply(ctx, r.Client, scrt)
}

// SetupWithManager sets up the controller with the Manager.
func (r *KeyPrefixReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&redisStandalone.KeyPrefix{}).
		Watches(
			&source.Kind{Type: &corev1.Secret{}}, handler.EnqueueRequestsFromMapFunc(
				func(obj client.Object) []reconcile.Request {
					list := new(redisStandalone.KeyPrefixList)
					if err := r.List(
						context.TODO(), list, &client.ListOptions{
							LabelSelector: labels.SelectorFromValidatedSet(obj.GetLabels()),
						},
					); err != nil {
						return nil
					}

					var reqs []reconcile.Request
					for _, item := range list.Items {
						reqs = append(
							reqs, reconcile.Request{NamespacedName: types.NamespacedName{Namespace: item.Namespace, Name: item.Name}},
						)
					}
					return reqs
				},
			),
		).
		Complete(r)
}
