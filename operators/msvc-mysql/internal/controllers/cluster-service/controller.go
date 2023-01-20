package clusterService

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ct "github.com/kloudlite/operator/apis/common-types"
	mysqlMsvcv1 "github.com/kloudlite/operator/apis/mysql.msvc/v1"
	"github.com/kloudlite/operator/operators/msvc-mysql/internal/env"
	"github.com/kloudlite/operator/operators/msvc-mysql/internal/types"
	"github.com/kloudlite/operator/pkg/conditions"
	"github.com/kloudlite/operator/pkg/constants"
	"github.com/kloudlite/operator/pkg/errors"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	"github.com/kloudlite/operator/pkg/templates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	logger     logging.Logger
	Name       string
	yamlClient *kubectl.YAMLClient
	Env        *env.Env
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	HelmReady        string = "helm-ready"
	StsReady         string = "sts-ready"
	AccessCredsReady string = "access-creds-ready"
	PVCReady         string = "pvc-ready"
)

const (
	KeyMsvcOutput     string = "msvc-output"
	KeyStsPvcInitSize string = "kl-sts-pvc-init-size"
)

// +kubebuilder:rbac:groups=mysql.msvc.kloudlite.io,resources=clusterServices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mysql.msvc.kloudlite.io,resources=clusterServices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mysql.msvc.kloudlite.io,resources=clusterServices/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, request.NamespacedName, &mysqlMsvcv1.ClusterService{})
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

	if step := r.reconHelm(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconSts(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconPVC(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) finalize(req *rApi.Request[*mysqlMsvcv1.ClusterService]) stepResult.Result {
	return req.Finalize()
}

func (r *Reconciler) reconAccessCreds(req *rApi.Request[*mysqlMsvcv1.ClusterService]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks

	check := rApi.Check{Generation: obj.Generation}
	secretName := "msvc-" + obj.Name
	scrt, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, secretName), &corev1.Secret{})
	if err != nil {
		req.Logger.Infof("secret %s does not exist yet, would be creating it ...", fn.NN(obj.Namespace, secretName).String())
	}

	if scrt == nil {
		rootPassword := fn.CleanerNanoid(30)
		mysqlHost := fmt.Sprintf("%s-primary-headless.%s.svc.cluster.local:3306", obj.Name, obj.Namespace)
		b, err := templates.Parse(
			templates.Secret, map[string]any{
				"name":       secretName,
				"namespace":  obj.Namespace,
				"labels":     obj.GetLabels(),
				"owner-refs": obj.GetOwnerReferences(),
				"string-data": types.MsvcOutput{
					ReplicationPassword: fn.CleanerNanoid(30),
					RootPassword:        rootPassword,
					Hosts:               mysqlHost,
					DSN:                 fmt.Sprintf("mysql://%s:%s@tcp(%s:3306)/%s", "root", rootPassword, mysqlHost, "mysql"),
					URI:                 fmt.Sprintf("mysql://%s:%s@%s/%s", "root", rootPassword, mysqlHost, "mysql"),
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

	output, err := fn.ParseFromSecret[types.MsvcOutput](scrt)
	if err != nil {
		return req.CheckFailed(AccessCredsReady, check, err.Error()).Err(nil)
	}

	if output == nil {
		return req.CheckFailed(AccessCredsReady, check, "output secret is nil").Err(nil)
	}

	// ensure existing secret for helm
	helmSecret, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, "helm-"+obj.Name), &corev1.Secret{})
	if err != nil {
		req.Logger.Infof("helm secret does not exist, will be creating it prior to helm installation")
		helmSecret = nil
	}

	if helmSecret == nil {
		if err := r.Create(
			ctx, &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "helm-" + obj.Name,
					Namespace:       obj.Namespace,
					OwnerReferences: []metav1.OwnerReference{fn.AsOwner(obj, true)},
				},
				StringData: map[string]string{
					KeyStsPvcInitSize:            obj.Spec.Resources.Storage.Size,
					"mysql-root-password":        output.RootPassword,
					"mysql-replication-password": output.ReplicationPassword,
					"mysql-password":             "",
				},
				Type: "",
			},
		); err != nil {
			return req.CheckFailed(AccessCredsReady, check, err.Error())
		}
	}

	rApi.SetLocal(req, KeyMsvcOutput, *output)
	rApi.SetLocal(req, KeyStsPvcInitSize, string(helmSecret.Data[KeyStsPvcInitSize]))
	return req.Next()
}

