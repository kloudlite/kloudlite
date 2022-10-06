package redisstandalonemsvc

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	redisStandalone "operators.kloudlite.io/apis/redis-standalone.msvc/v1"
	"operators.kloudlite.io/env"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/errors"
	fn "operators.kloudlite.io/lib/functions"
	"operators.kloudlite.io/lib/logging"
	rApi "operators.kloudlite.io/lib/operator"
	stepResult "operators.kloudlite.io/lib/operator/step-result"
	libRedis "operators.kloudlite.io/lib/redis"
	"operators.kloudlite.io/lib/templates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// ACLAccountReconciler reconciles a ACLAccount object
type ACLAccountReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	logger logging.Logger
	Name   string
}

func (r *ACLAccountReconciler) GetName() string {
	return r.Name
}

type Key string

const (
	KeyUserPassword = "user-password"
)

const (
	ACLUserExists   conditions.Type = "ACLUserExists"
	ACLConfigExists conditions.Type = "ACLConfigExists"
)

type MsvcOutputRef struct {
	Hosts        string
	RootPassword string
	ACLConfig    *corev1.ConfigMap
}

func parseMsvcOutput(s *corev1.Secret, aclCfg *corev1.ConfigMap) *MsvcOutputRef {
	return &MsvcOutputRef{
		Hosts:        string(s.Data["HOSTS"]),
		RootPassword: string(s.Data["ROOT_PASSWORD"]),
		ACLConfig:    aclCfg,
	}
}

// +kubebuilder:rbac:groups=redis-standalone.msvc.kloudlite.io,resources=aclaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=redis-standalone.msvc.kloudlite.io,resources=aclaccounts/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=redis-standalone.msvc.kloudlite.io,resources=aclaccounts/finalizers,verbs=update

