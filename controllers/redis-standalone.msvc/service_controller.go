package redisstandalonemsvc

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"encoding/json"
	"fmt"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	apiLabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	redisStandalone "operators.kloudlite.io/apis/redis-standalone.msvc/v1"
	"operators.kloudlite.io/controllers/crds"
	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/errors"
	"operators.kloudlite.io/lib/finalizers"
	fn "operators.kloudlite.io/lib/functions"
	reconcileResult "operators.kloudlite.io/lib/reconcile-result"
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
	logger    *zap.SugaredLogger
	redisSvc  *redisStandalone.Service
	stateData map[string]any
}

func (s *ServiceReconReq) SetStateData(k string, v any) {
	if s.stateData == nil {
		s.stateData = map[string]any{}
	}
	s.stateData[k] = v
}

func (s *ServiceReconReq) GetStateData(k string) any {
	if s.stateData == nil {
		s.stateData = map[string]any{}
	}
	return s.stateData[k]
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

func (r *ServiceReconciler) Reconcile(ctx context.Context, orgReq ctrl.Request) (ctrl.Result, error) {
	req := &ServiceReconReq{
		logger:   crds.GetLogger(orgReq.NamespacedName),
		redisSvc: new(redisStandalone.Service),
	}

	if err := r.Get(ctx, orgReq.NamespacedName, req.redisSvc); err != nil {
		if apiErrors.IsNotFound(err) {
			return reconcileResult.OK()
		}
		return reconcileResult.Failed()
	}

	if !req.redisSvc.HasLabels() {
		req.redisSvc.EnsureLabels()
		if err := r.Update(ctx, req.redisSvc); err != nil {
			return reconcileResult.FailedE(err)
		}
		return reconcileResult.OK()
	}

	if req.redisSvc.GetDeletionTimestamp() != nil {
		return r.finalize(ctx, req)
	}

	reconResult, err := r.reconcileStatus(ctx, req)

	if err != nil {
		return r.failWithErr(ctx, req, err)
	}
	if reconResult != nil {
		return *reconResult, nil
	}
	return r.reconcileOperations(ctx, req)
}

func (r *ServiceReconciler) finalize(ctx context.Context, req *ServiceReconReq) (ctrl.Result, error) {
	req.logger.Infof("finalizing: %+v", req.redisSvc.NameRef())
	if err := r.Delete(
		ctx, &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": constants.MsvcApiVersion,
				"kind":       constants.HelmRedisKind,
				"metadata": map[string]interface{}{
					"name":      req.redisSvc.Name,
					"namespace": req.redisSvc.Namespace,
				},
			},
		},
	); err != nil {
		req.logger.Infof("could not delete helm resource: %+v", err)
		if !apiErrors.IsNotFound(err) {
			return reconcileResult.FailedE(err)
		}
	}
	controllerutil.RemoveFinalizer(req.redisSvc, finalizers.MsvcCommonService.String())
	if err := r.Update(ctx, req.redisSvc); err != nil {
		return reconcileResult.FailedE(err)
	}
	return reconcileResult.OK()
}

func (r *ServiceReconciler) failWithErr(ctx context.Context, req *ServiceReconReq, err error) (ctrl.Result, error) {
	fn.Conditions2.MarkNotReady(&req.redisSvc.Status.Conditions, err, "ReconcileFailedWithErr")
	if err2 := r.Status().Update(ctx, req.redisSvc); err2 != nil {
		return ctrl.Result{}, err2
	}
	return reconcileResult.FailedE(err)
}

