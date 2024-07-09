package standalone_database

import (
	"context"
	"fmt"
	"strings"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	postgresv1 "github.com/kloudlite/operator/apis/postgres.msvc/v1"
	"github.com/kloudlite/operator/operators/msvc-postgres/internal/env"
	"github.com/kloudlite/operator/operators/msvc-postgres/internal/templates"
	"github.com/kloudlite/operator/operators/msvc-postgres/internal/types"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/yaml"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	logger     logging.Logger
	Name       string
	Env        *env.Env
	yamlClient kubectl.YAMLClient

	templateDBLifecycle []byte
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	createDBCreds       string = "create-db-creds"
	patchDefaults       string = "patch-defaults"
	createDBCreationJob string = "create-db-creation-job"
)

// +kubebuilder:rbac:groups=mysql.msvc.kloudlite.io,resources=databases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mysql.msvc.kloudlite.io,resources=databases/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mysql.msvc.kloudlite.io,resources=databases/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &postgresv1.StandaloneDatabase{})
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

	if step := req.EnsureCheckList([]rApi.CheckMeta{
		{Name: patchDefaults, Title: "Patch Defaults"},
		{Name: createDBCreds, Title: "Create DB Credentials"},
		{Name: createDBCreationJob, Title: "Create DB Creation Job"},
	}); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.patchDefaults(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.createDBCreds(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.createDBUserLifecycle(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *Reconciler) finalize(req *rApi.Request[*postgresv1.StandaloneDatabase]) stepResult.Result {
	// ctx, obj := req.Context(), req.Object
	//
	// check := rApi.Check{Generation: obj.Generation}
	//
	// if step := req.EnsureChecks(DBUserDeleted); !step.ShouldProceed() {
	// 	return step
	// }
	//
	// msvcSecret, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.MsvcRef.Namespace, "msvc-"+obj.Spec.MsvcRef.Name), &corev1.Secret{})
	// if err != nil {
	// 	return req.CheckFailed(DBUserDeleted, check, err.Error()).Err(nil)
	// }
	//
	// msvcOutput, err := fn.ParseFromSecret[types.MsvcOutput](msvcSecret)
	// if err != nil {
	// 	return req.CheckFailed(AccessCredsReady, check, errors.NewEf(err, "msvc output could not be parsed").Error()).Err(nil)
	// }
	//
	// if obj.Status.IsReady {
	// 	mysqlCli, err := libMysql.NewClient(msvcOutput.Hosts, "mysql", "root", msvcOutput.RootPassword)
	// 	if err != nil {
	// 		req.Logger.Infof("failed to create mysql client, retrying in 5 seconds")
	// 		return req.CheckFailed(DBUserReady, check, err.Error()).Err(nil).RequeueAfter(5 * time.Second)
	// 	}
	//
	// 	if err := mysqlCli.Connect(ctx); err != nil {
	// 		return req.CheckFailed(DBUserDeleted, check, err.Error())
	// 	}
	//
	// 	if err := mysqlCli.DropUser(sanitizeDbUsername(obj.Name)); err != nil {
	// 		return req.CheckFailed(DBUserDeleted, check, err.Error())
	// 	}
	// }

	return req.Finalize()
}

func sanitizeDbName(dbName string) string {
	return strings.ReplaceAll(dbName, "-", "_")
}

func sanitizeDbUsername(username string) string {
	return fn.Md5([]byte(username))
}

func (r *Reconciler) patchDefaults(req *rApi.Request[*postgresv1.StandaloneDatabase]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(patchDefaults, req)

	hasUpdate := false
	if obj.Output.CredentialsRef.Name == "" {
		hasUpdate = true
		obj.Output.CredentialsRef.Name = fmt.Sprintf("mres-%s-creds", obj.Name)
	}

	if hasUpdate {
		if err := r.Update(ctx, obj); err != nil {
			return check.Failed(err)
		}

		return check.StillRunning(fmt.Errorf("waiting for resource to re-sync"))
	}

	return check.Completed()
}

func (r *Reconciler) createDBCreds(req *rApi.Request[*postgresv1.StandaloneDatabase]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(createDBCreds, req)

	msvc, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.MsvcRef.Namespace, obj.Spec.MsvcRef.Name), &postgresv1.Standalone{})
	if err != nil {
		return check.Failed(err)
	}

	msvcCreds, err := rApi.Get(ctx, r.Client, fn.NN(msvc.Namespace, msvc.Output.CredentialsRef.Name), &corev1.Secret{})
	if err != nil {
		return check.Failed(err)
	}

	so, err := fn.ParseFromSecretData[types.StandaloneOutput](msvcCreds.Data)
	if err != nil {
		check.Failed(err)
	}

	creds := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: obj.Output.CredentialsRef.Name, Namespace: obj.Namespace}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, creds, func() error {
		creds.SetLabels(obj.GetLabels())
		if creds.Data == nil {
			username := sanitizeDbName(obj.Name)
			password := fn.CleanerNanoid(40)

			dbName := sanitizeDbName(obj.Name)

			// INFO: secret does not exist yet
			out := &types.StandaloneDatabaseOutput{
				Username:         username,
				Password:         password,
				DbName:           dbName,
				Port:             so.Port,
				Host:             so.Host,
				URI:              fmt.Sprintf("postgres://%s:%s@%s:%s/%s", username, password, so.Host, so.Port, dbName),
				ClusterLocalHost: so.ClusterLocalHost,
				ClusterLocalURI:  fmt.Sprintf("postgres://%s:%s@%s:%s/%s", username, password, so.ClusterLocalHost, so.Port, dbName),
			}

			m, err := out.ToMap()
			if err != nil {
				return err
			}

			creds.StringData = m
		}
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	// function-body
	return check.Completed()
}

