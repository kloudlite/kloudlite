package crds

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"operators.kloudlite.io/apis/crds/v1"
	"operators.kloudlite.io/lib"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/errors"
	fn "operators.kloudlite.io/lib/functions"
	rApi "operators.kloudlite.io/lib/operator"
	"operators.kloudlite.io/lib/templates"
)

// ManagedResourceReconciler reconciles a ManagedResource object
type ManagedResourceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	lib.MessageSender
}

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=managedresources,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=managedresources/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=managedresources/finalizers,verbs=update

func (r *ManagedResourceReconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	req := rApi.NewRequest(ctx, r.Client, oReq.NamespacedName, &v1.ManagedResource{})

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

func (r *ManagedResourceReconciler) notify(mres *v1.ManagedResource) error {
	return nil
	return r.SendMessage(
		fn.NamespacedName(mres).String(), lib.MessageReply{
			Key:        fn.NamespacedName(mres).String(),
			Conditions: mres.Status.Conditions,
			Status:     mres.Status.IsReady,
		},
	)
}

func (r *ManagedResourceReconciler) finalize(req *rApi.Request[*v1.ManagedResource]) rApi.StepResult {
	return req.Finalize()
}

func (r *ManagedResourceReconciler) reconcileStatus(req *rApi.Request[*v1.ManagedResource]) rApi.StepResult {
	// STEP: PRE if msvc is ready
	ctx := req.Context()
	mres := req.Object
	msvc, err := rApi.Get(ctx, r.Client, fn.NN(mres.Namespace, mres.Spec.ManagedSvcName), &v1.ManagedService{})
	if err != nil {
		return req.FailWithStatusError(err)
	}

	if !msvc.Status.IsReady {
		return req.FailWithStatusError(errors.Newf("msvc %s is not ready yet", msvc.Name))
	}

	rApi.SetLocal(req, "msvc", msvc)

	isReady := true
	var cs []metav1.Condition

	// STEP: fetch conditions from real managed resource
	resourceC, err := conditions.FromResource(
		ctx, r.Client, metav1.GroupVersionKind{
			Group:   strings.Split(mres.Spec.ApiVersion, "/")[0],
			Version: strings.Split(mres.Spec.ApiVersion, "/")[1],
			Kind:    mres.Spec.Kind,
		},
		"Res", fn.NN(mres.Namespace, mres.Name),
	)

	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.FailWithStatusError(err)
		}
		isReady = false
		cs = append(cs, conditions.New("MresResourceExists", false, "NotFound", err.Error()))
	}
	cs = append(cs, resourceC...)

	// STEP: resource output is ready
	mresOutput, err := rApi.Get(
		ctx,
		r.Client,
		fn.NN(mres.Namespace, fmt.Sprintf("mres-%s", mres.Name)),
		&corev1.Secret{},
	)
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.FailWithStatusError(err)
		}
		isReady = false
		cs = append(cs, conditions.New("MsvcOutputExists", false, "NotFound", err.Error()))
		mresOutput = nil
	}

	if mresOutput != nil {
		cs = append(cs, conditions.New("MresOutputExists", true, "SecretFound"))
	}

	newConditions, hasUpdated, err := conditions.Patch(mres.Status.Conditions, cs)
	if err != nil {
		return req.FailWithStatusError(errors.NewEf(err, "while patching conditions"))
	}

	if !hasUpdated {
		return req.Next()
	}

	mres.Status.IsReady = isReady
	mres.Status.Conditions = newConditions
	mres.Status.OpsConditions = []metav1.Condition{}
	if err := r.Status().Update(ctx, mres); err != nil {
		return req.FailWithStatusError(err)
	}
	return req.Done()
}

func (r *ManagedResourceReconciler) reconcileOperations(req *rApi.Request[*v1.ManagedResource]) rApi.StepResult {
	ctx := req.Context()
	mres := req.Object

	if !controllerutil.ContainsFinalizer(mres, constants.CommonFinalizer) {
		controllerutil.AddFinalizer(mres, constants.CommonFinalizer)
		controllerutil.AddFinalizer(mres, constants.ForegroundFinalizer)

		if err := r.Update(ctx, mres); err != nil {
			return req.FailWithStatusError(err)
		}
		return req.Next()
	}

	obj, err := templates.ParseObject(templates.CommonMres, req.Object)
	msvc, ok := rApi.GetLocal[*v1.ManagedService](req, "msvc")
	if !ok {
		return req.FailWithOpError(errors.Newf("%s key not found in locals", "msvc"))
	}
	obj.SetOwnerReferences(
		[]metav1.OwnerReference{
			fn.AsOwner(msvc, true),
		},
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

// SetupWithManager sets up the controller with the Manager.
func (r *ManagedResourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.ManagedResource{}).
		Watches(
			&source.Kind{
				Type: &v1.ManagedService{},
			}, handler.EnqueueRequestsFromMapFunc(
				func(obj client.Object) []reconcile.Request {

					var mresList v1.ManagedResourceList

					if err := r.List(
						context.TODO(), &mresList, &client.ListOptions{
							LabelSelector: labels.SelectorFromValidatedSet(
								map[string]string{
									"msvc.kloudlite.io/ref": obj.GetName(),
								},
							),
							Namespace: obj.GetNamespace(),
						},
					); err != nil {
						return nil
					}

					var reqs []reconcile.Request
					for _, item := range mresList.Items {
						reqs = append(reqs, reconcile.Request{NamespacedName: fn.NamespacedName(&item)})
					}
					return reqs
				},
			),
		).
		Complete(r)
}
