package mongodbstandalonemsvc

import (
	"context"
	"fmt"

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
	DbUser     string = "DB_USER"
	DbPassword string = "DB_PASSWORD"
	DbHosts    string = "DB_HOSTS"
	DbUrl      string = "DB_URL"
)

const (
	DbPasswordKey string = "db-password"
	RootUri       string = "DB_URL"
)

// +kubebuilder:rbac:groups=mongodb-standalone.msvc.kloudlite.io,resources=databases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mongodb-standalone.msvc.kloudlite.io,resources=databases/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mongodb-standalone.msvc.kloudlite.io,resources=databases/finalizers,verbs=update

func (r *DatabaseReconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	req := rApi.NewRequest(ctx, r.Client, oReq.NamespacedName, &mongodbStandalone.Database{})

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
		if !apiErrors.IsNotFound(err) {
			return req.FailWithStatusError(err)
		}
		isReady = false
		cs = append(cs, conditions.New("MsvcExists", false, "MsvcNotFound", err.Error()))
		msvc = nil
	}

	if msvc != nil {
		cs = append(cs, conditions.New("MsvcExists", true, "MsvcFound", ""))

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
			cs = append(cs, conditions.New("MsvcOutputExists", false, "SecretNotFound", err.Error()))
			isReady = false
		}

		if msvcOutput != nil {
			rootUri := string(msvcOutput.Data[RootUri])
			rApi.SetLocal(req, RootUri, rootUri)
			rApi.SetLocal(req, DbHosts, string(msvcOutput.Data[DbHosts]))

			// STEP: mongo user userExists
			mc, err := libMongo.NewClient(rootUri)
			if err != nil {
				return req.FailWithStatusError(err)
			}

			rApi.SetLocal(req, "MongoCli", mc)

			userExists, err := mc.UserExists(ctx, databaseObj.Name)
			if err != nil {
				return req.FailWithStatusError(err)
			}

			if !userExists {
				cs = append(cs, conditions.New("MongoUserExists", false, "NotFound"))
			}
			isReady = userExists
		}
	}

	// STEP: check generated vars
	if _, ok := databaseObj.Status.GeneratedVars.GetString(DbPasswordKey); !ok {
		cs = append(cs, conditions.New("GeneratedVars", false, "NotGeneratedYet"))
		isReady = false
	}

	newConditions, updated, err := conditions.Patch(databaseObj.Status.Conditions, cs)
	if err != nil {
		return req.FailWithStatusError(err)
	}

	if !updated && isReady == databaseObj.Status.IsReady {
		return req.FailWithStatusError(err)
	}

	databaseObj.Status.IsReady = isReady
	databaseObj.Status.Conditions = newConditions
	databaseObj.Status.OpsConditions = []metav1.Condition{}
	if err := r.Status().Update(ctx, databaseObj); err != nil {
		return req.FailWithStatusError(err)
	}
	return req.Done()
}

func (r *DatabaseReconciler) reconcileOperations(req *rApi.Request[*mongodbStandalone.Database]) rApi.StepResult {
	ctx := req.Context()
	databaseObj := req.Object

	if meta.IsStatusConditionFalse(databaseObj.Status.Conditions, "GeneratedVars") {
		if err := databaseObj.Status.GeneratedVars.Set(DbPasswordKey, fn.CleanerNanoid(40)); err != nil {
			return req.FailWithStatusError(err)
		}
		if err := r.Status().Update(ctx, databaseObj); err != nil {
			return req.FailWithStatusError(err)
		}
		return req.Done()
	}

	mc, ok := rApi.GetLocal[*libMongo.Client](req, "MongoCli")
	if !ok {
		return req.FailWithOpError(errors.Newf("mongo client not found in locals"))
	}

	dbPasswd, ok := databaseObj.Status.GeneratedVars.GetString(DbPasswordKey)
	if !ok {
		return req.FailWithOpError(errors.Newf("key %s not found in GeneratedVars", DbPasswordKey))
	}
	if err := mc.UpsertUser(ctx, databaseObj.Name, databaseObj.Name, dbPasswd); err != nil {
		return req.FailWithOpError(err)
	}

	hosts, ok := rApi.GetLocal[string](req, DbHosts)
	if !ok {
		return req.FailWithOpError(errors.Newf("key %s not found in locals", DbHosts))
	}
	outScrt := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: databaseObj.Namespace,
			Name:      fmt.Sprintf("mres-%s", databaseObj.Name),
			OwnerReferences: []metav1.OwnerReference{
				fn.AsOwner(databaseObj, true),
			},
		},
		StringData: map[string]string{
			DbPassword: dbPasswd,
			DbUser:     databaseObj.Name,
			DbHosts:    hosts,
			DbUrl: fmt.Sprintf(
				"mongodb://%s:%s@%s/%s",
				databaseObj.Name, dbPasswd, hosts, databaseObj.Name,
			),
		},
	}

	if err := fn.KubectlApply(ctx, r.Client, outScrt); err != nil {
		return req.FailWithOpError(err)
	}

	return req.Done()
}

// SetupWithManager sets up the controller with the Manager.
func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mongodbStandalone.Database{}).
		Complete(r)
}
