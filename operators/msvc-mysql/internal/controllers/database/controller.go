package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	mysqlMsvcv1 "github.com/kloudlite/operator/apis/mysql.msvc/v1"
	"github.com/kloudlite/operator/operators/msvc-mysql/internal/env"
	"github.com/kloudlite/operator/operators/msvc-mysql/internal/types"
	"github.com/kloudlite/operator/pkg/constants"
	"github.com/kloudlite/operator/pkg/errors"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	libMysql "github.com/kloudlite/operator/pkg/mysql"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	"github.com/kloudlite/operator/pkg/templates"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	logger     logging.Logger
	Name       string
	Env        *env.Env
	yamlClient kubectl.YAMLClient
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	DBUserReady      string = "db-user"
	AccessCredsReady string = "access-creds"
	IsOwnedByMsvc    string = "is-owned-by-msvc"
	DBUserDeleted    string = "db-user-deleted"
)

const (
	KeyMsvcOutput string = "msvc-output"
	KeyMresOutput string = "mres-output"
)

// +kubebuilder:rbac:groups=mysql.msvc.kloudlite.io,resources=databases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mysql.msvc.kloudlite.io,resources=databases/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mysql.msvc.kloudlite.io,resources=databases/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &mysqlMsvcv1.Database{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	req.PreReconcile()
	defer req.PostReconcile()

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.RestartIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureChecks(DBUserReady, AccessCredsReady, IsOwnedByMsvc); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconOwnership(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconDBCreds(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconDBUser(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	req.Object.Status.LastReconcileTime = &metav1.Time{Time: time.Now()}
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod * time.Second}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) finalize(req *rApi.Request[*mysqlMsvcv1.Database]) stepResult.Result {
	ctx, obj := req.Context(), req.Object

	check := rApi.Check{Generation: obj.Generation}

	if step := req.EnsureChecks(DBUserDeleted); !step.ShouldProceed() {
		return step
	}

	msvcSecret, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.MsvcRef.Namespace, "msvc-"+obj.Spec.MsvcRef.Name), &corev1.Secret{})
	if err != nil {
		return req.CheckFailed(DBUserDeleted, check, err.Error()).Err(nil)
	}

	msvcOutput, err := fn.ParseFromSecret[types.MsvcOutput](msvcSecret)
	if err != nil {
		return req.CheckFailed(AccessCredsReady, check, errors.NewEf(err, "msvc output could not be parsed").Error()).Err(nil)
	}

	if obj.Status.IsReady {
		mysqlCli, err := libMysql.NewClient(msvcOutput.Hosts, "mysql", "root", msvcOutput.RootPassword)
		if err != nil {
			req.Logger.Infof("failed to create mysql client, retrying in 5 seconds")
			return req.CheckFailed(DBUserReady, check, err.Error()).Err(nil).RequeueAfter(5 * time.Second)
		}

		if err := mysqlCli.Connect(ctx); err != nil {
			return req.CheckFailed(DBUserDeleted, check, err.Error())
		}

		if err := mysqlCli.DropUser(sanitizeDbUsername(obj.Name)); err != nil {
			return req.CheckFailed(DBUserDeleted, check, err.Error())
		}
	}

	return req.Finalize()
}

func (r *Reconciler) reconOwnership(req *rApi.Request[*mysqlMsvcv1.Database]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	msvc, err := rApi.Get(
		ctx, r.Client, fn.NN(obj.Spec.MsvcRef.Namespace, obj.Spec.MsvcRef.Name), fn.NewUnstructured(
			metav1.TypeMeta{
				Kind:       obj.Spec.MsvcRef.Kind,
				APIVersion: obj.Spec.MsvcRef.APIVersion,
			},
		),
	)

	if err != nil {
		return req.CheckFailed(IsOwnedByMsvc, check, err.Error())
	}

	if !fn.IsOwner(obj, fn.AsOwner(msvc)) {
		obj.SetOwnerReferences(append(obj.GetOwnerReferences(), fn.AsOwner(msvc)))
		if err := r.Update(ctx, obj); err != nil {
			return req.CheckFailed(IsOwnedByMsvc, check, err.Error())
		}
		return req.UpdateStatus()
	}

	check.Status = true
	if check != obj.Status.Checks[IsOwnedByMsvc] {
		obj.Status.Checks[IsOwnedByMsvc] = check
		return req.UpdateStatus()
	}

	return req.Next()
}

func sanitizeDbName(dbName string) string {
	return strings.ReplaceAll(dbName, "-", "_")
}

func sanitizeDbUsername(username string) string {
	return fn.Md5([]byte(username))
}

