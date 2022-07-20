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
	req, err := rApi.NewRequest(ctx, r.Client, oReq.NamespacedName, &influxDB.Bucket{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.Result(), x.Err()
		}
	}

	req.Logger.Infof("-------------------- NEW RECONCILATION------------------")

	if x := req.EnsureLabelsAndAnnotations(); !x.ShouldProceed() {
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
		msvc = nil
		if !apiErrors.IsNotFound(err) {
			return req.FailWithStatusError(err)
		}
		cs = append(cs, conditions.New(conditions.ManagedSvcExists, false, conditions.NotFound, err.Error()))
	} else {
		cs = append(cs, conditions.New(conditions.ManagedSvcExists, true, conditions.Found))
		cs = append(cs, conditions.New(conditions.ManagedSvcReady, msvc.Status.IsReady, conditions.Empty))
		if !msvc.Status.IsReady {
			isReady = false
			msvc = nil
		}
	}

	// STEP: 2. retrieve managed svc output (usually secret)
	if msvc != nil {
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
	return rApi.NewStepResult(&ctrl.Result{}, r.Status().Update(ctx, obj))

}

func (r *BucketReconciler) reconcileOperations(req *rApi.Request[*influxDB.Bucket]) rApi.StepResult {
	ctx := req.Context()
	obj := req.Object

	// STEP: 1. add finalizers if needed
	if !controllerutil.ContainsFinalizer(obj, constants.CommonFinalizer) {
		controllerutil.AddFinalizer(obj, constants.CommonFinalizer)
		controllerutil.AddFinalizer(obj, constants.ForegroundFinalizer)

		return rApi.NewStepResult(&ctrl.Result{}, r.Update(ctx, obj))
	}

	// STEP: 3. retrieve msvc output, need it in creating reconciler output
	msvcRef, ok := rApi.GetLocal[*MsvcOutputRef](req, "msvc-output-ref")
	if !ok {
		return req.FailWithOpError(errors.Newf("err=%s key not found in req locals", "msvc-output-ref"))
	}

	// STEP: 4. create child components like mongo-user, redis-acl etc.
	bucket, err4 := func() (*libInflux.Bucket, error) {
		influxClient := libInflux.NewClient(msvcRef.Uri, msvcRef.Token)
		defer influxClient.Close()
		return influxClient.UpsertBucket(ctx, msvcRef.OrgName, obj.Name)
	}()
	if err4 != nil {
		// TODO:(user) might need to reconcile with retry with timeout error
		return req.FailWithOpError(err4)
	}

	if meta.IsStatusConditionFalse(obj.Status.Conditions, HasBucketId.String()) {
		if err := obj.Status.DisplayVars.Set(BucketIdKey, bucket.BucketId); err != nil {
			return req.FailWithOpError(err)
		}

		return rApi.NewStepResult(&ctrl.Result{}, r.Status().Update(ctx, obj))
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
			return err
		}

		if _, err := fn.KubectlApplyExec(b); err != nil {
			return err
		}
		return nil
	}(); errt != nil {
		return req.FailWithOpError(errt)
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
