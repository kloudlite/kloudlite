package msvc

import (
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	crdsv1 "operators.kloudlite.io/api/v1"
	msvcv1 "operators.kloudlite.io/apis/msvc/v1"
	"operators.kloudlite.io/controllers"
	"operators.kloudlite.io/lib/errors"
	reconcileResult "operators.kloudlite.io/lib/reconcile-result"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// MongoDBReconciler reconciles a MongoDB object
type MongoDBReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	logger *zap.SugaredLogger
	msvc   *crdsv1.ManagedService
}

// +kubebuilder:rbac:groups=msvc.kloudlite.io,resources=mongodbs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=msvc.kloudlite.io,resources=mongodbs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=msvc.kloudlite.io,resources=mongodbs/finalizers,verbs=update

const (
	RootPassword string = "ROOT_PASSWORD"
	DbHosts      string = "HOSTS"
	DbUrl        string = "DB_URL"
)

func (r *MongoDBReconciler) notifyMsvc(ctx context.Context, c metav1.Condition) error {
	meta.SetStatusCondition(&r.msvc.Status.Conditions, c)
	if err := r.Status().Update(ctx, r.msvc); err != nil {
		return errors.NewEf(err, "could not update resource %s", r.msvc.LogRef())
	}
	return nil
}

func (r *MongoDBReconciler) IfDeployment(ctx context.Context, req ctrl.Request) (*metav1.Condition, error) {
	var deployment appsv1.Deployment
	if err := r.Get(ctx, req.NamespacedName, &deployment); err != nil {
		// not a deployment request
		return nil, nil
	}
	if deployment.Name == "" {
		return nil, nil
	}
	r.logger.Infof("request is a deployment request")
	return nil, nil
}

func (r *MongoDBReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.logger = controllers.GetLogger(req.NamespacedName)
	logger := r.logger.With("RECONCILE", true)

	mdb := &msvcv1.MongoDB{}
	if err := r.Get(ctx, req.NamespacedName, mdb); err != nil {
		if apiErrors.IsNotFound(err) {
			// INFO: might have been deleted
			return reconcileResult.OK()
		}
		return reconcileResult.Failed()
	}

	if mdb.GetDeletionTimestamp() != nil {
		return reconcileResult.OK()
	}

	r.IfDeployment(ctx, req)

	logger.Debugf("Reconilation started ...")

	depl := appsv1.Deployment{}
	if err := r.Client.Get(ctx, types.NamespacedName{Namespace: mdb.Namespace, Name: mdb.DeploymentName()}, &depl); err != nil {
		return reconcileResult.FailedE(errors.NewEf(err, "could not get deployment for helm resource"))
	}

	x := metav1.Condition{
		Type:    "Ready",
		Status:  "False",
		Reason:  "Initialized",
		Message: "deployment has not been created yet",
	}

	for _, cond := range depl.Status.Conditions {
		if cond.Type == appsv1.DeploymentAvailable {
			x.Status = metav1.ConditionStatus(cond.Status)
			x.Reason = cond.Reason
			x.Message = cond.Message
		}
	}

	var msvc crdsv1.ManagedService
	if err := r.Get(ctx, types.NamespacedName{Namespace: mdb.Namespace, Name: mdb.Name}, &msvc); err != nil {
		return reconcileResult.FailedE(err)
	}
	r.msvc = &msvc

	if err := r.notifyMsvc(ctx, x); err != nil {
		return reconcileResult.FailedE(err)
	}

	if x.Status != "True" {
		return reconcileResult.Retry()
	}

	var mongoCfg corev1.Secret

	if err := r.Get(ctx, types.NamespacedName{Namespace: mdb.Namespace, Name: mdb.DeploymentName()}, &mongoCfg); err != nil {
		return reconcileResult.FailedE(err)
	}

	body := map[string]string{
		RootPassword: string(mongoCfg.Data["mongodb-root-password"]),
		DbHosts:      fmt.Sprintf("%s.%s.svc.cluster.local", mdb.DeploymentName(), mdb.Namespace),
		DbUrl:        fmt.Sprintf("mongodb://%s:%s@%s.%s.svc.cluster.local/admin?authSource=admin", "root", string(mongoCfg.Data["mongodb-root-password"]), mdb.DeploymentName(), mdb.Namespace),
	}

	b, err := json.Marshal(body)
	if err != nil {
		return reconcileResult.FailedE(errors.NewEf(err, "could not marshal secret body into JSON"))
	}
	body["JSON"] = string(b)

	ts := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: mdb.Namespace,
			Name:      fmt.Sprintf("msvc-%s", mdb.Name),
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, &ts, func() error {
		ts.StringData = body
		if err := controllerutil.SetControllerReference(mdb, &ts, r.Scheme); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return reconcileResult.FailedE(err)
	}

	logger.Infof("Reconcile Completed ...", x)
	return reconcileResult.OK()
}

func (r *MongoDBReconciler) retry() (ctrl.Result, error) {
	return reconcileResult.Retry()
}

// SetupWithManager sets up the controller with the Manager.
func (r *MongoDBReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&msvcv1.MongoDB{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
