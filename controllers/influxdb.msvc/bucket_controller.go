package influxdbmsvc

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	influxDB "operators.kloudlite.io/apis/influxdb.msvc/v1"
	"operators.kloudlite.io/env"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/errors"
	fn "operators.kloudlite.io/lib/functions"
	libInflux "operators.kloudlite.io/lib/influx"
	"operators.kloudlite.io/lib/logging"
	rApi "operators.kloudlite.io/lib/operator"
	stepResult "operators.kloudlite.io/lib/operator/step-result"
	"operators.kloudlite.io/lib/templates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// BucketReconciler reconciles a Bucket object
type BucketReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	logger logging.Logger
}

func (r *BucketReconciler) GetName() string {
	return "influxdb-bucket"
}

const (
	BucketIdKey = "bucket-id"
)

const (
	BucketExists conditions.Type = "BucketExists"
	HasBucketId  conditions.Type = "HasBucketId"
)

type MsvcOutputRef struct {
	Token   string
	Uri     string
	OrgName string
}

func parseMsvcOutput(s *corev1.Secret) *MsvcOutputRef {
	return &MsvcOutputRef{
		Token:   string(s.Data["TOKEN"]),
		Uri:     string(s.Data["URI"]),
		OrgName: string(s.Data["ORG"]),
	}
}

// +kubebuilder:rbac:groups=influxdb.msvc.kloudlite.io,resources=buckets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=influxdb.msvc.kloudlite.io,resources=buckets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=influxdb.msvc.kloudlite.io,resources=buckets/finalizers,verbs=update

