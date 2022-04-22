package controllers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/client-go/kubernetes"
	crdsv1 "operators.kloudlite.io/api/v1"
	"operators.kloudlite.io/lib"
	"operators.kloudlite.io/lib/errors"
	"operators.kloudlite.io/lib/finalizers"
	reconcileResult "operators.kloudlite.io/lib/reconcile-result"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// ProjectReconciler reconciles a Project object
type ProjectReconciler struct {
	client.Client
	Scheme         *runtime.Scheme
	ClientSet      *kubernetes.Clientset
	SendMessage    func(key string, msg lib.MessageReply) error
	JobMgr         lib.Job
	logger         *zap.SugaredLogger
	HarborUserName string
	HarborPassword string
}

const (
	TypeSucceeded  = "Succeeded"
	TypeFailed     = "Failed"
	TypeInProgress = "InProgress"
)

//+kubebuilder:rbac:groups=crds.kloudlite.io,resources=projects,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=crds.kloudlite.io,resources=projects/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=crds.kloudlite.io,resources=projects/finalizers,verbs=update

const ImagePullSecretName = "kloudlite-docker-registry"

func (r *ProjectReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.logger = GetLogger(req.NamespacedName)

	project := &crdsv1.Project{}
	if err := r.Get(ctx, req.NamespacedName, project); err != nil {
		if apiErrors.IsNotFound(err) {
			// INFO: might have been deleted
			return reconcileResult.OK()
		}
		return reconcileResult.Failed()
	}

	if project.HasToBeDeleted() {
		r.logger.Debugf("project.HasToBeDeleted()")
		return r.finalizeProject(ctx, project)
	}

	if project.IsNewGeneration() {
		r.logger.Debugf("project.IsNewGeneration()")
		if project.Status.NamespaceCheck.IsRunning() {
			return reconcileResult.Retry()
		}

		project.DefaultStatus()
		return r.updateStatus(ctx, project)
	}

	if project.Status.NamespaceCheck.ShouldCheck() {
		r.logger.Debugf("project.Status.NamespaceCheck.ShouldCheck()")
		project.Status.NamespaceCheck.SetStarted()

		ns := corev1.Namespace{
			TypeMeta: TypeNamespace,
			ObjectMeta: metav1.ObjectMeta{
				Name: project.Name,
			},
		}

		controllerutil.CreateOrUpdate(ctx, r.Client, &ns, func() error {
			return nil
		})
		err := r.apply(ctx, &corev1.Namespace{}, &corev1.Namespace{
			TypeMeta: TypeNamespace,
			ObjectMeta: metav1.ObjectMeta{
				Name: project.Name,
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion:         project.APIVersion,
						Kind:               project.Kind,
						Name:               project.Name,
						UID:                project.UID,
						Controller:         newBool(true),
						BlockOwnerDeletion: newBool(true),
					},
				},
			},
		})

		if err != nil {
			project.Status.NamespaceCheck.SetFinishedWith(false, errors.NewEf(err, "while creating namespace").Error())
			return r.updateStatus(ctx, project)
		}
		project.Status.NamespaceCheck.SetFinishedWith(true, "namespace created")
		return r.updateStatus(ctx, project)
	}

	if !project.Status.NamespaceCheck.Status {
		r.logger.Infof("Namespace %s does not exist, aborting reconcilation ...", project.Name)
		return reconcileResult.Failed()
	}

	if project.Status.PullSecretCheck.ShouldCheck() {
		project.Status.PullSecretCheck.SetStarted()
		encAuthPass := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", r.HarborUserName, r.HarborPassword)))

		dockerConfigJson, err := json.Marshal(map[string]interface{}{
			"auths": map[string]interface{}{
				ImageRegistry: map[string]interface{}{
					"auth": encAuthPass,
				},
			},
		})

		if err != nil {
			e := errors.New("could not decode into dockerconfigjson file format")
			r.logger.Error(e)
			project.Status.PullSecretCheck.SetFinishedWith(false, e.Error())
			return r.updateStatus(ctx, project)
		}

		err = r.apply(ctx, &corev1.Secret{}, &corev1.Secret{
			TypeMeta:   TypeSecret,
			ObjectMeta: metav1.ObjectMeta{Name: ImagePullSecretName, Namespace: project.Name},
			Data:       map[string][]byte{".dockerconfigjson": dockerConfigJson},
			Type:       corev1.SecretTypeDockerConfigJson,
		})

		if err != nil {
			e := errors.NewEf(err, "could not create image pull secret")
			r.logger.Error(e)
			project.Status.PullSecretCheck.SetFinishedWith(false, e.Error())
			return r.updateStatus(ctx, project)
		}
		project.Status.PullSecretCheck.SetFinishedWith(true, "image pull secret has been created")
		return r.updateStatus(ctx, project)
	}

	if !project.Status.PullSecretCheck.Status {
		r.logger.Debug("Image pull secret could not been created ..., aborting")
		return reconcileResult.Failed()
	}

	if project.Status.SvcAccountCheck.ShouldCheck() {
		project.Status.SvcAccountCheck.SetStarted()

		err := r.apply(ctx, &rbacv1.Role{}, &rbacv1.Role{
			TypeMeta: TypeRole,
			ObjectMeta: metav1.ObjectMeta{
				Namespace: project.Name,
				Name:      NamespaceAdminRole,
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
		})
		if err != nil {
			project.Status.SvcAccountCheck.SetFinishedWith(false, errors.NewEf(err, "while creating admin role").Error())
			return r.updateStatus(ctx, project)
		}

		err = r.apply(ctx, &rbacv1.RoleBinding{}, &rbacv1.RoleBinding{
			TypeMeta: TypeRoleBinding,
			ObjectMeta: metav1.ObjectMeta{
				Namespace: project.Name,
				Name:      NamespaceAdminRoleBinding,
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
		})

		if err != nil {
			project.Status.SvcAccountCheck.SetFinishedWith(false, errors.NewEf(err, "while creating admin role binding").Error())
			return r.updateStatus(ctx, project)
		}

		err = r.apply(ctx, &corev1.ServiceAccount{}, &corev1.ServiceAccount{
			TypeMeta:         TypeSvcAccount,
			ObjectMeta:       metav1.ObjectMeta{Name: SvcAccountName, Namespace: project.Name},
			ImagePullSecrets: []corev1.LocalObjectReference{{Name: ImagePullSecretName}},
		})

		if err != nil {
			project.Status.SvcAccountCheck.SetFinishedWith(false, errors.NewEf(err, "while creating service account").Error())
			return r.updateStatus(ctx, project)
		}

		project.Status.SvcAccountCheck.SetFinishedWith(true, "svc account created")
		return r.updateStatus(ctx, project)
	}

	r.logger.Infof("Project (%s) Reconcile Successful\n", req.NamespacedName)
	return reconcileResult.OK()
}

func (r *ProjectReconciler) updateStatus(ctx context.Context, project *crdsv1.Project) (ctrl.Result, error) {
	project.BuildConditions()

	b, err := json.Marshal(project.Status.Conditions)
	if err != nil {
		r.logger.Debug(err)
		b = []byte("")
	}

	err = r.SendMessage(fmt.Sprintf("%s/%s/%s", project.Namespace, "project", project.Name), lib.MessageReply{
		Message: string(b),
		Status:  meta.FindStatusCondition(project.Status.Conditions, "Ready").Status == metav1.ConditionTrue,
	})

	if err != nil {
		r.logger.Infof("unable to send kafka reply message")
	}

	if err = r.Status().Update(ctx, project); err != nil {
		return reconcileResult.RetryE(2, errors.StatusUpdate(err))
	}
	r.logger.Debugf("project (name=%s) has been updated", project.Name)

	return reconcileResult.OK()
}

func (r *ProjectReconciler) finalizeProject(ctx context.Context, project *crdsv1.Project) (ctrl.Result, error) {
	logger := r.logger.With("FINALIZER", "true")
	logger.Debug("finalizing ...")
	if !controllerutil.ContainsFinalizer(project, finalizers.Project.String()) {
		if controllerutil.ContainsFinalizer(project, foregroundFinalizer) {

			//TODO: check if namespace has been deleted
			controllerutil.RemoveFinalizer(project, foregroundFinalizer)
			err := r.Update(ctx, project)
			if err != nil {
				r.logger.Debugf("could not update to remove foreground finalizer from resource")
				return reconcileResult.Retry()
			}
			r.logger.Infof("Removing foreground finalizer , Deletion successfull...")
			return reconcileResult.OK()
		}
		return reconcileResult.OK()
	}

	if project.Status.DelNamespaceCheck.ShouldCheck() {
		logger.Infof("project.Status.DelNamespaceCheck.ShouldCheck()")
		project.Status.DelNamespaceCheck.SetStarted()
		if err := r.ClientSet.CoreV1().Namespaces().Delete(ctx, project.Name, metav1.DeleteOptions{}); err != nil {
			logger.Infof(err.Error())
			if apiErrors.IsNotFound(err) {
				project.Status.DelNamespaceCheck.SetFinishedWith(true, errors.NewEf(err, "no namespace found to be deleted").Error())
				return r.updateStatus(ctx, project)
			}
			project.Status.DelNamespaceCheck.SetFinishedWith(false, errors.NewEf(err, "could not delete namespace").Error())
			return r.updateStatus(ctx, project)
		}
		project.Status.DelNamespaceCheck.SetFinishedWith(true, "namespace deleted")
		return r.updateStatus(ctx, project)
	}

	if !project.Status.DelNamespaceCheck.Status {
		logger.Debug("!project.Status.DelNamespaceCheck.ShouldCheck()")
		time.AfterFunc(time.Second*semiCoolingTime, func() {
			project.Status.DelNamespaceCheck = crdsv1.Recon{}
			_, err := r.updateStatus(ctx, project)
			if err != nil {
				logger.Error(err)
			}
		})
		return reconcileResult.OK()
	}

	logger.Infof("Removing projectfinalizer (%s) ...", finalizers.Project.String())
	controllerutil.RemoveFinalizer(project, finalizers.Project.String())
	logger.Infof("finalizers: %+v", project.GetFinalizers())
	err := r.Update(ctx, project)
	if err != nil {
		return reconcileResult.RetryE(minCoolingTime, err)
	}
	return reconcileResult.OK()
}

func (r *ProjectReconciler) apply(ctx context.Context, resourceRef client.Object, value client.Object) error {
	nameRef := types.NamespacedName{Namespace: value.GetNamespace(), Name: value.GetName()}
	err := r.Get(ctx, nameRef, resourceRef)
	if err != nil {
		if apiErrors.IsNotFound(err) {
			// STEP: create
			err = r.Client.Create(ctx, value, &client.CreateOptions{})
			if err != nil {
				return errors.NewEf(err, "could not update resource %s", toRefString(value))
			}
			return nil
		}
		return errors.NewEf(err, "could not get resource %s", toRefString(value))
	}
	// STEP: Update , but first check if it is in deletion phase then pause,
	if resourceRef.GetDeletionTimestamp() != nil {
		return errors.Newf("could not update resource(%s) as it has deletion timestamp on it, wait for it to be deleted first...", toRefString(resourceRef))
	}
	err = r.Client.Update(ctx, value, &client.UpdateOptions{})
	if err != nil {
		return errors.NewEf(err, "could not update resource %s", toRefString(value))
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ProjectReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crdsv1.Project{}).
		Owns(&corev1.Namespace{}).
		Complete(r)
}
