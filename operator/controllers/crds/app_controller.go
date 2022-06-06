package crds

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/constants"
	fn "operators.kloudlite.io/lib/functions"
	rApi "operators.kloudlite.io/lib/operator"

	"operators.kloudlite.io/lib"
	"operators.kloudlite.io/lib/templates"
)

// AppReconciler reconciles a Deployment object
type AppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	lib.MessageSender

	HarborUserName string
	HarborPassword string
}

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps/finalizers,verbs=update

func (r *AppReconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	req := rApi.NewRequest(ctx, r.Client, oReq.NamespacedName, &crdsv1.App{})

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

func (r *AppReconciler) finalize(req *rApi.Request[*crdsv1.App]) rApi.StepResult {
	return req.Finalize()
}

func (r *AppReconciler) reconcileStatus(req *rApi.Request[*crdsv1.App]) rApi.StepResult {
	ctx := req.Context()
	app := req.Object

	var cs []metav1.Condition
	isReady := true

	dConditions, err := conditions.FromResource(
		ctx,
		r.Client,
		constants.DeploymentGroup,
		"Deployment",
		fn.NN(app.Namespace, app.Name),
	)
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.FailWithStatusError(err)
		}
		isReady = false
		cs = append(cs, conditions.New("DeploymentExists", false, "NotFound", err.Error()))
	}
	cs = append(cs, dConditions...)

	if !meta.IsStatusConditionTrue(dConditions, "Deployment-Available") {
		isReady = false
	}

	newConditions, hasUpdated, err := conditions.Patch(app.Status.Conditions, cs)
	if err != nil {
		return req.FailWithStatusError(err)
	}

	if !hasUpdated && isReady == app.Status.IsReady {
		return req.Next()
	}

	app.Status.IsReady = isReady
	app.Status.Conditions = newConditions

	if err := r.Status().Update(ctx, app); err != nil {
		return req.FailWithStatusError(err)
	}
	return req.Done()
}

func (r *AppReconciler) reconcileOperations(req *rApi.Request[*crdsv1.App]) rApi.StepResult {
	ctx := req.Context()
	app := req.Object

	if !controllerutil.ContainsFinalizer(app, constants.CommonFinalizer) {
		controllerutil.AddFinalizer(app, constants.CommonFinalizer)
		controllerutil.AddFinalizer(app, constants.ForegroundFinalizer)
		if err := r.Status().Update(ctx, app); err != nil {
			return req.FailWithOpError(err)
		}
		return req.Next()
	}

	depl, err := templates.Parse(templates.Deployment, app)
	if err != nil {
		return req.FailWithOpError(err)
	}

	svc, err := templates.Parse(templates.Service, app)
	if err != nil {
		return req.FailWithOpError(err)
	}

	if _, err := fn.KubectlApplyExec(depl, svc); err != nil {
		return req.FailWithOpError(err)
	}

	return req.Done()
}

func (r *AppReconciler) notify(req *rApi.Request[*crdsv1.App]) error {
	app := req.Object
	return r.SendMessage(
		req.Object.LogRef(), lib.MessageReply{
			Key:        app.LogRef(),
			Conditions: app.Status.Conditions,
			Status:     app.Status.IsReady,
		},
	)
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
//		req.Header.Add("Content-Kind", "application/vnd.oci.image.index.v1+json")
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
		Owns(&corev1.Service{}).
		Complete(r)
}
