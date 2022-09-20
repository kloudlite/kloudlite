package mongodbexternal

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	mongodbexternalv1 "operators.kloudlite.io/apis/mongodb.external/v1"
	"operators.kloudlite.io/env"
	"operators.kloudlite.io/lib/conditions"
	fn "operators.kloudlite.io/lib/functions"
	"operators.kloudlite.io/lib/logging"
	rApi "operators.kloudlite.io/lib/operator"
	stepResult "operators.kloudlite.io/lib/operator/step-result"
	"operators.kloudlite.io/lib/templates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DatabaseReconciler reconciles a Database object
type DatabaseReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	logger logging.Logger
	Name   string
}

func (r *DatabaseReconciler) GetName() string {
	return r.Name
}

const (
	OutputExists conditions.Type = "mongo-external-db/output.exists"
)

// +kubebuilder:rbac:groups=mongodb.external.kloudlite.io,resources=databases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mongodb.external.kloudlite.io,resources=databases/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mongodb.external.kloudlite.io,resources=databases/finalizers,verbs=update

func (r *DatabaseReconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, oReq.NamespacedName, &mongodbexternalv1.Database{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	req.Logger.Infof("--------------------NEW RECONCILATION------------------")

	if x := req.EnsureLabelsAndAnnotations(); !x.ShouldProceed() {
		return x.ReconcilerResponse()
	}

	if x := r.reconcileStatus(req); !x.ShouldProceed() {
		return x.ReconcilerResponse()
	}

	if x := r.reconcileOperations(req); !x.ShouldProceed() {
		return x.ReconcilerResponse()
	}

	req.Logger.Infof("--------------------RECONCILATION FINISH------------------")

	return ctrl.Result{}, nil
}

func (r *DatabaseReconciler) finalize(req *rApi.Request[*mongodbexternalv1.Database]) stepResult.Result {
	return req.Finalize()
}

func (r *DatabaseReconciler) reconcileStatus(req *rApi.Request[*mongodbexternalv1.Database]) stepResult.Result {
	obj := req.Object
	ctx := req.Context()

	isReady := true
	var cs []metav1.Condition

	_, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Name), &corev1.Secret{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			cs = append(cs, conditions.New(OutputExists, false, conditions.NotFound, err.Error()))
			return req.FailWithStatusError(err, cs...)
		}
		isReady = false
		cs = append(cs, conditions.New(OutputExists, false, conditions.NotFound, err.Error()))
	}

	nConditions, hasUpdated, err := conditions.Patch(obj.Status.Conditions, cs)
	if err != nil {
		return req.FailWithStatusError(err, cs...)
	}

	if !hasUpdated && isReady == obj.Status.IsReady {
		return req.Next()
	}

	obj.Status.Conditions = nConditions
	obj.Status.IsReady = isReady

	if err := r.Status().Update(ctx, obj); err != nil {
		return req.FailWithStatusError(err)
	}

	return req.Done()
}

func (r *DatabaseReconciler) reconcileOperations(req *rApi.Request[*mongodbexternalv1.Database]) stepResult.Result {
	obj := req.Object
	ctx := req.Context()

	b, err := templates.Parse(
		templates.CoreV1.Secret, map[string]any{
			"name":       "mres-" + obj.Name,
			"namespace":  obj.Namespace,
			"labels":     obj.GetLabels(),
			"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},
			"string-data": map[string]string{
				"DB_URI":  obj.Spec.ConnectionUri,
				"DB_NAME": obj.Spec.DbName,
			},
		},
	)

	if err != nil {
		return req.FailWithOpError(err).Err(nil)
	}

	if err := fn.KubectlApplyExec(ctx, b); err != nil {
		return req.FailWithOpError(err).Err(nil)
	}

	obj.Status.OpsConditions = nil
	if err := r.Status().Update(ctx, obj); err != nil {
		return req.FailWithOpError(err)
	}

	return req.Next()
}

func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager, envVars *env.Env, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)

	return ctrl.NewControllerManagedBy(mgr).
		For(&mongodbexternalv1.Database{}).
		Complete(r)
}
