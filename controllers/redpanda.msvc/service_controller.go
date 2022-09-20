package redpandamsvc

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ct "operators.kloudlite.io/apis/common-types"
	"operators.kloudlite.io/env"
	"operators.kloudlite.io/lib/conditions"
	"operators.kloudlite.io/lib/constants"
	"operators.kloudlite.io/lib/errors"
	fn "operators.kloudlite.io/lib/functions"
	"operators.kloudlite.io/lib/logging"
	rApi "operators.kloudlite.io/lib/operator"
	stepResult "operators.kloudlite.io/lib/operator/step-result"
	"operators.kloudlite.io/lib/templates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	redpandamsvcv1 "operators.kloudlite.io/apis/redpanda.msvc/v1"
)

// ServiceReconciler reconciles a Service object
type ServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	logger logging.Logger
	Name   string
}

func (r *ServiceReconciler) GetName() string {
	return r.Name
}

const (
	MsvcExists    conditions.Type = "redpanda.msvc/exists"
	MsvcReady     conditions.Type = "redpanda.msvc/Ready"
	OutputExists  conditions.Type = "redpanda.output/Exists"
	ClusterExists conditions.Type = "redpanda.cluster/Exists"
	StsExists     conditions.Type = "redpanda.sts/Exists"
	StsReady      conditions.Type = "redpanda.sts/Ready"
)

const (
	OutputRedpandaHostsKey string = "REDPANDA_HOSTS"
)

// +kubebuilder:rbac:groups=redpanda.msvc.kloudlite.io,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=redpanda.msvc.kloudlite.io,resources=services/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=redpanda.msvc.kloudlite.io,resources=services/finalizers,verbs=update

func (r *ServiceReconciler) Reconcile(ctx context.Context, oReq ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, oReq.NamespacedName, &redpandamsvcv1.Service{})
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

func (r *ServiceReconciler) finalize(req *rApi.Request[*redpandamsvcv1.Service]) stepResult.Result {
	return req.Finalize()
}

func (r *ServiceReconciler) reconcileStatus(req *rApi.Request[*redpandamsvcv1.Service]) stepResult.Result {
	ctx := req.Context()
	obj := req.Object

	isReady := true
	cs := make([]metav1.Condition, 0, 4)

	// redpanda Cluster exists ?
	redpandaCR, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Name), fn.NewUnstructured(constants.RedpandaClusterType))
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.FailWithStatusError(err, cs...)
		}
		isReady = false
		cs = append(cs, conditions.New(ClusterExists, false, conditions.NotFound, err.Error()))
		redpandaCR = nil
	}

	if redpandaCR != nil {
		cs = append(cs, conditions.New(ClusterExists, true, conditions.Found))
		sts, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Name), &appsv1.StatefulSet{})
		if err != nil {
			if !apiErrors.IsNotFound(err) {
				return req.FailWithStatusError(err)
			}
			isReady = false
			cs = append(cs, conditions.New(StsExists, false, conditions.NotFound, err.Error()))
		}

		if sts != nil {
			cs = append(cs, conditions.New(StsExists, true, conditions.Found))
			stsConditions, err := conditions.ParseFromResource(sts, "redpanda.sts/")
			if err != nil {
				return req.FailWithStatusError(err, stsConditions...)
			}

			if sts.Status.Replicas != sts.Status.ReadyReplicas {
				isReady = false
				cs = append(cs, conditions.New(StsReady, false, conditions.Empty))
			} else {
				cs = append(cs, conditions.New(StsReady, true, conditions.Empty))
			}
		}
	}

	nConditions, hasUpdated, err := conditions.Patch(obj.Status.Conditions, cs)
	if err != nil {
		return req.FailWithOpError(err).Err(nil)
	}

	if !hasUpdated && isReady == obj.Status.IsReady {
		return req.Next()
	}

	obj.Status.IsReady = isReady
	obj.Status.Conditions = nConditions
	if err := r.Status().Update(ctx, obj); err != nil {
		return req.FailWithStatusError(err)
	}
	return req.Done()
}

func (r *ServiceReconciler) reconcileOperations(req *rApi.Request[*redpandamsvcv1.Service]) stepResult.Result {
	obj := req.Object
	ctx := req.Context()

	if !fn.ContainsFinalizers(obj, constants.CommonFinalizer, constants.ForegroundFinalizer) {
		controllerutil.AddFinalizer(obj, constants.CommonFinalizer)
		controllerutil.AddFinalizer(obj, constants.ForegroundFinalizer)

		if err := r.Update(ctx, obj); err != nil {
			return req.FailWithOpError(err)
		}
		return req.Done()
	}

	storageClass, err := func() (string, error) {
		if obj.Spec.Storage.StorageClass != "" {
			return obj.Spec.Storage.StorageClass, nil
		}
		return obj.Spec.CloudProvider.GetStorageClass(ct.Xfs)
	}()
	if err != nil {
		return req.FailWithOpError(errors.NewEf(err, "could not find storage class to use")).Err(nil)
	}

	b, err := templates.Parse(
		templates.RedpandaOneNodeCluster, map[string]any{
			"storage-class": storageClass,
			"obj":           obj,
		},
	)
	if err != nil {
		return req.FailWithOpError(err).Err(nil)
	}

	// access config
	b2, err := templates.Parse(
		templates.CoreV1.Secret, map[string]any{
			"name":       "msvc-" + obj.Name,
			"namespace":  obj.Namespace,
			"labels":     obj.GetLabels(),
			"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},
			"string-data": map[string]string{
				OutputRedpandaHostsKey: fmt.Sprintf("%s.%s.svc.cluster.local", obj.Name, obj.Namespace),
			},
		},
	)

	if err != nil {
		return req.FailWithOpError(err).Err(nil)
	}

	if err := fn.KubectlApplyExec(ctx, b, b2); err != nil {
		return req.FailWithOpError(errors.NewEf(err, "failed while kubectl apply for template=%s", templates.RedpandaOneNodeCluster)).Err(nil)
	}

	obj.Status.OpsConditions = []metav1.Condition{}
	if err := r.Status().Update(ctx, obj); err != nil {
		return req.FailWithOpError(err)
	}
	return req.Next()
}

// SetupWithManager sets up the controllers with the Manager.
func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager, envVars *env.Env, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)

	return ctrl.NewControllerManagedBy(mgr).
		For(&redpandamsvcv1.Service{}).
		Complete(r)
}