func (r *ServiceReconciler) reconcileStatus(ctx context.Context, req *ServiceReconReq) (*ctrl.Result, error) {
	var conditions []metav1.Condition

	if err := fn.Conditions2.BuildFromHelmMsvc(
		&conditions, ctx, r.Client, constants.HelmRedisKind, req.redisSvc.NamespacedName(),
	); err != nil {
		if !apiErrors.IsNotFound(err) {
			return nil, err
		}
	}

	if err := fn.Conditions2.BuildFromStatefulset(
		&conditions,
		ctx,
		r.Client,
		types.NamespacedName{Namespace: req.redisSvc.Namespace, Name: fmt.Sprintf("%s-master", req.redisSvc.Name)},
	); err != nil {
		if !apiErrors.IsNotFound(err) {
			return nil, err
		}
	}

	// STEP: helm secret
	helmSecret := new(corev1.Secret)
	if err := r.Get(ctx, req.redisSvc.NamespacedName(), helmSecret); err != nil {
		if !apiErrors.IsNotFound(err) {
			return nil, err
		}
		fn.Conditions2.Build(
			&conditions, "Helm", metav1.Condition{
				Type:    "OutputExists",
				Status:  metav1.ConditionFalse,
				Reason:  "SecretNotFound",
				Message: err.Error(),
			},
		)
		helmSecret = nil
	}

	fn.IfThenElseFn(helmSecret == nil, )

	if helmSecret != nil {
		passwd, ok := helmSecret.Data[RedisPasswordKey]
		req.SetStateData(RedisPasswordKey, fn.IfThenElse(ok, string(passwd), fn.CleanerNanoid(40)))
	}

	// STEP: ACL list
	aclCfg := new(corev1.ConfigMap)
	cfgName := fmt.Sprintf("msvc-%s-acl-accounts", req.redisSvc.Name)
	nn := types.NamespacedName{Namespace: req.redisSvc.Namespace, Name: cfgName}
	if err := r.Get(ctx, nn, aclCfg); err != nil {
		if !apiErrors.IsNotFound(err) {
			return nil, err
		}
		fn.Conditions2.Build(
			&conditions, "RedisACL", metav1.Condition{
				Type:    "ConfigmapExists",
				Status:  metav1.ConditionFalse,
				Reason:  "ConfigmapNotFound",
				Message: err.Error(),
			},
		)
		aclCfg = nil
	}

	if aclCfg != nil {
		req.SetStateData(AclAccountsKey, aclCfg.Data)
	}

	if fn.Conditions2.Equal(conditions, req.redisSvc.Status.Conditions) {
		req.logger.Infof("resource status is in sync ...")
		return nil, nil
	}

	req.logger.Infof("status is different, so updating status ...")

	req.logger.Debugf("conditions: %+v", conditions)
	req.logger.Debugf("req.redisSvc.Status.Conditions: %+v", req.redisSvc.Status.Conditions)

	req.redisSvc.Status.Conditions = conditions
	if err := r.Status().Update(ctx, req.redisSvc); err != nil {
		return nil, err
	}
	return reconcileResult.OKP()
}

func (r *ServiceReconciler) preOps(_ context.Context, req *ServiceReconReq) error {
	m, err := fn.Json.FromRawMessage(req.redisSvc.Spec.Inputs)
	if err != nil {
		return err
	}
	m[RedisPasswordKey] = req.GetStateData(RedisPasswordKey)
	aclMap, ok := req.GetStateData(AclAccountsKey).(map[string]string)
	if !ok {
		return errors.Newf("Bad state-data(key=%s)", AclAccountsKey)
	}
	marshal, err := json.Marshal(m)
	if err != nil {
		return err
	}
	req.redisSvc.Spec.Inputs = marshal
	if
	req.redisSvc.Spec.ACLAccounts = aclMap
	return nil
}

func (r *ServiceReconciler) reconcileOperations(ctx context.Context, req *ServiceReconReq) (ctrl.Result, error) {
	// hash := req.redisSvc.Hash()
	// if hash == req.redisSvc.Status.LastHash {
	// 	return reconcileResult.OK()
	// }

	if err := r.preOps(ctx, req); err != nil {
		return ctrl.Result{}, err
	}

	b, err := templates.Parse(templates.RedisStandalone, req.redisSvc)
	if err != nil {
		return reconcileResult.FailedE(err)
	}

	if _, err := fn.KubectlApply(b); err != nil {
		return reconcileResult.FailedE(errors.NewEf(err, "could not apply kubectl for redis standalone"))
	}

	if err := r.reconcileOutput(ctx, req); err != nil {
		return reconcileResult.FailedE(err)
	}
	//req.redisSvc.Status.LastHash = req.redisSvc.Hash()
	return reconcileResult.OK()
}

