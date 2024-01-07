package database

import (
	"context"
	"fmt"
	"time"

	mongodbMsvcv1 "github.com/kloudlite/operator/apis/mongodb.msvc/v1"
	"github.com/kloudlite/operator/operators/msvc-mongo/internal/env"
	"github.com/kloudlite/operator/operators/msvc-mongo/internal/types"
	"github.com/kloudlite/operator/pkg/constants"
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
	"sigs.k8s.io/controller-runtime/pkg/controller"
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

	templateJobUserCreate []byte
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

const (
	LabelResourceGeneration = "job-resource-generation"
	LabelUserCreateJob      = "user-create-job"
	LabelUserRemoveJob      = "user-remove-job"
)

func (r *Reconciler) newMongoContext(parent context.Context) (context.Context, context.CancelFunc) {
	if r.Env.IsDev {
		return context.WithCancel(parent)
	}
	return context.WithTimeout(parent, 5*time.Second)
}

// +kubebuilder:rbac:groups=mongodb.msvc.kloudlite.io,resources=databases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mongodb.msvc.kloudlite.io,resources=databases/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mongodb.msvc.kloudlite.io,resources=databases/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &mongodbMsvcv1.Database{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	req.PreReconcile()
	defer req.PostReconcile()

	if step := r.patchDefaults(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureChecks(AccessCredsReady, DBUserReady, IsOwnedByMsvc, DBUserDeleted, DefaultsPatched); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	// if step := r.reconOwnership(req); !step.ShouldProceed() {
	// 	return step.ReconcilerResponse()
	// }

	if step := r.reconDBCreds(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

// type MsvcMeta struct {
// 	Name      string
// 	Namespace string
// }
//
// func getMsvcMeta(res *crdsv1.ManagedResource) MsvcMeta {
// 	return MsvcMeta{
// 		Name:      res.Spec.ResourceTemplate.MsvcRef.Namespace,
// 		Namespace: res.Spec.ResourceTemplate.MsvcRef.Name,
// 	}
// }

func (r *Reconciler) patchDefaults(req *rApi.Request[*mongodbMsvcv1.Database]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(DefaultsPatched)
	defer req.LogPostCheck(DefaultsPatched)

	hasPatched := false

	if obj.Spec.Output.Credentials.Name == "" {
		hasPatched = true
		obj.Spec.Output.Credentials.Name = fmt.Sprintf("mres-%s-creds", obj.Name)
	}

	if obj.Spec.Output.Credentials.Namespace == "" {
		hasPatched = true
		obj.Spec.Output.Credentials.Namespace = obj.Namespace
	}

	if hasPatched {
		if err := r.Update(ctx, obj); err != nil {
			return req.CheckFailed(DefaultsPatched, check, err.Error())
		}
	}

	check.Status = true
	if check != obj.Status.Checks[DefaultsPatched] {
		fn.MapSet(&obj.Status.Checks, DefaultsPatched, check)
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
		return req.Done()
	}

	return req.Next()
}

func (r *Reconciler) finalize(req *rApi.Request[*mongodbMsvcv1.Database]) stepResult.Result {
	ctx, obj := req.Context(), req.Object

	if step := req.EnsureChecks(DBUserDeleted); !step.ShouldProceed() {
		return step
	}

	check := rApi.Check{Generation: obj.Generation}

	_, URI, err := r.getMsvcConnectionParams(ctx, obj)
	if err != nil {
		return req.CheckFailed(DBUserDeleted, check, err.Error()).Err(nil)
	}

	mctx, cancel := r.newMongoContext(ctx)
	defer cancel()
	mongoCli, err := libMongo.NewClient(mctx, URI)
	if err != nil {
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
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(IsOwnedByMsvc)
	defer req.LogPostCheck(IsOwnedByMsvc)

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
	}

	check.Status = true
	if check != obj.Status.Checks[IsOwnedByMsvc] {
		fn.MapSet(&obj.Status.Checks, IsOwnedByMsvc, check)
		if step := req.UpdateStatus(); !step.ShouldProceed() {
			return step
		}
	}

	return req.Next()
}

func (r *Reconciler) getMsvcConnectionParams(ctx context.Context, obj *mongodbMsvcv1.Database) (hosts string, URI string, err error) {
	switch obj.Spec.MsvcRef.Kind {
	case "StandaloneService":
		{
			msvc, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.MsvcRef.Namespace, obj.Spec.MsvcRef.Name), &mongodbMsvcv1.StandaloneService{})
			if err != nil {
				return "", "", err
			}

			s, err := rApi.Get(ctx, r.Client, fn.NN(msvc.Spec.Output.Credentials.Namespace, msvc.Spec.Output.Credentials.Name), &corev1.Secret{})
			if err != nil {
				return "", "", err
			}

			cso, err := fn.ParseFromSecret[types.StandaloneSvcOutput](s)
			if err != nil {
				return "", "", err
			}

			return cso.Hosts, cso.URI, err
		}
	case "ClusterService":
		{
			msvc, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.MsvcRef.Namespace, obj.Spec.MsvcRef.Name), &mongodbMsvcv1.ClusterService{})
			if err != nil {
				return "", "", err
			}

			s, err := rApi.Get(ctx, r.Client, fn.NN(msvc.Spec.Output.Credentials.Namespace, msvc.Spec.Output.Credentials.Name), &corev1.Secret{})
			if err != nil {
				return "", "", err
			}

			cso, err := fn.ParseFromSecret[types.ClusterSvcOutput](s)
			if err != nil {
				return "", "", err
			}

			return cso.Hosts, cso.URI, err
		}
	default:
		return "", "", fmt.Errorf("unknown msvc kind: %s", obj.Spec.MsvcRef.Kind)
	}
}

