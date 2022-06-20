package mysqlstandalonemsvc

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/errors"
	fn "operators.kloudlite.io/lib/functions"
	libMysql "operators.kloudlite.io/lib/mysql"
	rApi "operators.kloudlite.io/lib/operator"
	"operators.kloudlite.io/lib/templates"
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
	SvcKey             = "svc-key"
	SvcHostsKey        = "HOSTS"
	SvcRootPasswordKey = "ROOT_PASSWORD"
	MysqlClientKey     = "mysql-client"

	DbPasswordKey = "db-password"
)

// +kubebuilder:rbac:groups=mysql-standalone.msvc.kloudlite.io,resources=databases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mysql-standalone.msvc.kloudlite.io,resources=databases/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mysql-standalone.msvc.kloudlite.io,resources=databases/finalizers,verbs=update

func (r *DatabaseReconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	req := rApi.NewRequest(ctx, r.Client, oReq.NamespacedName, &mysqlStandalone.Database{})

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
	databaseObj := req.Object

	isReady := true
	var cs []metav1.Condition

	// STEP: check managed service is ready
	msvc, err := rApi.Get(
		ctx, r.Client, fn.NN(databaseObj.Namespace, databaseObj.Spec.ManagedSvcName),
		&mysqlStandalone.Service{},
	)

	if err != nil {
		return req.FailWithStatusError(err)
	}

	if !msvc.Status.IsReady {
		return req.FailWithStatusError(errors.Newf("msvc is not ready"))
	}

	rApi.SetLocal(req, SvcKey, msvc)

	// STEP: check managed service output is ready
	msvcSecret, err2 := rApi.Get(
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

	if msvcSecret != nil {
		hosts := string(msvcSecret.Data[SvcHostsKey])
		rootPassword := string(msvcSecret.Data[SvcRootPasswordKey])
		rApi.SetLocal(req, SvcHostsKey, hosts)
		rApi.SetLocal(req, SvcRootPasswordKey, rootPassword)

		// 2. check if user exists
		mysqlClient, err := libMysql.NewClient(hosts, "mysql", "root", rootPassword)
		if err != nil {
			isReady = false
			return req.FailWithStatusError(err)
		}
		if err := mysqlClient.Connect(ctx); err != nil {
			return req.FailWithStatusError(err)
		}
		rApi.SetLocal(req, MysqlClientKey, mysqlClient)

		userExists, err := mysqlClient.UserExists(databaseObj.Name)
		if err != nil {
			return req.FailWithStatusError(err)
		}

		if !userExists {
			cs = append(cs, conditions.New("MysqlUserExists", false, "NotFound"))
			isReady = false
		} else {
			cs = append(cs, conditions.New("MysqLUserExists", true, "Found"))
		}
	}

	// STEP: check generated vars
	if !databaseObj.Status.GeneratedVars.Exists(DbPasswordKey) {
		cs = append(cs, conditions.New("GeneratedVars", false, "NotGeneratedYet"))
		isReady = false
	} else {
		cs = append(cs, conditions.New("GeneratedVars", true, "Generated"))
	}

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

func (r *DatabaseReconciler) reconcileOperations(req *rApi.Request[*mysqlStandalone.Database]) rApi.StepResult {
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
	defer mysqlClient.Disconnect()

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
