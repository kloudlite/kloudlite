package mysqlstandalonemsvc

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mongodbStandalone "operators.kloudlite.io/apis/mongodb-standalone.msvc/v1"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/errors"
	fn "operators.kloudlite.io/lib/functions"
	libMysql "operators.kloudlite.io/lib/mysql"
	rApi "operators.kloudlite.io/lib/operator"
	"operators.kloudlite.io/lib/templates"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	mysqlStandalone "operators.kloudlite.io/apis/mysql-standalone.msvc/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DatabaseReconciler reconciles a Database object
type DatabaseReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const (
	DbPasswordKey = "db-password"
)

const (
	MysqlUserExists conditions.Type = "MysqlUserExists"
)

type msvcOutputRef struct {
	Hosts string
	RootPassword string
}


func parseMsvcOutput(scrt *corev1.Secret) *msvcOutputRef {
	data := scrt.Data
	return &msvcOutputRef{
		Hosts:        string(data["HOSTS"]),
		RootPassword: string(data["ROOT_PASSWORD"]),
	}
}

// +kubebuilder:rbac:groups=mysql-standalone.msvc.kloudlite.io,resources=databases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mysql-standalone.msvc.kloudlite.io,resources=databases/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mysql-standalone.msvc.kloudlite.io,resources=databases/finalizers,verbs=update

func (r *DatabaseReconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	req, _ := rApi.NewRequest(ctx, r.Client, oReq.NamespacedName, &mysqlStandalone.Database{})

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

func (r *DatabaseReconciler) finalize(req *rApi.Request[*mysqlStandalone.Database]) rApi.StepResult {
	return req.Finalize()
}

func formatDbName(dbName string) string {
	return strings.ReplaceAll(dbName, "-", "_")
}

func (r *DatabaseReconciler) reconcileStatus(req *rApi.Request[*mysqlStandalone.Database]) rApi.StepResult {
	ctx := req.Context()
	obj := req.Object

	isReady := true
	var cs []metav1.Condition

	// STEP: check managed service is ready
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
			parseMsvcOutput(msvcOutput).
			rApi.SetLocal(req, MsvcOutputKey, msvcOutput)
		}

		// STEP: check reconciler (child components e.g. mongo account, s3 bucket, redis ACL user) exists
		if err := func() error {
			mysqlClient, err := libMysql.NewClient(
				string(msvcOutput.Data[SvcHostsKey]), "mysql", "root", string(msvcOutput.Data[SvcRootPasswordKey]),
			)
			if err != nil {
				isReady = false
				return err
			}
			if err := mysqlClient.Connect(ctx); err != nil {
				return err
			}
			defer mysqlClient.Close()

			userExists, err := mysqlClient.UserExists(obj.Name)
			if err != nil {
				return err
			}
			if !userExists {
				isReady = false
				cs = append(cs, conditions.New(MysqlUserExists, false, conditions.NotFound))
				return nil
			}
			cs = append(cs, conditions.New(MysqlUserExists, true, conditions.Found))
			return nil
		}(); err != nil {
			return req.FailWithStatusError(err)
		}
	}

	// STEP: check generated vars
	if msvc != nil && !obj.Status.GeneratedVars.Exists(DbPasswordKey) {
		cs = append(cs, conditions.New(conditions.GeneratedVars, false, conditions.NotReconciledYet))
	} else {
		cs = append(cs, conditions.New(conditions.GeneratedVars, true, conditions.Found))
	}

	// STEP: patch conditions
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

