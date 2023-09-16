package standalone_service

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"strings"

	mongodbMsvcv1 "github.com/kloudlite/operator/apis/mongodb.msvc/v1"
	"github.com/kloudlite/operator/operators/msvc-mongo/internal/env"
	"github.com/kloudlite/operator/operators/msvc-mongo/internal/types"
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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	logger     logging.Logger
	Name       string
	Env        *env.Env
	yamlClient kubectl.YAMLClient

	templateHelmMongoDB     []byte
	templateHelmMongoDBAuth []byte
}

func (r *Reconciler) GetName() string {
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

var (
	//go:embed templates
	templatesDir embed.FS
)

func getHelmSecretName(name string) string {
	return "helm-" + name
}

// +kubebuilder:rbac:groups=mongodb.msvc.kloudlite.io,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mongodb.msvc.kloudlite.io,resources=services/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mongodb.msvc.kloudlite.io,resources=services/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &mongodbMsvcv1.StandaloneService{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	req.LogPreReconcile()
	defer req.LogPostReconcile()

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
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
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, nil
}

func (r *Reconciler) finalize(req *rApi.Request[*mongodbMsvcv1.StandaloneService]) stepResult.Result {
	return req.Finalize()
}

func (r *Reconciler) reconAccessCreds(req *rApi.Request[*mongodbMsvcv1.StandaloneService]) stepResult.Result {
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
				"name":      secretName,
				"namespace": obj.Namespace,
				"labels":    obj.GetLabels(),
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

		resourceRefs, err := r.yamlClient.ApplyYAML(ctx, b)
		if err != nil {
			return req.CheckFailed(AccessCredsReady, check, err.Error())
		}

		req.AddToOwnedResources(resourceRefs...)
	}

	// if !fn.IsOwner(obj, fn.AsOwner(scrt)) {
	// 	obj.SetOwnerReferences(append(obj.GetOwnerReferences(), fn.AsOwner(scrt)))
	// 	if err := r.Update(ctx, obj); err != nil {
	// 		return req.CheckFailed(AccessCredsReady, check, err.Error())
	// 	}
	// 	return req.Done().RequeueAfter(2 * time.Second)
	// }

	check.Status = true
	if check != checks[AccessCredsReady] {
		checks[AccessCredsReady] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	msvcOutput, err := fn.ParseFromSecret[types.MsvcOutput](scrt)
	if err != nil {
		return req.CheckFailed(AccessCredsReady, check, err.Error())
	}

	rApi.SetLocal(req, KeyMsvcOutput, *msvcOutput)
	return req.Next()
}

func (r *Reconciler) reconHelmSecret(req *rApi.Request[*mongodbMsvcv1.StandaloneService]) stepResult.Result {
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
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

func (r *Reconciler) reconHelm(req *rApi.Request[*mongodbMsvcv1.StandaloneService]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(HelmReady)
	defer req.LogPostCheck(HelmReady)

	//helmRes, err := rApi.Get(
	//	ctx, r.Client, fn.NN(obj.Namespace, obj.Name), fn.NewUnstructured(constants.HelmMongoDBType),
	//)
	//
	//if err != nil {
	//	if !apiErrors.IsNotFound(err) {
	//		return req.CheckFailed(HelmReady, check, err.Error()).Err(nil)
	//	}
	//	helmRes = nil
	//	req.Logger.Infof("helm resource (%s) not found, will be creating it", fn.NN(obj.Namespace, obj.Name).String())
	//}

	// TODO (nxtcoder17): when increasing pvc volume size, we can not trigger helm update, as it complains about forbidden field
	b, err := templates.ParseBytes(r.templateHelmMongoDB, map[string]any{
		"name":      obj.Name,
		"namespace": obj.Namespace,
		"labels": map[string]string{
			constants.MsvcNameKey: obj.Name,
		},
		"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj)},

		"storage-class": obj.Spec.Resources.Storage.StorageClass,
		"storage-size":  obj.Spec.Resources.Storage.Size,

		"requests-cpu": obj.Spec.Resources.Cpu.Min,
		"requests-mem": obj.Spec.Resources.Memory,

		"limits-cpu": obj.Spec.Resources.Cpu.Min,
		"limits-mem": obj.Spec.Resources.Memory,

		"existing-secret": getHelmSecretName(obj.Name),

		"freeze": false,
	})
	if err != nil {
		return req.CheckFailed(HelmReady, check, err.Error()).Err(nil)
	}

	//b, err := templates.Parse(
	//	templates.MongoDBStandalone, map[string]any{
	//		"object": obj,
	//		"freeze": obj.GetLabels()[constants.LabelKeys.Freeze] == "true",
	//		"storage-class": func() string {
	//			if obj.Spec.Resources.Storage.StorageClass != "" {
	//				return obj.Spec.Resources.Storage.StorageClass
	//			}
	//			return fmt.Sprintf("%s-%s", obj.Spec.Region, ct.Xfs)
	//		}(),
	//		"owner-refs":      obj.GetOwnerReferences(),
	//		"existing-secret": getHelmSecretName(obj.Name),
	//	},
	//)

	fmt.Printf("yamls: \n\n%s\n", b)

	rr, err := r.yamlClient.ApplyYAML(ctx, b)
	if err != nil {
		return req.CheckFailed(HelmReady, check, err.Error()).Err(nil)
	}

	req.AddToOwnedResources(rr...)

	//if _, err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
	//	return req.CheckFailed(HelmReady, check, err.Error()).Err(nil)
	//}

	//cds, err := conditions.FromObject(helmRes)
	//if err != nil {
	//	return req.CheckFailed(HelmReady, check, err.Error())
	//}

	//deployedC := meta.FindStatusCondition(cds, "Deployed")
	//if deployedC == nil {
	//	return req.Done()
	//}
	//
	//if deployedC.Status == metav1.ConditionFalse {
	//	return req.CheckFailed(HelmReady, check, deployedC.Message)
	//}
	//
	//if deployedC.Status == metav1.ConditionTrue {
	//	check.Status = true
	//}

	check.Status = true
	if check != checks[HelmReady] {
		checks[HelmReady] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

func (r *Reconciler) reconSts(req *rApi.Request[*mongodbMsvcv1.StandaloneService]) stepResult.Result {
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
				return req.CheckFailed(StsReady, check, err.Error())
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
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

	b, err := templatesDir.ReadFile("templates/helm-mongodb-standalone.yml.tpl")
	if err != nil {
		return err
	}
	r.templateHelmMongoDB = b

	b, err = templatesDir.ReadFile("templates/helm-mongodb-standalone-auth.yml.tpl")
	if err != nil {
		return err
	}
	r.templateHelmMongoDBAuth = b

	builder := ctrl.NewControllerManagedBy(mgr).For(&mongodbMsvcv1.StandaloneService{})
	// builder.Owns(fn.NewUnstructured(constants.HelmMongoDBType))
	builder.Owns(&corev1.Secret{})

	watchList := []client.Object{
		&appsv1.StatefulSet{},
	}

	for i := range watchList {
		builder.Watches(
			&source.Kind{Type: watchList[i]},
			handler.EnqueueRequestsFromMapFunc(
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

	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
