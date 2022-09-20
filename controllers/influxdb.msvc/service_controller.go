package influxdbmsvc

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ct "operators.kloudlite.io/apis/common-types"
	"operators.kloudlite.io/env"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/errors"
	fn "operators.kloudlite.io/lib/functions"
	"operators.kloudlite.io/lib/logging"
	rApi "operators.kloudlite.io/lib/operator"
	stepResult "operators.kloudlite.io/lib/operator/step-result"
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
	logger logging.Logger
	Name   string
}

func (r *ServiceReconciler) GetName() string {
	return r.Name
}

const (
	SvcAdminPasswordKey = "admin-password"
	SvcAdminTokenKey    = "admin-token"
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
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, oReq.NamespacedName, &influxdbmsvcv1.Service{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// STEP: cleaning up last run, clearing opsConditions
	if len(req.Object.Status.OpsConditions) > 0 {
		req.Object.Status.OpsConditions = []metav1.Condition{}
		return ctrl.Result{RequeueAfter: 0}, r.Status().Update(ctx, req.Object)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
	}

	req.Logger.Infof("-------------------- NEW RECONCILATION------------------")

	if x := req.EnsureLabelsAndAnnotations(); !x.ShouldProceed() {
		return x.ReconcilerResponse()
	}

	if x := r.reconcileStatus(req); !x.ShouldProceed() {
		return x.ReconcilerResponse()
	}

	if x := r.reconcileOperations(req); !x.ShouldProceed() {
		return x.ReconcilerResponse()
	}

	return ctrl.Result{}, nil
}

func (r *ServiceReconciler) finalize(req *rApi.Request[*influxdbmsvcv1.Service]) stepResult.Result {
	return req.Finalize()
}

func (r *ServiceReconciler) reconcileStatus(req *rApi.Request[*influxdbmsvcv1.Service]) stepResult.Result {
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

	if err := r.Status().Update(ctx, svcObj); err != nil {
		return req.FailWithStatusError(err)
	}
	return req.Done()
}

func (r *ServiceReconciler) reconcileOperations(req *rApi.Request[*influxdbmsvcv1.Service]) stepResult.Result {
	ctx := req.Context()
	svcObj := req.Object

	if meta.IsStatusConditionFalse(svcObj.Status.Conditions, "GeneratedVars") {
		if err := svcObj.Status.GeneratedVars.Set(SvcAdminPasswordKey, fn.CleanerNanoid(40)); err != nil {
			return req.FailWithOpError(err)
		}
		if err := svcObj.Status.GeneratedVars.Set(SvcAdminTokenKey, fn.CleanerNanoid(40)); err != nil {
			return req.FailWithOpError(err)
		}
		if err := r.Status().Update(ctx, svcObj); err != nil {
			return req.FailWithOpError(err)
		}
		return req.Done()
	}

	storageClass, err := svcObj.Spec.CloudProvider.GetStorageClass(ct.Ext4)
	if err != nil {
		return req.FailWithOpError(err)
	}
	b, err := templates.Parse(
		templates.InfluxDB, map[string]any{
			"object":        svcObj,
			"storage-class": storageClass,
			"owner-refs": []metav1.OwnerReference{
				fn.AsOwner(svcObj, true),
			},
		},
	)
	if err != nil {
		req.Logger.Errorf(err, "failed parsing template %s", templates.InfluxDB)
		return req.FailWithOpError(err).Err(nil)
	}

	if err := fn.KubectlApplyExec(ctx, b); err != nil {
		req.Logger.Errorf(err, "failed kubectl apply for template %s", templates.InfluxDB)
		return req.FailWithOpError(err).Err(nil)
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
		templates.CoreV1.Secret, map[string]any{
			"name":       "msvc-" + svcObj.Name,
			"namespace":  svcObj.Namespace,
			"labels":     svcObj.GetLabels(),
			"owner-refs": []metav1.OwnerReference{fn.AsOwner(svcObj, true)},
			"string-data": map[string]string{
				"USERNAME": svcObj.Spec.Admin.Username,
				"PASSWORD": adminPassword,
				"BUCKET":   svcObj.Spec.Admin.Bucket,
				"ORG":      svcObj.Spec.Admin.Org,
				"HOSTS":    host,
				"TOKEN":    adminToken,
				"URI":      fmt.Sprintf("http://%s", host),
			},
		},
	)
	if err != nil {
		req.Logger.Errorf(err, "failed parsing template %s", templates.Secret)
		return req.FailWithOpError(err).Err(nil)
	}

	if err := fn.KubectlApplyExec(ctx, b); err != nil {
		req.Logger.Errorf(err, "failed kubectl apply for emplate %s", templates.Secret)
		return req.FailWithOpError(err).Err(nil)
	}

	return req.Next()
}

// SetupWithManager sets up the controllers with the Manager.
func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager, envVars *env.Env, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)

	return ctrl.NewControllerManagedBy(mgr).
		For(&influxdbmsvcv1.Service{}).
		Complete(r)
}