func (r *ACLAccountReconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, oReq.NamespacedName, &redisStandalone.ACLAccount{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	req.Logger.Infof("-------------------- NEW RECONCILATION------------------")

	if x := req.EnsureLabelsAndAnnotations(); !x.ShouldProceed() {
		return x.ReconcilerResponse()
	}

	if x := r.reconcileStatus(req); !x.ShouldProceed() {
		return x.ReconcilerResponse()
	}

	if x := r.reconcileOperations(req); !x.ShouldProceed() {
		return x.ReconcilerResponse()
	}

	return ctrl.Result{}, nil
}

func (r *ACLAccountReconciler) finalize(req *rApi.Request[*redisStandalone.ACLAccount]) stepResult.Result {
	// TODO: ACL finalizer not deleting entry from ACL configmap
	ctx := req.Context()
	obj := req.Object

	// remove ACL Entry for user
	aclCfg, err := rApi.Get(
		ctx, r.Client, fn.NN(obj.GetNamespace(), getACLConfigmapName(obj.Name)),
		&corev1.ConfigMap{},
	)

	if err != nil {
		if apiErrors.IsNotFound(err) {
			return req.Finalize()
		}
		return req.FailWithOpError(err)
	}

	patch := client.MergeFrom(aclCfg)
	delete(aclCfg.Data, obj.Name)
	if err := r.Patch(ctx, obj, patch); err != nil {
		return req.FailWithOpError(err)
	}

	return req.Finalize()
}

func (r *ACLAccountReconciler) reconcileStatus(req *rApi.Request[*redisStandalone.ACLAccount]) stepResult.Result {
	ctx := req.Context()
	obj := req.Object

	isReady := true
	var cs []metav1.Condition

	// STEP: 1. check managed service is ready
	msvc, err := rApi.Get(
		ctx, r.Client, fn.NN(obj.Namespace, obj.Spec.ManagedSvcName),
		&redisStandalone.Service{},
	)

	if err != nil {
		isReady = false
		cs = append(cs, conditions.New(conditions.ManagedSvcExists, false, conditions.NotFound, err.Error()))
		return req.FailWithStatusError(err, cs...).Err(nil)
	}
	cs = append(cs, conditions.New(conditions.ManagedSvcExists, true, conditions.Found))
	cs = append(cs, conditions.New(conditions.ManagedSvcReady, msvc.Status.IsReady, conditions.Empty))
	if !msvc.Status.IsReady {
		isReady = false
		return req.FailWithStatusError(err, cs...).Err(nil)
	}

	// STEP: 2. retrieve managed svc output (usually secret)
	msvcRef, err2 := func() (*MsvcOutputRef, error) {
		msvcOutput, err := rApi.Get(
			ctx, r.Client, fn.NN(msvc.Namespace, fmt.Sprintf("msvc-%s", msvc.Name)),
			&corev1.Secret{},
		)
		if err != nil {
			isReady = false
			cs = append(cs, conditions.New(conditions.ManagedSvcOutputExists, false, conditions.NotFound, err.Error()))
			return nil, err
		}
		cs = append(cs, conditions.New(conditions.ManagedSvcOutputExists, true, conditions.Found))

		// acl-config
		aclCfg, err := rApi.Get(
			ctx, r.Client, fn.NN(msvc.Namespace, fmt.Sprintf("msvc-%s-acl-account", msvc.Name)),
			&corev1.ConfigMap{},
		)
		if err != nil {
			isReady = false
			cs = append(cs, conditions.New(ACLConfigExists, false, conditions.NotFound, err.Error()))
			return nil, err
		}
		cs = append(cs, conditions.New(ACLConfigExists, true, conditions.Found))

		outputRef := parseMsvcOutput(msvcOutput, aclCfg)
		rApi.SetLocal(req, "msvc-output-ref", outputRef)
		return outputRef, nil
	}()
	if err2 != nil {
		return req.FailWithStatusError(err2)
	}

	if err2 := func() error {
		// STEP: 3. check reconciler (child components e.g. mongo account, s3 bucket, redis ACL user) exists
		// TODO: (user) use msvcRef values
		redisCli, err := libRedis.NewClient(msvcRef.Hosts, "", msvcRef.RootPassword)
		if err != nil {
			return errors.NewEf(err, "could not create redis client")
		}
		defer redisCli.Close()

		exists, err := redisCli.UserExists(ctx, obj.Name)
		if err != nil {
			return err
		}
		if !exists {
			isReady = false
			cs = append(cs, conditions.New(ACLUserExists, false, conditions.NotFound))
			return nil
		}
		cs = append(cs, conditions.New(ACLUserExists, true, conditions.Found))
		return nil
	}(); err2 != nil {
		isReady = false
		return req.FailWithStatusError(err2)
	}

	// STEP: 4. check generated vars
	if msvc != nil && !obj.Status.GeneratedVars.Exists(KeyUserPassword) {
		cs = append(cs, conditions.New(conditions.GeneratedVars, false, conditions.NotReconciledYet))
	} else {
		cs = append(cs, conditions.New(conditions.GeneratedVars, true, conditions.Found))
	}

	// STEP: 5. reconciler output exists?
	_, err5 := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, fmt.Sprintf("mres-%s", obj.Name)), &corev1.Secret{})
	if err5 != nil {
		cs = append(cs, conditions.New(conditions.ReconcilerOutputExists, false, conditions.NotFound, err5.Error()))
	} else {
		cs = append(cs, conditions.New(conditions.ReconcilerOutputExists, true, conditions.Found))
	}

	// STEP: 6. patch conditions
	newConditions, updated, err := conditions.Patch(obj.Status.Conditions, cs)
	if err != nil {
		return req.FailWithStatusError(err)
	}

	if !updated && isReady == obj.Status.IsReady {
		return req.Next()
	}

	obj.Status.IsReady = isReady
	obj.Status.Conditions = newConditions
	if err := r.Status().Update(ctx, obj); err != nil {
		return req.FailWithStatusError(err)
	}
	return req.Done()
}

