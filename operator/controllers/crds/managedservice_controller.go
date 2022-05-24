package crds

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	"operators.kloudlite.io/lib"
	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/errors"
	"operators.kloudlite.io/lib/finalizers"
	fn "operators.kloudlite.io/lib/functions"
	reconcileResult "operators.kloudlite.io/lib/reconcile-result"
	"operators.kloudlite.io/lib/templates"
	t "operators.kloudlite.io/lib/types"
)

// ManagedServiceReconciler reconciles a ManagedService object
type ManagedServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	lib.MessageSender
	lt metav1.Time
}

type MsvcReconReq struct {
	t.ReconReq
	ctrl.Request
	condBuilder fn.StatusConditions
	logger      *zap.SugaredLogger
	msvc        *crdsv1.ManagedService
}

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=managedservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=managedservices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=managedservices/finalizers,verbs=update

func (r *ManagedServiceReconciler) Reconcile(ctx context.Context, orgReq ctrl.Request) (ctrl.Result, error) {
	req := &MsvcReconReq{Request: orgReq, logger: GetLogger(orgReq.NamespacedName), msvc: new(crdsv1.ManagedService)}
	if err := r.Get(ctx, req.NamespacedName, req.msvc); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	req.condBuilder = fn.Conditions.From(req.msvc.Status.Conditions)

	if !req.msvc.HasLabels() {
		req.msvc.EnsureLabels()
		if err := r.Update(ctx, req.msvc); err != nil {
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
	}

	if req.msvc.GetDeletionTimestamp() != nil {
		return r.finalize(ctx, req)
	}

	ctrlRequest, err := r.reconcileStatus(ctx, req)
	if err != nil {
		return r.failWithErr(ctx, req, err)
	}
	if ctrlRequest != nil {
		return *ctrlRequest, nil
	}

	return r.reconcileOperations(ctx, req)
}

func (r *ManagedServiceReconciler) finalize(ctx context.Context, req *MsvcReconReq) (ctrl.Result, error) {
	msvc := req.msvc
	msvcFinalizer := finalizers.ManagedService.String()

	if controllerutil.ContainsFinalizer(msvc, msvcFinalizer) {
		controllerutil.RemoveFinalizer(msvc, msvcFinalizer)
		return ctrl.Result{}, r.Update(ctx, msvc)
	}

	if controllerutil.ContainsFinalizer(msvc, finalizers.Foreground.String()) {
		resource := unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": msvc.Spec.ApiVersion,
				"kind":       "Service",
			},
		}

		if err := r.Get(ctx, types.NamespacedName{Namespace: msvc.Namespace, Name: msvc.Name}, &resource); err != nil {
			req.logger.Infof("ERR: %+v", err)
			if apiErrors.IsNotFound(err) {
				controllerutil.RemoveFinalizer(msvc, finalizers.Foreground.String())
				return ctrl.Result{}, r.Update(ctx, msvc)
			}
		}

		// HACK: this is a hack to get the service to be deleted, IDK ownerReferences somehow not getting it deleted
		if resource.GetName() != "" {
			if err := r.Delete(ctx, &resource); err != nil {
				return r.failWithErr(ctx, req, err)
			}
		}
	}

	return reconcileResult.OK()
}

func (r *ManagedServiceReconciler) statusUpdate(ctx context.Context, req *MsvcReconReq) error {
	req.msvc.Status.Conditions = req.condBuilder.GetAll()
	if err := r.notify(req); err != nil {
		return err
	}
	return r.Status().Update(ctx, req.msvc)
}

func (r *ManagedServiceReconciler) failWithErr(ctx context.Context, req *MsvcReconReq, err error) (
	ctrl.Result,
	error,
) {
	req.condBuilder.MarkNotReady(err)
	return ctrl.Result{}, r.statusUpdate(ctx, req)
}

