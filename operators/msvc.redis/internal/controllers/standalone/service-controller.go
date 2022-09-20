package standalone

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ct "operators.kloudlite.io/apis/common-types"
	redisMsvcv1 "operators.kloudlite.io/apis/redis.msvc/v1"
	"operators.kloudlite.io/env"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/constants"
	fn "operators.kloudlite.io/lib/functions"
	"operators.kloudlite.io/lib/harbor"
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

type ServiceReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	env       *env.Env
	harborCli *harbor.Client
	logger    logging.Logger
	Name      string
}

func (r *ServiceReconciler) GetName() string {
	return r.Name
}

const (
	HelmReady         string = "helm-ready"
	StsReady          string = "sts-ready"
	AccessCredsReady  string = "access-creds-ready"
	ACLConfigMapReady string = "acl-configmap-ready"
	OutputReady       string = "output-ready"
)

const (
	KeyRootPassword      string = "root-password"
	KeyAclConfigMapName  string = "acl-configmap-name"
	KeyAvailableReplicas string = "available-replicas"
)

// +kubebuilder:rbac:groups=redis.msvc.kloudlite.io,resources=standaloneservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=redis.msvc.kloudlite.io,resources=standaloneservices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=redis.msvc.kloudlite.io,resources=standaloneservices/finalizers,verbs=update

func (r *ServiceReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(
		context.WithValue(ctx, "logger", r.logger),
		r.Client,
		request.NamespacedName,
		&redisMsvcv1.StandaloneService{},
	)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	req.Logger.Infof("NEW RECONCILATION")

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.RestartIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	// TODO: initialize all checks here
	if step := req.EnsureChecks(HelmReady, StsReady, AccessCredsReady, ACLConfigMapReady); !step.ShouldProceed() {
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

	if step := r.reconOutput(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	req.Object.Status.LastReconcileTime = metav1.Time{Time: time.Now()}
	req.Logger.Infof("RECONCILATION COMPLETE")
	return ctrl.Result{RequeueAfter: r.env.ReconcilePeriod * time.Second}, r.Status().Update(ctx, req.Object)
}

func (r *ServiceReconciler) finalize(req *rApi.Request[*redisMsvcv1.StandaloneService]) stepResult.Result {
	return req.Finalize()
}

func (r *ServiceReconciler) reconACLConfigmap(req *rApi.Request[*redisMsvcv1.StandaloneService]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks

	check := rApi.Check{Generation: obj.Generation}

	aclConfigmapName := "msvc-" + obj.Name + "-acl"

	aclCfgMap, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, aclConfigmapName), &redisMsvcv1.ACLConfigMap{})
	if err != nil {
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
				Spec: redisMsvcv1.ACLConfigMapSpec{},
			},
		); err != nil {
			return req.CheckFailed(ACLConfigMapReady, check, err.Error())
		}

		checks[ACLConfigMapReady] = check
		return req.UpdateStatus()
	}

	if !aclCfgMap.Status.IsReady {
		b, err := json.Marshal(aclCfgMap.Status.Message)
		if err != nil {
			return req.CheckFailed(ACLConfigMapReady, check, err.Error())
		}
		return req.CheckFailed(ACLConfigMapReady, check, string(b))
	}

	check.Status = true
	if check != checks[ACLConfigMapReady] {
		checks[ACLConfigMapReady] = check
		return req.UpdateStatus()
	}

	rApi.SetLocal(req, KeyAclConfigMapName, aclCfgMap.Name)

	return req.Next()
}

func (r *ServiceReconciler) reconAccessCreds(req *rApi.Request[*redisMsvcv1.StandaloneService]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks

	check := rApi.Check{Generation: obj.Generation}
	secretName := "msvc-" + obj.Name
	scrt, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, secretName), &corev1.Secret{})
	if err != nil {
		req.Logger.Infof("secret %s does not exist yet, would be creating it ...", fn.NN(obj.Namespace, secretName).String())
	}

	if scrt == nil {
		rootPassword := fn.CleanerNanoid(40)
		b, err := templates.Parse(
			templates.Secret, map[string]any{
				"name":       secretName,
				"namespace":  obj.Namespace,
				"labels":     obj.GetLabels(),
				"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj)},
				"string-data": map[string]string{
					"ROOT_PASSWORD": rootPassword,
					"HOSTS":         "",
					"URI":           "",
				},
			},
		)

		if err != nil {
			return req.CheckFailed(AccessCredsReady, check, err.Error())
		}

		if err := fn.KubectlApplyExec(ctx, b); err != nil {
			return req.CheckFailed(AccessCredsReady, check, err.Error())
		}

		checks[AccessCredsReady] = check
		return req.UpdateStatus()
	}

	if !fn.IsOwner(obj, fn.AsOwner(scrt)) {
		obj.SetOwnerReferences(append(obj.GetOwnerReferences(), fn.AsOwner(scrt)))
		if err := r.Update(ctx, obj); err != nil {
			return req.FailWithOpError(err)
		}
		return req.Done().RequeueAfter(2 * time.Second)
	}

	check.Status = true
	if check != checks[AccessCredsReady] {
		checks[AccessCredsReady] = check
		return req.UpdateStatus()
	}

	rApi.SetLocal(req, KeyRootPassword, string(scrt.Data["ROOT_PASSWORD"]))
	return req.Next()
}

