package external_app

import (
	"context"
	"fmt"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/app-n-lambda/internal/env"
	"github.com/kloudlite/operator/operators/app-n-lambda/internal/templates"
	"github.com/kloudlite/operator/pkg/constants"
	"github.com/kloudlite/operator/pkg/errors"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/toolkit/reconciler"
	stepResult "github.com/kloudlite/operator/toolkit/reconciler/step-result"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	controllerutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type ExternalAppReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Env        *env.Env
	Name       string
	yamlClient kubectl.YAMLClient

	appInterceptTemplate []byte
}

func (r *ExternalAppReconciler) GetName() string {
	return r.Name
}

const (
	createExternalNameService = "createExternalNameService"
	createAppIntercept        = "createIntercept"
)

// +kubebuilder:rbac:groups=crdsv1,resources=external_apps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crdsv1,resources=external_apps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crdsv1,resources=external_apps/finalizers,verbs=update

func (r *ExternalAppReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := reconciler.NewRequest(ctx, r.Client, request.NamespacedName, &crdsv1.ExternalApp{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	req.PreReconcile()
	defer req.PostReconcile()

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureCheckList([]reconciler.CheckMeta{
		{Name: createExternalNameService, Title: "Creates External Name Service", Hide: req.Object.IsInterceptEnabled()},
		{Name: createAppIntercept, Title: "Create App Intercept", Hide: !req.Object.IsInterceptEnabled()},
	}); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.createExternalService(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.checkAppIntercept(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *ExternalAppReconciler) createExternalService(req *reconciler.Request[*crdsv1.ExternalApp]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := reconciler.NewRunningCheck(createExternalNameService, req)

	svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: obj.Name, Namespace: obj.Namespace}}

	spec := func() corev1.ServiceSpec {
		if obj.IsInterceptEnabled() {
			ports := make([]corev1.ServicePort, 0, len(obj.Spec.Intercept.PortMappings))
			for _, v := range obj.Spec.Intercept.PortMappings {
				ports = append(ports, corev1.ServicePort{
					Name:       fmt.Sprintf("p-%d", v.AppPort),
					Protocol:   corev1.ProtocolTCP,
					Port:       int32(v.AppPort),
					TargetPort: intstr.IntOrString{IntVal: int32(v.AppPort)},
					NodePort:   0,
				})
			}
			return corev1.ServiceSpec{
				Ports:    ports,
				Selector: fn.MapFilterWithPrefix(obj.Labels, "kloudlite.io/"),
			}
		}

		espec := corev1.ServiceSpec{Type: corev1.ServiceTypeExternalName}
		switch obj.Spec.RecordType {
		case crdsv1.ExternalAppRecordTypeCNAME:
			{
				espec.ExternalName = obj.Spec.Record
			}
		case crdsv1.ExternalAppRecordTypeIPAddr:
			{
				espec.ExternalIPs = []string{obj.Spec.Record}
			}
		}

		return espec
	}()

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, svc, func() error {
		svc.Spec = spec
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	return check.Completed()
}

func (r *ExternalAppReconciler) checkAppIntercept(req *reconciler.Request[*crdsv1.ExternalApp]) stepResult.Result {
	ctx, obj := req.Context(), req.Object

	podname := obj.Name + "-intercept"
	podns := obj.Namespace
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: podname, Namespace: podns}}

	check := reconciler.NewRunningCheck(createAppIntercept, req)

	if !obj.IsInterceptEnabled() {
		if err := r.Delete(ctx, pod); err != nil {
			if !apiErrors.IsNotFound(err) {
				return check.Failed(err)
			}
		}
		return check.Completed()
	}

	if obj.Spec.Intercept.PortMappings == nil {
		return check.Failed(errors.Newf(".spec.intercept.portMappings is required for intercept"))
	}

	appGenerationLabel := "kloudlite.io/app-generation"

	if err := r.Get(ctx, client.ObjectKeyFromObject(pod), pod); err != nil {
		if apiErrors.IsNotFound(err) {
			portMappings := make(map[uint16]uint16, len(obj.Spec.Intercept.PortMappings))
			for _, pm := range obj.Spec.Intercept.PortMappings {
				portMappings[pm.AppPort] = pm.DevicePort
			}

			deviceHostSuffix := "device.local"
			if obj.Spec.Intercept.DeviceHostSuffix != nil {
				deviceHostSuffix = *obj.Spec.Intercept.DeviceHostSuffix
			}

			b, err := templates.ParseBytes(r.appInterceptTemplate, map[string]any{
				"name":             podname,
				"namespace":        podns,
				"labels":           fn.MapMerge(fn.MapFilterWithPrefix(obj.Labels, "kloudlite.io/"), map[string]string{appGenerationLabel: fmt.Sprintf("%d", obj.Generation)}),
				"owner-references": []metav1.OwnerReference{fn.AsOwner(obj, true)},
				"device-host":      fmt.Sprintf("%s.%s", obj.Spec.Intercept.ToDevice, deviceHostSuffix),
				"port-mappings":    portMappings,
			})
			if err != nil {
				return check.Failed(err).NoRequeue()
			}

			if _, err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
				return check.Failed(err).NoRequeue()
			}

			return check.StillRunning(fmt.Errorf("waiting for intercept pod to start"))
		}
	}

	if pod.Labels[appGenerationLabel] != fmt.Sprintf("%d", obj.Generation) {
		if err := r.Delete(ctx, pod); err != nil {
			return check.Failed(err)
		}
		return check.StillRunning(fmt.Errorf("waiting for previous generation pod to be deleted"))
	}

	if pod.Status.Phase != corev1.PodRunning {
		return check.StillRunning(fmt.Errorf("waiting for pod to start running"))
	}

	return check.Completed()
}

func (r *ExternalAppReconciler) finalize(req *reconciler.Request[*crdsv1.ExternalApp]) stepResult.Result {
	return req.Finalize()
}

func (r *ExternalAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()

	var err error
	r.appInterceptTemplate, err = templates.Read(templates.AppIntercept)
	if err != nil {
		return err
	}

	builder := ctrl.NewControllerManagedBy(mgr).For(&crdsv1.ExternalApp{})
	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	return builder.Complete(r)
}
