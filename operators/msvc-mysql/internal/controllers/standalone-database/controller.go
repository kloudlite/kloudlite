package standalone_database

import (
	"context"
	"fmt"
	"strings"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	mysqlv1 "github.com/kloudlite/operator/apis/mysql.msvc/v1"
	"github.com/kloudlite/operator/operators/msvc-mysql/internal/env"
	"github.com/kloudlite/operator/operators/msvc-mysql/internal/templates"
	"github.com/kloudlite/operator/operators/msvc-mysql/internal/types"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
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
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &mysqlv1.StandaloneDatabase{})
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

func (r *Reconciler) finalize(req *rApi.Request[*mysqlv1.StandaloneDatabase]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck("finalizing", req)

	ss, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.MsvcRef.Namespace, obj.Spec.MsvcRef.Name), &mysqlv1.StandaloneService{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return check.Failed(err)
		}
		ss = nil
	}

	if ss == nil || ss.DeletionTimestamp != nil {
		if step := req.ForceCleanupOwnedResources(check); !step.ShouldProceed() {
			return step
		}
	} else {
		if step := req.CleanupOwnedResources(); !step.ShouldProceed() {
			return step
		}
	}

	if err := fn.DeleteAndWait(ctx, r.logger, r.Client, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Object.Output.CredentialsRef.Name,
			Namespace: req.Object.Namespace,
		},
	}); err != nil {
		if !apiErrors.IsNotFound(err) {
			return check.Failed(err)
		}
	}

	return req.Finalize()
}

func sanitizeDbName(dbName string) string {
	return strings.ReplaceAll(dbName, "-", "_")
}

func sanitizeDbUsername(username string) string {
	return fn.Md5([]byte(username))
}

func (r *Reconciler) patchDefaults(req *rApi.Request[*mysqlv1.StandaloneDatabase]) stepResult.Result {
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

func (r *Reconciler) createDBCreds(req *rApi.Request[*mysqlv1.StandaloneDatabase]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(createDBCreds, req)

	msvc, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.MsvcRef.Namespace, obj.Spec.MsvcRef.Name), &mysqlv1.StandaloneService{})
	if err != nil {
		return check.Failed(err)
	}

	msvcCreds, err := rApi.Get(ctx, r.Client, fn.NN(msvc.Namespace, msvc.Output.CredentialsRef.Name), &corev1.Secret{})
	if err != nil {
		return check.Failed(err)
	}

	so, err := fn.ParseFromSecretData[types.StandaloneServiceOutput](msvcCreds.Data)
	if err != nil {
		check.Failed(err)
	}

	creds := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: obj.Output.CredentialsRef.Name, Namespace: obj.Namespace}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, creds, func() error {
		for k, v := range fn.MapFilter(obj.GetLabels(), "kloudlite.io/") {
			fn.MapSet(&creds.Labels, k, v)
		}
		if creds.Data == nil {
			username := sanitizeDbName(obj.Name)
			password := fn.CleanerNanoid(40)

			dbName := sanitizeDbName(obj.Name)

			// INFO: secret does not exist yet
			out := &types.StandaloneDatabaseOutput{
				Username: username,
				Password: password,
				DbName:   dbName,
				Port:     so.Port,

				Host: so.Host,
				URI:  fmt.Sprintf("mysql://%s:%s@%s:%s/%s", username, password, so.Host, so.Port, dbName),
				DSN:  fmt.Sprintf("mysql://%s:%s@tcp(%s:%s)/%s", username, password, so.Host, so.Port, dbName),

				ClusterLocalHost: so.ClusterLocalHost,
				ClusterLocalURI:  fmt.Sprintf("mysql://%s:%s@%s:%s/%s", username, password, so.ClusterLocalHost, so.Port, dbName),
				ClusterLocalDSN:  fmt.Sprintf("mysql://%s:%s@tcp(%s:%s)/%s", username, password, so.ClusterLocalHost, so.Port, dbName),

				GlobalVPNHost: so.GlobalVPNHost,
				GlobalVPNURI:  fmt.Sprintf("mysql://%s:%s@%s:%s/%s", username, password, so.GlobalVPNHost, so.Port, dbName),
				GlobalVPNDSN:  fmt.Sprintf("mysql://%s:%s@tcp(%s:%s)/%s", username, password, so.GlobalVPNHost, so.Port, dbName),
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

	return check.Completed()
}

func (r *Reconciler) createDBUserLifecycle(req *rApi.Request[*mysqlv1.StandaloneDatabase]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(createDBCreationJob, req)

	msvc, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.MsvcRef.Namespace, obj.Spec.MsvcRef.Name), &mysqlv1.StandaloneService{})
	if err != nil {
		return check.Failed(err)
	}

	lf := &crdsv1.Lifecycle{ObjectMeta: metav1.ObjectMeta{Name: obj.Name, Namespace: obj.Namespace}}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, lf, func() error {
		for k, v := range fn.MapFilter(obj.Labels, "kloudlite.io/") {
			fn.MapSet(&lf.Labels, k, v)
		}
		lf.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})

		b, err := templates.ParseBytes(r.templateDBLifecycle, templates.DBLifecycleVars{
			NodeSelector:          map[string]string{},
			Tolerations:           []corev1.Toleration{},
			RootCredentialsSecret: msvc.Output.CredentialsRef.Name,
			NewCredentialsSecret:  obj.Output.CredentialsRef.Name,
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

	req.AddToOwnedResources(rApi.ParseResourceRef(lf))

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

	builder := ctrl.NewControllerManagedBy(mgr).For(&mysqlv1.StandaloneDatabase{})
	builder.Owns(&corev1.Secret{})

	watchList := []client.Object{
		&mysqlv1.StandaloneService{},
		&crdsv1.Lifecycle{},
	}

	for _, obj := range watchList {
		builder.Watches(
			obj, handler.EnqueueRequestsFromMapFunc(
				func(ctx context.Context, obj client.Object) []reconcile.Request {
					msvcName, ok := obj.GetLabels()[constants.MsvcNameKey]
					if !ok {
						return nil
					}

					var dbList mysqlv1.StandaloneDatabaseList
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
