package database

import (
	"context"
	"fmt"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"strings"
	"time"

	mongodbMsvcv1 "github.com/kloudlite/operator/apis/mongodb.msvc/v1"
	"github.com/kloudlite/operator/operators/msvc-mongo/internal/env"
	"github.com/kloudlite/operator/operators/msvc-mongo/internal/types"
	"github.com/kloudlite/operator/pkg/constants"
	"github.com/kloudlite/operator/pkg/errors"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	libMongo "github.com/kloudlite/operator/pkg/mongo"
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
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	logger     logging.Logger
	Name       string
	Env        *env.Env
	yamlClient *kubectl.YAMLClient
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	AccessCredsReady string = "access-creds"
	DBUserReady      string = "db-user-ready"
	IsOwnedByMsvc    string = "is-owned-by-msvc"

	DBUserDeleted   string = "db-user-deleted"
	DefaultsPatched string = "defaults-patched"
)

const (
	KeyMsvcOutput string = "msvc-output"
	KeyMresOutput string = "mres-output"
)

// +kubebuilder:rbac:groups=mongodb.msvc.kloudlite.io,resources=databases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mongodb.msvc.kloudlite.io,resources=databases/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mongodb.msvc.kloudlite.io,resources=databases/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	if strings.HasSuffix(request.Namespace, "-blueprint") {
		return ctrl.Result{}, nil
	}

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

	req.LogPreReconcile()
	defer req.LogPostReconcile()

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.RestartIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

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
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod * time.Second}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) finalize(req *rApi.Request[*mongodbMsvcv1.Database]) stepResult.Result {
	ctx, obj := req.Context(), req.Object

	if step := req.EnsureChecks(DBUserDeleted); !step.ShouldProceed() {
		return step
	}

	check := rApi.Check{Generation: obj.Generation}

	msvcSecret, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, "msvc-"+obj.Spec.MsvcRef.Name), &corev1.Secret{})
	if err != nil {
		req.Logger.Infof("msvc secret does not exist, means msvc does not exist, then why keep managed resource, finalizing ...")
		return req.Finalize()
	}

	msvcOutput, err := fn.ParseFromSecret[types.MsvcOutput](msvcSecret)
	if err != nil {
		return req.CheckFailed(DBUserDeleted, check, err.Error()).Err(nil)
	}

	mongoCli, err := libMongo.NewClient(msvcOutput.URI)
	if err != nil {
		return req.CheckFailed(DBUserDeleted, check, err.Error())
	}

	mCtx, _ := context.WithTimeout(ctx, 5*time.Second)
	if err := mongoCli.Connect(mCtx); err != nil {
		return req.CheckFailed(DBUserDeleted, check, err.Error())
	}
	defer mongoCli.Close()

	if err := mongoCli.DeleteUser(ctx, obj.Name, obj.Name); err != nil {
		return req.CheckFailed(DBUserDeleted, check, err.Error())
	}

	return req.Finalize()
}

// ensures ManagedResource is Owned by corresponding ManagedService
func (r *Reconciler) reconOwnership(req *rApi.Request[*mongodbMsvcv1.Database]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(IsOwnedByMsvc)
	defer req.LogPostCheck(IsOwnedByMsvc)

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

	req.LogPreCheck(AccessCredsReady)
	defer req.LogPostCheck(AccessCredsReady)

	secretName := "mres-" + obj.Name

	scrt, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, secretName), &corev1.Secret{})
	if err != nil {
		req.Logger.Infof("access credentials %s does not exist, will be creating it now...", fn.NN(obj.Namespace, secretName).String())
	}

	// msvc output ref
	msvcSecret, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, "msvc-"+obj.Spec.MsvcRef.Name), &corev1.Secret{})
	if err != nil {
		return req.CheckFailed(AccessCredsReady, check, errors.NewEf(err, "msvc output does not exist").Error())
	}

	msvcOutput, err := fn.ParseFromSecret[types.MsvcOutput](msvcSecret)
	if err != nil {
		return req.CheckFailed(AccessCredsReady, check, err.Error()).Err(nil)
	}

	if scrt == nil {
		dbPasswd := fn.CleanerNanoid(40)

		b2, err := templates.Parse(
			templates.Secret, map[string]any{
				"name":       secretName,
				"namespace":  obj.Namespace,
				"owner-refs": obj.GetOwnerReferences(),
				"string-data": types.MresOutput{
					Username: obj.Name,
					Password: dbPasswd,
					Hosts:    msvcOutput.Hosts,
					// DbName:   obj.Name,
					DbName: obj.Spec.ResourceName,
					URI:    fmt.Sprintf("mongodb://%s:%s@%s/%s", obj.Name, dbPasswd, msvcOutput.Hosts, obj.Spec.ResourceName),
				},
			},
		)
		if err != nil {
			return req.CheckFailed(AccessCredsReady, check, err.Error())
		}

		if _, err := r.yamlClient.ApplyYAML(ctx, b2); err != nil {
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

	mresOutput, err := fn.ParseFromSecret[types.MresOutput](scrt)
	if err != nil {
		return req.CheckFailed(AccessCredsReady, check, err.Error()).Err(nil)
	}

	rApi.SetLocal(req, KeyMsvcOutput, *msvcOutput)
	rApi.SetLocal(req, KeyMresOutput, *mresOutput)
	return req.Next()
}

func (r *Reconciler) reconDBUser(req *rApi.Request[*mongodbMsvcv1.Database]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(DBUserReady)
	defer req.LogPostCheck(DBUserReady)

	mresOutput, ok := rApi.GetLocal[types.MresOutput](req, KeyMresOutput)
	if !ok {
		return req.CheckFailed(DBUserReady, check, errors.NotInLocals(KeyMresOutput).Error())
	}

	msvcOutput, ok := rApi.GetLocal[types.MsvcOutput](req, KeyMsvcOutput)
	if !ok {
		return req.CheckFailed(DBUserReady, check, errors.NotInLocals(KeyMsvcOutput).Error())
	}

	mongoCli, err := libMongo.NewClient(msvcOutput.URI)
	if err != nil {
		return req.CheckFailed(DBUserReady, check, err.Error())
	}

	mCtx, _ := context.WithTimeout(ctx, 5*time.Second)
	if err := mongoCli.Connect(mCtx); err != nil {
		return req.CheckFailed(DBUserReady, check, err.Error())
	}
	defer mongoCli.Close()

	exists, err := mongoCli.UserExists(ctx, mresOutput.DbName, obj.Name)
	if err != nil {
		return req.CheckFailed(DBUserReady, check, err.Error())
	}

	if !exists {
		if err := mongoCli.UpsertUser(ctx, mresOutput.DbName, mresOutput.Username, mresOutput.Password); err != nil {
			return req.CheckFailed(DBUserReady, check, err.Error())
		}
		checks[DBUserReady] = check
		return req.UpdateStatus()
	}

	check.Status = true
	if check != checks[DBUserReady] {
		checks[DBUserReady] = check
		return req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

	builder := ctrl.NewControllerManagedBy(mgr).For(&mongodbMsvcv1.Database{})
	builder.WithOptions(controller.Options{
		MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles,
	})
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
