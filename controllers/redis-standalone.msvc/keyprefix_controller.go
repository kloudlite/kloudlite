package redisstandalonemsvc

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/go-redis/redis/v8"

	redisStandalone "operators.kloudlite.io/apis/redis-standalone.msvc/v1"
	"operators.kloudlite.io/controllers/crds"
	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/finalizers"
	fn "operators.kloudlite.io/lib/functions"
	reconcileResult "operators.kloudlite.io/lib/reconcile-result"
	"operators.kloudlite.io/lib/templates"
	t "operators.kloudlite.io/lib/types"
)

// KeyPrefixReconciler reconciles a KeyPrefix object
type KeyPrefixReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

type KeyPrefixReconReq struct {
	t.ReconReq
	ctrl.Request
	logger      *zap.SugaredLogger
	condBuilder fn.StatusConditions
	keyPrefix   *redisStandalone.KeyPrefix
	redisCli    *redis.Client
}

// +kubebuilder:rbac:groups=redis-standalone.msvc.kloudlite.io,resources=keyprefixes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=redis-standalone.msvc.kloudlite.io,resources=keyprefixes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=redis-standalone.msvc.kloudlite.io,resources=keyprefixes/finalizers,verbs=update

func (r *KeyPrefixReconciler) Reconcile(ctx context.Context, orgReq ctrl.Request) (ctrl.Result, error) {
	req := &KeyPrefixReconReq{
		Request:   orgReq,
		logger:    crds.GetLogger(orgReq.NamespacedName),
		keyPrefix: new(redisStandalone.KeyPrefix),
	}

	if err := r.Get(ctx, orgReq.NamespacedName, req.keyPrefix); err != nil {
		if apiErrors.IsNotFound(err) {
			return reconcileResult.OK()
		}
		return reconcileResult.Failed()
	}

	req.condBuilder = fn.Conditions.From(req.keyPrefix.Status.Conditions)

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

func (r *KeyPrefixReconciler) statusUpdate(ctx context.Context, req *KeyPrefixReconReq) error {
	req.keyPrefix.Status.Conditions = req.condBuilder.GetAll()
	req.logger.Infof("Conditions: %+v", req.keyPrefix.Status.Conditions)
	return r.Status().Update(ctx, req.keyPrefix)
}

func (r *KeyPrefixReconciler) failWithErr(ctx context.Context, req *KeyPrefixReconReq, err error) (ctrl.Result, error) {
	req.condBuilder.MarkNotReady(err)
	if err2 := r.statusUpdate(ctx, req); err2 != nil {
		return ctrl.Result{}, err2
	}
	return reconcileResult.FailedE(err)
}

func (r *KeyPrefixReconciler) reconcileStatus(ctx context.Context, req *KeyPrefixReconReq) (*ctrl.Result, error) {
	prevConditions := req.keyPrefix.Status.Conditions
	req.condBuilder.Reset()

	redisMsvcSecret := new(corev1.Secret)
	nn := types.NamespacedName{
		Namespace: req.keyPrefix.Namespace, Name: fmt.Sprintf("msvc-%s", req.keyPrefix.Spec.ManagedSvcName),
	}
	if err := r.Get(ctx, nn, redisMsvcSecret); err != nil {
		if !apiErrors.IsNotFound(err) {
			return nil, err
		}
		req.logger.Infof("secret for managed resource (%s) is not available yet, aborting ...", nn)
		req.condBuilder.MarkNotReady(err, "MsvcSecretNotFound")
	}

	rootPassword := string(redisMsvcSecret.Data["ROOT_PASSWORD"])
	hosts := string(redisMsvcSecret.Data["HOSTS"])

	req.SetStateData("ROOT_PASSWORD", rootPassword)
	req.SetStateData("HOSTS", hosts)

	rCli := redis.NewClient(
		&redis.Options{
			Addr:     hosts,
			Password: rootPassword,
		},
	)

	if err := rCli.Ping(ctx).Err(); err != nil {
		req.condBuilder.MarkNotReady(err, "IrrReconcilable")
		req.condBuilder.Build(
			"Redis", metav1.Condition{
				Type:    "PingFailed",
				Status:  "True",
				Reason:  "BadHostAddr",
				Message: fmt.Sprintf("could not reach to redis hosts %s", req.GetStateData("HOSTS")),
			},
		)
		req.logger.Debugf("ping failed, so aborting  conditions: ... %+v", req.condBuilder.GetAll())
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
		req.condBuilder.Build(
			"Output", metav1.Condition{
				Type:    "NotFound",
				Status:  "True",
				Reason:  "NotExists",
				Message: fmt.Sprintf("keyPrefix output does not exist yet"),
			},
		)
		cfgMap = nil
	}

	if cfgMap != nil {
		if _, ok := cfgMap.Data[req.keyPrefix.Name]; !ok {
			req.SetStateData("UserPassword", fn.CleanerNanoid(40))
		} else {
			req.SetStateData("UserExists", "true")
			req.condBuilder.Build(
				"ACLAccounts", metav1.Condition{
					Type:    "EntryExists",
					Status:  metav1.ConditionTrue,
					Reason:  "Exists",
					Message: "",
				},
			)
		}
	}

	req.redisCli = rCli

	req.logger.Debugf("prevConditions: %+v", prevConditions)
	req.logger.Debugf("req.condBuilder.GetAll(): %+v", req.condBuilder.GetAll())

	if req.condBuilder.Equal(prevConditions) {
		req.logger.Infof("Status is already in sync, so moving forward with ops")
		return nil, nil
	}

	req.logger.Infof("status is different, so updating status ...")
	if err := r.statusUpdate(ctx, req); err != nil {
		return nil, err
	}

	return reconcileResult.OKP()
}

func (r *KeyPrefixReconciler) createACLAcc(
	ctx context.Context, redisCli *redis.Client, username, prefix,
	password string,
) error {
	return redisCli.Do(
		ctx,
		"ACL",
		"SETUSER",
		username,
		"on",
		fmt.Sprintf("~%s:*", prefix),
		fmt.Sprintf("+@all -@dangerous +info resetpass >%s", password),
	).Err()
}

func (r *KeyPrefixReconciler) reconcileOperations(ctx context.Context, req *KeyPrefixReconReq) (ctrl.Result, error) {
	password := req.GetStateData("UserPassword")
	userExists := req.GetStateData("UserExists")
	if userExists == "true" {
		return reconcileResult.OK()
	}

	inputs, err := fn.Json.FromRawMessage(req.keyPrefix.Spec.Inputs)
	if err != nil {
		return reconcileResult.FailedE(err)
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
		inputs["prefix"],
		password,
	)

	b, err := templates.Parse(templates.ConfigMap, cfgMap)
	if err != nil {
		return reconcileResult.FailedE(err)
	}
	req.logger.Debugf("b: %s", b)
	stdout, err := fn.KubectlApply(b)
	if err != nil {
		return ctrl.Result{}, err
	}
	req.logger.Debugf("stdout: %+v", stdout)
	return reconcileResult.OK()
}

// SetupWithManager sets up the controller with the Manager.
func (r *KeyPrefixReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&redisStandalone.KeyPrefix{}).
		Complete(r)
}
