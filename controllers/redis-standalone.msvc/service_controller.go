package redisstandalonemsvc

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ct "operators.kloudlite.io/apis/common-types"
	redisStandalone "operators.kloudlite.io/apis/redis-standalone.msvc/v1"
	"operators.kloudlite.io/env"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/errors"
	fn "operators.kloudlite.io/lib/functions"
	"operators.kloudlite.io/lib/logging"
	rApi "operators.kloudlite.io/lib/operator"
	stepResult "operators.kloudlite.io/lib/operator/step-result"
	"operators.kloudlite.io/lib/templates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// ServiceReconciler reconciles a Service object
type ServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	logger logging.Logger
	Name   string
}

func (r *ServiceReconciler) GetName() string {
	return r.Name
}

const (
	RedisPasswordKey  string = "redis-password"
	KeyAclAccountsMap string = "acl-accounts-map"
)

const (
	ACLConfigMapExists conditions.Type = "ACLConfigMapExists"
)

func getACLConfigmapName(name string) string {
	return fmt.Sprintf("msvc-%s-acl-accounts", name)
}

// +kubebuilder:rbac:groups=redis-standalone.msvc.kloudlite.io,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=redis-standalone.msvc.kloudlite.io,resources=services/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=redis-standalone.msvc.kloudlite.io,resources=services/finalizers,verbs=update

func (r *ServiceReconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, oReq.NamespacedName, &redisStandalone.Service{})
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

func (r *ServiceReconciler) finalize(req *rApi.Request[*redisStandalone.Service]) stepResult.Result {
	return req.Finalize()
}

func (r *ServiceReconciler) reconcileStatus(req *rApi.Request[*redisStandalone.Service]) stepResult.Result {
	ctx := req.Context()
	obj := req.Object

	isReady := true
	var cs []metav1.Condition
	var childC []metav1.Condition

	// STEP: 1. sync conditions from CRs of helm/custom controllers
	helmResource, err := rApi.Get(
		ctx, r.Client, fn.NN(obj.Namespace, obj.Name), fn.NewUnstructured(constants.HelmRedisType),
	)

	if err != nil {
		isReady = false
		if !apiErrors.IsNotFound(err) {
			return req.FailWithStatusError(err)
		}
		cs = append(cs, conditions.New(conditions.HelmResourceExists, false, conditions.NotFound, err.Error()))
	} else {
		cs = append(cs, conditions.New(conditions.HelmResourceExists, true, conditions.Found))

		rConditions, err := conditions.ParseFromResource(helmResource, "Helm")
		if err != nil {
			return req.FailWithStatusError(err)
		}
		childC = append(childC, rConditions...)
		rReady := meta.IsStatusConditionTrue(rConditions, "HelmDeployed")
		if !rReady {
			isReady = false
		}
		cs = append(
			cs, conditions.New(conditions.HelmResourceReady, rReady, conditions.Empty),
		)
	}

	// STEP: 2. sync conditions from deployments/statefulsets
	stsRes, err := rApi.Get(
		ctx, r.Client, fn.NN(obj.Namespace, fmt.Sprintf("%s-master", obj.Name)), &appsv1.StatefulSet{},
	)
	if err != nil {
		isReady = false
		if !apiErrors.IsNotFound(err) {
			return req.FailWithStatusError(err)
		}
		cs = append(cs, conditions.New(conditions.StsExists, false, conditions.NotFound, err.Error()))
	} else {
		cs = append(cs, conditions.New(conditions.StsExists, true, conditions.Found))
		rConditions, err := conditions.ParseFromResource(stsRes, "Sts")
		if err != nil {
			return req.FailWithStatusError(err)
		}
		childC = append(childC, rConditions...)

		if stsRes.Status.Replicas != stsRes.Status.ReadyReplicas {
			isReady = false
			cs = append(cs, conditions.New(conditions.StsReady, false, conditions.Empty))
		} else {
			cs = append(cs, conditions.New(conditions.StsReady, true, conditions.Empty))
		}
	}

	// STEP: 3. if vars generated ?
	if !obj.Status.GeneratedVars.Exists(RedisPasswordKey) {
		isReady = false
		cs = append(
			cs, conditions.New(
				conditions.GeneratedVars, false, conditions.NotReconciledYet,
			),
		)
	} else {
		cs = append(cs, conditions.New(conditions.GeneratedVars, true, conditions.Found))
	}

	// STEP: 4. if reconciler output exists
	_, err = rApi.Get(
		ctx, r.Client, fn.NN(obj.Namespace, fmt.Sprintf("msvc-%s", obj.Name)), &corev1.Secret{},
	)
	if err != nil {
		isReady = false
		cs = append(cs, conditions.New(conditions.ReconcilerOutputExists, false, conditions.NotFound, err.Error()))
	} else {
		cs = append(cs, conditions.New(conditions.ReconcilerOutputExists, true, conditions.Found))
	}

	// acl config exists ?
	aclCfg, err := rApi.Get(
		ctx, r.Client, fn.NN(obj.Namespace, getACLConfigmapName(obj.Name)), &corev1.ConfigMap{},
	)
	if err != nil {
		isReady = false
		cs = append(cs, conditions.New(ACLConfigMapExists, false, conditions.NotFound, err.Error()))
		rApi.SetLocal(req, KeyAclAccountsMap, map[string]string{})
	} else {
		cs = append(cs, conditions.New(ACLConfigMapExists, true, conditions.Found))
		rApi.SetLocal(req, KeyAclAccountsMap, aclCfg.Data)
	}

	// STEP: 5. patch aggregated conditions
	nConditionsC, hasUpdatedC, err := conditions.Patch(obj.Status.ChildConditions, childC)
	if err != nil {
		return req.FailWithStatusError(err)
	}

	nConditions, hasSUpdated, err := conditions.Patch(obj.Status.Conditions, cs)
	if err != nil {
		return req.FailWithStatusError(err)
	}

	if !hasUpdatedC && !hasSUpdated && isReady == obj.Status.IsReady {
		return req.Next()
	}

	obj.Status.IsReady = isReady
	obj.Status.Conditions = nConditions
	obj.Status.ChildConditions = nConditionsC
	if err := r.Status().Update(ctx, obj); err != nil {
		return req.FailWithStatusError(err)
	}
	return req.Done()
}