func (r *Reconciler) createDBUserLifecycle(req *rApi.Request[*postgresv1.StandaloneDatabase]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(createDBCreationJob, req)

	msvc, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.MsvcRef.Namespace, obj.Spec.MsvcRef.Name), &postgresv1.Standalone{})
	if err != nil {
		return check.Failed(err)
	}

	lf := crdsv1.Lifecycle{ObjectMeta: metav1.ObjectMeta{Name: obj.Name, Namespace: obj.Namespace}}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, &lf, func() error {
		lf.SetLabels(obj.GetLabels())
		lf.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})

		b, err := templates.ParseBytes(r.templateDBLifecycle, templates.DBLifecycleVars{
			Metadata: metav1.ObjectMeta{
				Name:      obj.Name,
				Namespace: obj.Namespace,
				Labels:    obj.GetLabels(),
			},
			NodeSelector:                   map[string]string{},
			Tolerations:                    []corev1.Toleration{},
			PostgressRootCredentialsSecret: msvc.Output.CredentialsRef.Name,
			PostgressNewCredentialsSecret:  obj.Output.CredentialsRef.Name,
		})
		if err != nil {
			return err
		}

		var lfres crdsv1.Lifecycle
		if err := yaml.Unmarshal(b, &lfres); err != nil {
			return err
		}

		lf.Spec = lfres.Spec
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	if !lf.HasCompleted() {
		return check.StillRunning(fmt.Errorf("waiting for lifecycle job to complete"))
	}

	if lf.Status.Phase == crdsv1.JobPhaseFailed {
		return check.Failed(fmt.Errorf("lifecycle job failed"))
	}

	return check.Completed()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig(), kubectl.YAMLClientOpts{Logger: r.logger})

	var err error
	r.templateDBLifecycle, err = templates.Read(templates.DBLifecycleTemplate)
	if err != nil {
		return err
	}

	builder := ctrl.NewControllerManagedBy(mgr).For(&postgresv1.StandaloneDatabase{})
	builder.Owns(&corev1.Secret{})

	watchList := []client.Object{
		&postgresv1.Standalone{},
	}

	for _, obj := range watchList {
		builder.Watches(
			obj, handler.EnqueueRequestsFromMapFunc(
				func(ctx context.Context, obj client.Object) []reconcile.Request {
					msvcName, ok := obj.GetLabels()[constants.MsvcNameKey]
					if !ok {
						return nil
					}

					var dbList postgresv1.StandaloneDatabaseList
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
