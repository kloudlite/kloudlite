package backup

import (
	"context"
	"fmt"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	mongov1 "github.com/kloudlite/operator/apis/mongodb.msvc/v1"
	"github.com/kloudlite/operator/operators/msvc-mongo/internal/env"
	"github.com/kloudlite/operator/operators/msvc-mongo/internal/templates"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	batchv1 "k8s.io/api/batch/v1"
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
	patchDefaults   string = "patch-defaults"
	createBackupJob string = "create-backup-job"
)

// +kubebuilder:rbac:groups=mysql.msvc.kloudlite.io,resources=databases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mysql.msvc.kloudlite.io,resources=databases/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mysql.msvc.kloudlite.io,resources=databases/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &mongov1.Backup{})
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
		{Name: createBackupJob, Title: "Create DB Creation Job"},
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

	if step := r.createBackupLifecycle(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *Reconciler) finalize(req *rApi.Request[*mongov1.Backup]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck("finalizing", req)

	ss, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.MsvcRef.Namespace, obj.Spec.MsvcRef.Name), &mongov1.StandaloneService{})
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
		return check.Failed(err)
	}

	return req.Finalize()
}

func (r *Reconciler) createBackupLifecycle(req *rApi.Request[*mongov1.Backup]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(createBackupJob, req)

	msvc, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.MsvcRef.Namespace, obj.Spec.MsvcRef.Name), &mongov1.StandaloneService{})
	if err != nil {
		return check.Failed(err)
	}

	if !msvc.Status.IsReady {
		return check.Failed(fmt.Errorf("waiting for service to be ready"))
	}

	/*
	  TODO: create a job/cronjob that creates a backup
	*/

	switch obj.Spec.BackupType {
	case mongov1.BackupTypeOneShot:
		{
		}
		// return r.createOneShotBackup(req)
	case mongov1.BackupTypeCron:
		{
			if obj.Spec.Cron == nil {
				return check.Failed(fmt.Errorf("cron backup type requires .spec.cron to be set"))
			}

			cron := &batchv1.CronJob{ObjectMeta: metav1.ObjectMeta{Name: obj.Name, Namespace: obj.Namespace}}
			if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, cron, func() error {
				lb := cron.GetLabels()
				for k, v := range fn.MapFilter(obj.GetLabels(), "kloudlite.io/") {
					lb[k] = v
				}
				cron.SetLabels(lb)
				cron.SetOwnerReferences(append(cron.GetOwnerReferences(), fn.AsOwner(obj, true)))

				cron.Spec.Schedule = obj.Spec.Cron.Schedule
				cron.Spec.JobTemplate = batchv1.JobTemplateSpec{
					Spec: batchv1.JobSpec{
						Parallelism:      fn.New(int32(1)),
						PodFailurePolicy: &batchv1.PodFailurePolicy{},
						BackoffLimit:     fn.New(int32(1)),
						Template: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{},
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name:       "backup",
										Image:      "mongo:latest",
										Command:    []string{
										  "sh",
										  "-c",
										  "mongodump --host $MONGODB_HOST --port $MONGODB_PORT --username $MONGODB_USERNAME --password $MONGODB_PASSWORD --authenticationDatabase $MONGODB_AUTHENTICATION_DATABASE --db $MONGODB_DATABASE --out /backups/$BACKUP_NAME"
										},
										Args:       []string{},
										WorkingDir: "",
										Ports:      []corev1.ContainerPort{},
										EnvFrom:    []corev1.EnvFromSource{},
									},
								},
							},
						},
					},
				}

				return nil
			}); err != nil {
				return check.Failed(err)
			}
		}
	}

	lf := &crdsv1.Lifecycle{ObjectMeta: metav1.ObjectMeta{Name: obj.Name, Namespace: obj.Namespace}}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, lf, func() error {
		lf.SetOwnerReferences(append(lf.GetOwnerReferences(), fn.AsOwner(obj, true)))

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

	builder := ctrl.NewControllerManagedBy(mgr).For(&mongov1.Backup{})
	builder.Owns(&corev1.Secret{})
	builder.Owns(&crdsv1.Lifecycle{})

	watchList := []client.Object{
		&mongov1.StandaloneService{},
	}

	for _, obj := range watchList {
		builder.Watches(
			obj, handler.EnqueueRequestsFromMapFunc(
				func(ctx context.Context, obj client.Object) []reconcile.Request {
					msvcName, ok := obj.GetLabels()[constants.MsvcNameKey]
					if !ok {
						return nil
					}

					var dbList mongov1.BackupList
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
	return builder.Complete(r)
}
