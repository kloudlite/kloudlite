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
	MsvcOutputKey   = "msvc-output"
	MsvcOutputHosts = "DB_HOSTS"
	MsvcOutputURL   = "DB_URL"
)

const (
	MongoUserExists conditions.Type = "MongoUserExists"
)

const (
	DbPasswordKey string = "db-password"
)

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

	req.Logger.Info("----------------database reconciler -- NEW RECONCILATION------------------")

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
	databaseObj := req.Object

	isReady := true
	var cs []metav1.Condition

	// STEP: check managed service is ready
	msvc, err := rApi.Get(
		ctx, r.Client, fn.NN(databaseObj.Namespace, databaseObj.Spec.ManagedSvcName),
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
	}

	// STEP: retrieve managed svc output (usually secret)
	if msvc != nil {
		msvcOutput, err := rApi.Get(
			ctx, r.Client, fn.NN(msvc.Namespace, fmt.Sprintf("msvc-%s", msvc.Name)),
			&corev1.Secret{},
		)
		if err != nil {
			isReady = false
			if !apiErrors.IsNotFound(err) {
				return req.FailWithStatusError(err)
			}
			cs = append(cs, conditions.New(conditions.ManagedSvcOutputExists, false, conditions.NotFound, err.Error()))
		} else {
			cs = append(cs, conditions.New(conditions.ManagedSvcOutputExists, true, conditions.Found))
			rApi.SetLocal(req, MsvcOutputKey, msvcOutput)
		}

		// STEP: check reconciler (child components e.g. mongo account, s3 bucket, redis ACL user) exists
		mc, err := libMongo.NewClient(string(msvcOutput.Data["DB_URL"]))
		if err != nil {
			return req.FailWithStatusError(err)
		}
		if err := mc.Connect(ctx); err != nil {
			return req.FailWithStatusError(err)
		}
		defer mc.Close()

		func() {
			userExists, err := mc.UserExists(ctx, databaseObj.Name)
			if err != nil {
				cs = append(cs, conditions.New(MongoUserExists, false, conditions.NotFound, err.Error()))
				return
			}
			if !userExists {
				cs = append(cs, conditions.New(MongoUserExists, false, conditions.NotFound))
				return
			}
			cs = append(cs, conditions.New(MongoUserExists, true, conditions.Found))
		}()
	}

	// STEP: check generated vars
	if msvc != nil && !databaseObj.Status.GeneratedVars.Exists(DbPasswordKey) {
		cs = append(cs, conditions.New(conditions.GeneratedVars, false, conditions.NotReconciledYet))
	} else {
		cs = append(cs, conditions.New(conditions.GeneratedVars, true, conditions.Found))
	}

	// STEP: patch conditions
	newConditions, updated, err := conditions.Patch(databaseObj.Status.Conditions, cs)
	if err != nil {
		return req.FailWithStatusError(err)
	}

	if !updated && isReady == databaseObj.Status.IsReady {
		return req.Next()
	}

	databaseObj.Status.IsReady = isReady
	databaseObj.Status.Conditions = newConditions
	databaseObj.Status.OpsConditions = []metav1.Condition{}
	return rApi.NewStepResult(&ctrl.Result{}, r.Status().Update(ctx, databaseObj))
}

