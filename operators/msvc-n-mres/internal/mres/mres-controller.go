package mres

import (
	"context"
	"encoding/json"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	influxdbMsvcv1 "operators.kloudlite.io/apis/influxdb.msvc/v1"
	mongodbMsvcv1 "operators.kloudlite.io/apis/mongodb.msvc/v1"
	mysqlMsvcv1 "operators.kloudlite.io/apis/mysql.msvc/v1"
	redisMsvcv1 "operators.kloudlite.io/apis/redis.msvc/v1"
	"operators.kloudlite.io/lib/constants"
	fn "operators.kloudlite.io/lib/functions"
	"operators.kloudlite.io/lib/harbor"
	"operators.kloudlite.io/lib/logging"
	rApi "operators.kloudlite.io/lib/operator"
	stepResult "operators.kloudlite.io/lib/operator/step-result"
	"operators.kloudlite.io/lib/templates"
	env2 "operators.kloudlite.io/operators/msvc-n-mres/internal/env"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ManagedResourceReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	harborCli *harbor.Client
	logger    logging.Logger
	Name      string
	Env       *env2.Env
}

func (r *ManagedResourceReconciler) GetName() string {
	return r.Name
}

const (
	RealMresReady string = "real-mres-ready"
	MsvcIsOwner   string = "msvc-is-owner"
)

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=crds/finalizers,verbs=update

func (r *ManagedResourceReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, request.NamespacedName, &crdsv1.ManagedResource{})
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

	if step := req.EnsureChecks(RealMresReady); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconOwnership(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconRealMres(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	req.Logger.Infof("RECONCILATION COMPLETE")
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod * time.Second}, r.Status().Update(ctx, req.Object)
}

func (r *ManagedResourceReconciler) finalize(req *rApi.Request[*crdsv1.ManagedResource]) stepResult.Result {
	return req.Finalize()
}

func (r *ManagedResourceReconciler) reconOwnership(req *rApi.Request[*crdsv1.ManagedResource]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks

	check := rApi.Check{Generation: obj.Generation}
	msvc, err := rApi.Get(
		ctx, r.Client, fn.NN(obj.Namespace, obj.Spec.MsvcRef.Name), &crdsv1.ManagedService{},
	)

	if err != nil {
		return req.CheckFailed(MsvcIsOwner, check, err.Error())
	}

	if msvc != nil && !fn.IsOwner(obj, fn.AsOwner(msvc)) {
		obj.SetOwnerReferences(append(obj.GetOwnerReferences(), fn.AsOwner(msvc)))
		if err := r.Update(ctx, obj); err != nil {
			return req.FailWithOpError(err)
		}
		return req.Done()
	}

	check.Status = true
	if check != checks[MsvcIsOwner] {
		checks[MsvcIsOwner] = check
		return req.UpdateStatus()
	}

	return req.Next()
}

func (r *ManagedResourceReconciler) reconRealMres(req *rApi.Request[*crdsv1.ManagedResource]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	realMres, err := rApi.Get(
		ctx, r.Client, fn.NN(obj.Namespace, obj.Name), fn.NewUnstructured(
			metav1.TypeMeta{
				APIVersion: obj.Spec.MsvcRef.APIVersion,
				Kind:       obj.Spec.MresKind.Kind,
			},
		),
	)
	if err != nil {
		req.Logger.Infof("real managed resource (%s) does not exist, creating it now...", fn.NN(obj.Namespace, obj.Name).String())
	}

	if realMres == nil || check.Generation > checks[RealMresReady].Generation {
		b, err := templates.Parse(
			templates.CommonMres, map[string]any{
				"object": obj,
				"owner-refs": []metav1.OwnerReference{
					fn.AsOwner(obj, true),
				},
			},
		)
		if err != nil {
			return req.CheckFailed(RealMresReady, check, err.Error()).Err(nil)
		}
		if err := fn.KubectlApplyExec(ctx, b); err != nil {
			return req.CheckFailed(RealMresReady, check, err.Error()).Err(nil)
		}

		checks[RealMresReady] = check
		return req.UpdateStatus()
	}

	b, err := json.Marshal(realMres)
	if err != nil {
		return req.CheckFailed(RealMresReady, check, err.Error()).Err(nil)
	}

	var realMresObj struct {
		Status rApi.Status `json:"status"`
	}
	if err := json.Unmarshal(b, &realMresObj); err != nil {
		return req.CheckFailed(RealMresReady, check, err.Error()).Err(nil)
	}

	if !realMresObj.Status.IsReady {
		b, err := json.Marshal(realMresObj.Status.Message)
		if err != nil {
			return req.CheckFailed(RealMresReady, check, err.Error()).Err(nil)
		}
		return req.CheckFailed(RealMresReady, check, string(b))
	}

	check.Status = true
	if check != checks[RealMresReady] {
		checks[RealMresReady] = check
		return req.UpdateStatus()
	}

	return req.Next()
}

func (r *ManagedResourceReconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)

	builder := ctrl.NewControllerManagedBy(mgr).For(&crdsv1.ManagedResource{})
	builder.Owns(&corev1.Secret{})

	children := []client.Object{
		&redisMsvcv1.StandaloneService{},
		&redisMsvcv1.ClusterService{},
		&redisMsvcv1.ACLAccount{},

		&mongodbMsvcv1.StandaloneService{},
		&mongodbMsvcv1.ClusterService{},
		&mongodbMsvcv1.Database{},

		&mysqlMsvcv1.ClusterService{},
		&mysqlMsvcv1.StandaloneService{},
		&mysqlMsvcv1.Database{},

		&influxdbMsvcv1.Bucket{},
		&influxdbMsvcv1.Service{},
	}

	for _, child := range children {
		builder.Owns(child)
	}

	return builder.Complete(r)
}
