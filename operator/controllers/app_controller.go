package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	crdsv1 "operators.kloudlite.io/api/v1"
	"operators.kloudlite.io/lib"
	"operators.kloudlite.io/lib/errors"
	reconcileResult "operators.kloudlite.io/lib/reconcile-result"
	_ "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// AppReconciler reconciles a App object
type AppReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	ClientSet *kubernetes.Clientset
	JobMgr    lib.Job
}

//+kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps/finalizers,verbs=update

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *AppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
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
			case "":
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
func (r *AppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crdsv1.App{}).
		Complete(r)
}
