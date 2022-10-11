package driver

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ct "operators.kloudlite.io/apis/common-types"
	csiv1 "operators.kloudlite.io/apis/csi/v1"
	"operators.kloudlite.io/lib/constants"
	fn "operators.kloudlite.io/lib/functions"
	"operators.kloudlite.io/lib/logging"
	rApi "operators.kloudlite.io/lib/operator"
	stepResult "operators.kloudlite.io/lib/operator/step-result"
	"operators.kloudlite.io/lib/templates"
	"operators.kloudlite.io/operators/csi-drivers/internal/env"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
	logger logging.Logger
	Name   string
	Env    *env.Env
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	CSIDriversReady     string = "csi-drivers-ready"
	StorageClassesReady string = "storage-classes-ready"
)

// +kubebuilder:rbac:groups=csi.kloudlite.io,resources=drivers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=csi.kloudlite.io,resources=drivers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=csi.kloudlite.io,resources=drivers/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, request.NamespacedName, &csiv1.Driver{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	req.Logger.Infof("NEW RECONCILATION")

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.RestartIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	// TODO: initialize all checks here
	if step := req.EnsureChecks(CSIDriversReady); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconCSIDriver(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconStorageClasses(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	req.Logger.Infof("RECONCILATION COMPLETE")
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) finalize(req *rApi.Request[*csiv1.Driver]) stepResult.Result {
	return req.Finalize()
}

func (r *Reconciler) reconCSIDriver(req *rApi.Request[*csiv1.Driver]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	if obj.Spec.Provider == "aws" {
		accessSecret, err := rApi.Get(ctx, r.Client, fn.NN("kl-core", obj.Spec.SecretRef), &corev1.Secret{})
		if err != nil {
			return req.CheckFailed(CSIDriversReady, check, err.Error()).Err(nil)
		}

		b, err := templates.Parse(
			templates.AwsEbsCsiDriver, map[string]any{
				"name":            fmt.Sprintf("%s-%s-csi", obj.Name, obj.Spec.Provider),
				"namespace":       obj.Spec.SecretRef,
				"aws-secret-name": obj.Spec.SecretRef,
				"aws-key":         string(accessSecret.Data["accessKey"]),
				"aws-secret":      string(accessSecret.Data["accessSecret"]),
				"owner-refs":      []metav1.OwnerReference{fn.AsOwner(obj, true)},
				"node-selector": map[string]string{
					"kloudlite.io/provider-ref": obj.Name,
				},
			},
		)
		if err != nil {
			return req.CheckFailed(CSIDriversReady, check, err.Error()).Err(nil)
		}
		if err := fn.KubectlApplyExec(ctx, b); err != nil {
			return req.CheckFailed(CSIDriversReady, check, err.Error()).Err(nil)
		}
	}

	check.Status = true
	if check != checks[CSIDriversReady] {
		checks[CSIDriversReady] = check
		return req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) reconStorageClasses(req *rApi.Request[*csiv1.Driver]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	edgesList := unstructured.UnstructuredList{
		Object: map[string]any{
			"apiVersion": constants.EdgeInfraType.APIVersion,
			"kind":       constants.EdgeInfraType.Kind,
		},
	}

	if obj.Spec.Provider == "aws" {
		if err := r.List(
			ctx, &edgesList, &client.ListOptions{
				LabelSelector: labels.SelectorFromValidatedSet(
					map[string]string{
						"kloudlite.io/provider-ref": obj.Name,
					},
				),
			},
		); err != nil {
			return req.CheckFailed(StorageClassesReady, check, err.Error())
		}

		for i := range edgesList.Items {
			b, err := templates.Parse(
				templates.AwsEbsStorageClass, map[string]any{
					"name":        edgesList.Items[i].GetName(),
					"driver-name": obj.Name,
					"fs-types":    []ct.FsType{ct.Ext4, ct.Xfs},
				},
			)
			if err != nil {
				return req.CheckFailed(StorageClassesReady, check, err.Error()).Err(nil)
			}

			if err := fn.KubectlApplyExec(ctx, b); err != nil {
				return req.CheckFailed(StorageClassesReady, check, err.Error()).Err(nil)
			}
		}
	}

	check.Status = true
	if check != checks[StorageClassesReady] {
		checks[StorageClassesReady] = check
		return req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)

	builder := ctrl.NewControllerManagedBy(mgr).For(&csiv1.Driver{})
	// builder.Watches(
	// 	&source.Kind{Type: fn.NewUnstructured(constants.EdgeInfraType)}, handler.EnqueueRequestsFromMapFunc(
	// 		func(obj client.Object) []reconcile.Request {
	// 			var drivers csiv1.DriverList
	// 			if err := r.List(
	// 				context.TODO(), &drivers, &client.ListOptions{
	// 					LabelSelector: labels.SelectorFromValidatedSet(
	// 						map[string]string{
	// 							constants.LabelKeys.CsiForEdge: obj.GetName(),
	// 						},
	// 					),
	// 				},
	// 			); err != nil {
	// 				return nil
	// 			}
	//
	// 			for i := range drivers.Items {
	//
	// 			}
	//
	// 			return nil
	// 		},
	// 	),
	// )
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
