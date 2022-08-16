package mysqlstandalonemsvc

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"operators.kloudlite.io/env"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/errors"
	fn "operators.kloudlite.io/lib/functions"
	"operators.kloudlite.io/lib/logging"
	libMysql "operators.kloudlite.io/lib/mysql"
	rApi "operators.kloudlite.io/lib/operator"
	stepResult "operators.kloudlite.io/lib/operator/step-result"
	"operators.kloudlite.io/lib/templates"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"k8s.io/apimachinery/pkg/runtime"
	mysqlStandalone "operators.kloudlite.io/apis/mysql-standalone.msvc/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DatabaseReconciler reconciles a Database object
type DatabaseReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	logger logging.Logger
	Name   string
}

func (r *DatabaseReconciler) GetName() string {
	return r.Name
}

const (
	DbPasswordKey = "db-password"
)

const (
	MysqlUserExists conditions.Type = "MysqlUserExists"
)

type MsvcOutputRef struct {
	Hosts        string
	RootPassword string
}

func parseMsvcOutput(s *corev1.Secret) *MsvcOutputRef {
	return &MsvcOutputRef{
		Hosts:        string(s.Data["HOSTS"]),
		RootPassword: string(s.Data["ROOT_PASSWORD"]),
	}
}

// +kubebuilder:rbac:groups=mysql-standalone.msvc.kloudlite.io,resources=databases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mysql-standalone.msvc.kloudlite.io,resources=databases/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mysql-standalone.msvc.kloudlite.io,resources=databases/finalizers,verbs=update

func (r *DatabaseReconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, oReq.NamespacedName, &mysqlStandalone.Database{})
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

func (r *DatabaseReconciler) finalize(req *rApi.Request[*mysqlStandalone.Database]) stepResult.Result {
	return req.Finalize()
}

func formatDbName(dbName string) string {
	return strings.ReplaceAll(dbName, "-", "_")
}

func (r *DatabaseReconciler) reconcileStatus(req *rApi.Request[*mysqlStandalone.Database]) stepResult.Result {
	ctx := req.Context()
	obj := req.Object

	isReady := true
	var cs []metav1.Condition

	// STEP: 1. check managed service is ready
	msvc, err := rApi.Get(
		ctx, r.Client, fn.NN(obj.Namespace, obj.Spec.ManagedSvcName),
		&mysqlStandalone.Service{},
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
			mysqlClient, err := libMysql.NewClient(msvcRef.Hosts, "mysql", "root", msvcRef.RootPassword)
			if err != nil {
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
		}(); err2 != nil {
			isReady = false
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

	if err := r.Status().Update(ctx, obj); err != nil {
		return req.FailWithStatusError(err)
	}
	return req.Done()
}

func (r *DatabaseReconciler) reconcileOperations(req *rApi.Request[*mysqlStandalone.Database]) stepResult.Result {
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

	// STEP: 2. generate vars if needed to
	if meta.IsStatusConditionFalse(obj.Status.Conditions, conditions.GeneratedVars.String()) {
		if err := obj.Status.GeneratedVars.Set(DbPasswordKey, fn.CleanerNanoid(40)); err != nil {
			return req.FailWithStatusError(err)
		}
		if err := r.Status().Update(ctx, obj); err != nil {
			return req.FailWithOpError(err)
		}
		return req.Done()
	}

	// STEP: 3. retrieve msvc output, need it in creating reconciler output
	msvcRef, ok := rApi.GetLocal[*MsvcOutputRef](req, "msvc-output-ref")
	if !ok {
		return req.FailWithOpError(errors.Newf("err=%s key not found in req locals", "msvc-output-ref"))
	}

	dbPasswd, ok := obj.Status.GeneratedVars.GetString(DbPasswordKey)
	if !ok {
		return req.FailWithOpError(errors.Newf("key=%s must be present in .Status.GeneratedVars", DbPasswordKey))
	}
	dbName := formatDbName(obj.Name)
	// STEP: 4. create child components like mongo-user, redis-acl etc.
	err4 := func() error {
		mysqlClient, err := libMysql.NewClient(msvcRef.Hosts, "mysql", "root", msvcRef.RootPassword)
		if err != nil {
			return err
		}
		if err := mysqlClient.Connect(ctx); err != nil {
			return err
		}
		defer mysqlClient.Close()

		return mysqlClient.UpsertUser(dbName, obj.Name, dbPasswd)
	}()
	if err4 != nil {
		// TODO:(user) might need to reconcile with retry with timeout error
		return req.FailWithOpError(err4)
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
				"PASSWORD": dbPasswd,
				"HOSTS":    msvcRef.Hosts,
				"DB_NAME":  formatDbName(obj.Name),
				"DSN":      fmt.Sprintf("%s:%s@tcp(%s)/%s", obj.Name, dbPasswd, msvcRef.Hosts, dbName),
				"URI":      fmt.Sprintf("mysqlx://%s:%s@%s/%s", obj.Name, dbPasswd, msvcRef.Hosts, dbName),
			},
		},
	)
	if err != nil {
		req.Logger.Errorf(err, "failed parsing template %s", templates.Secret)
		return req.FailWithOpError(err).Err(nil)
	}

	if err := fn.KubectlApplyExec(ctx, b); err != nil {
		req.Logger.Errorf(err, "failed kubectl apply for template %s", templates.Secret)
		return req.FailWithOpError(err).Err(nil)
	}

	return req.Next()
}

func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager, envVars *env.Env, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)

	return ctrl.NewControllerManagedBy(mgr).
		For(&mysqlStandalone.Database{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
