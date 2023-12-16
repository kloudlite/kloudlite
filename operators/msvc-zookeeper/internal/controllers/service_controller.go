package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	ct "github.com/kloudlite/operator/apis/common-types"
	zookeeperMsvcv1 "github.com/kloudlite/operator/apis/zookeeper.msvc/v1"
	"github.com/kloudlite/operator/operators/msvc-zookeeper/internal/env"
	"github.com/kloudlite/operator/operators/msvc-zookeeper/internal/types"
	"github.com/kloudlite/operator/pkg/conditions"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	"github.com/kloudlite/operator/pkg/templates"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ServiceReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	logger     logging.Logger
	Name       string
	Env        *env.Env
	yamlClient kubectl.YAMLClient
}

func (r *ServiceReconciler) GetName() string {
	return r.Name
}

const (
	HelmReady        string = "helm-ready"
	StsReady         string = "sts-ready"
	AccessCredsReady string = "access-creds-ready"
	HelmSecretReady  string = "helm-secret-ready"
)

const (
	KeyOutput string = "output"
)

func getHelmSecretName(name string) string {
	return "helm-" + name
}

func getMsvcSecretName(name string) string {
	return "msvc-" + name
}

// +kubebuilder:rbac:groups=zookeeper.msvc.kloudlite.io,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=zookeeper.msvc.kloudlite.io,resources=services/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=zookeeper.msvc.kloudlite.io,resources=services/finalizers,verbs=update

func (r *ServiceReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, request.NamespacedName, &zookeeperMsvcv1.Service{})
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
	defer func() {
		req.Logger.Infof("RECONCILATION COMPLETE (isReady=%v)", req.Object.Status.IsReady)
	}()

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.RestartIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureChecks(HelmReady, StsReady, AccessCredsReady, HelmSecretReady); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconAccessCreds(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureHelmSecret(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconHelm(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconSts(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	req.Object.Status.LastReconcileTime = &metav1.Time{Time: time.Now()}
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
}

func (r *ServiceReconciler) finalize(req *rApi.Request[*zookeeperMsvcv1.Service]) stepResult.Result {
	return req.Finalize()
}

func (r *ServiceReconciler) reconAccessCreds(req *rApi.Request[*zookeeperMsvcv1.Service]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	scrt, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, getMsvcSecretName(obj.Name)), &corev1.Secret{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.CheckFailed(AccessCredsReady, check, err.Error())
		}
		req.Logger.Infof("secret %s does not exist yet, would be creating it ...", getMsvcSecretName(obj.Name))
	}

	if scrt == nil {
		b, err := templates.Parse(
			templates.Secret, map[string]any{
				"name":       getMsvcSecretName(obj.Name),
				"namespace":  obj.Namespace,
				"labels":     obj.GetLabels(),
				"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},
				"string-data": types.MsvcOutput{
					RootUsername: "",
					RootPassword: "",
					Host:         fmt.Sprintf("%s-headless.%s.svc.cluster.local", obj.Name, obj.Namespace),
					Port:         "2181",
				},
			},
		)

		if err != nil {
			return req.CheckFailed(AccessCredsReady, check, err.Error())
		}

		if _, err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
			return req.CheckFailed(AccessCredsReady, check, err.Error())
		}

		checks[AccessCredsReady] = check
		return req.UpdateStatus()
	}

	if !fn.IsOwner(obj, fn.AsOwner(scrt)) {
		obj.SetOwnerReferences(append(obj.GetOwnerReferences(), fn.AsOwner(scrt)))
		if err := r.Update(ctx, obj); err != nil {
			return req.CheckFailed(AccessCredsReady, check, err.Error())
		}
		return req.Done().RequeueAfter(2 * time.Second)
	}

	check.Status = true
	if check != checks[AccessCredsReady] {
		checks[AccessCredsReady] = check
		return req.UpdateStatus()
	}

	msvcOutput, err := fn.ParseFromSecret[types.MsvcOutput](scrt)
	if err != nil {
		return req.CheckFailed(AccessCredsReady, check, err.Error()).Err(nil)
	}

	rApi.SetLocal(req, KeyOutput, msvcOutput)
	return req.Next()
}