func (r *ManagedServiceReconciler) reconcileStatus(ctx context.Context, req *MsvcReconReq) (*ctrl.Result, error) {
	prevStatus := req.msvc.Status
	req.condBuilder.Reset()

	svcObj := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": req.msvc.Spec.ApiVersion,
			"kind":       "Service",
		},
	}

	nn := types.NamespacedName{Namespace: req.msvc.Namespace, Name: req.msvc.Name}
	if err := r.Get(ctx, nn, &svcObj); err != nil {
		if !apiErrors.IsNotFound(err) {
			return nil, err
		}
		req.condBuilder.Build(
			"", metav1.Condition{
				Type:   "NotFound",
				Status: metav1.ConditionTrue,
				Reason: "ResourceNotFound",
				Message: fmt.Sprintf(
					"resource (ApiVersion=%s, Kind=%s) %s not found", req.msvc.Spec.ApiVersion,
					"Service", nn.String(),
				),
			},
		)
	}

	mj, err := svcObj.MarshalJSON()

	if err != nil {
		return nil, err
	}

	var j struct {
		Status struct {
			Conditions []metav1.Condition `json:"conditions,omitempty"`
		} `json:"status,omitempty"`
	}

	if err := json.Unmarshal(mj, &j); err != nil {
		return nil, err
	}
	req.condBuilder.Build("", j.Status.Conditions...)
	if req.condBuilder.Equal(prevStatus.Conditions) {
		req.logger.Infof("Status is already in sync, so moving forward with ops")
		return nil, nil
	}

	req.logger.Infof("status is different, so updating status ...")
	if err := r.statusUpdate(ctx, req); err != nil {
		return nil, err
	}
	return reconcileResult.OKP()
}

func (r *ManagedServiceReconciler) reconcileOperations(ctx context.Context, req *MsvcReconReq) (ctrl.Result, error) {
	hash, err := req.msvc.Hash()
	req.logger.Infof("Hash == req.msvc.Status.LastHash %v", hash == req.msvc.Status.LastHash)
	if hash == req.msvc.Status.LastHash {
		return reconcileResult.OK()
	}

	b, err := templates.Parse(templates.CommonMsvc, req.msvc)
	if err != nil {
		return r.failWithErr(ctx, req, err)
	}

	if _, err := fn.KubectlApply(b); err != nil {
		return reconcileResult.FailedE(
			errors.NewEf(
				err, "could not apply service %s:%s", req.msvc.APIVersion,
				req.msvc.Kind,
			),
		)
	}

	req.msvc.Status.LastHash = hash
	return ctrl.Result{}, r.Status().Update(ctx, req.msvc)
}

func (r *ManagedServiceReconciler) notify(req *MsvcReconReq) error {
	return r.SendMessage(
		req.msvc.NameRef(), lib.MessageReply{
			Key:        req.msvc.NameRef(),
			Conditions: req.condBuilder.GetAll(),
			Status:     req.condBuilder.IsTrue(constants.ConditionReady.Type),
		},
	)
}

func (r *ManagedServiceReconciler) watcherFuncMap(c client.Object) []reconcile.Request {
	var msvcList crdsv1.ManagedServiceList

	key, _ := crdsv1.ManagedService{}.LabelRef()
	if err := r.List(
		context.TODO(), &msvcList, &client.ListOptions{
			LabelSelector: labels.SelectorFromValidatedSet(map[string]string{key: c.GetObjectKind().GroupVersionKind().Group}),
		},
	); err != nil {
		return nil
	}
	var reqs []reconcile.Request
	for _, item := range msvcList.Items {
		nn := types.NamespacedName{Namespace: item.GetNamespace(), Name: item.GetName()}
		for _, req := range reqs {
			if req.NamespacedName.String() == nn.String() {
				return nil
			}
		}
		reqs = append(reqs, reconcile.Request{NamespacedName: nn})
	}
	return reqs
}

// SetupWithManager sets up the controller with the Manager.
func (r *ManagedServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	builder := ctrl.NewControllerManagedBy(mgr).For(&crdsv1.ManagedService{})

	allMsvcs := []string{
		"mongodb-standalone",
		"mongodb-cluster",
		"mysql-standalone",
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
			}, handler.EnqueueRequestsFromMapFunc(r.watcherFuncMap),
		)
	}

	return builder.Complete(r)
}
