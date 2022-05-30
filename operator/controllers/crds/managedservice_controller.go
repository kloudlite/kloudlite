package crds

import (
	"context"
	"encoding/json"
	"fmt"

	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	v1 "operators.kloudlite.io/apis/crds/v1"
	"operators.kloudlite.io/lib"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/errors"
	fn "operators.kloudlite.io/lib/functions"
	rApi "operators.kloudlite.io/lib/operator"
	"operators.kloudlite.io/lib/templates"
)

// ManagedServiceReconciler reconciles a ManagedService object
type ManagedServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	lib.MessageSender
	lt metav1.Time
}

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=managedservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=managedservices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=managedservices/finalizers,verbs=update

func (r *ManagedServiceReconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	req := rApi.NewRequest(ctx, r.Client, oReq.NamespacedName, &v1.ManagedService{})
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
		if err := r.notify(req); err != nil {
			return ctrl.Result{}, err
		}
		return x.Result(), x.Err()
	}

	if x := r.reconcileOperations(req); !x.ShouldProceed() {
		if err := r.notify(req); err != nil {
			return ctrl.Result{}, err
		}
		return x.Result(), x.Err()
	}

	return ctrl.Result{}, nil
}

func (r *ManagedServiceReconciler) reconcileStatus(req *rApi.Request[*v1.ManagedService]) rApi.StepResult {
	ctx := req.Context()
	msvc := req.Object

	isReady := true
	var cs []metav1.Condition

	svcObj, err := rApi.Get(
		ctx, r.Client, fn.NN(msvc.Namespace, msvc.Name), fn.NewUnstructured(
			metav1.TypeMeta{Kind: "Service", APIVersion: msvc.Spec.ApiVersion},
		),
	)

	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.FailWithStatusError(err)
		}
		isReady = false
		cs = append(cs, conditions.New("ServiceExists", false, "ObjectNotFound", err.Error()))
	}

	mj, err := svcObj.MarshalJSON()
	if err != nil {
		return req.FailWithStatusError(err)
	}

	var j struct {
		Status rApi.Status `json:"status"`
	}

	if err := json.Unmarshal(mj, &j); err != nil {
		return req.FailWithStatusError(err)
	}

	cs = append(cs, j.Status.Conditions...)

	if !j.Status.IsReady {
		isReady = false
	}

	newConditions, hasUpdated, err := conditions.Patch(msvc.Status.Conditions, cs)
	if err != nil {
		return req.FailWithStatusError(errors.NewEf(err, "while patching conditions"))
	}

	if !hasUpdated {
		return req.Next()
	}

	msvc.Status.Conditions = newConditions
	msvc.Status.IsReady = isReady

	if err := r.Status().Update(ctx, msvc); err != nil {
		return req.FailWithStatusError(err)
	}

	return req.Done()
}

func (r *ManagedServiceReconciler) reconcileOperations(req *rApi.Request[*v1.ManagedService]) rApi.StepResult {
	ctx := req.Context()
	msvc := req.Object

	if !controllerutil.ContainsFinalizer(msvc, constants.CommonFinalizer) {
		controllerutil.AddFinalizer(msvc, constants.CommonFinalizer)
		controllerutil.AddFinalizer(msvc, constants.ForegroundFinalizer)

		if err := r.Update(ctx, msvc); err != nil {
			return req.FailWithStatusError(err)
		}
		return req.Next()
	}

	obj, err := templates.ParseObject(templates.CommonMsvc, req.Object)
	obj.SetLabels(
		fn.MapMerge(obj.GetLabels(), msvc.GetWatchLabels()),
	)
	if err != nil {
		return req.FailWithOpError(err)
	}
	err = fn.KubectlApply(req.Context(), r.Client, obj)
	if err != nil {
		return req.FailWithOpError(err)
	}
	return req.Done()
}

func (r *ManagedServiceReconciler) notify(req *rApi.Request[*v1.ManagedService]) error {
	return nil
	// return r.SendMessage(
	// 	req.msvc.NameRef(), lib.MessageReply{
	// 		Key:        req.msvc.NameRef(),
	// 		Conditions: req.condBuilder.GetAll(),
	// 		Status:     req.condBuilder.IsTrue(constants.ConditionReady.Type),
	// 	},
	// )
}

func (r *ManagedServiceReconciler) finalize(req *rApi.Request[*v1.ManagedService]) rApi.StepResult {
	return req.Finalize()
}

// SetupWithManager sets up the controller with the Manager.
func (r *ManagedServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	builder := ctrl.NewControllerManagedBy(mgr).For(&v1.ManagedService{})

	allMsvcs := []string{
		"mongodb-standalone",
		"mongodb-cluster",
		"mysql-standalone",
		"redis-standalone",
	}

	for _, msvc := range allMsvcs {
		builder.Watches(
			&source.Kind{
				Type: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": fmt.Sprintf("%s.%s", msvc, constants.MsvcApiVersion),
						"kind":       "Service",
					},
				},
			}, handler.EnqueueRequestsFromMapFunc(
				func(obj client.Object) []reconcile.Request {
					labels := obj.GetLabels()
					s, ok := labels["msvc.kloudlite.io/ref"]
					if !ok {
						return nil
					}
					return []reconcile.Request{{NamespacedName: fn.NN(obj.GetNamespace(), s)}}
				},
			),
		)
	}

	return builder.Complete(r)
}