func (r *ServiceReconciler) ensureHelmSecret(req *rApi.Request[*zookeeperMsvcv1.Service]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	helmScrt, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, getHelmSecretName(obj.Name)), &corev1.Secret{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.CheckFailed(HelmSecretReady, check, err.Error())
		}
		req.Logger.Infof("helm secret (%s) does not exist, will be creating it", getHelmSecretName(obj.Name))
	}

	if helmScrt == nil {
		if err := r.Create(ctx, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:            getHelmSecretName(obj.Name),
				Namespace:       obj.Namespace,
				OwnerReferences: []metav1.OwnerReference{fn.AsOwner(obj, true)},
				Labels:          obj.GetLabels(),
			},
			StringData: map[string]string{
				"client-password": "",
				"server-password": "",
			},
		}); err != nil {
			return req.CheckFailed(HelmSecretReady, check, err.Error())
		}
	}

	check.Status = true
	if check != checks[HelmSecretReady] {
		checks[HelmSecretReady] = check
		return req.UpdateStatus()
	}
	return req.Next()
}

func (r *ServiceReconciler) reconHelm(req *rApi.Request[*zookeeperMsvcv1.Service]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	helmRes, err := rApi.Get(
		ctx, r.Client, fn.NN(obj.Namespace, obj.Name), fn.NewUnstructured(constants.HelmZookeeperType),
	)
	if err != nil {
		req.Logger.Infof("helm resource (%s) not found, will be creating it", fn.NN(obj.Namespace, obj.Name).String())
	}

	output, ok := rApi.GetLocal[*types.MsvcOutput](req, KeyOutput)
	if !ok {
		return req.CheckFailed(HelmReady, check, err.Error())
	}

	if helmRes == nil || check.Generation > checks[HelmReady].Generation {
		b, err := templates.Parse(
			templates.MsvcHelmZookeeper, map[string]any{
				"obj":        obj,
				"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},
				"storage-class": func() string {
					if obj.Spec.Resources.Storage.StorageClass != "" {
						return obj.Spec.Resources.Storage.StorageClass
					}
					return fmt.Sprintf("%s-%s", obj.Spec.Region, ct.Ext4)
				}(),
				"root-password":   output.RootPassword,
				"existing-secret": getHelmSecretName(obj.Name),
			},
		)
		if err != nil {
			return req.CheckFailed(HelmReady, check, err.Error()).Err(nil)
		}

		if _, err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
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

func (r *ServiceReconciler) reconSts(req *rApi.Request[*zookeeperMsvcv1.Service]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}
	var stsList appsv1.StatefulSetList

	if err := r.List(
		ctx, &stsList, &client.ListOptions{
			LabelSelector: labels.SelectorFromValidatedSet(map[string]string{constants.MsvcNameKey: obj.Name}),
			Namespace:     obj.Namespace,
		},
	); err != nil {
		return req.CheckFailed(StsReady, check, err.Error())
	}

	for i := range stsList.Items {
		item := stsList.Items[i]
		if item.Status.AvailableReplicas != item.Status.Replicas {
			req.Logger.Infof("Waiting for Statefulset to be available ...")
			check.Status = false

			var podsList corev1.PodList
			if err := r.List(
				ctx, &podsList, &client.ListOptions{
					LabelSelector: labels.SelectorFromValidatedSet(
						map[string]string{
							constants.MsvcNameKey: obj.Name,
						},
					),
				},
			); err != nil {
				return req.CheckFailed(StsReady, check, err.Error())
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
	}

	check.Status = true
	if check != checks[StsReady] {
		checks[StsReady] = check
		return req.UpdateStatus()
	}

	return req.Next()
}

func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

	builder := ctrl.NewControllerManagedBy(mgr).For(&zookeeperMsvcv1.Service{})
	builder.Owns(fn.NewUnstructured(constants.HelmZookeeperType))
	builder.Owns(&corev1.Secret{})
	builder.Watches(&appsv1.StatefulSet{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
		value, ok := obj.GetLabels()[constants.MsvcNameKey]
		if !ok {
			return nil
		}
		return []reconcile.Request{
			{NamespacedName: fn.NN(obj.GetNamespace(), value)},
		}
	}))
	return builder.Complete(r)
}
