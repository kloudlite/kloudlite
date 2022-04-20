package controllers

import (
	"context"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	_ "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	// applyCoreV1 "k8s.io/client-go/applyconfigurations/core/v1"
	// applyMetaV1 "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/client-go/kubernetes"
	crdsv1 "operators.kloudlite.io/api/v1"
	"operators.kloudlite.io/lib"
	"operators.kloudlite.io/lib/errors"
	reconcileResult "operators.kloudlite.io/lib/reconcile-result"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
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

func (b Bool) Condition() metav1.ConditionStatus {
	if !b {
		return metav1.ConditionFalse
	}
	return metav1.ConditionTrue
}

func GetLogger(_ types.NamespacedName) *zap.SugaredLogger {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()
	return sugar
	// return sugar.With(
	// 	"NAME", args.String(),
	// )
}

//+kubebuilder:rbac:groups=crds.kloudlite.io,resources=projects,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=crds.kloudlite.io,resources=projects/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=crds.kloudlite.io,resources=projects/finalizers,verbs=update

const projectFinalizer = "finalizers.kloudlite.io/project"
const coolingTime = 5

func (r *ProjectReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx).WithValues("resourceName:", req.Name, "namespace:", req.Namespace)
	logger := GetLogger(req.NamespacedName)
	r.logger = logger

	project := &crdsv1.Project{}
	if err := r.Get(ctx, req.NamespacedName, project); err != nil {
		if apiErrors.IsNotFound(err) {
			// INFO: might have been deleted
			return reconcileResult.OK()
		}
		return reconcileResult.Failed()
	}

	if project.GetDeletionTimestamp() != nil {
		logger.Debug("has deletiontimestamp")
		return r.finalizeProject(ctx, project)
	}

	if project.IsNewGeneration() || project.Status.Namespace != project.Name {
		logger.Infof("Status.Namespace %v project.Name %v", project.Status.Namespace, project.Name)
		logger.Debugf("project.Status.Namespace != project.Name")

		// ASSERT: On Updating
		if ns, ok := r.namespaceExists(ctx, project.Name); ok {
			_, err := r.ClientSet.CoreV1().Namespaces().Update(ctx, ns, metav1.UpdateOptions{})
			if err != nil {
				return reconcileResult.RetryE(2, errors.NewEf(err, "could not update namspace %v", ns.Name))
			}
			logger.Debugf("namespace %v has been updated", ns.Name)
			return r.updateStatus(ctx, project)
		}

		// ASSERT: New Creation
		_, err := r.ClientSet.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: project.Name,
			},
		}, metav1.CreateOptions{})

		if err != nil {
			return reconcileResult.RetryE(maxCoolingTime, errors.NewEf(err, "could not create namespace"))
		}
		logger.Infof("created namespace (%s)\n", project.Name)
		project.Status.Namespace = project.Name
		return r.updateStatus(ctx, project)
	}

	return reconcileResult.OK()
}

func (r *ProjectReconciler) updateStatus(ctx context.Context, project *crdsv1.Project) (ctrl.Result, error) {
	project.Status.Generation = project.Generation
	err := r.Status().Update(ctx, project)
	if err != nil {
		return reconcileResult.RetryE(2, errors.StatusUpdate(err))
	}
	r.logger.Debugf("project (name=%s) has been updated", project.Name)
	r.SendMessage("KEY", lib.MessageReply{
		Message: "HHII",
		Status:  true,
	})
	return reconcileResult.OK()
}

func (r *ProjectReconciler) finalizeProject(ctx context.Context, project *crdsv1.Project) (ctrl.Result, error) {
	logger := GetLogger(types.NamespacedName{Name: project.Name})

	if controllerutil.ContainsFinalizer(project, projectFinalizer) {
		if err := r.ClientSet.CoreV1().Namespaces().Delete(ctx, project.Name, metav1.DeleteOptions{}); err != nil {
			return reconcileResult.RetryE(2, errors.NewEf(err, "could not delete namespace"))
		}

		controllerutil.RemoveFinalizer(project, projectFinalizer)
		logger.Infof("finalizers: %+v", project.GetFinalizers())
		err := r.Update(ctx, project)
		if err != nil {
			return reconcileResult.RetryE(maxCoolingTime, errors.Newf("finalized project (name=%s)", project.Name))
		}
		return reconcileResult.OK()
	}

	return reconcileResult.OK()
}

func (r *ProjectReconciler) namespaceExists(ctx context.Context, namespace string) (*corev1.Namespace, bool) {
	var ns corev1.Namespace
	err := r.Get(ctx, types.NamespacedName{Name: namespace}, &ns)
	return &ns, err == nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ProjectReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crdsv1.Project{}).
		Owns(&corev1.Namespace{}).
		Complete(r)
}
