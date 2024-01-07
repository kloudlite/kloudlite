package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kloudlite/operator/pkg/kubectl"

	influxdbMsvcv1 "github.com/kloudlite/operator/apis/influxdb.msvc/v1"
	"github.com/kloudlite/operator/operators/msvc-influx/internal/env"
	"github.com/kloudlite/operator/operators/msvc-influx/internal/types"
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
	Scheme     *runtime.Scheme
	logger     logging.Logger
	Name       string
	Env        *env.Env
	yamlClient kubectl.YAMLClient
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	DefaultsPatched  string = "defaults-patched"
	HelmReady        string = "helm-ready"
	StsReady         string = "sts-ready"
	AccessCredsReady string = "access-creds-ready"
)

const (
	KeyMsvcOutput string = "msvc-output"
)

// +kubebuilder:rbac:groups=influxdb.msvc.kloudlite.io,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=influxdb.msvc.kloudlite.io,resources=services/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=influxdb.msvc.kloudlite.io,resources=services/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, request.NamespacedName, &influxdbMsvcv1.Service{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	req.PreReconcile()
	defer req.PostReconcile()

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.RestartIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	// TODO: initialize all checks here
	if step := req.EnsureChecks(DefaultsPatched, HelmReady, StsReady, AccessCredsReady); !step.ShouldProceed() {
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

func (r *Reconciler) finalize(req *rApi.Request[*influxdbMsvcv1.Service]) stepResult.Result {
	return req.Finalize()
}

func (r *Reconciler) reconDefaults(req *rApi.Request[*influxdbMsvcv1.Service]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(DefaultsPatched)
	defer req.LogPostCheck(DefaultsPatched)

	hasUpdated := false

	if obj.Spec.Admin == nil {
		hasUpdated = true
		obj.Spec.Admin = &influxdbMsvcv1.Admin{
			Username: "admin",
			Bucket:   "admin",
			Org:      "admin",
		}
	}

	if hasUpdated {
		if err := r.Update(ctx, obj); err != nil {
			return req.CheckFailed(DefaultsPatched, check, err.Error())
		}
		return req.Done().RequeueAfter(1 * time.Second)
	}

	check.Status = true
	if check == checks[DefaultsPatched] {
		checks[DefaultsPatched] = check
		req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) reconAccessCreds(req *rApi.Request[*influxdbMsvcv1.Service]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(AccessCredsReady)
	defer req.LogPostCheck(AccessCredsReady)

	secretName := "msvc-" + obj.Name
	scrt, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, secretName), &corev1.Secret{})
	if err != nil {
		req.Logger.Infof("secret %s does not exist yet, would be creating it ...", fn.NN(obj.Namespace, secretName).String())
	}

	fail := func(err error) stepResult.Result {
		return req.CheckFailed(AccessCredsReady, check, err.Error())
	}

	if scrt == nil {
		adminPassword := fn.CleanerNanoid(40)
		adminToken := fn.CleanerNanoid(40)

		host := fmt.Sprintf("%s.%s.svc.cluster.local:8086", obj.Name, obj.Namespace)
		b, err := templates.Parse(
			templates.Secret, map[string]any{
				"name":       secretName,
				"namespace":  obj.Namespace,
				"labels":     obj.GetLabels(),
				"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj)},
				"string-data": types.MsvcOutput{
					Username: obj.Spec.Admin.Username,
					Password: adminPassword,
					Bucket:   obj.Spec.Admin.Bucket,
					Org:      obj.Spec.Admin.Org,
					Host:     host,
					Token:    adminToken,
					Uri:      fmt.Sprintf("http://%s", host),
				},
			},
		)
		if err != nil {
			return fail(err)
		}

		if err := fn.KubectlApplyExec(ctx, b); err != nil {
			return fail(err)
		}

		checks[AccessCredsReady] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	if !fn.IsOwner(obj, fn.AsOwner(scrt)) {
		obj.SetOwnerReferences(append(obj.GetOwnerReferences(), fn.AsOwner(scrt)))
		if err := r.Update(ctx, obj); err != nil {
			return fail(err)
		}
		return req.Done().RequeueAfter(2 * time.Second)
	}

	check.Status = true
	if check != checks[AccessCredsReady] {
		checks[AccessCredsReady] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	b, err := json.Marshal(scrt.Data)
	if err != nil {
		return req.CheckFailed(AccessCredsReady, check, err.Error()).Err(nil)
	}

	var m types.MsvcOutput
	if err := json.Unmarshal(b, &m); err != nil {
		return req.CheckFailed(AccessCredsReady, check, err.Error()).Err(nil)
	}
	rApi.SetLocal(req, KeyMsvcOutput, m)
	return req.Next()
}

func (r *Reconciler) reconHelm(req *rApi.Request[*influxdbMsvcv1.Service]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(HelmReady)
	defer req.LogPostCheck(HelmReady)

	fail := func(err error) stepResult.Result {
		return req.CheckFailed(HelmReady, check, err.Error())
	}

	helmRes, err := rApi.Get(
		ctx, r.Client, fn.NN(obj.Namespace, obj.Name), fn.NewUnstructured(constants.HelmInfluxDBType),
	)
	if err != nil {
		req.Logger.Infof("helm reosurce (%s) not found, will be creating it", fn.NN(obj.Namespace, obj.Name).String())
	}

	msvcOutput, ok := rApi.GetLocal[types.MsvcOutput](req, KeyMsvcOutput)
	if !ok {
		return req.CheckFailed(HelmReady, check, fmt.Sprintf("key %s is not available in req-locals", KeyMsvcOutput))
	}

	b, err := templates.Parse(
		templates.InfluxDB, map[string]any{
			"obj":            obj,
			"owner-refs":     obj.GetOwnerReferences(),
			"storage-class":  obj.Spec.Resources.Storage.StorageClass,
			"admin-password": msvcOutput.Password,
			"admin-token":    msvcOutput.Token,
		},
	)
	if err != nil {
		return req.CheckFailed(HelmReady, check, err.Error()).Err(nil)
	}

	rr, err := r.yamlClient.ApplyYAML(ctx, b)
	if err != nil {
		return fail(err)
	}

	req.AddToOwnedResources(rr...)

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
		req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) reconSts(req *rApi.Request[*influxdbMsvcv1.Service]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(StsReady)
	defer req.LogPostCheck(StsReady)

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
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig(), kubectl.YAMLClientOpts{Logger: r.logger})

	builder := ctrl.NewControllerManagedBy(mgr).For(&influxdbMsvcv1.Service{})
	builder.Owns(&corev1.Secret{})
	builder.Owns(fn.NewUnstructured(constants.HelmInfluxDBType))

	builder.Watches(
		&appsv1.StatefulSet{}, handler.EnqueueRequestsFromMapFunc(
			func(_ context.Context, obj client.Object) []reconcile.Request {
				v, ok := obj.GetLabels()[constants.MsvcNameKey]
				if !ok {
					return nil
				}
				return []reconcile.Request{{NamespacedName: fn.NN(obj.GetNamespace(), v)}}
			},
		),
	)
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