func (r *DatabaseReconciler) reconcileOperations(req *rApi.Request[*mongodbStandalone.Database]) rApi.StepResult {
	ctx := req.Context()
	databaseObj := req.Object

	// STEP: 1. add finalizers if needed
	if !controllerutil.ContainsFinalizer(databaseObj, constants.CommonFinalizer) {
		controllerutil.AddFinalizer(databaseObj, constants.CommonFinalizer)
		controllerutil.AddFinalizer(databaseObj, constants.ForegroundFinalizer)

		return rApi.NewStepResult(&ctrl.Result{}, r.Update(ctx, databaseObj))
	}

	// STEP: 2. generate vars if needed to
	if meta.IsStatusConditionFalse(databaseObj.Status.Conditions, conditions.GeneratedVars.String()) {
		if err := databaseObj.Status.GeneratedVars.Set(DbPasswordKey, fn.CleanerNanoid(40)); err != nil {
			return req.FailWithStatusError(err)
		}
		return rApi.NewStepResult(&ctrl.Result{}, r.Status().Update(ctx, databaseObj))
	}

	// STEP: 3. retrieve msvc output, need it in creating reconciler output
	msvcOutput, ok := rApi.GetLocal[corev1.Secret](req, MsvcOutputKey)
	if !ok {
		return req.FailWithOpError(errors.Newf("err=%s key not found in req locals", MsvcOutputKey))
	}

	// STEP: 4. create child components like mongo-user, redis-acl etc.
	mc, err := libMongo.NewClient(string(msvcOutput.Data[MsvcOutputURL]))
	if err != nil {
		return req.FailWithStatusError(err)
	}
	if err := mc.Connect(ctx); err != nil {
		return req.FailWithStatusError(err)
	}
	defer mc.Close()

	dbPasswd, ok := databaseObj.Status.GeneratedVars.GetString(DbPasswordKey)
	if !ok {
		return req.FailWithOpError(errors.Newf("key %s not found in GeneratedVars", DbPasswordKey))
	}
	if err := mc.UpsertUser(ctx, databaseObj.Name, databaseObj.Name, dbPasswd); err != nil {
		return req.FailWithOpError(err)
	}

	// STEP: 5. create reconciler output (eg. secret)
	if errt := func() error {
		b, err := templates.Parse(
			templates.Secret, &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("mres-%s", databaseObj.Name),
					Namespace: databaseObj.Namespace,
					OwnerReferences: []metav1.OwnerReference{
						fn.AsOwner(databaseObj, true),
					},
				},
				StringData: map[string]string{
					"DB_PASSWORD": dbPasswd,
					"DB_USER":     databaseObj.Name,
					"DB_HOSTS":    string(msvcOutput.Data[MsvcOutputHosts]),
					"DB_URL": fmt.Sprintf(
						"mongodb://%s:%s@%s/%s",
						databaseObj.Name, dbPasswd, string(msvcOutput.Data[MsvcOutputHosts]), databaseObj.Name,
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

func (r *DatabaseReconciler) reconcileOperations2(req *rApi.Request[*mongodbStandalone.Database]) rApi.StepResult {
	ctx := req.Context()
	databaseObj := req.Object

	if !controllerutil.ContainsFinalizer(databaseObj, constants.CommonFinalizer) {
		controllerutil.AddFinalizer(databaseObj, constants.CommonFinalizer)
		controllerutil.AddFinalizer(databaseObj, constants.ForegroundFinalizer)

		return rApi.NewStepResult(&ctrl.Result{}, r.Update(ctx, databaseObj))
	}

	if meta.IsStatusConditionFalse(databaseObj.Status.Conditions, conditions.GeneratedVars.String()) {
		if err := databaseObj.Status.GeneratedVars.Set(DbPasswordKey, fn.CleanerNanoid(40)); err != nil {
			return req.FailWithStatusError(err)
		}
		return rApi.NewStepResult(&ctrl.Result{}, r.Status().Update(ctx, databaseObj))
	}

	msvcOutput, ok := rApi.GetLocal[corev1.Secret](req, MsvcOutputKey)
	if !ok {
		return req.FailWithOpError(errors.Newf("err=%s key not found in req locals"))
	}

	mc, err := libMongo.NewClient(string(msvcOutput.Data[MsvcOutputURL]))
	if err != nil {
		return req.FailWithStatusError(err)
	}
	if err := mc.Connect(ctx); err != nil {
		return req.FailWithStatusError(err)
	}
	defer mc.Close()

	dbPasswd, ok := databaseObj.Status.GeneratedVars.GetString(DbPasswordKey)
	if !ok {
		return req.FailWithOpError(errors.Newf("key %s not found in GeneratedVars", DbPasswordKey))
	}
	if err := mc.UpsertUser(ctx, databaseObj.Name, databaseObj.Name, dbPasswd); err != nil {
		return req.FailWithOpError(err)
	}

	outScrt := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: databaseObj.Namespace,
			Name:      fmt.Sprintf("mres-%s", databaseObj.Name),
			OwnerReferences: []metav1.OwnerReference{
				fn.AsOwner(databaseObj, true),
			},
		},
		StringData: map[string]string{
			"DB_PASSWORD": dbPasswd,
			"DB_USER":     databaseObj.Name,
			"DB_HOSTS":    string(msvcOutput.Data[MsvcOutputHosts]),
			"DB_URL": fmt.Sprintf(
				"mongodb://%s:%s@%s/%s",
				databaseObj.Name, dbPasswd, string(msvcOutput.Data[MsvcOutputHosts]), databaseObj.Name,
			),
		},
	}

	if err := fn.KubectlApply(ctx, r.Client, fn.ParseSecret(outScrt)); err != nil {
		return req.FailWithOpError(err)
	}

	return req.Done()
}

// SetupWithManager sets up the controller with the Manager.
func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mongodbStandalone.Database{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
