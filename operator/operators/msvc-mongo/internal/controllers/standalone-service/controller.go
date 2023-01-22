package standalone_service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	ct "github.com/kloudlite/operator/apis/common-types"
	mongodbMsvcv1 "github.com/kloudlite/operator/apis/mongodb.msvc/v1"
	"github.com/kloudlite/operator/operators/msvc-mongo/internal/env"
	"github.com/kloudlite/operator/operators/msvc-mongo/internal/types"
	"github.com/kloudlite/operator/pkg/conditions"
	"github.com/kloudlite/operator/pkg/constants"
	"github.com/kloudlite/operator/pkg/errors"
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
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type ServiceReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	logger     logging.Logger
	Name       string
	Env        *env.Env
	yamlClient *kubectl.YAMLClient
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
	KeyMsvcOutput string = "msvc-output"
)

func getHelmSecretName(name string) string {
	return "helm-" + name
}

// +kubebuilder:rbac:groups=mongodb.msvc.kloudlite.io,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mongodb.msvc.kloudlite.io,resources=services/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mongodb.msvc.kloudlite.io,resources=services/finalizers,verbs=update

func (r *ServiceReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	if strings.HasSuffix(request.Namespace, "-blueprint") {
		return ctrl.Result{}, nil
	}

	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &mongodbMsvcv1.StandaloneService{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	req.LogPostReconcile()
	defer req.LogPostReconcile()

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	// TODO: initialize all checks here
	if step := req.EnsureChecks(HelmReady, StsReady, AccessCredsReady); !step.ShouldProceed() {
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

	if step := r.reconHelmSecret(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconHelm(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconSts(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
}

func (r *ServiceReconciler) finalize(req *rApi.Request[*mongodbMsvcv1.StandaloneService]) stepResult.Result {
	return req.Finalize()
}

func (r *ServiceReconciler) reconAccessCreds(req *rApi.Request[*mongodbMsvcv1.StandaloneService]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(AccessCredsReady)
	defer req.LogPostCheck(AccessCredsReady)

	secretName := "msvc-" + obj.Name
	scrt, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, secretName), &corev1.Secret{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.CheckFailed(AccessCredsReady, check, err.Error()).Err(nil)
		}
		req.Logger.Infof("secret %s does not exist yet, would be creating it ...", fn.NN(obj.Namespace, secretName).String())
	}

	if scrt == nil {
		var hosts []string
		for i := 0; i < obj.Spec.ReplicaCount; i++ {
			hosts = append(hosts, fmt.Sprintf("%s-%d.%s.%s.svc.cluster.local:27017", obj.Name, i, obj.Name, obj.Namespace))
		}

		rootPassword := fn.CleanerNanoid(40)
		b, err := templates.Parse(
			templates.Secret, map[string]any{
				"name":       secretName,
				"namespace":  obj.Namespace,
				"labels":     obj.GetLabels(),
				"owner-refs": obj.GetOwnerReferences(),
				"string-data": types.MsvcOutput{
					RootPassword: rootPassword,
					Hosts:        strings.Join(hosts, ","),
					URI:          fmt.Sprintf("mongodb://%s:%s@%s/admin?authSource=admin", "root", rootPassword, strings.Join(hosts, ",")),
				},
			},
		)

		if err != nil {
			return req.CheckFailed(AccessCredsReady, check, err.Error())
		}

		if err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
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

	msvcOutput, err := fn.ParseFromSecret[types.MsvcOutput](scrt)
	if err != nil {
		return req.CheckFailed(AccessCredsReady, check, err.Error())
	}

	rApi.SetLocal(req, KeyMsvcOutput, *msvcOutput)
	return req.Next()
}

func (r *ServiceReconciler) reconHelmSecret(req *rApi.Request[*mongodbMsvcv1.StandaloneService]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(HelmSecretReady)
	defer req.LogPostCheck(HelmSecretReady)

	helmSecret, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, getHelmSecretName(obj.Name)), &corev1.Secret{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.CheckFailed(HelmSecretReady, check, err.Error())
		}
		req.Logger.Infof("helm secret (%s) does not exist, will be creating now...", getHelmSecretName(obj.Name))
		helmSecret = nil
	}

	msvcOutput, ok := rApi.GetLocal[types.MsvcOutput](req, KeyMsvcOutput)
	if !ok {
		return req.CheckFailed(HelmSecretReady, check, errors.NotInLocals(KeyMsvcOutput).Error()).Err(nil)
	}

	if helmSecret == nil {
		if err := r.Create(
			ctx, &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:            getHelmSecretName(obj.Name),
					Namespace:       obj.Namespace,
					OwnerReferences: obj.GetOwnerReferences(),
				},
				StringData: map[string]string{
					"mongodb-passwords":        "",
					"mongodb-root-password":    msvcOutput.RootPassword,
					"mongodb-metrics-password": "",
					"mongodb-replica-set-key":  "",
				},
			},
		); err != nil {
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

func (r *ServiceReconciler) reconHelm(req *rApi.Request[*mongodbMsvcv1.StandaloneService]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(HelmReady)
	defer req.LogPostCheck(HelmReady)

	helmRes, err := rApi.Get(
		ctx, r.Client, fn.NN(obj.Namespace, obj.Name), fn.NewUnstructured(constants.HelmMongoDBType),
	)

	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.CheckFailed(HelmReady, check, err.Error()).Err(nil)
		}
		helmRes = nil
		req.Logger.Infof("helm resource (%s) not found, will be creating it", fn.NN(obj.Namespace, obj.Name).String())
	}

	if helmRes == nil || check.Generation > checks[HelmReady].Generation {
		b, err := templates.Parse(
			templates.MongoDBStandalone, map[string]any{
				"object": obj,
				"freeze": obj.GetLabels()[constants.LabelKeys.Freeze] == "true",
				"storage-class": func() string {
					if obj.Spec.Resources.Storage.StorageClass != "" {
						return obj.Spec.Resources.Storage.StorageClass
					}
					return fmt.Sprintf("%s-%s", obj.Spec.Region, ct.Xfs)
				}(),
				"owner-refs":      obj.GetOwnerReferences(),
				"existing-secret": getHelmSecretName(obj.Name),
			},
		)
		if err != nil {
			return req.CheckFailed(HelmReady, check, err.Error()).Err(nil)
		}

		if err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
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

func (r *ServiceReconciler) reconSts(req *rApi.Request[*mongodbMsvcv1.StandaloneService]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(StsReady)
	defer req.LogPostCheck(StsReady)

	var stsList appsv1.StatefulSetList
	if err := r.List(
		ctx, &stsList, &client.ListOptions{
			LabelSelector: labels.SelectorFromValidatedSet(
				map[string]string{constants.MsvcNameKey: obj.Name},
			),
			Namespace: obj.Namespace,
		},
	); err != nil {
		return req.CheckFailed(StsReady, check, err.Error()).Err(nil)
	}

	for i := range stsList.Items {
		item := stsList.Items[i]
		if item.Status.AvailableReplicas != item.Status.Replicas {
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
				return req.FailWithOpError(err)
			}

			messages := rApi.GetMessagesFromPods(podsList.Items...)
			if len(messages) > 0 {
				b, err := json.Marshal(messages)
				if err != nil {
					return req.CheckFailed(StsReady, check, err.Error()).Err(nil)
				}
				return req.CheckFailed(StsReady, check, string(b)).Err(nil)
			}

			return req.CheckFailed(StsReady, check, "waiting for pods to start ...").Err(nil)
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

	builder := ctrl.NewControllerManagedBy(mgr).For(&mongodbMsvcv1.StandaloneService{})
	builder.Owns(fn.NewUnstructured(constants.HelmMongoDBType))
	builder.Owns(&corev1.Secret{})

	watchList := []client.Object{
		&appsv1.StatefulSet{},
	}

	for i := range watchList {
		builder.Watches(
			&source.Kind{Type: watchList[i]}, handler.EnqueueRequestsFromMapFunc(
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
	}

	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
