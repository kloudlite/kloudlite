package controllers

import (
	"context"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	crdsv1 "operators.kloudlite.io/api/v1"
	"operators.kloudlite.io/lib"
	"operators.kloudlite.io/lib/errors"
	"operators.kloudlite.io/lib/reconcile-result"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	_ "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ProjectReconciler reconciles a Project object
type ProjectReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	ClientSet *kubernetes.Clientset
	JobMgr    lib.Job
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

func GetLogger(args types.NamespacedName) *zap.SugaredLogger {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()
	return sugar.With(
		"NAME", args.String(),
	)
}

//+kubebuilder:rbac:groups=crds.kloudlite.io,resources=projects,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=crds.kloudlite.io,resources=projects/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=crds.kloudlite.io,resources=projects/finalizers,verbs=update

const projectFinalizer = "finalizers.kloudlite.io/project"
const coolingTime = 5

func (r *ProjectReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx).WithValues("resourceName:", req.Name, "namespace:", req.Namespace)
	logger := GetLogger(req.NamespacedName)

	project := &crdsv1.Project{}
	if err := r.Get(ctx, req.NamespacedName, project); err != nil {
		if apiErrors.IsNotFound(err) {
			// INFO: might have been deleted
			return reconcileResult.OK()
		}
		return reconcileResult.Failed()
	}

	logger.Infof("received update conditions: %+v", project.Status.Conditions)
	if project.IsNextGeneration() {
		return reconcileResult.Retry(coolingTime)
	}
	logger.Info("HERE")

	if project.IsReady() {
		logger.Info("HERE 2")
		if project.HasConditionsMet() {
			project.SetReady()
			err := r.Status().Update(ctx, project)
			if err != nil {
				return reconcileResult.RetryE(coolingTime, errors.ConditionUpdate(err))
			}
		}

		for _, condition := range project.ExpectedConditions() {
			switch condition {

			case "namespace-exists":
				{
					var ns corev1.Namespace
					if err := r.Get(ctx, types.NamespacedName{Name: project.Name}, &ns); err == nil {
						meta.SetStatusCondition(&project.Status.Conditions, metav1.Condition{
							Type:               "namespace-exists",
							Status:             metav1.ConditionTrue,
							ObservedGeneration: project.Generation,
							Reason:             "already-exists",
							Message:            "",
						})

						err = r.Status().Update(ctx, project)
						if err != nil {
							return reconcileResult.RetryE(10, errors.ConditionUpdate(err))
						}
						return reconcileResult.Retry(1)
					}

					_, err := r.ClientSet.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
						ObjectMeta: metav1.ObjectMeta{
							Name:       project.Name,
							Finalizers: []string{projectFinalizer},
						},
					}, metav1.CreateOptions{})
					logger.Debugf("error %+w", err)

					if err != nil {
						return reconcileResult.RetryE(coolingTime, errors.NewEf(err, "could not create namespace"))
					}

					meta.SetStatusCondition(&project.Status.Conditions, metav1.Condition{
						Type:               "namespace-exists",
						Status:             metav1.ConditionTrue,
						ObservedGeneration: project.Generation,
						Reason:             "done",
						Message:            "",
					})

					err = r.Status().Update(ctx, project)
					if err != nil {
						return reconcileResult.RetryE(10, errors.ConditionUpdate(err))
					}

					return reconcileResult.Retry(1)
				}
			}
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ProjectReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crdsv1.Project{}).
		Complete(r)
}