func (r *ServiceReconciler) reconcileOperations(req *rApi.Request[*redisStandalone.Service]) stepResult.Result {
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
		if err := obj.Status.GeneratedVars.Set(RedisPasswordKey, fn.CleanerNanoid(40)); err != nil {
			return req.FailWithStatusError(err)
		}
		if err := r.Status().Update(ctx, obj); err != nil {
			return req.FailWithOpError(err)
		}
		return req.Done()
	}

	// STEP: 3. apply CRs of helm/custom controller

	aclAccountsMap, ok := rApi.GetLocal[map[string]string](req, KeyAclAccountsMap)
	if !ok {
		return req.FailWithOpError(rApi.ErrNotInReqLocals.Format(KeyAclAccountsMap))
	}

	storageClass, err := obj.Spec.CloudProvider.GetStorageClass(ct.Ext4)
	if err != nil {
		return req.FailWithOpError(err)
	}
	b1, err := templates.Parse(
		templates.RedisStandalone, map[string]any{
			"object":           obj,
			"storage-class":    storageClass,
			"freeze":           obj.GetLabels()[constants.LabelKeys.Freeze],
			"acl-accounts-map": aclAccountsMap,
			"owner-refs": []metav1.OwnerReference{
				fn.AsOwner(obj, true),
			},
		},
	)

	if err != nil {
		req.Logger.Errorf(err, "failed processing template %s", templates.RedisStandalone)
		return req.FailWithOpError(err).Err(nil)
	}

	// STEP: 4. create output

	redisPasswd, ok := obj.Status.GeneratedVars.GetString(RedisPasswordKey)
	if !ok {
		return req.FailWithOpError(errors.NewEf(err, "key=%s not in GeneratedVars", RedisPasswordKey))
	}

	hostUrl := fmt.Sprintf("%s-headless.%s.svc.cluster.local:6379", obj.Name, obj.Namespace)
	b2, err := templates.Parse(
		templates.Secret, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("msvc-%s", obj.Name),
				Namespace: obj.Namespace,
				OwnerReferences: []metav1.OwnerReference{
					fn.AsOwner(obj, true),
				},
			},
			StringData: map[string]string{
				"ROOT_PASSWORD": redisPasswd,
				"HOSTS":         hostUrl,
				"URI":           fmt.Sprintf("redis://:%s@%s?allowUsernameInURI=true", redisPasswd, hostUrl),
			},
		},
	)
	if err != nil {
		req.Logger.Errorf(err, "failed parsing tempalte %s", templates.Secret)
		return req.FailWithOpError(err).Err(nil)
	}

	if err := fn.KubectlApplyExec(ctx, b1, b2); err != nil {
		req.Logger.Errorf(err, "failed kubectl apply for template %s", templates.Secret)
		return req.FailWithOpError(err).Err(nil)
	}

	// create acl configmap
	if meta.IsStatusConditionFalse(obj.Status.Conditions, ACLConfigMapExists.String()) {
		if err := r.Create(
			ctx, &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("msvc-%s-acl-accounts", obj.Name),
					Namespace: obj.Namespace,
					OwnerReferences: []metav1.OwnerReference{
						fn.AsOwner(obj, true),
					},
				},
			},
		); err != nil {
			return req.FailWithOpError(err)
		}
	}

	return req.Next()
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager, envVars *env.Env, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)

	builder := ctrl.NewControllerManagedBy(mgr).For(&redisStandalone.Service{})

	builder.Owns(fn.NewUnstructured(constants.HelmRedisType))
	builder.Owns(&corev1.Secret{})
	builder.Owns(&corev1.ConfigMap{})
	builder.Owns(&redisStandalone.ACLAccount{})
	builder.Owns(&appsv1.StatefulSet{})

	refWatchList := []client.Object{
		&appsv1.StatefulSet{},
		&corev1.Pod{},
	}

	for _, item := range refWatchList {
		builder.Watches(
			&source.Kind{Type: item}, handler.EnqueueRequestsFromMapFunc(
				func(obj client.Object) []reconcile.Request {

					var services redisStandalone.ServiceList
					if err := r.List(
						context.TODO(), &services, &client.ListOptions{
							LabelSelector: labels.SelectorFromValidatedSet(
								map[string]string{
									"kloudlite.io/msvc.name": obj.GetLabels()["kloudlite.io/msvc.name"],
								},
							),
						},
					); err != nil {
						return nil
					}

					requests := make([]reconcile.Request, 0, len(services.Items))
					for _, service := range services.Items {
						requests = append(requests, reconcile.Request{NamespacedName: fn.NN(service.GetNamespace(), service.GetName())})
					}

					return requests
				},
			),
		)
	}

	return builder.Complete(r)
}
