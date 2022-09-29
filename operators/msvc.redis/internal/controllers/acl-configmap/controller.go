package acl_configmap

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	redisMsvcv1 "operators.kloudlite.io/apis/redis.msvc/v1"
	"operators.kloudlite.io/env"
	"operators.kloudlite.io/lib/constants"
	fn "operators.kloudlite.io/lib/functions"
	"operators.kloudlite.io/lib/harbor"
	"operators.kloudlite.io/lib/logging"
	rApi "operators.kloudlite.io/lib/operator"
	stepResult "operators.kloudlite.io/lib/operator/step-result"
	"operators.kloudlite.io/lib/templates"
	"operators.kloudlite.io/operators/msvc.redis/internal/controllers/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type Reconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	env       *env.Env
	harborCli *harbor.Client
	logger    logging.Logger
	Name      string
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	ACLConfigMapExists string = "acl-configmap-exists"
	ACLConfigMapReady  string = "acl-configmap-ready"
)

// +kubebuilder:rbac:groups=redis.msvc.kloudlite.io,resources=aclaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=redis.msvc.kloudlite.io,resources=aclaccounts/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=redis.msvc.kloudlite.io,resources=aclaccounts/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(
		context.WithValue(ctx, "logger", r.logger), r.Client, request.NamespacedName,
		&redisMsvcv1.ACLConfigMap{},
	)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if step := req.EnsureChecks(ACLConfigMapExists, ACLConfigMapReady); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconRedisConfigmap(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.buildRedisConf(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Logger.Infof("RECONCILATION COMPLETE")
	req.Object.Status.IsReady = true
	return req.UpdateStatus().ReconcilerResponse()
}

func (r *Reconciler) reconRedisConfigmap(req *rApi.Request[*redisMsvcv1.ACLConfigMap]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks

	check := rApi.Check{Generation: obj.Generation}

	aclCfgMap, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Name), &corev1.ConfigMap{})
	if err != nil {
		req.Logger.Infof("acl config map (%s) does not exist, will be creating it...", fn.NN(obj.Namespace, obj.Name).String())
	}

	if aclCfgMap == nil {
		b, err := templates.Parse(
			templates.RedisACLConfigMap, map[string]any{
				"name":       obj.Name,
				"namespace":  obj.Namespace,
				"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},
			},
		)
		if err != nil {
			return req.CheckFailed(ACLConfigMapExists, check, err.Error())
		}

		if err := fn.KubectlApplyExec(ctx, b); err != nil {
			return req.CheckFailed(ACLConfigMapExists, check, err.Error())
		}

		checks[ACLConfigMapExists] = check
		return req.UpdateStatus()
	}

	if !fn.IsOwner(obj, fn.AsOwner(aclCfgMap)) {
		obj.SetOwnerReferences(append(obj.GetOwnerReferences(), fn.AsOwner(aclCfgMap)))
		if err := r.Update(ctx, obj); err != nil {
			return req.FailWithOpError(err)
		}
	}

	check.Status = true
	if check != checks[ACLConfigMapExists] {
		checks[ACLConfigMapExists] = check
		return req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) buildRedisConf(req *rApi.Request[*redisMsvcv1.ACLConfigMap]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks

	check := rApi.Check{Generation: obj.Generation}

	var scrtList corev1.SecretList
	if err := r.List(
		ctx, &scrtList, &client.ListOptions{
			LabelSelector: labels.SelectorFromValidatedSet(
				map[string]string{
					constants.MsvcNameKey:  obj.Spec.MsvcName,
					constants.IsMresOutput: "true",
				},
			),
			Namespace: obj.Namespace,
		},
	); err != nil {
		return req.CheckFailed(ACLConfigMapReady, check, err.Error())
	}

	aclSecrets := make([]types.MresOutput, len(scrtList.Items))

	for i := range scrtList.Items {
		aclSecrets[i] = types.MresOutput{
			Password: string(scrtList.Items[i].Data["PASSWORD"]),
			Username: string(scrtList.Items[i].Data["PREFIX"]),
			Prefix:   string(scrtList.Items[i].Data["USERNAME"]),
		}
	}

	b, err := templates.Parse(
		templates.RedisACLConfigMap, map[string]any{
			"name":        obj.Name,
			"namespace":   obj.Namespace,
			"owner-refs":  []metav1.OwnerReference{fn.AsOwner(obj, true)},
			"acl-secrets": aclSecrets,
		},
	)

	if err != nil {
		return req.CheckFailed(ACLConfigMapReady, check, err.Error()).Err(nil)
	}

	if err := fn.KubectlApplyExec(ctx, b); err != nil {
		return req.CheckFailed(ACLConfigMapReady, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != checks[ACLConfigMapReady] {
		checks[ACLConfigMapReady] = check
		return req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, envVars *env.Env, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.env = envVars

	builder := ctrl.NewControllerManagedBy(mgr).For(&redisMsvcv1.ACLConfigMap{}).Owns(&corev1.ConfigMap{})
	builder.Watches(
		&source.Kind{Type: &corev1.Secret{}}, handler.EnqueueRequestsFromMapFunc(
			func(obj client.Object) []reconcile.Request {
				msvcName, ok := obj.GetLabels()[constants.MsvcNameKey]
				if !ok {
					return nil
				}
				if _, ok := obj.GetLabels()[constants.IsMresOutput]; !ok {
					return nil
				}
				return []reconcile.Request{{NamespacedName: fn.NN(obj.GetNamespace(), "msvc-"+msvcName+"-acl")}}
			},
		),
	)

	return builder.Complete(r)
}
