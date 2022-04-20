package controllers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	_ "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/client-go/kubernetes"
	crdsv1 "operators.kloudlite.io/api/v1"
	"operators.kloudlite.io/lib"
	"operators.kloudlite.io/lib/errors"
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

const projectFinalizer = "finalizers.kloudlite.io/project"
const coolingTime = 5
const FieldManager = "kl-operator"

const ImagePullSecretName = "kloudlite-docker-registry"

var svcAccountName string = "hotspot-svc-account"

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
			return reconcileResult.Retry(minCoolingTime)
		}

		project.DefaultStatus()
		return r.updateStatus(ctx, project)
	}

	if project.Status.NamespaceCheck.ShouldCheck() {
		r.logger.Debugf("project.Status.NamespaceCheck.ShouldCheck()")
		project.Status.NamespaceCheck.SetStarted()

		tRef := &corev1.Namespace{}
		err := r.apply(ctx, tRef, &corev1.Namespace{
			TypeMeta:   tRef.TypeMeta,
			ObjectMeta: metav1.ObjectMeta{Name: project.Name},
		})
		if err != nil {
			project.Status.NamespaceCheck.SetFinishedWith(false, errors.NewEf(err, "while creating namespace").Error())
			return r.updateStatus(ctx, project)
		}
		project.Status.NamespaceCheck.SetFinishedWith(true, "namespace created")
		return r.updateStatus(ctx, project)
	}

	if !project.Status.NamespaceCheck.Status {
		r.logger.Infof("Namespace %s does not exist", project.Name)
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

		tRef := &corev1.Secret{}
		err = r.apply(ctx, tRef, &corev1.Secret{
			TypeMeta:   tRef.TypeMeta,
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

		// r.ClientSet.CoreV1().ServiceAccounts(project.Name).Create(ctx, &corev1.ServiceAccount{}, metav1.CreateOptions{})
		tRef := &corev1.ServiceAccount{}
		err := r.apply(ctx, tRef, &corev1.ServiceAccount{
			TypeMeta:         tRef.TypeMeta,
			ObjectMeta:       metav1.ObjectMeta{Name: svcAccountName, Namespace: project.Name},
			ImagePullSecrets: []corev1.LocalObjectReference{{Name: ImagePullSecretName}},
		})

		if err != nil {
			e := errors.NewEf(err, "could not apply service account")
			r.logger.Error(e)
			project.Status.SvcAccountCheck.SetFinishedWith(false, e.Error())
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
	if !controllerutil.ContainsFinalizer(project, projectFinalizer) {
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

	logger.Infof("Removing projectfinalizer (%s) ...", projectFinalizer)
	controllerutil.RemoveFinalizer(project, projectFinalizer)
	logger.Infof("finalizers: %+v", project.GetFinalizers())
	err := r.Update(ctx, project)
	if err != nil {
		return reconcileResult.RetryE(minCoolingTime, err)
	}
	return reconcileResult.OK()
}

func (r *ProjectReconciler) apply(ctx context.Context, typeRef client.Object, value client.Object) error {
	nameRef := types.NamespacedName{Namespace: value.GetNamespace(), Name: value.GetName()}
	err := r.Get(ctx, nameRef, typeRef)
	if err != nil {
		if apiErrors.IsNotFound(err) {
			// STEP: create
			err = r.Client.Create(ctx, value, &client.CreateOptions{})
			if err != nil {
				return errors.NewEf(err, "could not update resource %s (kind=%s)", nameRef.String(), value.GetObjectKind())
			}
			return nil
		}
		return errors.NewEf(err, "could not get resource %s (kind=%s)", nameRef.String(), value.GetObjectKind())
	}
	// STEP: Update
	err = r.Client.Update(ctx, value, &client.UpdateOptions{})
	if err != nil {
		return errors.NewEf(err, "could not update resource %s (kind=%s)", nameRef.String(), value.GetObjectKind())
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
