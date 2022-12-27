package project

import (
	"context"
	"encoding/json"
	"fmt"
	"operators.kloudlite.io/pkg/templates"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	artifactsv1 "operators.kloudlite.io/apis/artifacts/v1"
	"operators.kloudlite.io/apis/crds/v1"
	"operators.kloudlite.io/operators/project/internal/env"
	"operators.kloudlite.io/pkg/constants"
	fn "operators.kloudlite.io/pkg/functions"
	"operators.kloudlite.io/pkg/harbor"
	"operators.kloudlite.io/pkg/kubectl"
	"operators.kloudlite.io/pkg/logging"
	rApi "operators.kloudlite.io/pkg/operator"
	stepResult "operators.kloudlite.io/pkg/operator/step-result"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	harborCli  *harbor.Client
	logger     logging.Logger
	Name       string
	Env        *env.Env
	yamlClient *kubectl.YAMLClient
	IsDev      bool
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	BlueprintCreated      string = "blueprint-created"
	DefaultEnvCreated     string = "default-env-created"
	HarborAccessAvailable string = "harbor-creds-available"
	NamespacedRBACsReady  string = "namespaced-rbacs-ready"
)

func getBlueprintName(projName string) string {
	return projName + "-blueprint"
}

func getDefaultEnvName(projName string) string {
	return projName + "-default"
}

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=projects,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=projects/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=projects/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, request.NamespacedName, &v1.Project{})
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

	// TODO: initialize all checks here
	if step := req.EnsureChecks(BlueprintCreated, DefaultEnvCreated, HarborAccessAvailable); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureBlueprint(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureNamespacedRBACs(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureDefaultEnv(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconHarborAccess(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	req.Object.Status.LastReconcileTime = metav1.Time{Time: time.Now()}
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) finalize(req *rApi.Request[*v1.Project]) stepResult.Result {
	return req.Finalize()
}

func (r *Reconciler) ensureBlueprint(req *rApi.Request[*v1.Project]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(BlueprintCreated)
	defer req.LogPostCheck(BlueprintCreated)

	blueprintNs := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-blueprint", obj.Name)}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, blueprintNs, func() error {
		if !fn.IsOwner(blueprintNs, fn.AsOwner(obj)) {
			blueprintNs.SetOwnerReferences(append(blueprintNs.GetOwnerReferences(), fn.AsOwner(obj, true)))
		}

		if blueprintNs.ObjectMeta.Labels == nil {
			blueprintNs.ObjectMeta.Labels = make(map[string]string, 4)
		}

		blueprintNs.ObjectMeta.Labels[constants.ProjectNameKey] = obj.Name
		blueprintNs.ObjectMeta.Labels[constants.AccountRef] = obj.Spec.AccountRef
		blueprintNs.ObjectMeta.Labels[constants.IsBluePrintKey] = "true"
		blueprintNs.ObjectMeta.Labels[constants.MarkedAsBlueprint] = "true"
		return nil
	}); err != nil {
		return req.CheckFailed(BlueprintCreated, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != checks[BlueprintCreated] {
		checks[BlueprintCreated] = check
		return req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureNamespacedRBACs(req *rApi.Request[*v1.Project]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(NamespacedRBACsReady)
	defer req.LogPreCheck(NamespacedRBACsReady)

	b, err := templates.Parse(
		templates.ProjectRBAC, map[string]any{
			"namespace":          getBlueprintName(obj.Name),
			"role-name":          r.Env.AdminRoleName,
			"role-binding-name":  r.Env.AdminRoleName + "-rb",
			"svc-account-name":   r.Env.SvcAccountName,
			"docker-secret-name": r.Env.DockerSecretName,
			"owner-refs":         []metav1.OwnerReference{fn.AsOwner(obj, true)},
		},
	)
	if err != nil {
		return req.CheckFailed(NamespacedRBACsReady, check, err.Error()).Err(nil)
	}

	if err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
		return req.CheckFailed(NamespacedRBACsReady, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != checks[NamespacedRBACsReady] {
		checks[NamespacedRBACsReady] = check
		return req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) ensureDefaultEnv(req *rApi.Request[*v1.Project]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(DefaultEnvCreated)
	defer req.LogPostCheck(DefaultEnvCreated)

	defaultEnv := &v1.Env{ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-default", obj.Name)}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, defaultEnv, func() error {
		if !fn.IsOwner(defaultEnv, fn.AsOwner(obj)) {
			defaultEnv.SetOwnerReferences(append(defaultEnv.GetOwnerReferences(), fn.AsOwner(obj, true)))
		}
		defaultEnv.Spec = v1.EnvSpec{
			ProjectName:   obj.Name,
			BlueprintName: obj.Name + "-blueprint",
			AccountRef:    obj.Spec.AccountRef,
			Primary:       true,
		}
		return nil
	}); err != nil {
		return req.CheckFailed(DefaultEnvCreated, check, err.Error())
	}

	check.Status = true
	if check != checks[DefaultEnvCreated] {
		checks[DefaultEnvCreated] = check
		return req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) reconHarborAccess(req *rApi.Request[*v1.Project]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(HarborAccessAvailable)
	defer req.LogPostCheck(HarborAccessAvailable)

	harborProject := &artifactsv1.HarborProject{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.AccountRef}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, harborProject, func() error {
		if harborProject.Labels == nil {
			harborProject.Labels = make(map[string]string, 1)
		}
		harborProject.Labels[constants.AccountRef] = obj.Spec.AccountRef
		return nil
	}); err != nil {
		return req.CheckFailed(HarborAccessAvailable, check, err.Error())
	}

	harborUserAcc := &artifactsv1.HarborUserAccount{ObjectMeta: metav1.ObjectMeta{Name: r.Env.DockerSecretName, Namespace: getBlueprintName(obj.Name)}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, harborUserAcc, func() error {
		if !fn.IsOwner(harborUserAcc, fn.AsOwner(obj)) {
			harborUserAcc.SetOwnerReferences(append(harborUserAcc.OwnerReferences, fn.AsOwner(obj, true)))
		}
		harborUserAcc.Spec.ProjectRef = harborProject.Name
		return nil
	}); err != nil {
		return req.CheckFailed(HarborAccessAvailable, check, err.Error()).Err(nil)
	}

	if !harborProject.Status.IsReady {
		bMessage, err := json.Marshal(harborProject.Status.Message)
		if err != nil {
			return req.CheckFailed(HarborAccessAvailable, check, err.Error()).Err(nil)
		}
		return req.CheckFailed(HarborAccessAvailable, check, string(bMessage)).Err(nil)
	}

	if !harborUserAcc.Status.IsReady {
		bMessage, err := json.Marshal(harborUserAcc.Status.Message)
		if err != nil {
			return req.CheckFailed(HarborAccessAvailable, check, err.Error()).Err(nil)
		}
		return req.CheckFailed(HarborAccessAvailable, check, string(bMessage)).Err(nil)
	}

	check.Status = true
	if check != checks[HarborAccessAvailable] {
		checks[HarborAccessAvailable] = check
		return req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

	builder := ctrl.NewControllerManagedBy(mgr).For(&v1.Project{})
	builder.Owns(&corev1.Namespace{})
	builder.Owns(&corev1.ServiceAccount{})
	builder.Owns(&rbacv1.Role{})
	builder.Owns(&rbacv1.RoleBinding{})
	builder.Owns(&artifactsv1.HarborUserAccount{})

	builder.WithEventFilter(rApi.ReconcileFilter())

	return builder.Complete(r)
}
