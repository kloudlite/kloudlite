package standaloneService

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	ct "github.com/kloudlite/operator/apis/common-types"
	neo4jMsvcv1 "github.com/kloudlite/operator/apis/neo4j.msvc/v1"
	"github.com/kloudlite/operator/operators/msvc-neo4j/internal/env"
	"github.com/kloudlite/operator/operators/msvc-neo4j/internal/types"
	"github.com/kloudlite/operator/pkg/conditions"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	"github.com/kloudlite/operator/pkg/templates"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
	logger logging.Logger
	Name   string
	Env    *env.Env
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	HelmReady        string = "helm-ready"
	StsReady         string = "sts-ready"
	AccessCredsReady string = "access-creds-ready"
)

const (
	KeyAccessCreds  string = "access-creds"
	KeyRootPassword string = "root-password"
)

// +kubebuilder:rbac:groups=neo4j.msvc.kloudlite.io,resources=standaloneServices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=neo4j.msvc.kloudlite.io,resources=standaloneServices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=neo4j.msvc.kloudlite.io,resources=standaloneServices/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(
		rApi.NewReconcilerCtx(ctx, r.logger),
		r.Client,
		request.NamespacedName,
		&neo4jMsvcv1.StandaloneService{},
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

	req.Object.Status.IsReady = true
	req.Object.Status.LastReconcileTime = &metav1.Time{Time: time.Now()}
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) finalize(req *rApi.Request[*neo4jMsvcv1.StandaloneService]) stepResult.Result {
	return req.Finalize()
}

func (r *Reconciler) reconAccessCreds(req *rApi.Request[*neo4jMsvcv1.StandaloneService]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks

	check := rApi.Check{Generation: obj.Generation}
	secretName := "msvc-" + obj.Name
	scrt, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, secretName), &corev1.Secret{})
	if err != nil {
		req.Logger.Infof("secret %s does not exist yet, would be creating it ...", fn.NN(obj.Namespace, secretName).String())
	}

	if scrt == nil {
		b, err := templates.Parse(
			templates.Secret, map[string]any{
				"name":       secretName,
				"namespace":  obj.Namespace,
				"labels":     obj.GetLabels(),
				"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj)},
				"string-data": types.MsvcOutput{
					RootPassword: fn.CleanerNanoid(40),
					Hosts:        fmt.Sprintf("%s-0.%s.%s.svc.cluster.local", obj.Name, obj.Name, obj.Namespace),
					AdminHosts:   fmt.Sprintf("%s-0.%s-admin.%s.svc.cluster.local", obj.Name, obj.Name, obj.Namespace),
					PortBolt:     "7687",
					PortHttp:     "7474",
					PortHttps:    "7473",
					PortBackup:   "6362",
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
			return req.CheckFailed(AccessCredsReady, check, err.Error())
		}
		return req.Done().RequeueAfter(2 * time.Second)
	}

	check.Status = true
	if check != checks[AccessCredsReady] {
		checks[AccessCredsReady] = check
		return req.UpdateStatus()
	}

	accessCreds, err := fn.ParseFromSecret[types.MsvcOutput](scrt)
	if err != nil {
		return req.CheckFailed(KeyAccessCreds, check, err.Error())
	}

	rApi.SetLocal(req, KeyAccessCreds, accessCreds)
	return req.Next()
}

func (r *Reconciler) reconHelm(req *rApi.Request[*neo4jMsvcv1.StandaloneService]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	helmRes, err := rApi.Get(
		ctx, r.Client, fn.NN(obj.Namespace, obj.Name), fn.NewUnstructured(constants.HelmNeo4JStandaloneType),
	)
	if err != nil {
		req.Logger.Infof("helm resource (%s) not found, will be creating it", fn.NN(obj.Namespace, obj.Name).String())
	}

	accessCreds, ok := rApi.GetLocal[*types.MsvcOutput](req, KeyAccessCreds)
	if !ok {
		return req.CheckFailed(HelmReady, check, err.Error())
	}

	if helmRes == nil || check.Generation > checks[HelmReady].Generation {
		b, err := templates.Parse(
			templates.MsvcHelmNeo4jStandalone, map[string]any{
				"obj":           obj,
				"storage-class": fmt.Sprintf("%s-%s", obj.Spec.Region, ct.Ext4),
				"owner-refs":    []metav1.OwnerReference{fn.AsOwner(obj, true)},
				"root-password": accessCreds.RootPassword,
				"labels": map[string]string{
					constants.MsvcNameKey: obj.Name,
				},
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

func (r *Reconciler) reconSts(req *rApi.Request[*neo4jMsvcv1.StandaloneService]) stepResult.Result {
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

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)

	builder := ctrl.NewControllerManagedBy(mgr).For(&neo4jMsvcv1.StandaloneService{})
	builder.Owns(&corev1.Secret{})
	builder.Owns(fn.NewUnstructured(constants.HelmNeo4JStandaloneType))

	builder.Watches(
		&appsv1.StatefulSet{}, handler.EnqueueRequestsFromMapFunc(
			func(ctx context.Context, obj client.Object) []reconcile.Request {
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
