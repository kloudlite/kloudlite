package database

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	mongodbMsvcv1 "operators.kloudlite.io/apis/mongodb.msvc/v1"
	"operators.kloudlite.io/env"
	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/errors"
	fn "operators.kloudlite.io/lib/functions"
	"operators.kloudlite.io/lib/harbor"
	"operators.kloudlite.io/lib/logging"
	libMongo "operators.kloudlite.io/lib/mongo"
	rApi "operators.kloudlite.io/lib/operator"
	stepResult "operators.kloudlite.io/lib/operator/step-result"
	"operators.kloudlite.io/lib/templates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type Reconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	env       *env.Env
	harborCli *harbor.Client
	logger    logging.Logger
	Name      string
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	AccessCredsReady string = "access-creds"
	DBUserReady      string = "db-user-ready"
	IsOwnedByMsvc    string = "is-owned-by-msvc"
)

// +kubebuilder:rbac:groups=mongodb.msvc.kloudlite.io,resources=databases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mongodb.msvc.kloudlite.io,resources=databases/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mongodb.msvc.kloudlite.io,resources=databases/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, request.NamespacedName, &mongodbMsvcv1.Database{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	req.Logger.Infof("NEW RECONCILATION")

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.RestartIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	// TODO: initialize all checks here
	if step := req.EnsureChecks(AccessCredsReady, DBUserReady); !step.ShouldProceed() {
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
	req.Logger.Infof("RECONCILATION COMPLETE")
	return ctrl.Result{RequeueAfter: r.env.ReconcilePeriod * time.Second}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) finalize(req *rApi.Request[*mongodbMsvcv1.Database]) stepResult.Result {
	return req.Finalize()
}

func (r *Reconciler) reconOwnership(req *rApi.Request[*mongodbMsvcv1.Database]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks

	check := rApi.Check{Generation: obj.Generation}

	msvc, err := rApi.Get(
		ctx, r.Client, fn.NN(obj.Namespace, obj.Spec.MsvcRef.Name), fn.NewUnstructured(
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
			return req.FailWithOpError(err)
		}
		return req.UpdateStatus()
	}

	check.Status = true
	if check != checks[IsOwnedByMsvc] {
		checks[IsOwnedByMsvc] = check
		return req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) reconDBCreds(req *rApi.Request[*mongodbMsvcv1.Database]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	accessSecretName := "mres-" + obj.Name

	accessSecret, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, accessSecretName), &corev1.Secret{})
	if err != nil {
		req.Logger.Infof("access credentials %s does not exist, will be creating it now...", fn.NN(obj.Namespace, accessSecretName).String())
	}

	// msvc output ref
	msvcOutput, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, "msvc-"+obj.Spec.MsvcRef.Name), &corev1.Secret{})
	if err != nil {
		return req.CheckFailed(AccessCredsReady, check, errors.NewEf(err, "msvc output does not exist").Error())
	}

	if accessSecret == nil {
		dbPasswd := fn.CleanerNanoid(40)
		hosts := string(msvcOutput.Data["DB_HOSTS"])

		b, err := templates.Parse(
			templates.Secret, map[string]any{
				"name":       accessSecretName,
				"namespace":  obj.Namespace,
				"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},
				"string-data": map[string]string{
					"DB_PASSWORD": dbPasswd,
					"DB_USER":     obj.Name,
					"DB_HOSTS":    hosts,
					"DB_NAME":     obj.Name,
					"DB_URL":      fmt.Sprintf("mongodb://%s:%s@%s/%s", obj.Name, dbPasswd, hosts, obj.Name),
				},
			},
		)
		if err != nil {
			return req.CheckFailed(AccessCredsReady, check, err.Error())
		}

		if err := fn.KubectlApplyExec(ctx, b); err != nil {
			return req.CheckFailed(AccessCredsReady, check, err.Error())
		}

		checks[AccessCredsReady] = check
		return req.UpdateStatus()
	}

	check.Status = true
	if check != checks[AccessCredsReady] {
		checks[AccessCredsReady] = check
		return req.UpdateStatus()
	}

	rApi.SetLocal(req, "db-name", obj.Name)
	rApi.SetLocal(req, "db-user", obj.Name)
	rApi.SetLocal(req, "db-password", string(accessSecret.Data["DB_PASSWORD"]))
	rApi.SetLocal(req, "db-root-uri", string(msvcOutput.Data["DB_URL"]))

	return req.Next()
}

func (r *Reconciler) reconDBUser(req *rApi.Request[*mongodbMsvcv1.Database]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	dbName, ok1 := rApi.GetLocal[string](req, "db-name")
	dbUser, ok2 := rApi.GetLocal[string](req, "db-user")
	dbPassword, ok3 := rApi.GetLocal[string](req, "db-password")
	dbRootUri, ok4 := rApi.GetLocal[string](req, "db-root-uri")
	if !ok1 || !ok2 || !ok3 || !ok4 {
		return req.CheckFailed(DBUserReady, check, "key db-name/db-user/db-password/db-root-uri does not exist in req-locals")
	}

	mongoClient, err := libMongo.NewClient(dbRootUri)
	if err != nil {
		return req.CheckFailed(AccessCredsReady, check, err.Error())
	}

	mCtx, _ := context.WithTimeout(ctx, 5*time.Second)
	if err := mongoClient.Connect(mCtx); err != nil {
		return req.CheckFailed(DBUserReady, check, err.Error())
	}
	defer mongoClient.Close()

	userExists, err := mongoClient.UserExists(ctx, obj.Name)
	if err != nil {
		return req.CheckFailed(DBUserReady, check, err.Error())
	}

	if !userExists {
		if err := mongoClient.UpsertUser(ctx, dbName, dbUser, dbPassword); err != nil {
			return req.CheckFailed(DBUserReady, check, err.Error())
		}
	}

	check.Status = true
	if check != checks[DBUserReady] {
		checks[DBUserReady] = check
		return req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, envVars *env.Env, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.env = envVars

	builder := ctrl.NewControllerManagedBy(mgr).For(&mongodbMsvcv1.Database{})
	builder.Owns(&corev1.Secret{})

	watchList := []client.Object{
		&mongodbMsvcv1.StandaloneService{},
		&mongodbMsvcv1.ClusterService{},
	}

	for i := range watchList {
		builder.Watches(
			&source.Kind{Type: watchList[i]}, handler.EnqueueRequestsFromMapFunc(
				func(obj client.Object) []reconcile.Request {
					msvcName, ok := obj.GetLabels()[constants.MsvcNameKey]
					if !ok {
						return nil
					}

					var dbList mongodbMsvcv1.DatabaseList
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

	return builder.Complete(r)
}