func (r *ACLAccountReconciler) reconcileOperations(req *rApi.Request[*redisStandalone.ACLAccount]) stepResult.Result {
	ctx := req.Context()
	obj := req.Object

	// STEP: 1. add finalizers if needed
	if !controllerutil.ContainsFinalizer(obj, constants.CommonFinalizer) {
		controllerutil.AddFinalizer(obj, constants.CommonFinalizer)
		controllerutil.AddFinalizer(obj, constants.ForegroundFinalizer)

		if err := r.Update(ctx, obj); err != nil {
			return req.FailWithOpError(err)
		}
		return req.Done()
	}

	// STEP: 2. generate vars if needed to
	if meta.IsStatusConditionFalse(obj.Status.Conditions, conditions.GeneratedVars.String()) {
		if err := obj.Status.GeneratedVars.Set(KeyUserPassword, fn.CleanerNanoid(40)); err != nil {
			return req.FailWithStatusError(err)
		}
		if err := r.Status().Update(ctx, obj); err != nil {
			return req.FailWithOpError(err)
		}
		return req.Done()
	}

	// STEP: 3. retrieve msvc output, need it in creating reconciler output
	msvcRef, ok := rApi.GetLocal[*MsvcOutputRef](req, "msvc-output-ref")
	if !ok {
		return req.FailWithOpError(errors.Newf("err=%s key not found in req locals", "msvc-output-ref")).Err(nil)
	}

	userPassword, ok := obj.Status.GeneratedVars.GetString(KeyUserPassword)
	if !ok {
		return req.FailWithOpError(errors.Newf("key=%s not present in .Status.GeneratedVars", KeyUserPassword)).Err(nil)
	}

	// STEP: 4. create child components like mongo-user, redis-acl etc.
	err4 := func() error {
		redisCli, err := libRedis.NewClient(msvcRef.Hosts, "", msvcRef.RootPassword)
		if err != nil {
			return errors.NewEf(err, "could not create redis client")
		}
		defer redisCli.Close()

		return redisCli.UpsertUser(ctx, obj.Spec.KeyPrefix, obj.Name, userPassword)
	}()
	if err4 != nil {
		// TODO:(user) might need to reconcile with retry with timeout error
		return req.FailWithOpError(err4)
	}

	// STEP: 5. create reconciler output (eg. secret)

	b, err := templates.Parse(
		templates.CoreV1.Secret, map[string]any{
			"name":       "mres-" + obj.Name,
			"namespace":  obj.Namespace,
			"labels":     obj.GetLabels(),
			"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},
			"string-data": map[string]string{
				"HOSTS":    msvcRef.Hosts,
				"PASSWORD": userPassword,
				"USERNAME": obj.Name,
				"PREFIX":   obj.Spec.KeyPrefix,
				"URI":      fmt.Sprintf("redis://%s:%s@%s?allowUsernameInURI=true", obj.Name, userPassword, msvcRef.Hosts),
			},
		},
	)
	if err != nil {
		return req.FailWithOpError(err).Err(nil)
	}

	if err := fn.KubectlApplyExec(ctx, b); err != nil {
		return req.FailWithOpError(err).Err(nil)
	}

	// patch := client.StrategicMergeFrom(msvcRef.ACLConfig)
	if msvcRef.ACLConfig.Data == nil {
		msvcRef.ACLConfig.Data = map[string]string{}
	}
	msvcRef.ACLConfig.Data[obj.Name] = fmt.Sprintf(
		"user %s on ~%s:* +@all -@dangerous +info resetpass >%s", obj.Name, obj.Spec.KeyPrefix, userPassword,
	)

	if err := r.Update(ctx, msvcRef.ACLConfig); err != nil {
		return req.FailWithOpError(err)
	}

	// if err := r.Client.Patch(ctx, msvcRef.ACLConfig, patch); err != nil {
	// 	return req.FailWithOpError(err)
	// }

	obj.Status.OpsConditions = []metav1.Condition{}
	if err := r.Status().Update(ctx, obj); err != nil {
		return req.FailWithOpError(err)
	}
	return req.Next()
}

// SetupWithManager sets up the controllers with the Manager.
func (r *ACLAccountReconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)

	builder := ctrl.NewControllerManagedBy(mgr).For(&redisStandalone.ACLAccount{})
	builder.Owns(&corev1.Secret{})
	builder.Watches(
		&source.Kind{Type: &redisStandalone.Service{}}, handler.EnqueueRequestsFromMapFunc(
			func(obj client.Object) []reconcile.Request {

				var aclList redisStandalone.ACLAccountList
				if err := r.List(
					context.TODO(), &aclList, &client.ListOptions{
						Namespace: obj.GetNamespace(),
						LabelSelector: labels.SelectorFromValidatedSet(
							map[string]string{
								constants.MsvcNameKey: obj.GetLabels()[constants.MsvcNameKey],
							},
						),
					},
				); err != nil {
					return nil
				}

				requests := make([]reconcile.Request, 0, len(aclList.Items))
				for _, service := range aclList.Items {
					requests = append(requests, reconcile.Request{NamespacedName: fn.NN(service.GetNamespace(), service.GetName())})
				}

				return requests
			},
		),
	)
	return builder.Complete(r)
}
