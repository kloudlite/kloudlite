package controllers

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels2 "k8s.io/apimachinery/pkg/labels"
	fn "operators.kloudlite.io/lib/functions"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	crdsv1 "operators.kloudlite.io/api/v1"
	"operators.kloudlite.io/lib"
	"operators.kloudlite.io/lib/errors"
	"operators.kloudlite.io/lib/finalizers"
	reconcileResult "operators.kloudlite.io/lib/reconcile-result"
	"operators.kloudlite.io/lib/templates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// AppReconciler reconciles a Deployment object
type AppReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	ClientSet *kubernetes.Clientset
	JobMgr    lib.Job
	logger    *zap.SugaredLogger
	lib.MessageSender
	app *crdsv1.App

	HarborUserName string
	HarborPassword string
}

func (r *AppReconciler) notifyAndDie(ctx context.Context, err error) (ctrl.Result, error) {
	meta.SetStatusCondition(&r.app.Status.Conditions, metav1.Condition{
		Type:    "Ready",
		Status:  "False",
		Reason:  "ErrUnknown",
		Message: err.Error(),
	})

	return r.notify(ctx)
}

func (r *AppReconciler) notify(ctx context.Context) (ctrl.Result, error) {
	r.logger.Infof("Notify conditions: %+v", r.app.Status.Conditions)
	err := r.SendMessage(r.app.LogRef(), lib.MessageReply{
		Key:        r.app.LogRef(),
		Conditions: r.app.Status.Conditions,
		Status:     meta.IsStatusConditionTrue(r.app.Status.Conditions, "Ready"),
	})
	if err != nil {
		return reconcileResult.FailedE(errors.NewEf(err, "could not send message into kafka"))
	}

	if err := r.Status().Update(ctx, r.app); err != nil {
		return reconcileResult.FailedE(errors.NewEf(err, "could not update status for (app=%s)", r.app.LogRef()))
	}
	return reconcileResult.OK()
}

func (r *AppReconciler) IfDeployment(ctx context.Context, req ctrl.Request) (*metav1.Condition, error) {
	var deployment appsv1.Deployment
	if err := r.Get(ctx, req.NamespacedName, &deployment); err != nil {
		//not a deployment request
		return nil, nil
	}

	// ASSERT: not a deployment request
	if deployment.Name == "" {
		return nil, nil
	}

	// ASSERT: deployment request
	r.logger.Infof("resource request is a Deployment")
	//r.logger.Infof("Deployment.Status: %+v", deployment.Status)
	opts := &client.ListOptions{
		LabelSelector: labels2.SelectorFromValidatedSet(deployment.Spec.Template.GetLabels()),
		Namespace:     req.Namespace,
	}
	// TODO [x]: read container status from pod to figure out the real status
	p := corev1.PodList{}
	err := r.List(ctx, &p, opts)
	if err != nil {
		return nil, nil
	}

	readyCond := metav1.Condition{
		Type:   "Ready",
		Status: metav1.ConditionFalse,
		//ObservedGeneration: 0,
		//LastTransitionTime: metav1.Time{},
		//Reason:  "Initialized",
		//Message: "Initialized",
	}

	if len(p.Items) == 0 {
		readyCond.Reason = "Initialized"
		readyCond.Message = "Init"
		return &readyCond, nil
	}

	for _, item := range p.Items {
		//r.logger.Infof("----------")
		for _, condition := range item.Status.Conditions {
			if condition.Type == corev1.PodReady {
				readyCond.Reason = "PodIsReady"
				readyCond.Status = metav1.ConditionTrue
				readyCond.Message = "all pods are ready"
			}
			//r.logger.Infof("Pod.Status.Condition: %+v", condition)
			if condition.Status != corev1.ConditionTrue {
				readyCond.Status = metav1.ConditionFalse
				for _, status := range item.Status.ContainerStatuses {
					//r.logger.Infof("Pod.Status.ContainerStatus: %+v", status)
					if status.State.Waiting != nil {
						readyCond.Reason = status.State.Waiting.Reason
						readyCond.Message = status.State.Waiting.Message
						return &readyCond, nil
					}
				}
			}
		}
		//r.logger.Infof("----------")
	}

	// TODO: try to read conditions from replicaset also, to aggregate SYSTEM errors
	return &readyCond, nil
}

//+kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps/finalizers,verbs=update

