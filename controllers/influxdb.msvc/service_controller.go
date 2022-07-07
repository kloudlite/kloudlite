package influxdbmsvc

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/errors"
	fn "operators.kloudlite.io/lib/functions"
	rApi "operators.kloudlite.io/lib/operator"
	"operators.kloudlite.io/lib/templates"

	"k8s.io/apimachinery/pkg/runtime"
	influxdbmsvcv1 "operators.kloudlite.io/apis/influxdb.msvc/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ServiceReconciler reconciles a Service object
type ServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const (
	SvcAdminPasswordKey = "admin-password"
	SvcAdminTokenKey    = "admin-token"
)
const (
	InfluxAdminBucket   = "primary"
	InfluxAdminOrg      = "primary"
	InfluxAdminUsername = "admin"
)

const (
	HelmInfluxDBExists     conditions.Type = "HelmInfluxDBExists"
	ReconcilerOutputExists conditions.Type = "ReconcilerOutputExists"
	GeneratedVars          conditions.Type = "GeneratedVars"
)

// +kubebuilder:rbac:groups=influxdb.msvc.kloudlite.io,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=influxdb.msvc.kloudlite.io,resources=services/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=influxdb.msvc.kloudlite.io,resources=services/finalizers,verbs=update

func (r *ServiceReconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	req, _ := rApi.NewRequest(ctx, r.Client, oReq.NamespacedName, &influxdbmsvcv1.Service{})

	if req == nil {
		return ctrl.Result{}, nil
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.Result(), x.Err()
		}
	}

	req.Logger.Infof("-------------------- NEW RECONCILATION------------------")

	if x := req.EnsureLabels(); !x.ShouldProceed() {
		return x.Result(), x.Err()
	}

	if x := r.reconcileStatus(req); !x.ShouldProceed() {
		return x.Result(), x.Err()
	}

	if x := r.reconcileOperations(req); !x.ShouldProceed() {
		return x.Result(), x.Err()
	}

	return ctrl.Result{}, nil
}

func (r *ServiceReconciler) finalize(req *rApi.Request[*influxdbmsvcv1.Service]) rApi.StepResult {
	return req.Finalize()
}

func (r *ServiceReconciler) reconcileStatus(req *rApi.Request[*influxdbmsvcv1.Service]) rApi.StepResult {
	ctx := req.Context()
	svcObj := req.Object

	isReady := true
	var cs []metav1.Condition

	// 1. check if helm service is available
	helmConditions, err := conditions.FromResource(
		ctx,
		r.Client,
		constants.HelmInfluxDBType,
		"Helm",
		fn.NN(svcObj.Namespace, svcObj.Name),
	)
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.FailWithStatusError(err)
		}
		isReady = false
		cs = append(cs, conditions.New(HelmInfluxDBExists, false, conditions.NotFound, err.Error()))
	}
	cs = append(cs, conditions.New(HelmInfluxDBExists, true, conditions.Found))
	cs = append(cs, helmConditions...)

	// 2. check if reconciler output exists
	_, err = rApi.Get(ctx, r.Client, fn.NN(svcObj.Namespace, fmt.Sprintf("msvc-%s", svcObj.Name)), &corev1.Secret{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.FailWithStatusError(err)
		}
		isReady = false
		cs = append(cs, conditions.New(ReconcilerOutputExists, false, conditions.NotFound, err.Error()))
	} else {
		cs = append(cs, conditions.New(ReconcilerOutputExists, true, conditions.Found))
	}

	// 3. generated vars
	if svcObj.Status.GeneratedVars.Exists(SvcAdminPasswordKey, SvcAdminTokenKey) {
		cs = append(cs, conditions.New(GeneratedVars, true, conditions.Found))
	} else {
		isReady = false
		cs = append(cs, conditions.New(GeneratedVars, false, conditions.NotReconciledYet))
	}

	newConditions, hasUpdated, err := conditions.Patch(svcObj.Status.Conditions, cs)
	if err != nil {
		return req.FailWithStatusError(err)
	}
	if !hasUpdated && isReady == svcObj.Status.IsReady {
		return req.Next()
	}

	svcObj.Status.IsReady = isReady
	svcObj.Status.Conditions = newConditions
	svcObj.Status.OpsConditions = []metav1.Condition{}

	return rApi.NewStepResult(&ctrl.Result{}, r.Status().Update(ctx, svcObj))
}

func (r *ServiceReconciler) reconcileOperations(req *rApi.Request[*influxdbmsvcv1.Service]) rApi.StepResult {
	ctx := req.Context()
	svcObj := req.Object

	if meta.IsStatusConditionFalse(svcObj.Status.Conditions, "GeneratedVars") {
		if err := svcObj.Status.GeneratedVars.Set(SvcAdminPasswordKey, fn.CleanerNanoid(40)); err != nil {
			return req.FailWithOpError(err)
		}
		if err := svcObj.Status.GeneratedVars.Set(SvcAdminTokenKey, fn.CleanerNanoid(40)); err != nil {
			return req.FailWithOpError(err)
		}
		return rApi.NewStepResult(&ctrl.Result{}, r.Status().Update(ctx, svcObj))
	}

	b, err := templates.Parse(
		templates.InfluxDB, map[string]any{
			"object": svcObj,
			// TODO: switch to dynamic storage class name
			"storage-class":  constants.DoBlockStorage,
			"admin-bucket":   InfluxAdminBucket,
			"admin-org":      InfluxAdminOrg,
			"admin-username": InfluxAdminUsername,
			"owner-refs": []metav1.OwnerReference{
				fn.AsOwner(svcObj, true),
			},
		},
	)
	if err != nil {
		return req.FailWithOpError(err)
	}

	if _, err := fn.KubectlApplyExec(b); err != nil {
		return req.FailWithOpError(err)
	}

	adminPassword, ok := svcObj.Status.GeneratedVars.GetString(SvcAdminPasswordKey)
	if !ok {
		return req.FailWithOpError(
			errors.Newf(
				"key=%s must have been present in .Status.GeneratedVars",
				SvcAdminPasswordKey,
			),
		)
	}

	adminToken, ok := svcObj.Status.GeneratedVars.GetString(SvcAdminTokenKey)
	if !ok {
		return req.FailWithOpError(
			errors.Newf(
				"key=%s must have been present in .Status.GeneratedVars",
				SvcAdminTokenKey,
			),
		)
	}

	host := fmt.Sprintf("%s.%s.svc.cluster.local:8086", svcObj.Name, svcObj.Namespace)
	b, err = templates.Parse(
		templates.Secret, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("msvc-%s", svcObj.Name),
				Namespace: svcObj.Namespace,
				OwnerReferences: []metav1.OwnerReference{
					fn.AsOwner(svcObj, true),
				},
			},
			StringData: map[string]string{
				"USERNAME": InfluxAdminUsername,
				"PASSWORD": adminPassword,
				"BUCKET":   InfluxAdminBucket,
				"ORG":      InfluxAdminOrg,
				"HOSTS":    host,
				"TOKEN":    adminToken,
				"URI":      fmt.Sprintf("http://%s", host),
			},
		},
	)
	if err != nil {
		return req.FailWithOpError(err)
	}

	if _, err := fn.KubectlApplyExec(b); err != nil {
		return req.FailWithOpError(err)
	}

	return req.Done()
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&influxdbmsvcv1.Service{}).
		Complete(r)
}
