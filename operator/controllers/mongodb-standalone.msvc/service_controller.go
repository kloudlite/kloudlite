package mongodbstandalonemsvc

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	mongodbStandalone "operators.kloudlite.io/apis/mongodb-standalone.msvc/v1"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/errors"
	fn "operators.kloudlite.io/lib/functions"
	rApi "operators.kloudlite.io/lib/operator"
)

// ServiceReconciler reconciles a Service object

type ServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const (
	MongoDbRootPasswordKey = "mongodb-root-password"
	StorageClassKey        = "storage-class"
)

// +kubebuilder:rbac:groups=mongodb-standalone.msvc.kloudlite.io,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mongodb-standalone.msvc.kloudlite.io,resources=services/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mongodb-standalone.msvc.kloudlite.io,resources=services/finalizers,verbs=update

func (r *ServiceReconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	req := rApi.NewRequest(ctx, r.Client, oReq.NamespacedName, &mongodbStandalone.Service{})

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

func (r *ServiceReconciler) finalize(req *rApi.Request[*mongodbStandalone.Service]) rApi.StepResult {
	return req.Finalize()
}

func (r *ServiceReconciler) reconcileStatus(req *rApi.Request[*mongodbStandalone.Service]) rApi.StepResult {
	ctx := req.Context()
	svcObj := req.Object

	isReady := true

	cs, err := conditions.FromResource(
		ctx, r.Client, constants.HelmMongoDBGroup, "Helm", fn.NamespacedName(svcObj),
	)

	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.FailWithStatusError(err)
		}
		isReady = false
	}

	deploymentConditions, err := conditions.FromResource(
		ctx, r.Client, constants.DeploymentGroup, "Deployment", fn.NamespacedName(svcObj),
	)

	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.FailWithStatusError(err)
		}
		isReady = false
	}

	isReady = meta.IsStatusConditionTrue(deploymentConditions, "Deployment-Available")

	cs = append(cs, deploymentConditions...)

	// STEP: Helm Secret check
	_, err = rApi.Get(ctx, r.Client, fn.NamespacedName(svcObj), &corev1.Secret{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.FailWithStatusError(err)
		}
		cs = append(
			cs, conditions.New("HelmSecretAvailable", false, "HelmSecretNotFound", err.Error()),
		)
		isReady = false
	}

	// STEP: Generated Vars check
	if _, ok := svcObj.Status.GeneratedVars.GetString(MongoDbRootPasswordKey); !ok {
		cs = append(cs, conditions.New("GeneratedVars", false, "NotGeneratedYet"))
		isReady = false
	}

	// STEP: output check
	_, err = rApi.Get(
		ctx, r.Client, fn.NN(svcObj.Namespace, fmt.Sprintf("msvc-%s", svcObj.Name)), &corev1.Secret{},
	)
	if err != nil {
		isReady = false
		cs = append(cs, conditions.New("OutputExists", false, "SecretNotFound"))
	}

	// req.logger.Debugf("req.mongoSvc.Status: %+v", req.mongoSvc.Status)
	newConditions, updated, err := conditions.Patch(svcObj.Status.Conditions, cs)

	if !updated && isReady == svcObj.Status.IsReady {
		return req.Next()
	}

	svcObj.Status.IsReady = isReady
	svcObj.Status.Conditions = newConditions
	svcObj.Status.OpsConditions = []metav1.Condition{}

	if err := r.Status().Update(ctx, svcObj); err != nil {
		req.Logger.Error(err)
		req.FailWithStatusError(err)
	}

	return req.Done()
}

func (r *ServiceReconciler) reconcileOperations(req *rApi.Request[*mongodbStandalone.Service]) rApi.StepResult {
	ctx := req.Context()
	svcObj := req.Object
	if meta.IsStatusConditionFalse(svcObj.Status.Conditions, "GeneratedVars") {
		if err := svcObj.Status.GeneratedVars.Merge(
			map[string]any{
				MongoDbRootPasswordKey: fn.CleanerNanoid(40),
				// StorageClassKey:        "do-block-storage-xfs",
				StorageClassKey: "local-path-xfs",
			},
		); err != nil {
			return req.FailWithOpError(err)
		}
		return rApi.NewStepResult(&ctrl.Result{}, r.Status().Update(ctx, req.Object))
	}

	hostUrl := fmt.Sprintf("%s.%s.svc.cluster.local", svcObj.Name, svcObj.Namespace)
	authPasswd, ok := svcObj.Status.GeneratedVars.GetString(MongoDbRootPasswordKey)
	if !ok {
		return req.FailWithStatusError(errors.Newf("%s key not found in generated vars", MongoDbRootPasswordKey))
	}

	scrt := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("msvc-%s", svcObj.Name),
			Namespace: svcObj.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				fn.AsOwner(svcObj, true),
			},
		},
		StringData: map[string]string{
			"ROOT_PASSWORD": authPasswd,
			"DB_HOSTS":      hostUrl,
			"DB_URL":        fmt.Sprintf("mongodb://%s:%s@%s/admin?authSource=admin", "root", authPasswd, hostUrl),
		},
	}

	return rApi.NewStepResult(&ctrl.Result{}, fn.KubectlApply(ctx, r.Client, scrt))
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	builder := ctrl.NewControllerManagedBy(mgr).For(&mongodbStandalone.Service{})

	kinds := []client.Object{
		&appsv1.Deployment{},
		&corev1.Pod{},
		fn.NewUnstructured(
			metav1.TypeMeta{Kind: constants.HelmMongoDBGroup.Kind, APIVersion: constants.MsvcApiVersion},
		),
	}

	for _, kind := range kinds {
		builder.Watches(
			&source.Kind{Type: kind},
			handler.EnqueueRequestsFromMapFunc(
				func(c client.Object) []reconcile.Request {
					s, ok := c.GetLabels()[mongodbStandalone.GroupVersion.Group]
					if !ok {
						return nil
					}
					return []reconcile.Request{
						{NamespacedName: fn.NN(c.GetNamespace(), s)},
					}
				},
			),
		)
	}

	return builder.Complete(r)
}