func (r *Reconciler) reconDBCreds(req *rApi.Request[*mongodbMsvcv1.Database]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(AccessCredsReady)
	defer req.LogPostCheck(AccessCredsReady)

	secretName := obj.Spec.Output.Credentials.Name
	secretNamespace := obj.Spec.Output.Credentials.Namespace

	scrt, err := rApi.Get(ctx, r.Client, fn.NN(secretNamespace, secretName), &corev1.Secret{})
	if err != nil {
		req.Logger.Infof("access credentials %s/%s does not exist, will be creating it now...", secretNamespace, secretName)
	}

	msvcHosts, msvcURI, err := r.getMsvcConnectionParams(ctx, obj)
	if err != nil {
		return req.CheckFailed(AccessCredsReady, check, err.Error()).Err(nil)
	}

	shouldGeneratePassword := scrt == nil

	if scrt != nil {
		mresOutput, err := fn.ParseFromSecret[types.MresOutput](scrt)
		if err != nil {
			return req.CheckFailed(AccessCredsReady, check, err.Error()).Err(nil)
		}

		err = libMongo.ConnectAndPing(ctx, mresOutput.URI)
		if err != nil {
			if !libMongo.FailsWithAuthError(err) {
				return req.CheckFailed(AccessCredsReady, check, err.Error())
			}
			req.Logger.Infof("Invalid Credentials in secret's .data.URI, would need to be regenerated as connection failed with auth error")

			shouldGeneratePassword = true
		}
	}

	if shouldGeneratePassword {
		dbPasswd := fn.CleanerNanoid(40)

		mresOutput := types.MresOutput{
			Username: obj.Name,
			Password: dbPasswd,
			Hosts:    msvcHosts,
			DbName:   obj.Spec.ResourceName,
			URI: func() string {
				baseURI := fmt.Sprintf("mongodb://%s:%s@%s/%s?authSource=%s", obj.Name, dbPasswd, msvcHosts, obj.Spec.ResourceName, obj.Spec.ResourceName)
				if obj.Spec.MsvcRef.Kind == "ClusterService" {
					return baseURI + "&replicaSet=rs"
				}
				return baseURI
			}(),
		}

		b2, err := templates.Parse(
			templates.Secret, map[string]any{
				"name":        secretName,
				"namespace":   secretNamespace,
				"owner-refs":  obj.GetOwnerReferences(),
				"string-data": mresOutput,
			},
		)
		if err != nil {
			return req.CheckFailed(AccessCredsReady, check, err.Error())
		}

		if _, err := r.yamlClient.ApplyYAML(ctx, b2); err != nil {
			return req.CheckFailed(AccessCredsReady, check, err.Error())
		}

		mctx, cancel := r.newMongoContext(ctx)
		defer cancel()
		mongoCli, err := libMongo.NewClient(mctx, msvcURI)
		if err != nil {
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
			fn.MapSet(&obj.Status.Checks, DBUserReady, check)
			return req.UpdateStatus()
		}

		if exists {
			if err := mongoCli.UpdateUserPassword(ctx, mresOutput.DbName, mresOutput.Username, mresOutput.Password); err != nil {
				return req.CheckFailed(DBUserReady, check, err.Error())
			}
		}
	}

	check.Status = true
	if check != obj.Status.Checks[AccessCredsReady] {
		fn.MapSet(&obj.Status.Checks, AccessCredsReady, check)
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig(), kubectl.YAMLClientOpts{Logger: r.logger})

	builder := ctrl.NewControllerManagedBy(mgr).For(&mongodbMsvcv1.Database{})
	builder.WithOptions(controller.Options{
		MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles,
	})
	builder.Owns(&corev1.Secret{})

	watchList := []client.Object{
		&mongodbMsvcv1.StandaloneService{},
		&mongodbMsvcv1.ClusterService{},
	}

	for _, obj := range watchList {
		builder.Watches(
			obj,
			handler.EnqueueRequestsFromMapFunc(
				func(ctx context.Context, obj client.Object) []reconcile.Request {
					msvcName, ok := obj.GetLabels()[constants.MsvcNameKey]
					if !ok {
						return nil
					}

					var dbList mongodbMsvcv1.DatabaseList
					if err := r.List(ctx, &dbList, &client.ListOptions{
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
	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	return builder.Complete(r)
}