func (r *DatabaseReconciler) reconcileOperations(req *rApi.Request[*mysqlStandalone.Database]) rApi.StepResult {
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
	msvcOutput, ok := rApi.GetLocal[corev1.Secret](req, MsvcOutputKey)
	if !ok {
		return req.FailWithOpError(errors.Newf("err=%s key not found in req locals", MsvcOutputKey))
	}
	hosts, rootPassword := parseMsvcSecret(msvcOutput)

	if errt := func() error {
		// STEP: 4. create child components like mongo-user, redis-acl etc.
		mysqlClient, err := libMysql.NewClient(
			string(msvcOutput.Data[SvcHostsKey]), "mysql", "root", string(msvcOutput.Data[SvcRootPasswordKey]),
		)
		if err != nil {
			return err
		}
		if err := mysqlClient.Connect(ctx); err != nil {
			return err
		}
		defer mysqlClient.Close()

		dbPassword, ok := obj.Status.GeneratedVars.GetString(DbPasswordKey)
		if !ok {
			return errors.Newf("key=%s not found in .Status.GeneratedVars", DbPasswordKey)
		}
		if err := mysqlClient.UpsertUser(formatDbName(obj.Name), obj.Name, dbPassword); err != nil {
			return err
		}

		// STEP: 5. create reconciler output (eg. secret)
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
					"USERNAME": obj.Name,
					"PASSWORD": dbPassword,
					"HOSTS":    string(msvcOutput.Data[SvcHostsKey]),
					"DB_NAME":  formatDbName(obj.Name),
					"DSN": fmt.Sprintf(
						"%s:%s@tcp(%s)/%s",
						obj.Name,
						dbPassword,
						,
						formatDbName(databaseObj.Name),
					), "URI": fmt.Sprintf(
						"mysqlx://%s:%s@%s/%s", databaseObj.Name, dbPassword, mysqlHost,
						formatDbName(databaseObj.Name),
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

func (r *DatabaseReconciler) reconcileOperations2(req *rApi.Request[*mysqlStandalone.Database]) rApi.StepResult {
	ctx := req.Context()
	databaseObj := req.Object

	if meta.IsStatusConditionFalse(databaseObj.Status.Conditions, "GeneratedVars") {
		if err := databaseObj.Status.GeneratedVars.Set(DbPasswordKey, fn.CleanerNanoid(40)); err != nil {
			return req.FailWithOpError(err)
		}
		return rApi.NewStepResult(&ctrl.Result{}, r.Status().Update(ctx, databaseObj))
	}

	mysqlClient, ok := rApi.GetLocal[*libMysql.Client](req, MysqlClientKey)
	if !ok {
		return req.FailWithOpError(errors.Newf("key=%s must be present in req locals", MysqlClientKey))
	}
	if err := mysqlClient.Connect(ctx); err != nil {
		return req.FailWithOpError(err)
	}
	defer mysqlClient.Close()

	dbPassword, ok := databaseObj.Status.GeneratedVars.GetString(DbPasswordKey)
	if !ok {
		return req.FailWithOpError(errors.Newf("key=%s must be present in .Status.GeneratedVars", MysqlClientKey))
	}

	if err := mysqlClient.UpsertUser(databaseObj.Name, databaseObj.Name, dbPassword); err != nil {
		return req.FailWithOpError(errors.NewEf(err, "creating user=%s", databaseObj.Name))
	}

	mysqlHost, ok := rApi.GetLocal[string](req, SvcHostsKey)
	if !ok {
		return req.FailWithOpError(errors.Newf("key=%ss must be present in req.locals", SvcHostsKey))
	}

	// create secret for this resource

	b, err := templates.Parse(
		templates.Secret, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("mres-%s", databaseObj.Name),
				Namespace: databaseObj.Namespace,
			},
			// Immutable:  fn.NewBool(true),
			StringData: map[string]string{
				"USERNAME": databaseObj.Name,
				"PASSWORD": dbPassword,
				"HOSTS":    mysqlHost,
				"DB_NAME":  formatDbName(databaseObj.Name),
				"DSN": fmt.Sprintf(
					"%s:%s@tcp(%s)/%s",
					databaseObj.Name,
					dbPassword,
					mysqlHost,
					formatDbName(databaseObj.Name),
				), "URI": fmt.Sprintf(
					"mysqlx://%s:%s@%s/%s", databaseObj.Name, dbPassword, mysqlHost,
					formatDbName(databaseObj.Name),
				),
			},
		},
	)
	if err != nil {
		return req.FailWithOpError(errors.NewEf(err, "parsing template=%s", templates.Secret))
	}

	if _, err := fn.KubectlApplyExec(b); err != nil {
		return req.FailWithOpError(errors.NewEf(err, "kubectl apply"))
	}

	return req.Done()
}

// SetupWithManager sets up the controller with the Manager.
func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mysqlStandalone.Database{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
