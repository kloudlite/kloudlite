package mongodbstandalonemsvc

import (
	"context"
	"fmt"
	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/templates"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mongodbStandalone "operators.kloudlite.io/apis/mongodb-standalone.msvc/v1"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/errors"
	fn "operators.kloudlite.io/lib/functions"
	libMongo "operators.kloudlite.io/lib/mongo"
	rApi "operators.kloudlite.io/lib/operator"
)

// DatabaseReconciler reconciles a Database object
type DatabaseReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const (
	MongoUserExists conditions.Type = "MongoUserExists"
)

const (
	DbPasswordKey string = "db-password"
)

type MsvcOutputRef struct {
	Hosts string
	Uri   string
}

func parseMsvcOutput(s *corev1.Secret) *MsvcOutputRef {
	return &MsvcOutputRef{
		Hosts: string(s.Data["DB_HOSTS"]),
		Uri:   string(s.Data["DB_URL"]),
	}
}

// +kubebuilder:rbac:groups=mongodb-standalone.msvc.kloudlite.io,resources=databases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mongodb-standalone.msvc.kloudlite.io,resources=databases/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mongodb-standalone.msvc.kloudlite.io,resources=databases/finalizers,verbs=update

func (r *DatabaseReconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(ctx, r.Client, oReq.NamespacedName, &mongodbStandalone.Database{})
	if err != nil {
		return ctrl.Result{}, err
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.Result(), x.Err()
		}
	}

	req.Logger.Infof("----------------database reconciler -- NEW RECONCILATION------------------")

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

func (r *DatabaseReconciler) finalize(req *rApi.Request[*mongodbStandalone.Database]) rApi.StepResult {
	return req.Finalize()
}

func (r *DatabaseReconciler) reconcileStatus(req *rApi.Request[*mongodbStandalone.Database]) rApi.StepResult {
	ctx := req.Context()
	obj := req.Object

	isReady := true
	var cs []metav1.Condition

	// STEP: 1. check managed service is ready
	msvc, err := rApi.Get(
		ctx, r.Client, fn.NN(obj.Namespace, obj.Spec.ManagedSvcName),
		&mongodbStandalone.Service{},
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
			mc, err := libMongo.NewClient(msvcRef.Uri)
			if err != nil {
				return err
			}
			if err := mc.Connect(ctx); err != nil {
				return err
			}
			defer mc.Close()

			userExists, err := mc.UserExists(ctx, obj.Name)
			if err != nil {
				cs = append(cs, conditions.New(MongoUserExists, false, conditions.NotFound, err.Error()))
				return err
			}
			if !userExists {
				cs = append(cs, conditions.New(MongoUserExists, false, conditions.NotFound))
				return nil
			}
			cs = append(cs, conditions.New(MongoUserExists, true, conditions.Found))
			return nil
		}(); err2 != nil {
			// TODO: (user) might need to reconcile with retry in case of connection errors
			return req.FailWithStatusError(err2)
		}
	}

	// STEP: 4. check generated vars
	if msvc != nil && !obj.Status.GeneratedVars.Exists(DbPasswordKey) {
		cs = append(cs, conditions.New(conditions.GeneratedVars, false, conditions.NotReconciledYet))
	} else {
		cs = append(cs, conditions.New(conditions.GeneratedVars, true, conditions.Found))
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

func (r *DatabaseReconciler) reconcileOperations(req *rApi.Request[*mongodbStandalone.Database]) rApi.StepResult {
	ctx := req.Context()
	obj := req.Object

	// STEP: 1. add finalizers if needed
	if !controllerutil.ContainsFinalizer(obj, constants.CommonFinalizer) {
		controllerutil.AddFinalizer(obj, constants.CommonFinalizer)
		controllerutil.AddFinalizer(obj, constants.ForegroundFinalizer)

		return rApi.NewStepResult(&ctrl.Result{}, r.Update(ctx, obj))
	}

	// STEP: 2. generate vars if needed to
	if meta.IsStatusConditionFalse(obj.Status.Conditions, conditions.GeneratedVars.String()) {
		if err := obj.Status.GeneratedVars.Set(DbPasswordKey, fn.CleanerNanoid(40)); err != nil {
			return req.FailWithStatusError(err)
		}
		return rApi.NewStepResult(&ctrl.Result{}, r.Status().Update(ctx, obj))
	}

	// STEP: 3. retrieve msvc output, need it in creating reconciler output
	msvcRef, ok := rApi.GetLocal[*MsvcOutputRef](req, "msvc-output-ref")
	if !ok {
		return req.FailWithOpError(errors.Newf("err=%s key not found in req locals", "msvc-output-ref"))
	}

	// STEP: 4. create child components like mongo-user, redis-acl etc.
	dbPasswd, ok := obj.Status.GeneratedVars.GetString(DbPasswordKey)
	if !ok {
		return req.FailWithOpError(errors.Newf("key %s not found in GeneratedVars", DbPasswordKey))
	}
	err4 := func() error {
		mc, err := libMongo.NewClient(msvcRef.Uri)
		if err != nil {
			return err
		}
		if err := mc.Connect(ctx); err != nil {
			return err
		}
		defer mc.Close()

		return mc.UpsertUser(ctx, obj.Name, obj.Name, dbPasswd)
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
					"DB_PASSWORD": dbPasswd,
					"DB_USER":     obj.Name,
					"DB_HOSTS":    msvcRef.Hosts,
					"DB_NAME":     obj.Name,
					"DB_URL": fmt.Sprintf(
						"mongodb://%s:%s@%s/%s",
						obj.Name, dbPasswd, msvcRef.Hosts, obj.Name,
					),
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

func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mongodbStandalone.Database{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
