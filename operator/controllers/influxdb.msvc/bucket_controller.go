package influxdbmsvc

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	influxDB "operators.kloudlite.io/apis/influxdb.msvc/v1"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/errors"
	fn "operators.kloudlite.io/lib/functions"
	libInflux "operators.kloudlite.io/lib/influx"
	rApi "operators.kloudlite.io/lib/operator"
	"operators.kloudlite.io/lib/templates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// BucketReconciler reconciles a Bucket object
type BucketReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const (
	MsvcTokenKey = "TOKEN"
	MsvcUri      = "URI"
	MsvcOrgName  = "ORG"
)

const (
	BucketIdKey = "bucket-id"
)

const (
	BucketExists    conditions.Type = "BucketExists"
	HasBucketId     conditions.Type = "HasBucketId"
	MsvcOuputExists conditions.Type = "MsvcOutputExists"
)

// +kubebuilder:rbac:groups=influxdb.msvc.kloudlite.io,resources=buckets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=influxdb.msvc.kloudlite.io,resources=buckets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=influxdb.msvc.kloudlite.io,resources=buckets/finalizers,verbs=update

func (r *BucketReconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	req, _ := rApi.NewRequest(ctx, r.Client, oReq.NamespacedName, &influxDB.Bucket{})

	if req == nil {
		return ctrl.Result{}, nil
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.Result(), x.Err()
		}
	}

	req.Logger.Info("-------------------- NEW RECONCILATION------------------")

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

func (r *BucketReconciler) finalize(req *rApi.Request[*influxDB.Bucket]) rApi.StepResult {
	return req.Finalize()
}

func (r *BucketReconciler) reconcileStatus(req *rApi.Request[*influxDB.Bucket]) rApi.StepResult {
	ctx := req.Context()
	bucketObj := req.Object

	isReady := true
	var cs []metav1.Condition

	// 1 . managed svc ready
	msvc, err := rApi.Get(
		ctx, r.Client, fn.NN(bucketObj.Namespace, bucketObj.Spec.ManagedSvcName),
		&influxDB.Service{},
	)

	if err != nil {
		return req.FailWithStatusError(err)
	}

	if !msvc.Status.IsReady {
		return req.FailWithStatusError(errors.Newf("msvc is not ready"))
	}

	// STEP: check managed service output is ready
	msvcOutput, err2 := rApi.Get(
		ctx,
		r.Client,
		fn.NN(msvc.Namespace, fmt.Sprintf("msvc-%s", msvc.Name)),
		&corev1.Secret{},
	)
	if err2 != nil {
		if !apiErrors.IsNotFound(err) {
			return req.FailWithStatusError(err)
		}
		cs = append(cs, conditions.New(MsvcOuputExists, false, conditions.NotFound, err.Error()))
		isReady = false
	}

	if msvcOutput != nil {
		msvcToken := string(msvcOutput.Data[MsvcTokenKey])
		msvcUri := string(msvcOutput.Data[MsvcUri])
		rApi.SetLocal(req, MsvcTokenKey, msvcToken)
		rApi.SetLocal(req, MsvcUri, msvcUri)
		rApi.SetLocal(req, MsvcOrgName, string(msvcOutput.Data[MsvcOrgName]))

		// STEP: influxdb bucket exists
		influxClient := libInflux.NewClient(msvcUri, msvcToken)
		if err := influxClient.Connect(ctx); err != nil {
			return req.FailWithStatusError(err)
		}
		defer influxClient.Close()

		func() {
			bucketId, ok := bucketObj.Status.DisplayVars.GetString(BucketIdKey)
			if !ok {
				isReady = false
				cs = append(cs, conditions.New(HasBucketId, false, conditions.NotFound))
				return
			}
			cs = append(cs, conditions.New(HasBucketId, true, conditions.Found))
			if err := influxClient.BucketExists(ctx, bucketId); err != nil {
				cs = append(cs, conditions.New(BucketExists, false, conditions.NotFound, err.Error()))
				isReady = false
				return
			}
			cs = append(cs, conditions.New(BucketExists, true, conditions.Found))
		}()
	}

	newConditions, updated, err := conditions.Patch(bucketObj.Status.Conditions, cs)
	if err != nil {
		return req.FailWithStatusError(err)
	}

	if !updated && isReady == bucketObj.Status.IsReady {
		return req.Next()
	}

	bucketObj.Status.IsReady = isReady
	bucketObj.Status.Conditions = newConditions
	bucketObj.Status.OpsConditions = []metav1.Condition{}
	return rApi.NewStepResult(&ctrl.Result{}, r.Status().Update(ctx, bucketObj))
}

func (r *BucketReconciler) reconcileOperations(req *rApi.Request[*influxDB.Bucket]) rApi.StepResult {
	ctx := req.Context()
	bucketObj := req.Object

	if !controllerutil.ContainsFinalizer(bucketObj, constants.CommonFinalizer) {
		controllerutil.AddFinalizer(bucketObj, constants.CommonFinalizer)
		controllerutil.AddFinalizer(bucketObj, constants.ForegroundFinalizer)

		if err := r.Update(ctx, bucketObj); err != nil {
			return req.FailWithStatusError(err)
		}
		return req.Next()
	}

	msvcUri, ok := rApi.GetLocal[string](req, MsvcUri)
	if !ok {
		return req.FailWithOpError(errors.Newf("key=%s not present in req.locals", MsvcUri))
	}
	msvcToken, ok := rApi.GetLocal[string](req, MsvcTokenKey)
	if !ok {
		return req.FailWithOpError(errors.Newf("key=%s not present in req.locals", msvcToken))
	}
	msvcOrgName, ok := rApi.GetLocal[string](req, MsvcOrgName)
	if !ok {
		return req.FailWithOpError(errors.Newf("key=%s not present in req.locals", MsvcOrgName))
	}

	influxClient := libInflux.NewClient(msvcUri, msvcToken)
	defer influxClient.Close()

	bucket, err := influxClient.UpsertBucket(ctx, msvcOrgName, bucketObj.Name)
	if err != nil {
		return req.FailWithOpError(err)
	}

	if meta.IsStatusConditionFalse(bucketObj.Status.Conditions, HasBucketId.String()) {
		if err := bucketObj.Status.DisplayVars.Set(BucketIdKey, bucket.BucketId); err != nil {
			return req.FailWithOpError(err)
		}

		return rApi.NewStepResult(&ctrl.Result{}, r.Status().Update(ctx, bucketObj))
	}

	if err := func() error {
		b, err := templates.Parse(
			templates.Secret, &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: bucketObj.Namespace,
					Name:      fmt.Sprintf("mres-%s", bucketObj.Name),
					OwnerReferences: []metav1.OwnerReference{
						fn.AsOwner(bucketObj, true),
					},
				},
				StringData: map[string]string{
					"BUCKET_NAME": bucketObj.Name,
					"BUCKET_ID":   bucket.BucketId,
					"ORG_ID":      bucket.OrgId,
					"ORG_NAME":    msvcOrgName,
					"TOKEN":       msvcToken,
					"URI":         msvcUri,
				},
			},
		)
		if err != nil {
			return err
		}

		if _, err := fn.KubectlApplyExec(b); err != nil {
			return err
		}
		return nil
	}(); err != nil {
		return req.FailWithOpError(err)
	}

	return req.Done()
}

// SetupWithManager sets up the controller with the Manager.
func (r *BucketReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&influxDB.Bucket{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
