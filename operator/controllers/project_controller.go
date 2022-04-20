package controllers

import (
	"context"
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
	Scheme      *runtime.Scheme
	ClientSet   *kubernetes.Clientset
	SendMessage func(key string, msg lib.MessageReply) error
	JobMgr      lib.Job
	logger      *zap.SugaredLogger
}

const (
	TypeSucceeded  = "Succeeded"
	TypeFailed     = "Failed"
	TypeInProgress = "InProgress"
)

type Bool bool

//+kubebuilder:rbac:groups=crds.kloudlite.io,resources=projects,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=crds.kloudlite.io,resources=projects/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=crds.kloudlite.io,resources=projects/finalizers,verbs=update

const projectFinalizer = "finalizers.kloudlite.io/project"
const coolingTime = 5

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

		// does exist
		var ns corev1.Namespace
		err := r.Get(ctx, types.NamespacedName{Name: project.Name}, &ns)
		if err != nil {
			if apiErrors.IsNotFound(err) {
				_, err = r.ClientSet.CoreV1().Namespaces().Create(
					ctx,
					&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: project.Name}},
					metav1.CreateOptions{},
				)
				if err != nil {
					project.Status.NamespaceCheck.SetFinishedWith(false, errors.NewEf(err, "while creating namespace").Error())
					return r.updateStatus(ctx, project)
				}
				project.Status.NamespaceCheck.SetFinishedWith(true, "namespace created")
			}
		}

		_, err = r.ClientSet.CoreV1().Namespaces().Update(
			ctx,
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: project.Name}},
			metav1.UpdateOptions{},
		)
		if err != nil {
			project.Status.NamespaceCheck.SetFinishedWith(false, errors.NewEf(err, "while creating namespace").Error())
			return r.updateStatus(ctx, project)
		}
		project.Status.NamespaceCheck.SetFinishedWith(true, "namespace created")
		return r.updateStatus(ctx, project)
	}

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
	if !controllerutil.ContainsFinalizer(project, projectFinalizer) {
		return reconcileResult.OK()
	}

	if project.Status.DelNamespaceCheck.ShouldCheck() {
		project.Status.DelNamespaceCheck.SetStarted()
		if err := r.ClientSet.CoreV1().Namespaces().Delete(ctx, project.Name, metav1.DeleteOptions{}); err != nil {
			if apiErrors.IsNotFound(err) {
				project.Status.DelNamespaceCheck.SetFinishedWith(true, errors.NewEf(err, "no namespace found to be deleted").Error())
				return r.updateStatus(ctx, project)
			}
			project.Status.DelNamespaceCheck.SetFinishedWith(false, errors.NewEf(err, "could not delete namespace").Error())
			return r.updateStatus(ctx, project)
		}
	}

	if !project.Status.DelNamespaceCheck.Status {
		time.AfterFunc(time.Second*semiCoolingTime, func() {
			project.Status.DelNamespaceCheck = crdsv1.Recon{}
			r.Status().Update(ctx, project)
		})
		return reconcileResult.OK()
	}

	controllerutil.RemoveFinalizer(project, projectFinalizer)
	r.logger.Infof("finalizers: %+v", project.GetFinalizers())
	err := r.Update(ctx, project)
	if err != nil {
		return reconcileResult.RetryE(minCoolingTime, err)
	}
	return reconcileResult.OK()
}

// SetupWithManager sets up the controller with the Manager.
func (r *ProjectReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crdsv1.Project{}).
		Owns(&corev1.Namespace{}).
		Complete(r)
}