func (r *ServiceReconciler) reconcileOutput(ctx context.Context, req *ServiceReconReq) error {
	m, err := fn.Json.FromRawMessage(req.redisSvc.Spec.Inputs)
	if err != nil {
		return err
	}

	hostUrl := fmt.Sprintf("%s-headless.%s.svc.cluster.local:6379", req.redisSvc.Name, req.redisSvc.Namespace)
	out := Output{
		RootPassword: m[RedisPasswordKey].(string),
		Hosts:        hostUrl,
		Uri:          fmt.Sprintf("redis://:%s@%s?allowUsernameInURI=true", m[RedisPasswordKey], hostUrl),
	}

	scrt := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("msvc-%s", req.redisSvc.Name),
			Namespace: req.redisSvc.Namespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(
		ctx, r.Client, scrt, func() error {
			var outMap map[string]string
			if err := fn.Json.FromTo(out, &outMap); err != nil {
				return err
			}
			scrt.StringData = outMap
			return controllerutil.SetControllerReference(req.redisSvc, scrt, r.Scheme)
		},
	); err != nil {
		return err
	}

	if len(req.redisSvc.Spec.ACLAccounts) == 0 {
		cfgName := fmt.Sprintf("msvc-%s-acl-accounts", req.redisSvc.Name)
		nCfg := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cfgName,
				Namespace: req.redisSvc.Namespace,
				OwnerReferences: []metav1.OwnerReference{
					fn.AsOwner(req.redisSvc, true),
				},
			},
		}
		err := r.Create(ctx, nCfg)
		req.logger.Error("#######################", err)
		return err
	}

	return nil
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
	return ctrl.NewControllerManagedBy(mgr).
		For(&redisStandalone.Service{}).
		Watches(
			&source.Kind{
				Type: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": constants.MsvcApiVersion,
						"kind":       constants.HelmRedisKind,
					},
				},
			}, handler.EnqueueRequestsFromMapFunc(
				func(c client.Object) []reconcile.Request {
					var svcList redisStandalone.ServiceList
					key, value := redisStandalone.Service{}.LabelRef()
					if err := r.List(
						context.TODO(), &svcList, &client.ListOptions{
							LabelSelector: apiLabels.SelectorFromValidatedSet(map[string]string{key: value}),
						},
					); err != nil {
						return nil
					}
					var reqs []reconcile.Request
					for _, item := range svcList.Items {
						nn := types.NamespacedName{
							Name:      item.Name,
							Namespace: item.Namespace,
						}

						for _, req := range reqs {
							if req.NamespacedName.String() == nn.String() {
								return nil
							}
						}

						reqs = append(reqs, reconcile.Request{NamespacedName: nn})
					}
					return reqs
				},
			),
		).
		Watches(&source.Kind{Type: &appsv1.Deployment{}}, handler.EnqueueRequestsFromMapFunc(r.kWatcherMap)).
		Watches(&source.Kind{Type: &corev1.Pod{}}, handler.EnqueueRequestsFromMapFunc(r.kWatcherMap)).
		Watches(
			&source.Kind{Type: &corev1.ConfigMap{}}, handler.EnqueueRequestsFromMapFunc(
				func(o client.Object) []reconcile.Request {
					name, ok := o.GetLabels()["msvc.kloudlite.io/ref"]
					if !ok {
						return nil
					}
					return []reconcile.Request{
						{NamespacedName: types.NamespacedName{Name: name, Namespace: o.GetNamespace()}},
					}
				},
			),
		).
		Complete(r)
}
