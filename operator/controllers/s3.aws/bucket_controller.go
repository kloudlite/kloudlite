package s3aws

import (
	"context"
	"fmt"
	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	"operators.kloudlite.io/env"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"operators.kloudlite.io/lib/aws"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/constants"
	fn "operators.kloudlite.io/lib/functions"
	rApi "operators.kloudlite.io/lib/operator"
	"operators.kloudlite.io/lib/templates"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"k8s.io/apimachinery/pkg/runtime"
	s3awsv1 "operators.kloudlite.io/apis/s3.aws/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// BucketReconciler reconciles a Bucket object
type BucketReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Env    *env.Env
}

const (
	KeyBucketName   string = "KeyBucketName"
	KeyAccessSecret string = "KeyAccessSecret"

	KeyAccessKeyId     string = "AWS_ACCESS_KEY_ID"
	KeySecretAccessKey string = "AWS_SECRET_ACCESS_KEY"
)

type Credentials struct {
	AccessKeyId     string
	SecretAccessKey string
}

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

	req.Logger.Infof("----------------[Type: s3awsv1.Bucket] NEW RECONCILATION ----------------")

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

func (r *BucketReconciler) finalize(req *rApi.Request[*s3awsv1.Bucket]) rApi.StepResult {
	obj := req.Object

	s3Client, err := aws.NewS3Client(obj.Spec.Region)
	if err != nil {
		return req.FailWithOpError(err)
	}

	bucketName, ok := obj.Status.GeneratedVars.GetString(KeyBucketName)
	if !ok {
		return req.FailWithOpError(err)
	}

	if err := s3Client.EmptyBucket(bucketName); err != nil {
		return nil
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
	reconOutput, err5 := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, fmt.Sprintf("mres-%s", obj.Name)), &corev1.Secret{})
	if err5 != nil {
		cs = append(cs, conditions.New(conditions.ReconcilerOutputExists, false, conditions.NotFound, err5.Error()))
	} else {
		rApi.SetLocal(
			req, KeyAccessSecret, &Credentials{
				AccessKeyId:     string(reconOutput.Data[KeyAccessKeyId]),
				SecretAccessKey: string(reconOutput.Data[KeySecretAccessKey]),
			},
		)
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

	s3Client, err := aws.NewS3Client(obj.Spec.Region)
	if err != nil {
		return req.FailWithOpError(err)
	}

	iamClient, err := aws.NewIAMClient()
	if err != nil {
		return req.FailWithOpError(err)
	}

	accessCreds, err4 := func() (*Credentials, error) {
		if err := s3Client.CreateBucket(bucketName); err != nil {
			return nil, err
		}
		if _, err := iamClient.CreateUser(bucketName); err != nil {
			return nil, err
		}

		if meta.IsStatusConditionTrue(obj.Status.Conditions, conditions.ReconcilerOutputExists.String()) {
			creds, ok := rApi.GetLocal[*Credentials](req, KeyAccessSecret)
			if !ok {
				return nil, err
			}
			return creds, nil
		}

		if err := iamClient.DeleteAccessKey(bucketName); err != nil {
			return nil, err
		}
		keyId, secretKey, err := iamClient.CreateAccessKey(bucketName)
		if err != nil {
			return nil, err
		}
		return &Credentials{
			AccessKeyId:     keyId,
			SecretAccessKey: secretKey,
		}, nil
	}()
	if err4 != nil {
		// TODO:(user) might need to reconcile with retry with timeout error
		return req.FailWithOpError(err4)
	}

	if err5 := func() error {
		user, err := iamClient.GetUser(bucketName)
		if err != nil {
			return err
		}

		policies, err := s3Client.AddOwnerPolicy(bucketName, user)
		if err != nil {
			return err
		}

		publicObjectPolicies, err := s3Client.MakeObjectsDirsPublic(bucketName, obj.Spec.PublicFolders...)
		if err != nil {
			return err
		}
		policies = append(policies, publicObjectPolicies...)
		return s3Client.ApplyPolicies(bucketName, policies...)
	}(); err5 != nil {
		return req.FailWithOpError(err5)
	}

	// STEP: 5. create reconciler output (eg. secret)
	if errt := func() error {
		svcExternalName := fmt.Sprintf("%s.s3.%s.amazonaws.com", bucketName, obj.Spec.Region)

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
					"AWS_ACCESS_KEY_ID":     accessCreds.AccessKeyId,
					"AWS_SECRET_ACCESS_KEY": accessCreds.SecretAccessKey,
					"AWS_REGION":            obj.Spec.Region,
					"INTERNAL_BUCKET_HOST":  fmt.Sprintf("%s.%s.svc.cluster.local", obj.Name, obj.Namespace),
					"EXTERNAL_BUCKET_HOST":  svcExternalName,
				},
			},
		)
		if err != nil {
			return err
		}

		// create external name service

		b2, err := templates.Parse(
			templates.CoreV1.ExternalNameSvc, map[string]any{
				"name":      obj.Name,
				"namespace": obj.Namespace,
				"owner-refs": []metav1.OwnerReference{
					fn.AsOwner(obj, true),
				},
				"external-name": svcExternalName,
			},
		)
		if err != nil {
			return err
		}

		b3, err := templates.Parse(
			templates.CoreV1.Ingress, map[string]any{
				"ingress-class":  constants.DefaultIngressClass,
				"cluster-issuer": constants.DefaultClusterIssuer,
				"owner-refs": []metav1.OwnerReference{
					fn.AsOwner(obj, true),
				},

				"virtual-hostname": bucketName,

				"name":      obj.Name,
				"namespace": obj.Namespace,
				"router-ref": crdsv1.Router{
					ObjectMeta: metav1.ObjectMeta{
						Name:      obj.Name,
						Namespace: obj.Namespace,
					},
					Spec: crdsv1.RouterSpec{
						Https: crdsv1.Https{
							Enabled:       true,
							ForceRedirect: true,
						},
						Domains: []string{
							fmt.Sprintf("%s.s3.dev.kloudlite.io", obj.Name),
						},
					},
				},
				"wildcard-domain-suffix":      r.Env.WildcardDomainSuffix,
				"wildcard-domain-certificate": r.Env.WildcardDomainCertificate,

				"routes": map[string]crdsv1.Route{
					"/": {
						App:  obj.Name,
						Port: 443,
					},
				},
				"annotations": map[string]string{
					"nginx.ingress.kubernetes.io/backend-protocol": "https",
				},
			},
		)
		if err != nil {
			return err
		}

		if _, err := fn.KubectlApplyExec(b, b2, b3); err != nil {
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
