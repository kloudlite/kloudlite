package mongodbstandalonemsvc

import (
	"context"
	"fmt"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	mongodbStandalone "operators.kloudlite.io/apis/mongodb-standalone.msvc/v1"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/errors"
	fn "operators.kloudlite.io/lib/functions"
	rApi "operators.kloudlite.io/lib/operator"
	"operators.kloudlite.io/lib/templates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ServiceReconciler reconciles a Service object

type ServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *ServiceReconciler) GetName() string {
	return "mongo-standalone-service"
}

const (
	SvcRootPasswordKey = "root-password"
)

// +kubebuilder:rbac:groups=mongodb-standalone.msvc.kloudlite.io,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mongodb-standalone.msvc.kloudlite.io,resources=services/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mongodb-standalone.msvc.kloudlite.io,resources=services/finalizers,verbs=update

func (r *ServiceReconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(ctx, r.Client, oReq.NamespacedName, &mongodbStandalone.Service{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// STEP: cleaning up last run, clearing opsConditions
	if len(req.Object.Status.OpsConditions) > 0 {
		req.Object.Status.OpsConditions = []metav1.Condition{}
		return ctrl.Result{RequeueAfter: 0}, r.Status().Update(ctx, req.Object)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.Result(), x.Err()
		}
	}

	req.Logger.Infof("--------------------NEW RECONCILATION------------------")

	if x := req.EnsureLabelsAndAnnotations(); !x.ShouldProceed() {
		return x.Result(), x.Err()
	}

	if x := r.reconcileStatus(req); !x.ShouldProceed() {
		return x.Result(), x.Err()
	}

	if x := r.reconcileOperations(req); !x.ShouldProceed() {
		return x.Result(), x.Err()
	}

	req.Logger.Infof("--------------------RECONCILATION FINISH------------------")

	return ctrl.Result{}, nil
}

func (r *ServiceReconciler) finalize(req *rApi.Request[*mongodbStandalone.Service]) rApi.StepResult {
	return req.Finalize()
}

func (r *ServiceReconciler) reconcileStatus(req *rApi.Request[*mongodbStandalone.Service]) rApi.StepResult {
	ctx := req.Context()
	svcObj := req.Object

	isReady := true
	var cs []metav1.Condition
	var childC []metav1.Condition

	// STEP:  helm resource
	helmResource, err := rApi.Get(
		ctx, r.Client, fn.NN(svcObj.Namespace, svcObj.Name), fn.NewUnstructured(constants.HelmMongoDBType),
	)

	if err != nil {
		isReady = false
		if !apiErrors.IsNotFound(err) {
			return req.FailWithStatusError(err)
		}
		cs = append(cs, conditions.New(conditions.HelmResourceExists, false, conditions.NotFound, err.Error()))
	} else {
		cs = append(cs, conditions.New(conditions.HelmResourceExists, true, conditions.Found))

		rConditions, err := conditions.ParseFromResource(helmResource, "Helm")
		if err != nil {
			return req.FailWithStatusError(err)
		}
		childC = append(childC, rConditions...)
		rReady := meta.IsStatusConditionTrue(rConditions, "HelmDeployed")
		if !rReady {
			isReady = false
		}
		cs = append(
			cs, conditions.New(conditions.HelmResourceReady, rReady, conditions.Empty),
		)
	}

	// 2. deployment/sts
	deploymentRes, err := rApi.Get(ctx, r.Client, fn.NN(svcObj.Namespace, svcObj.Name), &appsv1.Deployment{})
	if err != nil {
		isReady = false
		if !apiErrors.IsNotFound(err) {
			return req.FailWithStatusError(err)
		}
		cs = append(cs, conditions.New(conditions.DeploymentExists, false, conditions.NotFound, err.Error()))
	} else {
		cs = append(cs, conditions.New(conditions.DeploymentExists, true, conditions.Found))
		rConditions, err := conditions.ParseFromResource(deploymentRes, "Deployment")
		if err != nil {
			return req.FailWithStatusError(err)
		}
		childC = append(childC, rConditions...)
		rReady := meta.IsStatusConditionTrue(rConditions, "DeploymentAvailable")
		if !rReady {
			isReady = false
		}
		cs = append(
			cs, conditions.New(conditions.DeploymentReady, rReady, conditions.Empty),
		)
	}

	// STEP: if vars generated ?
	if !svcObj.Status.GeneratedVars.Exists(SvcRootPasswordKey) {
		isReady = false
		cs = append(
			cs, conditions.New(
				conditions.GeneratedVars, false, conditions.NotReconciledYet,
			),
		)
	} else {
		cs = append(cs, conditions.New(conditions.GeneratedVars, true, conditions.Found))
	}

	// STEP: if reconciler output exists
	_, err = rApi.Get(
		ctx, r.Client, fn.NN(svcObj.Namespace, fmt.Sprintf("msvc-%s", svcObj.Name)), &corev1.Secret{},
	)
	if err != nil {
		isReady = false
		cs = append(cs, conditions.New(conditions.ReconcilerOutputExists, false, conditions.NotFound, err.Error()))
	} else {
		cs = append(cs, conditions.New(conditions.ReconcilerOutputExists, true, conditions.Found))
	}

	// STEP: patch aggregated conditions
	newChildConditions, hasUpdated, err := conditions.Patch(svcObj.Status.ChildConditions, childC)
	if err != nil {
		return req.FailWithStatusError(err)
	}

	newConditions, hasUpdated2, err := conditions.Patch(svcObj.Status.Conditions, cs)
	if err != nil {
		return req.FailWithStatusError(err)
	}

	if !hasUpdated && !hasUpdated2 && isReady == svcObj.Status.IsReady {
		return req.Next()
	}

	svcObj.Status.IsReady = isReady
	svcObj.Status.Conditions = newConditions
	svcObj.Status.ChildConditions = newChildConditions
	svcObj.Status.OpsConditions = []metav1.Condition{}

	return rApi.NewStepResult(&ctrl.Result{}, r.Status().Update(ctx, svcObj))
}

func (r *ServiceReconciler) reconcileOperations(req *rApi.Request[*mongodbStandalone.Service]) rApi.StepResult {
	ctx := req.Context()
	svcObj := req.Object

	if meta.IsStatusConditionFalse(svcObj.Status.Conditions, conditions.GeneratedVars.String()) {
		if err := svcObj.Status.GeneratedVars.Set(SvcRootPasswordKey, fn.CleanerNanoid(40)); err != nil {
			return req.FailWithOpError(err)
		}
		return rApi.NewStepResult(&ctrl.Result{}, r.Status().Update(ctx, svcObj))
	}

	if errP := func() error {
		b1, err := templates.Parse(
			templates.MongoDBStandalone, map[string]any{
				"object": svcObj,
				// TODO: storage-class
				"storage-class": constants.DoBlockStorageXFS,
				"owner-refs": []metav1.OwnerReference{
					fn.AsOwner(svcObj, true),
				},
			},
		)

		if err != nil {
			return err
		}

		hostUrl := fmt.Sprintf("%s.%s.svc.cluster.local", svcObj.Name, svcObj.Namespace)
		authPasswd, ok := svcObj.Status.GeneratedVars.GetString(SvcRootPasswordKey)
		if !ok {
			return errors.Newf("%s key not found in generated vars", SvcRootPasswordKey)
		}

		b2, err := templates.Parse(
			templates.Secret, &corev1.Secret{
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
			},
		)
		if err != nil {
			return err
		}

		if _, err := fn.KubectlApplyExec(b1, b2); err != nil {
			return err
		}
		return nil
	}(); errP != nil {
		return req.FailWithOpError(errP)
	}

	return req.Done()
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	builder := ctrl.NewControllerManagedBy(mgr).For(&mongodbStandalone.Service{})

	builder.Owns(fn.NewUnstructured(constants.HelmMongoDBType))
	builder.Owns(&corev1.Secret{})

	refWatchList := []client.Object{
		&appsv1.Deployment{},
		&corev1.Pod{},
	}

	for _, item := range refWatchList {
		builder.Watches(
			&source.Kind{Type: item}, handler.EnqueueRequestsFromMapFunc(
				func(obj client.Object) []reconcile.Request {
					value, ok := obj.GetLabels()[fmt.Sprintf("%s/ref", mongodbStandalone.GroupVersion.Group)]
					if !ok {
						return nil
					}
					return []reconcile.Request{
						{NamespacedName: fn.NN(obj.GetNamespace(), value)},
					}
				},
			),
		)
	}

	return builder.Complete(r)
}
