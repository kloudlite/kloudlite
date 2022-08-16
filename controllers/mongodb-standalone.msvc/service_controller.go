package mongodbstandalonemsvc

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	ct "operators.kloudlite.io/apis/common-types"
	"operators.kloudlite.io/env"
	"operators.kloudlite.io/lib/logging"
	stepResult "operators.kloudlite.io/lib/operator/step-result"
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
	Env    *env.Env
	logger logging.Logger
	Name   string
}

func (r *ServiceReconciler) GetName() string {
	return r.Name
}

const (
	SvcRootPasswordKey = "root-password"
)

// +kubebuilder:rbac:groups=mongodb-standalone.msvc.kloudlite.io,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mongodb-standalone.msvc.kloudlite.io,resources=services/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mongodb-standalone.msvc.kloudlite.io,resources=services/finalizers,verbs=update

func (r *ServiceReconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, oReq.NamespacedName, &mongodbStandalone.Service{})
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

func (r *ServiceReconciler) finalize(req *rApi.Request[*mongodbStandalone.Service]) stepResult.Result {
	return req.Finalize()
}

func (r *ServiceReconciler) reconcileStatus(req *rApi.Request[*mongodbStandalone.Service]) stepResult.Result {
	ctx := req.Context()
	svcObj := req.Object

	isReady := true
	var cs []metav1.Condition
	var childC []metav1.Condition

	// STEP: helm resource
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

	// STEP 2. read conditions from actual statefulset/deployment

	stsRes, err := rApi.Get(ctx, r.Client, fn.NN(svcObj.Namespace, svcObj.Name), &appsv1.StatefulSet{})
	if err != nil {
		isReady = false
		if !apiErrors.IsNotFound(err) {
			return req.FailWithStatusError(err)
		}
		cs = append(cs, conditions.New(conditions.StsExists, false, conditions.NotFound, err.Error()))
	} else {
		cs = append(cs, conditions.New(conditions.StsExists, true, conditions.Found))
		rConditions, err := conditions.ParseFromResource(stsRes, "Sts")
		if err != nil {
			return req.FailWithStatusError(err)
		}
		childC = append(childC, rConditions...)
		if stsRes.Status.Replicas != stsRes.Status.ReadyReplicas {
			isReady = false
			cs = append(cs, conditions.New(conditions.StsReady, false, conditions.Empty))
		} else {
			cs = append(cs, conditions.New(conditions.StsReady, true, conditions.Empty))
		}
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

	if err := r.Status().Update(ctx, svcObj); err != nil {
		return req.FailWithStatusError(err)
	}

	return req.Done()
}

func (r *ServiceReconciler) reconcileOperations(req *rApi.Request[*mongodbStandalone.Service]) stepResult.Result {
	ctx := req.Context()
	svcObj := req.Object

	if meta.IsStatusConditionFalse(svcObj.Status.Conditions, conditions.GeneratedVars.String()) {
		if err := svcObj.Status.GeneratedVars.Set(SvcRootPasswordKey, fn.CleanerNanoid(40)); err != nil {
			return req.FailWithOpError(err)
		}
		if err := r.Status().Update(ctx, svcObj); err != nil {
			return req.FailWithOpError(err)
		}
		return req.Done()
	}

	if errP := func() error {
		storageClass, err := svcObj.Spec.CloudProvider.GetStorageClass(ct.Xfs)
		if err != nil {
			return err
		}

		if freezeVal, ok := svcObj.GetLabels()[constants.LabelKeys.Freeze]; ok && freezeVal == strconv.FormatBool(true) {
			// kubectl.Scale(
			// 	kubectl.Deployments, svcObj.Namespace, map[string]string{
			// 		// "kloudlite.io/msvc.name": svcObj.Spec.
			// 	}
			// )
		}

		b1, err := templates.Parse(
			templates.MongoDBStandalone, map[string]any{
				"object":        svcObj,
				"freeze":        svcObj.GetLabels()[constants.LabelKeys.Freeze] == "true",
				"storage-class": storageClass,
				"owner-refs": []metav1.OwnerReference{
					fn.AsOwner(svcObj, true),
				},
			},
		)

		if err != nil {
			return err
		}

		hosts := make([]string, 0, svcObj.Spec.ReplicaCount)
		for idx := 0; idx < svcObj.Spec.ReplicaCount; idx += 1 {
			hosts = append(hosts, fmt.Sprintf("%s-%d.%s.%s.svc.cluster.local", svcObj.Name, idx, svcObj.Name, svcObj.Namespace))
		}

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
					"DB_HOSTS":      strings.Join(hosts, ","),
					"DB_URL":        fmt.Sprintf("mongodb://%s:%s@%s/admin?authSource=admin", "root", authPasswd, strings.Join(hosts, ",")),
				},
			},
		)
		if err != nil {
			req.Logger.Errorf(err, "failed parsing template %s", templates.Secret)
			return nil
		}

		if err := fn.KubectlApplyExec(ctx, b1, b2); err != nil {
			req.Logger.Errorf(err, "failed kubect apply")
			return nil
		}
		return nil
	}(); errP != nil {
		return req.FailWithOpError(errP)
	}

	return req.Next()
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager, envVars *env.Env, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	builder := ctrl.NewControllerManagedBy(mgr).For(&mongodbStandalone.Service{})

	builder.Owns(fn.NewUnstructured(constants.HelmMongoDBType))
	builder.Owns(&corev1.Secret{})

	refWatchList := []client.Object{
		&appsv1.StatefulSet{},
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
