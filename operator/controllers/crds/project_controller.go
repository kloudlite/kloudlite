package crds

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/constants"
	fn "operators.kloudlite.io/lib/functions"
	rApi "operators.kloudlite.io/lib/operator"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"operators.kloudlite.io/lib"
)

// ProjectReconciler reconciles a Project object
type ProjectReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	lib.MessageSender
	HarborUserName string
	HarborPassword string
}

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=projects,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=projects/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=projects/finalizers,verbs=update

func (r *ProjectReconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	req := rApi.NewRequest(ctx, r.Client, oReq.NamespacedName, &crdsv1.Project{})

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

func (r *ProjectReconciler) finalize(req *rApi.Request[*crdsv1.Project]) rApi.StepResult {
	return req.Finalize()
}

func (r *ProjectReconciler) reconcileStatus(req *rApi.Request[*crdsv1.Project]) rApi.StepResult {
	ctx := req.Context()
	project := req.Object

	isReady := true
	var cs []metav1.Condition

	ns, err := rApi.Get(ctx, r.Client, fn.NN(project.Namespace, project.Name), &corev1.Namespace{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.FailWithStatusError(err)
		}
		isReady = false
		cs = append(cs, conditions.New("NamespaceExists", false, "NotFound", err.Error()))
		ns = nil
	}

	if ns != nil {
		cs = append(cs, conditions.New("NamespaceExists", true, "Found"))
	}

	newConditions, hasUpdated, err := conditions.Patch(project.Status.Conditions, cs)
	if err != nil {
		return req.FailWithStatusError(err)
	}
	if !hasUpdated && isReady == project.Status.IsReady {
		return req.Next()
	}

	project.Status.IsReady = isReady
	project.Status.Conditions = newConditions
	if err := r.Status().Update(ctx, project); err != nil {
		return req.FailWithOpError(err)
	}

	return req.Done()
}

func (r *ProjectReconciler) reconcileOperations(req *rApi.Request[*crdsv1.Project]) rApi.StepResult {
	ctx := req.Context()
	project := req.Object

	if !controllerutil.ContainsFinalizer(project, constants.CommonFinalizer) ||
		!controllerutil.ContainsFinalizer(project, constants.ForegroundFinalizer) {
		controllerutil.AddFinalizer(project, constants.CommonFinalizer)
		controllerutil.AddFinalizer(project, constants.ForegroundFinalizer)
		if err := r.Update(ctx, project); err != nil {
			return req.FailWithOpError(err)
		}
		return req.Done()
	}

	if err := fn.KubectlApply(
		ctx, r.Client, &corev1.Namespace{
			TypeMeta:   metav1.TypeMeta{Kind: "Namespace", APIVersion: "v1"},
			ObjectMeta: metav1.ObjectMeta{Name: project.Name},
		},
	); err != nil {
		return req.FailWithOpError(err)
	}

	encAuthPass := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", r.HarborUserName, r.HarborPassword)))
	dockerConfigJson, err := json.Marshal(
		map[string]interface{}{
			"auths": map[string]interface{}{
				ImageRegistry: map[string]interface{}{
					"auth": encAuthPass,
				},
			},
		},
	)
	if err != nil {
		return req.FailWithOpError(err)
	}

	imgPullSecret := corev1.Secret{
		// TypeMeta: TypeSecret,
		ObjectMeta: metav1.ObjectMeta{
			Namespace: project.Name,
			Name:      ImagePullSecretName,
			OwnerReferences: []metav1.OwnerReference{
				fn.AsOwner(project, true),
			},
		},
		Data: map[string][]byte{
			".dockerconfigjson": dockerConfigJson,
		},
		Type: corev1.SecretTypeDockerConfigJson,
	}

	if err := fn.KubectlApply(ctx, r.Client, &imgPullSecret); err != nil {
		return req.FailWithOpError(err)
	}

	adminRole := rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: project.Name,
			Name:      NamespaceAdminRole,
			OwnerReferences: []metav1.OwnerReference{
				fn.AsOwner(project, true),
			},
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"", "extensions", "apps"},
				Resources: []string{"*"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{"batch"},
				Resources: []string{"jobs", "cronjobs"},
				Verbs:     []string{"*"},
			},
		},
	}

	if err := fn.KubectlApply(ctx, r.Client, &adminRole); err != nil {
		return req.FailWithOpError(err)
	}

	adminRoleBinding := rbacv1.RoleBinding{
		TypeMeta: TypeRoleBinding,
		ObjectMeta: metav1.ObjectMeta{
			Namespace: project.Name,
			Name:      NamespaceAdminRoleBinding,
			OwnerReferences: []metav1.OwnerReference{
				fn.AsOwner(project, true),
			},
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      TypeSvcAccount.Kind,
				APIGroup:  "",
				Name:      SvcAccountName,
				Namespace: project.Name,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "",
			Kind:     TypeRole.Kind,
			Name:     NamespaceAdminRole,
		},
	}

	if err := fn.KubectlApply(ctx, r.Client, &adminRoleBinding); err != nil {
		return req.FailWithOpError(err)
	}

	svcAccount := corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kloudlite-svc-account",
			Namespace: project.Name,
			OwnerReferences: []metav1.OwnerReference{
				fn.AsOwner(project, true),
			},
		},
		ImagePullSecrets: []corev1.LocalObjectReference{
			{
				Name: ImagePullSecretName,
			},
		},
	}

	if err := fn.KubectlApply(ctx, r.Client, &svcAccount); err != nil {
		return req.FailWithOpError(err)
	}

	return req.Done()
}

func (r *ProjectReconciler) notify(req *rApi.Request[*crdsv1.Project]) error {
	project := req.Object

	return nil

	return r.SendMessage(
		project.LogRef(), lib.MessageReply{
			Key:        project.LogRef(),
			Conditions: project.Status.Conditions,
			Status:     project.Status.IsReady,
		},
	)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ProjectReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crdsv1.Project{}).
		Owns(&corev1.Namespace{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&corev1.Secret{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.RoleBinding{}).
		Complete(r)
}