func (r *Reconciler) reconDBCreds(req *rApi.Request[*mysqlMsvcv1.Database]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	accessSecretName := "mres-" + obj.Name

	accessSecret, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, accessSecretName), &corev1.Secret{})
	if err != nil {
		req.Logger.Infof("access credentials %s does not exist, will be creating it now...", fn.NN(obj.Namespace, accessSecretName).String())
	}

	// msvc output ref
	msvcSecret, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.MsvcRef.Namespace, "msvc-"+obj.Spec.MsvcRef.Name), &corev1.Secret{})
	if err != nil {
		return req.CheckFailed(AccessCredsReady, check, err.Error()).Err(nil)
	}

	msvcOutput, err := fn.ParseFromSecret[types.MsvcOutput](msvcSecret)
	if err != nil {
		return req.CheckFailed(AccessCredsReady, check, errors.NewEf(err, "msvc output could not be parsed").Error()).Err(nil)
	}

	if accessSecret == nil {
		dbPasswd := fn.CleanerNanoid(40)
		dbName := sanitizeDbName(obj.Spec.ResourceName)
		dbUsername := sanitizeDbUsername(obj.Spec.ResourceName)

		mresOutput := types.MresOutput{
			Username: dbUsername,
			Password: dbPasswd,
			Hosts:    msvcOutput.Hosts,
			DbName:   dbName,
			DSN:      fmt.Sprintf("mysql://%s:%s@tcp(%s)/%s", dbUsername, dbPasswd, msvcOutput.Hosts, dbName),
			URI:      fmt.Sprintf("mysql://%s:%s@%s/%s", dbUsername, dbPasswd, msvcOutput.Hosts, dbName),
		}

		if msvcOutput.ExternalHost != "" {
			mresOutput.ExternalDSN = fmt.Sprintf("mysql://%s:%s@tcp(%s)/%s", dbUsername, dbPasswd, msvcOutput.ExternalHost, dbName)
		}

		b, err := templates.Parse(templates.Secret, map[string]any{
			"name":        accessSecretName,
			"namespace":   obj.Namespace,
			"owner-refs":  obj.GetOwnerReferences(),
			"string-data": mresOutput,
		})
		if err != nil {
			return req.CheckFailed(AccessCredsReady, check, err.Error())
		}

		if _, err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
			return req.CheckFailed(AccessCredsReady, check, err.Error())
		}

		checks[AccessCredsReady] = check
		req.UpdateStatus()
	}

	check.Status = true
	if check != obj.Status.Checks[AccessCredsReady] {
		obj.Status.Checks[AccessCredsReady] = check
		req.UpdateStatus()
	}

	mresOutput, err := fn.ParseFromSecret[types.MresOutput](accessSecret)
	if err != nil {
		return req.CheckFailed(AccessCredsReady, check, err.Error())
	}

	rApi.SetLocal(req, KeyMsvcOutput, *msvcOutput)
	rApi.SetLocal(req, KeyMresOutput, *mresOutput)

	return req.Next()
}

func (r *Reconciler) reconDBUser(req *rApi.Request[*mysqlMsvcv1.Database]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	msvcOutput, ok := rApi.GetLocal[types.MsvcOutput](req, KeyMsvcOutput)
	if !ok {
		return req.CheckFailed(DBUserReady, check, errors.NotInLocals(KeyMsvcOutput).Error()).Err(nil)
	}

	mysqlCli, err := libMysql.NewClient(msvcOutput.Hosts, "mysql", "root", msvcOutput.RootPassword)
	if err != nil {
		req.Logger.Infof("failed to create mysql client, retrying in 5 seconds")
		return req.CheckFailed(DBUserReady, check, err.Error()).Err(nil).RequeueAfter(5 * time.Second)
	}

	mresOutput, ok := rApi.GetLocal[types.MresOutput](req, KeyMresOutput)
	if !ok {
		return req.CheckFailed(DBUserReady, check, errors.NotInLocals(KeyMresOutput).Error()).Err(nil)
	}

	if err := mysqlCli.Connect(ctx); err != nil {
		req.Logger.Errorf(err, "failed to connect to mysql db instance, retrying in 5 seconds because")
		return req.CheckFailed(DBUserReady, check, errors.NewEf(err, "failed to connect to db").Error()).Err(nil).RequeueAfter(5 * time.Second)
	}
	defer mysqlCli.Close()

	exists, err := mysqlCli.UserExists(mresOutput.Username)
	if err != nil {
		return req.CheckFailed(DBUserReady, check, err.Error())
	}

	if !exists {
		if err := mysqlCli.UpsertUser(mresOutput.DbName, mresOutput.Username, mresOutput.Password); err != nil {
			return req.CheckFailed(DBUserReady, check, err.Error()).Err(nil)
		}

		checks[DBUserReady] = check
		req.UpdateStatus()
	}

	check.Status = true
	if check != obj.Status.Checks[DBUserReady] {
		obj.Status.Checks[DBUserReady] = check
		req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

	builder := ctrl.NewControllerManagedBy(mgr).For(&mysqlMsvcv1.Database{})
	builder.Owns(&corev1.Secret{})

	watchList := []client.Object{
		&mysqlMsvcv1.ClusterService{},
		&mysqlMsvcv1.StandaloneService{},
	}

	for _, obj := range watchList {
		builder.Watches(
			obj, handler.EnqueueRequestsFromMapFunc(
				func(ctx context.Context, obj client.Object) []reconcile.Request {
					msvcName, ok := obj.GetLabels()[constants.MsvcNameKey]
					if !ok {
						return nil
					}

					var dbList mysqlMsvcv1.DatabaseList
					if err := r.List(
						context.TODO(), &dbList, &client.ListOptions{
							LabelSelector: labels.SelectorFromValidatedSet(
								map[string]string{constants.MsvcNameKey: msvcName},
							),
							Namespace: obj.GetNamespace(),
						},
					); err != nil {
						return nil
					}

					reqs := make([]reconcile.Request, 0, len(dbList.Items))
					for j := range dbList.Items {
						reqs = append(reqs, reconcile.Request{NamespacedName: fn.NN(dbList.Items[j].GetNamespace(), dbList.Items[j].GetName())})
					}

					return reqs
				},
			),
		)
	}

	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