func (r *ServiceReconciler) reconHelm(req *rApi.Request[*redisMsvcv1.StandaloneService]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	helmRes, err := rApi.Get(
		ctx, r.Client, fn.NN(obj.Namespace, obj.Name), fn.NewUnstructured(constants.HelmRedisType),
	)
	if err != nil {
		req.Logger.Infof("helm resource (%s) not found, will be creating it", fn.NN(obj.Namespace, obj.Name).String())
	}

	rootPassword, ok1 := rApi.GetLocal[string](req, KeyRootPassword)
	aclConfigmapName, ok2 := rApi.GetLocal[string](req, KeyAclConfigMapName)
	if !ok1 || !ok2 {
		return req.CheckFailed(HelmReady, check, fmt.Sprintf("key %s/%s not found in req locals", KeyRootPassword, KeyAclConfigMapName))
	}

	if helmRes == nil || check.Generation > checks[HelmReady].Generation {
		storageClass, err := obj.Spec.CloudProvider.GetStorageClass(ct.Xfs)
		if err != nil {
			return req.CheckFailed(HelmReady, check, err.Error())
		}

		b, err := templates.Parse(
			templates.RedisStandalone, map[string]any{
				"object":             obj,
				"freeze":             obj.GetLabels()[constants.LabelKeys.Freeze] == "true",
				"storage-class":      storageClass,
				"owner-refs":         []metav1.OwnerReference{fn.AsOwner(obj, true)},
				"acl-configmap-name": aclConfigmapName,
				"root-password":      rootPassword,
			},
		)

		if err != nil {
			return req.CheckFailed(HelmReady, check, err.Error()).Err(nil)
		}

		if err := fn.KubectlApplyExec(ctx, b); err != nil {
			return req.CheckFailed(HelmReady, check, err.Error()).Err(nil)
		}

		checks[HelmReady] = check
		return req.UpdateStatus()
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

	if check != checks[HelmReady] {
		checks[HelmReady] = check
		return req.UpdateStatus()
	}

	return req.Next()
}

func getStsName(objName string) string {
	return objName + "-master"
}

func (r *ServiceReconciler) reconSts(req *rApi.Request[*redisMsvcv1.StandaloneService]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	sts, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, getStsName(obj.Name)), &appsv1.StatefulSet{})
	if err != nil {
		return req.CheckFailed(StsReady, check, err.Error())
	}

	rApi.SetLocal(req, KeyAvailableReplicas, int(sts.Status.AvailableReplicas))

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
	}

	check.Status = true
	if check != checks[StsReady] {
		checks[StsReady] = check
		return req.UpdateStatus()
	}

	return req.Next()
}

func (r *ServiceReconciler) reconOutput(req *rApi.Request[*redisMsvcv1.StandaloneService]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	rootPassword, ok := rApi.GetLocal[string](req, KeyRootPassword)
	if !ok {
		return req.CheckFailed(OutputReady, check, "no root password found")
	}

	avReplicas, ok := rApi.GetLocal[int](req, KeyAvailableReplicas)
	if !ok {
		return req.CheckFailed(OutputReady, check, fmt.Sprintf("key %s not available in req locals", KeyAvailableReplicas))
	}
	hosts := make([]string, avReplicas)
	for i := 0; i < avReplicas; i += 1 {
		hosts[i] = fmt.Sprintf("%s-%d.%s.%s.svc.cluster.local", getStsName(obj.Name), i, getStsName(obj.Name), obj.Namespace)
	}

	accessCreds, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, "msvc-"+obj.Name), &corev1.Secret{})
	if err != nil {
		return req.CheckFailed(OutputReady, check, err.Error())
	}

	if _, err := controllerutil.CreateOrUpdate(
		ctx, r.Client, accessCreds, func() error {
			accessCreds.Data["HOSTS"] = []byte(strings.Join(hosts, ","))
			accessCreds.Data["URI"] = []byte(fmt.Sprintf("redis://:%s@%s?allowUsernameInURI=true", rootPassword, strings.Join(hosts, ",")))
			return nil
		},
	); err != nil {
		return req.CheckFailed(OutputReady, check, err.Error())
	}

	check.Status = true
	if check != checks[OutputReady] {
		checks[OutputReady] = check
		return req.UpdateStatus()
	}

	return req.Next()
}

func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager, envVars *env.Env, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.env = envVars

	builder := ctrl.NewControllerManagedBy(mgr).For(&redisMsvcv1.StandaloneService{})
	builder.Owns(&corev1.Secret{})
	builder.Owns(&redisMsvcv1.ACLConfigMap{})

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

	return builder.Complete(r)
}