func (r *Reconciler) reconHelm(req *rApi.Request[*mysqlMsvcv1.ClusterService]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	helmRes, err := rApi.Get(
		ctx, r.Client, fn.NN(obj.Namespace, obj.Name), fn.NewUnstructured(constants.HelmMysqlType),
	)

	if err != nil {
		req.Logger.Infof("helm resource (%s) not found, will be creating it", fn.NN(obj.Namespace, obj.Name).String())
	}

	stsPvcInitSize, ok := rApi.GetLocal[string](req, KeyStsPvcInitSize)
	if !ok {
		return req.CheckFailed(HelmReady, check, errors.NotInLocals(KeyStsPvcInitSize).Error()).Err(nil)
	}

	if helmRes == nil || check.Generation > checks[HelmReady].Generation {
		b, err := templates.Parse(
			templates.MysqlCluster, map[string]any{
				"obj":             obj,
				"existing-secret": "helm-" + obj.Name,
				"storage-class": func() string {
					if obj.Spec.Resources.Storage.StorageClass != "" {
						return obj.Spec.Resources.Storage.StorageClass
					}
					return fmt.Sprintf("%s-%s", obj.Spec.Region, ct.Ext4)
				}(),
				KeyStsPvcInitSize: stsPvcInitSize,
				"owner-refs":      []metav1.OwnerReference{fn.AsOwner(obj, true)},
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

func (r *Reconciler) reconSts(req *rApi.Request[*mysqlMsvcv1.ClusterService]) stepResult.Result {
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

func (r *Reconciler) reconPVC(req *rApi.Request[*mysqlMsvcv1.ClusterService]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	var pvcList corev1.PersistentVolumeClaimList
	if err := r.List(
		ctx, &pvcList, &client.ListOptions{
			LabelSelector: labels.SelectorFromValidatedSet(
				map[string]string{
					constants.MsvcNameKey: obj.Name,
				},
			),
			Namespace: obj.Namespace,
		},
	); err != nil {
		return req.CheckFailed(PVCReady, check, err.Error())
	}

	hasResized := false

	for _, pvc := range pvcList.Items {
		currSize, ok := pvc.Spec.Resources.Requests.Storage().AsInt64()
		if !ok {
			return req.CheckFailed(PVCReady, check, "storage can't be converted to int64")
		}

		newSize, err := obj.Spec.Resources.Storage.ToInt()
		if err != nil {
			return req.CheckFailed(PVCReady, check, err.Error()).Err(nil)
		}

		if currSize < newSize {
			hasResized = true
			pvc.Spec.Resources.Requests = corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse(obj.Spec.Resources.Storage.Size),
			}

			if err := r.Update(ctx, &pvc); err != nil {
				return req.CheckFailed(PVCReady, check, err.Error())
			}
		}
	}

	if hasResized {
		var podsList corev1.PodList
		if err := r.List(
			ctx, &podsList, &client.ListOptions{
				LabelSelector: labels.SelectorFromValidatedSet(
					map[string]string{
						constants.MsvcNameKey: obj.Name,
					},
				),
				Namespace: obj.Namespace,
			},
		); err != nil {
			return req.CheckFailed(PVCReady, check, err.Error()).Err(nil)
		}

		for _, pod := range podsList.Items {
			if err := r.Delete(ctx, &pod); err != nil {
				return req.CheckFailed(PVCReady, check, err.Error())
			}
		}
	}

	check.Status = true
	if check != checks[PVCReady] {
		checks[PVCReady] = check
		return req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

	builder := ctrl.NewControllerManagedBy(mgr).For(&mysqlMsvcv1.ClusterService{})
	builder.Owns(&corev1.Secret{})
	builder.Owns(fn.NewUnstructured(constants.HelmMysqlType))

	builder.Watches(
		&source.Kind{Type: &appsv1.StatefulSet{}}, handler.EnqueueRequestsFromMapFunc(
			func(obj client.Object) []reconcile.Request {
				v, ok := obj.GetLabels()[constants.MsvcNameKey]
				if !ok {
					return nil
				}
				return []reconcile.Request{{NamespacedName: fn.NN(obj.GetNamespace(), v)}}
			},
		),
	)

	return builder.Complete(r)
}
