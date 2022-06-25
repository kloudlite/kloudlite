package s3aws

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"operators.kloudlite.io/lib/aws"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/constants"
	fn "operators.kloudlite.io/lib/functions"
	rApi "operators.kloudlite.io/lib/operator"
	"operators.kloudlite.io/lib/templates"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	s3awsv1 "operators.kloudlite.io/apis/s3.aws/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// BucketReconciler reconciles a Bucket object
type BucketReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const (
	KeyBucketName string = "KeyBucketName"
)

// +kubebuilder:rbac:groups=s3.aws.kloudlite.io,resources=buckets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=s3.aws.kloudlite.io,resources=buckets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=s3.aws.kloudlite.io,resources=buckets/finalizers,verbs=update

func (r *BucketReconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(ctx, r.Client, oReq.NamespacedName, &s3awsv1.Bucket{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.Result(), x.Err()
		}
	}

	req.Logger.Info("----------------[Type: s3awsv1.Bucket] NEW RECONCILATION ----------------")

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

func (r *BucketReconciler) finalize(req *rApi.Request[*s3awsv1.Bucket]) rApi.StepResult {
	obj := req.Object

	s3Client, err := aws.NewS3Client("ap-south-1")
	if err != nil {
		return req.FailWithOpError(err)
	}

	bucketName, ok := obj.Status.GeneratedVars.GetString(KeyBucketName)
	if !ok {
		return req.FailWithOpError(err)
	}

	if err := s3Client.DeleteBucket(bucketName); err != nil {
		return req.FailWithOpError(err)
	}

	iamClient, err := aws.NewIAMClient()
	if err != nil {
		return req.FailWithOpError(err)
	}

	if err := iamClient.DeleteUser(bucketName); err != nil {
		return req.FailWithOpError(err)
	}

	return req.Finalize()
}

func (r *BucketReconciler) reconcileStatus(req *rApi.Request[*s3awsv1.Bucket]) rApi.StepResult {
	ctx := req.Context()
	obj := req.Object

	isReady := true
	var cs []metav1.Condition

	// STEP: 4. check generated vars
	if !obj.Status.GeneratedVars.Exists(KeyBucketName) {
		cs = append(cs, conditions.New(conditions.GeneratedVars, false, conditions.NotReconciledYet))
	} else {
		isReady = false
		cs = append(cs, conditions.New(conditions.GeneratedVars, true, conditions.Found))
	}

	// STEP: 5. reconciler output exists
	_, err5 := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, fmt.Sprintf("mres-%s", obj.Name)), &corev1.Secret{})
	if err5 != nil {
		cs = append(cs, conditions.New(conditions.ReconcilerOutputExists, false, conditions.NotFound, err5.Error()))
	} else {
		cs = append(cs, conditions.New(conditions.ReconcilerOutputExists, true, conditions.Found))
	}

	// STEP: 6. patch conditions
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

func (r *BucketReconciler) reconcileOperations(req *rApi.Request[*s3awsv1.Bucket]) rApi.StepResult {
	ctx := req.Context()
	obj := req.Object

	// STEP: 1. add finalizers if needed
	if !controllerutil.ContainsFinalizer(obj, constants.CommonFinalizer) {
		controllerutil.AddFinalizer(obj, constants.CommonFinalizer)
		controllerutil.AddFinalizer(obj, constants.ForegroundFinalizer)

		return rApi.NewStepResult(&ctrl.Result{}, r.Update(ctx, obj))
	}

	if meta.IsStatusConditionFalse(obj.Status.Conditions, conditions.GeneratedVars.String()) {
		if err := obj.Status.GeneratedVars.Set(
			KeyBucketName,
			fmt.Sprintf("%s-%s", obj.Name, strings.ToLower(fn.CleanerNanoid(40))),
		); err != nil {
			return nil
		}
		return rApi.NewStepResult(&ctrl.Result{}, r.Status().Update(ctx, obj))
	}

	// STEP: 4. create child components like mongo-user, redis-acl etc.
	bucketName, ok := obj.Status.GeneratedVars.GetString(KeyBucketName)
	if !ok {
		return req.FailWithOpError(rApi.ErrNotInGeneratedVars.Format(KeyBucketName))
	}

	accKeyId, secretAccKey, err4 := func() (string, string, error) {
		s3Client, err := aws.NewS3Client(os.Getenv("AWS_REGION"))
		if err != nil {
			return "", "", err
		}
		if err := s3Client.CreateBucket(bucketName); err != nil {
			return "", "", err
		}
		iamClient, err := aws.NewIAMClient()
		if err != nil {
			return "", "", err
		}

		if err := iamClient.CreateUser(bucketName); err != nil {
			return "", "", err
		}

		if meta.IsStatusConditionFalse(obj.Status.Conditions, conditions.ReconcilerOutputExists.String()) {
			key, secretKey, err := iamClient.CreateAccessKey(bucketName)
			if err != nil {
				return "", "", err
			}
			return key, secretKey, nil
		}
		if err := s3Client.MakePublicReadable(bucketName); err != nil {
			return "", "", err
		}

		if err := s3Client.MakeObjectsPublic(bucketName, "public"); err != nil {
			return "", "", err
		}

		return "", "", nil
	}()
	if err4 != nil {
		// TODO:(user) might need to reconcile with retry with timeout error
		return req.FailWithOpError(err4)
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
					"AWS_ACCESS_KEY_ID":     accKeyId,
					"AWS_SECRET_ACCESS_KEY": secretAccKey,
					"AWS_REGION":            "ap-south-1",
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

func (r *BucketReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).For(&s3awsv1.Bucket{}).Complete(r)
}
