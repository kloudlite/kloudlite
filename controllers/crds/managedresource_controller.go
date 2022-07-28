package crds

import (
	"context"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"operators.kloudlite.io/apis/crds/v1"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/errors"
	fn "operators.kloudlite.io/lib/functions"
	rApi "operators.kloudlite.io/lib/operator"
	"operators.kloudlite.io/lib/templates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// ManagedResourceReconciler reconciles a ManagedResource object
type ManagedResourceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *ManagedResourceReconciler) GetName() string {
	return "managed-resource"
}

const (
	RealMresExists conditions.Type = "RealMresExists"
)

func (r *ManagedResourceReconciler) UpdateConditionsAndExit(
	req *rApi.Request[*v1.ManagedResource], cs ...metav1.Condition,
) rApi.StepResult {
	newConditions, hasUpdated, err := conditions.Patch(req.Object.Status.Conditions, cs)
	if err != nil {
		return req.FailWithStatusError(errors.NewEf(err, "while patching conditions"))
	}
	if !hasUpdated {
		return rApi.NewStepResult(nil, nil)
	}
	req.Object.Status.Conditions = newConditions
	return rApi.NewStepResult(nil, r.Status().Update(req.Context(), req.Object))
}

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=managedresources,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=managedresources/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=managedresources/finalizers,verbs=update

func (r *ManagedResourceReconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(ctx, r.Client, oReq.NamespacedName, &v1.ManagedResource{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if step := req.CleanupLastRun(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
	}

	req.Logger.Infof("-------------------- NEW RECONCILATION------------------")

	if x := req.EnsureLabelsAndAnnotations(); !x.ShouldProceed() {
		return x.ReconcilerResponse()
	}

	if x := r.reconcileStatus(req); !x.ShouldProceed() {
		return x.ReconcilerResponse()
	}

	if x := r.reconcileOperations(req); !x.ShouldProceed() {
		return x.ReconcilerResponse()
	}

	return ctrl.Result{}, nil
}

func (r *ManagedResourceReconciler) finalize(req *rApi.Request[*v1.ManagedResource]) rApi.StepResult {
	return req.Finalize()
}

func (r *ManagedResourceReconciler) reconcileStatus(req *rApi.Request[*v1.ManagedResource]) rApi.StepResult {
	// STEP: PRE if msvc is ready
	ctx := req.Context()
	obj := req.Object

	isReady := true
	var cs []metav1.Condition

	msvc, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Spec.MsvcRef.Name), &v1.ManagedService{})
	if err != nil {
		cs = append(cs, conditions.New(conditions.ManagedSvcExists, false, conditions.NotFound, err.Error()))
		return req.FailWithStatusError(err, cs...).NoErr()
	}

	if !msvc.Status.IsReady {
		cs = append(cs, conditions.New(conditions.ManagedSvcReady, false, conditions.NotReady))
		return req.FailWithStatusError(err, cs...).NoErr()
	}

	rApi.SetLocal(req, "msvc", msvc)

	// STEP: fetch conditions from real managed resource
	resourceC, err := conditions.FromResource(
		ctx, r.Client, metav1.TypeMeta{
			APIVersion: obj.Spec.MsvcRef.APIVersion,
			Kind:       obj.Spec.MresKind.Kind,
		},
		"", fn.NN(obj.Namespace, obj.Name),
	)

	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.FailWithStatusError(err)
		}
		isReady = false
		cs = append(cs, conditions.New(RealMresExists, false, conditions.NotFound, err.Error()))
		resourceC = nil
	} else {
		cs = append(cs, conditions.New(RealMresExists, true, conditions.Found))
	}

	cs = append(cs, resourceC...)

	newConditions, hasUpdated, err := conditions.Patch(obj.Status.Conditions, cs)
	if err != nil {
		return req.FailWithStatusError(errors.NewEf(err, "while patching conditions"))
	}

	if !hasUpdated && isReady == obj.Status.IsReady {
		return req.Next()
	}

	obj.Status.IsReady = isReady
	obj.Status.Conditions = newConditions
	return rApi.NewStepResult(&ctrl.Result{}, r.Status().Update(ctx, obj))
}

func (r *ManagedResourceReconciler) reconcileOperations(req *rApi.Request[*v1.ManagedResource]) rApi.StepResult {
	ctx := req.Context()
	mres := req.Object

	msvc, ok := rApi.GetLocal[*v1.ManagedService](req, "msvc")
	if !ok {
		return req.FailWithOpError(errors.Newf("%s key not found in locals", "msvc"))
	}

	if !fn.IsOwner(mres, fn.AsOwner(msvc, true)) {
		mres.SetOwnerReferences(
			[]metav1.OwnerReference{
				fn.AsOwner(msvc, true),
			},
		)

		return rApi.NewStepResult(&ctrl.Result{}, r.Update(ctx, mres))
	}

	if !controllerutil.ContainsFinalizer(mres, constants.CommonFinalizer) {
		controllerutil.AddFinalizer(mres, constants.CommonFinalizer)
		controllerutil.AddFinalizer(mres, constants.ForegroundFinalizer)

		return rApi.NewStepResult(&ctrl.Result{}, r.Update(ctx, mres))
	}

	b, err := templates.Parse(
		templates.CommonMres, map[string]any{
			"object": mres,
			"owner-refs": []metav1.OwnerReference{
				fn.AsOwner(mres, true),
			},
		},
	)
	if err != nil {
		return req.FailWithOpError(err)
	}
	if _, err := fn.KubectlApplyExec(b); err != nil {
		return req.FailWithOpError(err)
	}
	return req.Done()
}

// SetupWithManager sets up the controller with the Manager.
func (r *ManagedResourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	builder := ctrl.NewControllerManagedBy(mgr).For(&v1.ManagedResource{})

	resources := []metav1.TypeMeta{
		{Kind: "ACLAccount", APIVersion: "redis-standalone.msvc.kloudlite.io/v1"},
		{Kind: "Database", APIVersion: "mongodb-standalone.msvc.kloudlite.io/v1"},
		{Kind: "Service", APIVersion: "mongodb-standalone.msvc.kloudlite.io/v1"},
	}

	for _, resource := range resources {
		builder.Owns(fn.NewUnstructured(resource))
	}

	builder.Watches(
		&source.Kind{Type: &v1.ManagedService{}},
		handler.EnqueueRequestsFromMapFunc(
			func(obj client.Object) []reconcile.Request {
				var mresList v1.ManagedResourceList
				if err := r.List(
					context.TODO(), &mresList, &client.ListOptions{
						LabelSelector: labels.SelectorFromValidatedSet(
							map[string]string{
								"msvc.kloudlite.io/ref.name":    obj.GetName(),
								"msvc.kloudlite.io/ref.group":   obj.GetObjectKind().GroupVersionKind().Group,
								"msvc.kloudlite.io/ref.version": obj.GetObjectKind().GroupVersionKind().Version,
							},
						),
						Namespace: obj.GetNamespace(),
					},
				); err != nil {
					return nil
				}

				var reqs []reconcile.Request
				for _, item := range mresList.Items {
					reqs = append(reqs, reconcile.Request{NamespacedName: fn.NN(item.Namespace, item.Name)})
				}
				return reqs
			},
		),
	)

	return builder.Complete(r)
}