func (r *AppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := GetLogger(req.NamespacedName)
	r.logger = logger.With("Name", req.NamespacedName)

	app := &crdsv1.App{}

	if err := r.Get(ctx, req.NamespacedName, app); err != nil {
		if apiErrors.IsNotFound(err) {
			return reconcileResult.OK()
		}
		return reconcileResult.Failed()
	}
	r.app = app

	deplCondition, err := r.IfDeployment(ctx, req)
	if err != nil {
		fmt.Println()
	}
	if deplCondition != nil {
		r.logger.Infof("deployment condition received: %+v", *deplCondition)
		meta.SetStatusCondition(&app.Status.Conditions, *deplCondition)
	}

	if app.GetDeletionTimestamp() != nil {
		return r.finalize(ctx, app)
	}

	depl, err := templates.Parse(templates.Deployment, app)
	if err != nil {
		logger.Info(err)
		return r.notifyAndDie(ctx, err)
	}

	svc, err := templates.Parse(templates.Service, app)
	if err != nil {
		logger.Info(err)
		return r.notifyAndDie(ctx, err)
	}

	if err2 := fn.KubectlApply(depl, svc); err2 != nil {
		return r.notifyAndDie(ctx, errors.NewEf(err2, "could not apply app"))
	}
	logger.Info("App has been applied")
	return r.notify(ctx)
}

func (r *AppReconciler) finalize(ctx context.Context, app *crdsv1.App) (ctrl.Result, error) {
	if controllerutil.ContainsFinalizer(app, finalizers.App.String()) {
		controllerutil.RemoveFinalizer(app, finalizers.App.String())
		if err := r.Update(ctx, app); err != nil {
			return reconcileResult.FailedE(errors.NewEf(err, "could not update app to remove finalizer"))
		}
		return reconcileResult.OK()
	}
	return reconcileResult.OK()
}

const ImageRegistry = "harbor.dev.madhouselabs.io"

//func (r *AppReconciler) checkImagesAvailable(ctx context.Context, app *crdsv1.App) (map[string]bool, error) {
//	checks := map[string]bool{}
//	for _, c := range app.Spec.Containers {
//		if !strings.HasPrefix(c.Image, ImageRegistry) {
//			checks[c.Image] = true
//			continue
//		}
//
//		imageName := strings.Replace(c.Image, fmt.Sprintf("%s/ci/", ImageRegistry), "", 1)
//		artifact := strings.Split(imageName, ":")
//		artifactName := artifact[0]
//		tag := artifact[1]
//
//		req, err := http.NewRequest(
//			http.MethodGet,
//			fmt.Sprintf("https://harbor.dev.madhouselabs.io/api/v2.0/projects/ci/repositories/%s/artifacts/%s/tags", url.QueryEscape(artifactName), tag),
//			nil,
//		)
//		if err != nil {
//			return nil, errors.NewEf(err, "could not create http object")
//		}
//		req.Header.Add("Content-Type", "application/vnd.oci.image.index.v1+json")
//		req.SetBasicAuth(r.HarborUserName, r.HarborPassword)
//
//		resp, err := http.DefaultClient.Do(req)
//		if err != nil {
//			return nil, errors.NewEf(err, "could not make http request")
//		}
//		r.logger.Infof("resp.StatusCode=%v", resp.StatusCode)
//		checks[c.Image] = resp.StatusCode == 200
//	}
//
//	return checks, nil
//}
//
//func (r *AppReconciler) checkDependency(ctx context.Context, app *crdsv1.App) *map[string]string {
//	checks := map[string]string{}
//	for _, container := range app.Spec.Containers {
//		for _, env := range container.Env {
//			if env.Value != "" {
//				continue
//			}
//
//			sp := strings.Split(env.RefName, "/")
//			if sp[0] == "config" {
//				var cfg corev1.ConfigMap
//				err := r.Get(ctx, types.NamespacedName{Namespace: app.Namespace, Name: sp[1]}, &cfg)
//				if err != nil {
//					if apiErrors.IsNotFound(err) {
//						checks[env.RefName] = "NotFound"
//						return &checks
//					}
//					checks[env.RefName] = err.Error()
//					return &checks
//				}
//				if _, ok := cfg.Data[env.RefKey]; !ok {
//					checks[fmt.Sprintf("%s/%s", env.RefName, env.RefKey)] = "NotFound"
//					return &checks
//				}
//			}
//
//			if sp[0] == "secret" {
//				var scrt corev1.Secret
//				err := r.Get(ctx, types.NamespacedName{Namespace: app.Namespace, Name: sp[1]}, &scrt)
//				if err != nil {
//					if apiErrors.IsNotFound(err) {
//						checks[env.RefName] = "NotFound"
//						return &checks
//					}
//					checks[env.RefName] = err.Error()
//					return &checks
//				}
//				if _, ok := scrt.Data[env.RefKey]; !ok {
//					checks[fmt.Sprintf("%s/%s", env.RefName, env.RefKey)] = "NotFound"
//					return &checks
//				}
//			}
//		}
//	}
//	return nil
//}

// SetupWithManager sets up the controller with the Manager.
func (r *AppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crdsv1.App{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}