func (r *BucketReconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, oReq.NamespacedName, &influxDB.Bucket{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
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

func (r *BucketReconciler) finalize(req *rApi.Request[*influxDB.Bucket]) stepResult.Result {
	return req.Finalize()
}

func (r *BucketReconciler) reconcileStatus(req *rApi.Request[*influxDB.Bucket]) stepResult.Result {
	ctx := req.Context()
	obj := req.Object

	isReady := true
	var cs []metav1.Condition

	// STEP: 1. check managed service is ready
	msvc, err := rApi.Get(
		ctx, r.Client, fn.NN(obj.Namespace, obj.Spec.ManagedSvcName),
		&influxDB.Service{},
	)

	if err != nil {
		isReady = false
		cs = append(cs, conditions.New(conditions.ManagedSvcExists, false, conditions.NotFound, err.Error()))
		return req.FailWithStatusError(err, cs...).Err(nil)
	}
	cs = append(cs, conditions.New(conditions.ManagedSvcExists, true, conditions.Found))
	cs = append(cs, conditions.New(conditions.ManagedSvcReady, msvc.Status.IsReady, conditions.Empty))
	if !msvc.Status.IsReady {
		isReady = false
		return req.FailWithStatusError(err, cs...).Err(nil)
	}

	// STEP: 2. retrieve managed svc output (usually secret)
	msvcRef, err2 := func() (*MsvcOutputRef, error) {
		msvcOutput, err := rApi.Get(
			ctx, r.Client, fn.NN(msvc.Namespace, fmt.Sprintf("msvc-%s", msvc.Name)),
			&corev1.Secret{},
		)
		if err != nil {
			isReady = false
			cs = append(cs, conditions.New(conditions.ManagedSvcOutputExists, false, conditions.NotFound, err.Error()))
			return nil, err
		}
		cs = append(cs, conditions.New(conditions.ManagedSvcOutputExists, true, conditions.Found))
		outputRef := parseMsvcOutput(msvcOutput)
		rApi.SetLocal(req, "msvc-output-ref", outputRef)
		return outputRef, nil
	}()
	if err2 != nil {
		return req.FailWithStatusError(err2)
	}

	if err2 := func() error {
		// STEP: 3. check reconciler (child components e.g. mongo account, s3 bucket, redis ACL user) exists
		// TODO: (user) use msvcRef values

		influxClient := libInflux.NewClient(msvcRef.Uri, msvcRef.Token)
		if err := influxClient.Connect(ctx); err != nil {
			return err
		}
		defer influxClient.Close()
		bucketId, ok := obj.Status.DisplayVars.GetString(BucketIdKey)
		if !ok {
			isReady = false
			cs = append(cs, conditions.New(HasBucketId, false, conditions.NotFound))
			return nil
		}
		cs = append(cs, conditions.New(HasBucketId, true, conditions.Found))
		if err := influxClient.BucketExists(ctx, bucketId); err != nil {
			cs = append(cs, conditions.New(BucketExists, false, conditions.NotFound, err.Error()))
			isReady = false
			return nil
		}
		return nil
	}(); err2 != nil {
		return req.FailWithStatusError(err2)
	}

	// STEP: 5. patch conditions
	newConditions, updated, err := conditions.Patch(obj.Status.Conditions, cs)
	if err != nil {
		return req.FailWithStatusError(err)
	}

	if !updated && isReady == obj.Status.IsReady {
		return req.Next()
	}

	obj.Status.IsReady = isReady
	obj.Status.Conditions = newConditions
	if err := r.Status().Update(ctx, obj); err != nil {
		return req.FailWithStatusError(err)
	}
	return req.Done()

}

func (r *BucketReconciler) reconcileOperations(req *rApi.Request[*influxDB.Bucket]) stepResult.Result {
	ctx := req.Context()
	obj := req.Object

	// STEP: 1. add finalizers if needed
	if !controllerutil.ContainsFinalizer(obj, constants.CommonFinalizer) {
		controllerutil.AddFinalizer(obj, constants.CommonFinalizer)
		controllerutil.AddFinalizer(obj, constants.ForegroundFinalizer)

		if err := r.Update(ctx, obj); err != nil {
			return req.FailWithOpError(err)
		}
		return req.Done()
	}

	// STEP: 3. retrieve msvc output, need it in creating reconciler output
	msvcRef, ok := rApi.GetLocal[*MsvcOutputRef](req, "msvc-output-ref")
	if !ok {
		return req.FailWithOpError(errors.Newf("err=%s key not found in req locals", "msvc-output-ref"))
	}

	// STEP: 4. create child components like mongo-user, redis-acl etc.

	influxClient := libInflux.NewClient(msvcRef.Uri, msvcRef.Token)
	defer influxClient.Close()
	bucket, err := influxClient.UpsertBucket(ctx, msvcRef.OrgName, obj.Name)
	if err != nil {
		return req.FailWithOpError(err)
	}

	if meta.IsStatusConditionFalse(obj.Status.Conditions, HasBucketId.String()) {
		if err := obj.Status.DisplayVars.Set(BucketIdKey, bucket.BucketId); err != nil {
			return req.FailWithOpError(err)
		}

		if err := r.Status().Update(ctx, obj); err != nil {
			return req.FailWithOpError(err)
		}
		return req.Done()
	}

	// STEP: 5. create reconciler output (eg. secret)
	if errt := func() error {
		b, err := templates.Parse(
			templates.Secret, &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("mres-%s", obj.Name),
					Namespace: obj.Namespace,
					OwnerReferences: []metav1.OwnerReference{
						fn.AsOwner(obj, true),
					},
				},
				StringData: map[string]string{
					"BUCKET_NAME": obj.Name,
					"BUCKET_ID":   bucket.BucketId,
					"ORG_ID":      bucket.OrgId,
					"ORG_NAME":    msvcRef.OrgName,
					"TOKEN":       msvcRef.Token,
					"URI":         msvcRef.Uri,
				},
			},
		)
		if err != nil {
			req.Logger.Errorf(err, "failed parsing template %s", templates.Secret)
			return nil
		}
		if err := fn.KubectlApplyExec(ctx, b); err != nil {
			req.Logger.Errorf(err, "failed kubectl apply template %s", templates.Secret)
			return nil
		}
		return nil
	}(); errt != nil {
		return req.FailWithOpError(errt)
	}

	return req.Next()
}

// SetupWithManager sets up the controller with the Manager.
func (r *BucketReconciler) SetupWithManager(mgr ctrl.Manager, envVars *env.Env, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()

	r.logger = logger.WithName("influx-standalone-bucket")
	return ctrl.NewControllerManagedBy(mgr).
		For(&influxDB.Bucket{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
