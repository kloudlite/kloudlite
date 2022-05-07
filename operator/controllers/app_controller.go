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
	"k8s.io/apimachinery/pkg/types"
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
	lt  metav1.Time

	HarborUserName string
	HarborPassword string
}

func (r *AppReconciler) notifyAndDie(ctx context.Context, err error) (ctrl.Result, error) {
	r.buildConditions("", metav1.Condition{
		Type:    "Ready",
		Status:  "False",
		Reason:  "ErrWhileReconcilation",
		Message: err.Error(),
	})

	return r.notify(ctx)
}

func (r *AppReconciler) notify(ctx context.Context, retry ...bool) (ctrl.Result, error) {
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
	if len(retry) > 0 {
		// TODO: find a better way to handle image from deployments
		return reconcileResult.Retry(1)
	}
	return reconcileResult.OK()
}

func (r *AppReconciler) buildConditions(source string, conditions ...metav1.Condition) {
	meta.SetStatusCondition(&r.app.Status.Conditions, metav1.Condition{
		Type:               "Ready",
		Status:             "False",
		Reason:             "ChecksNotCompleted",
		LastTransitionTime: r.lt,
		Message:            "Not All Checks completed",
	})
	for _, c := range conditions {
		if c.Reason == "" {
			c.Reason = "NotSpecified"
		}
		if !c.LastTransitionTime.IsZero() {
			if c.LastTransitionTime.Time.Sub(r.lt.Time).Seconds() > 0 {
				r.lt = c.LastTransitionTime
			}
		}
		if c.LastTransitionTime.IsZero() {
			c.LastTransitionTime = r.lt
		}
		if source != "" {
			c.Reason = fmt.Sprintf("%s:%s", source, c.Reason)
			c.Type = fmt.Sprintf("%s%s", source, c.Type)
		}
		meta.SetStatusCondition(&r.app.Status.Conditions, c)
	}
}

func (r *AppReconciler) HandleDeployments(ctx context.Context) error {
	var depl appsv1.Deployment
	if err := r.Get(ctx, types.NamespacedName{Name: r.app.Name, Namespace: r.app.Namespace}, &depl); err != nil {
		return err
	}

	var deplConditions []metav1.Condition
	for _, cond := range depl.Status.Conditions {
		deplConditions = append(deplConditions, metav1.Condition{
			Type:               string(cond.Type),
			Status:             metav1.ConditionStatus(cond.Status),
			LastTransitionTime: cond.LastTransitionTime,
			Reason:             cond.Reason,
			Message:            cond.Message,
		})
	}

	r.buildConditions("Deployment", deplConditions...)
	if !meta.IsStatusConditionTrue(deplConditions, string(appsv1.DeploymentAvailable)) {
		var podsList corev1.PodList
		if err := r.List(ctx, &podsList, &client.ListOptions{
			LabelSelector: labels2.SelectorFromValidatedSet(depl.Spec.Template.GetLabels()),
			Namespace:     depl.Namespace,
		}); err != nil {
			return errors.NewEf(err, "could not list pods for deployment")
		}

		for _, pod := range podsList.Items {
			var podC []metav1.Condition
			for _, condition := range pod.Status.Conditions {
				podC = append(podC, metav1.Condition{
					Type:               string(condition.Type),
					Status:             metav1.ConditionStatus(condition.Status),
					LastTransitionTime: condition.LastTransitionTime,
					Reason:             "NotSpecified",
					Message:            condition.Message,
				})
			}
			r.buildConditions("Pod", podC...)
			var containerC []metav1.Condition
			for _, cs := range pod.Status.ContainerStatuses {
				p := metav1.Condition{
					Type:   fmt.Sprintf("Name-%s", cs.Name),
					Status: fn.StatusFromBool(cs.Ready),
				}
				if cs.State.Waiting != nil {
					p.Reason = cs.State.Waiting.Reason
					p.Message = cs.State.Waiting.Message
				}
				if cs.State.Running != nil {
					p.Reason = "Running"
					p.Message = fmt.Sprintf("Container running since %s", cs.State.Running.StartedAt.String())
				}
				containerC = append(containerC, p)
			}
			r.buildConditions("Container", containerC...)
		}
		return nil
	}

	r.buildConditions("", metav1.Condition{
		Type:    "Ready",
		Status:  metav1.ConditionTrue,
		Reason:  "AllChecksPassed",
		Message: "Deployment is ready",
	})

	return nil
}

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps/finalizers,verbs=update

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
	app.Status.Conditions = []metav1.Condition{}
	r.app = app

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

	if err := r.HandleDeployments(ctx); err != nil {
		e := errors.NewEf(err, "err handling deployments")
		r.logger.Error(e)
		return r.notifyAndDie(ctx, e)
	}

	if meta.IsStatusConditionFalse(app.Status.Conditions, "Ready") {
		return r.notify(ctx, true)
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

// func (r *AppReconciler) checkImagesAvailable(ctx context.Context, app *crdsv1.App) (map[string]bool, error) {
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
// }
//
// func (r *AppReconciler) checkDependency(ctx context.Context, app *crdsv1.App) *map[string]string {
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
// }

// SetupWithManager sets up the controller with the Manager.
func (r *AppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crdsv1.App{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
