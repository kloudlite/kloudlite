package driver

import (
	"context"
	"fmt"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"time"

	ct "github.com/kloudlite/operator/apis/common-types"
	csiv1 "github.com/kloudlite/operator/apis/csi/v1"
	"github.com/kloudlite/operator/operators/csi-drivers/internal/env"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	"github.com/kloudlite/operator/pkg/templates"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	logger     logging.Logger
	Name       string
	Env        *env.Env
	yamlClient *kubectl.YAMLClient
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	CSIDriversReady     string = "csi-drivers-ready"
	StorageClassesReady string = "storage-classes-ready"
	Finalizing          string = "finalizing"
)

func getDriverName(obj *csiv1.Driver) string {
	return fmt.Sprintf("%s-%s-csi", fn.Md5([]byte(obj.Name)), obj.Spec.Provider)
}

func getDriverNs(obj *csiv1.Driver) string {
	return "kl-" + obj.Name
}

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

	req.LogPreReconcile()
	defer req.LogPostReconcile()

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
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
	req.Object.Status.LastReconcileTime = &metav1.Time{Time: time.Now()}
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) finalize(req *rApi.Request[*csiv1.Driver]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	// 1. Ensure Deleting of CSI Drivers
	if obj.Spec.Provider == "aws" {
		nn := fn.NN(getDriverNs(obj), getDriverName(obj))
		helmDriver := fn.NewUnstructured(constants.HelmAwsEbsCsiKind, metav1.ObjectMeta{Namespace: nn.Namespace, Name: nn.Name})
		err := r.Delete(ctx, helmDriver)
		if !apiErrors.IsNotFound(err) {
			if err != nil {
				return req.CheckFailed(Finalizing, check, err.Error()).Err(nil)
			}
			return req.CheckFailed(Finalizing, check, fmt.Sprintf("waiting for AwsEbsCsiDriver (%s) to be fully deleted", nn.String())).Err(nil)
		}
	}

	if obj.Spec.Provider == "do" {
		nn := fn.NN(getDriverNs(obj), getDriverName(obj))
		helmDriver := fn.NewUnstructured(constants.HelmDigitaloceanCsiKind, metav1.ObjectMeta{Namespace: nn.Namespace, Name: nn.Name})
		err := r.Delete(ctx, helmDriver)
		if !apiErrors.IsNotFound(err) {
			if err != nil {
				return req.CheckFailed(Finalizing, check, err.Error()).Err(nil)
			}
			return req.CheckFailed(Finalizing, check, fmt.Sprintf("waiting for DigitaloceanCsi (%s) to be fully deleted", nn.String())).Err(nil)
		}
	}

	// 2. Ensure All Storage Classes to be Deleted
	var scList storagev1.StorageClassList
	if err := r.List(ctx, &scList, &client.ListOptions{
		LabelSelector: labels.SelectorFromValidatedSet(obj.GetEnsuredLabels()),
	}); err != nil {
		return req.CheckFailed(Finalizing, check, err.Error()).Err(nil)
	}

	for i := range scList.Items {
		if scList.Items[i].GetDeletionTimestamp() != nil {
			if err := r.Delete(ctx, &scList.Items[i]); err != nil {
				return req.CheckFailed(Finalizing, check, err.Error()).Err(nil)
			}
		}
	}

	if len(scList.Items) != 0 {
		return req.CheckFailed(Finalizing, check, "waiting for storage classes to be fully deleted").Err(nil)
	}

	check.Status = true
	if check != obj.Status.Checks[Finalizing] {
		obj.Status.Checks[Finalizing] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) reconCSIDriver(req *rApi.Request[*csiv1.Driver]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	if obj.Spec.Provider == "aws" {
		accessSecret, err := rApi.Get(ctx, r.Client, fn.NN("kl-core", obj.Spec.SecretRef), &corev1.Secret{})
		if err != nil {
			return req.CheckFailed(CSIDriversReady, check, err.Error()).Err(nil)
		}

		b, err := templates.Parse(
			templates.AwsEbsCsiDriver, map[string]any{
				"name":       fmt.Sprintf("%s-%s-csi", fn.Md5([]byte(obj.Name)), obj.Spec.Provider),
				"namespace":  `kl-` + obj.Name,
				"aws-key":    string(accessSecret.Data["accessKey"]),
				"aws-secret": string(accessSecret.Data["accessSecret"]),
				"labels":     obj.GetEnsuredLabels(),
				"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},
				"node-selector": map[string]string{
					constants.ProviderRef: obj.Name,
				},
			},
		)
		if err != nil {
			return req.CheckFailed(CSIDriversReady, check, err.Error()).Err(nil)
		}
		if err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
			return req.CheckFailed(CSIDriversReady, check, err.Error()).Err(nil)
		}
	}

	if obj.Spec.Provider == "do" {
		accessSecret, err := rApi.Get(ctx, r.Client, fn.NN("kl-core", obj.Spec.SecretRef), &corev1.Secret{})
		if err != nil {
			return req.CheckFailed(CSIDriversReady, check, err.Error()).Err(nil)
		}

		b, err := templates.Parse(
			templates.DigitaloceanCSIDriver, map[string]any{
				"name":      fmt.Sprintf("%s-%s-csi", fn.Md5([]byte(obj.Name)), obj.Spec.Provider),
				"namespace": "kl-" + obj.Name,
				"node-selector": map[string]string{
					constants.ProviderRef: obj.Spec.SecretRef,
				},
				"labels":          obj.GetEnsuredLabels(),
				"owner-refs":      []metav1.OwnerReference{fn.AsOwner(obj, true)},
				"do-access-token": string(accessSecret.Data["apiToken"]),
			},
		)

		if err != nil {
			return req.CheckFailed(CSIDriversReady, check, err.Error()).Err(nil)
		}

		if err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
			return req.CheckFailed(CSIDriversReady, check, err.Error()).Err(nil)
		}
	}

	check.Status = true
	if check != obj.Status.Checks[CSIDriversReady] {
		obj.Status.Checks[CSIDriversReady] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) reconStorageClasses(req *rApi.Request[*csiv1.Driver]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	edgesList := unstructured.UnstructuredList{
		Object: map[string]any{
			"apiVersion": constants.EdgeInfraType.APIVersion,
			"kind":       constants.EdgeInfraType.Kind,
		},
	}

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

	if obj.Spec.Provider == "aws" {
		for i := range edgesList.Items {
			b, err := templates.Parse(
				templates.AwsEbsStorageClass, map[string]any{
					"name":        edgesList.Items[i].GetName(),
					"fs-types":    []ct.FsType{ct.Ext4, ct.Xfs},
					"owner-refs":  []metav1.OwnerReference{fn.AsOwner(obj, true)},
					"provisioner": fmt.Sprintf("%s-%s-csi", fn.Md5([]byte(obj.Name)), obj.Spec.Provider),
					"labels":      obj.GetEnsuredLabels(),
				},
			)
			if err != nil {
				return req.CheckFailed(StorageClassesReady, check, err.Error()).Err(nil)
			}

			if err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
				return req.CheckFailed(StorageClassesReady, check, err.Error()).Err(nil)
			}
		}
	}

	if obj.Spec.Provider == "do" {
		for i := range edgesList.Items {
			b, err := templates.Parse(
				templates.DigitaloceanStorageClass, map[string]any{
					"name":        edgesList.Items[i].GetName(),
					"fs-types":    []ct.FsType{ct.Ext4, ct.Xfs},
					"owner-refs":  []metav1.OwnerReference{fn.AsOwner(obj, true)},
					"provisioner": fmt.Sprintf("%s-%s-csi", fn.Md5([]byte(obj.Name)), obj.Spec.Provider),
					"labels":      obj.GetEnsuredLabels(),
				},
			)

			if err != nil {
				return req.CheckFailed(StorageClassesReady, check, err.Error()).Err(nil)
			}

			if err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
				return req.CheckFailed(StorageClassesReady, check, err.Error()).Err(nil)
			}
		}
	}

	check.Status = true
	if check != obj.Status.Checks[StorageClassesReady] {
		obj.Status.Checks[StorageClassesReady] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

	builder := ctrl.NewControllerManagedBy(mgr).For(&csiv1.Driver{})
	builder.Owns(fn.NewUnstructured(constants.HelmAwsEbsCsiKind))
	builder.Owns(fn.NewUnstructured(constants.HelmDigitaloceanCsiKind))
	builder.Watches(
		&source.Kind{Type: &storagev1.StorageClass{}}, handler.EnqueueRequestsFromMapFunc(
			func(obj client.Object) []reconcile.Request {
				s, ok := obj.GetLabels()[constants.CsiDriverNameKey]
				if !ok {
					return nil
				}
				return []reconcile.Request{{NamespacedName: fn.NN("", s)}}
			},
		),
	)
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
