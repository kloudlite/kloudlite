package standalone_service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	mongodbMsvcv1 "github.com/kloudlite/operator/apis/mongodb.msvc/v1"
	"github.com/kloudlite/operator/operators/msvc-mongo/internal/env"
	"github.com/kloudlite/operator/operators/msvc-mongo/internal/templates"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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

const (
	// secret keys
	RootPassword string = "ROOT_PASSWORD"
	Hosts        string = "HOSTS"
	URI          string = "URI"
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

	// if step := r.reconHelmSecret(req); !step.ShouldProceed() {
	// 	return step.ReconcilerResponse()
	// }

	if step := r.reconHelm(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconSts(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
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
	scrt := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: secretName, Namespace: obj.Namespace}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, scrt, func() error {
		scrt.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		scrt.SetFinalizers(obj.GetFinalizers())

		scrt.Labels = obj.GetLabels()

		if scrt.Data == nil {
			scrt.Data = map[string][]byte{
				"ROOT_PASSWORD": []byte(fn.CleanerNanoid(40)),
				"HOSTS": func() []byte {
					var hosts []string
					for i := 0; i < obj.Spec.ReplicaCount; i++ {
						hosts = append(hosts, fmt.Sprintf("%s-%d.%s.%s.svc.cluster.local:27017", obj.Name, i, obj.Name, obj.Namespace))
					}
					return []byte(strings.Join(hosts, ","))
				}(),
				"URI": []byte(fmt.Sprintf("mongodb://%s:%s@%s/admin?authSource=admin", "root", scrt.Data[RootPassword], scrt.Data[Hosts])),
			}
		}

		return nil
	}); err != nil {
		return req.CheckFailed(AccessCredsReady, check, err.Error()).Err(nil)
	}

	req.AddToOwnedResources(rApi.ParseResourceRef(scrt))

	// creating helm secrets
	helmAdminSecret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: getHelmSecretName(obj.Name), Namespace: obj.Namespace}}
	controllerutil.CreateOrUpdate(ctx, r.Client, helmAdminSecret, func() error {
		helmAdminSecret.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})

		if helmAdminSecret.Data == nil {
			helmAdminSecret.Data = map[string][]byte{
				"mongodb-passwords":        []byte(""),
				"mongodb-root-password":    scrt.Data[RootPassword],
				"mongodb-metrics-password": []byte(""),
				"mongodb-replica-set-key":  []byte(""),
			}
		}
		return nil
	})

	check.Status = true
	if check != checks[AccessCredsReady] {
		checks[AccessCredsReady] = check
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

	// TODO (nxtcoder17): when increasing pvc volume size, we can not trigger helm update, as it complains about forbidden field
	b, err := templates.ParseBytes(r.templateHelmMongoDB, map[string]any{
		"name":      obj.Name,
		"namespace": obj.Namespace,
		"labels": map[string]string{
			constants.MsvcNameKey: obj.Name,
		},
		"owner-refs":    []metav1.OwnerReference{fn.AsOwner(obj)},
		"node-selector": obj.Spec.NodeSelector,

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

	rr, err := r.yamlClient.ApplyYAML(ctx, b)
	if err != nil {
		return req.CheckFailed(HelmReady, check, err.Error()).Err(nil)
	}

	req.AddToOwnedResources(rr...)

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

	var err error
	r.templateHelmMongoDB, err = templates.Read(templates.HelmMongoDBStandalone)
	if err != nil {
		return err
	}

	r.templateHelmMongoDBAuth, err = templates.Read(templates.HelmMongoDBStandaloneAuth)
	if err != nil {
		return err
	}

	builder := ctrl.NewControllerManagedBy(mgr).For(&mongodbMsvcv1.StandaloneService{})
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
