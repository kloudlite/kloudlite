package standalone

import (
	"context"
	"fmt"

	redisMsvcv1 "github.com/kloudlite/operator/apis/redis.msvc/v1"
	"github.com/kloudlite/operator/operators/msvc-redis/internal/env"
	"github.com/kloudlite/operator/operators/msvc-redis/internal/types"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type ServiceReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	logger     logging.Logger
	Name       string
	Env        *env.Env
	yamlClient kubectl.YAMLClient
}

func (r *ServiceReconciler) GetName() string {
	return r.Name
}

const (
	patchDefaults           string = "patch-defaults"
	createService           string = "create-service"
	createPVC               string = "create-pvc"
	createAccessCredentials string = "create-access-credentials"
	createStatefulSet       string = "create-statefulset"
)

const (
	kloudliteMsvcComponent string = "kloudlite.io/msvc.component"
)

// +kubebuilder:rbac:groups=redis.msvc.kloudlite.io,resources=standaloneservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=redis.msvc.kloudlite.io,resources=standaloneservices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=redis.msvc.kloudlite.io,resources=standaloneservices/finalizers,verbs=update

func (r *ServiceReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &redisMsvcv1.StandaloneService{})
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

	if step := req.EnsureCheckList([]rApi.CheckMeta{
		{Name: patchDefaults, Title: "Patch Defaults"},
		{Name: createService, Title: "Create Service"},
		{Name: createPVC, Title: "Create PVC"},
		{Name: createAccessCredentials, Title: "Create Access Credentials"},
		{Name: createStatefulSet, Title: "Create StatefulSet"},
	}); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.patchDefaults(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.patchDefaults(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.createService(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.createPVC(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.createAccessCredentials(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.createStatefulSet(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *ServiceReconciler) finalize(req *rApi.Request[*redisMsvcv1.StandaloneService]) stepResult.Result {
	checkName := "finalizing"

	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	if step := req.EnsureCheckList([]rApi.CheckMeta{
		{Name: "resources_deleted", Title: "resources deleted"},
	}); !step.ShouldProceed() {
		return step
	}

	if step := req.CleanupOwnedResources(); !step.ShouldProceed() {
		return step
	}

	return req.Finalize()
}

func (r *ServiceReconciler) patchDefaults(req *rApi.Request[*redisMsvcv1.StandaloneService]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(patchDefaults, req)

	hasUpdate := false

	if obj.Output.CredentialsRef.Name == "" {
		hasUpdate = true
		obj.Output.CredentialsRef.Name = obj.Name
	}

	if hasUpdate {
		if err := r.Update(ctx, obj); err != nil {
			return check.Failed(err)
		}
	}

	return check.Completed()
}

func getKloudliteDNSHostname(obj *redisMsvcv1.StandaloneService) string {
	return fmt.Sprintf("%s.svc", obj.Name)
}

func (r *ServiceReconciler) createService(req *rApi.Request[*redisMsvcv1.StandaloneService]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(createService, req)

	svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: obj.Name, Namespace: obj.Namespace}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, svc, func() error {
		fn.MapSet(&svc.Labels, constants.KloudliteDNSHostname, getKloudliteDNSHostname(obj))
		for k, v := range fn.MapFilter(obj.GetLabels(), "kloudlite.io/") {
			fn.MapSet(&svc.Labels, k, v)
		}

		svc.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		svc.Spec.Ports = []corev1.ServicePort{
			{
				Name:     "redis",
				Protocol: corev1.ProtocolTCP,
				Port:     6379,
			},
		}
		svc.Spec.Selector = fn.MapFilter(obj.GetLabels(), "kloudlite.io/")
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	return check.Completed()
}

func (r *ServiceReconciler) createPVC(req *rApi.Request[*redisMsvcv1.StandaloneService]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(createPVC, req)

	pvc := &corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: obj.Name, Namespace: obj.Namespace}}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, pvc, func() error {
		for k, v := range fn.MapFilter(obj.GetLabels(), "kloudlite.io/") {
			fn.MapSet(&pvc.Labels, k, v)
		}

		pvc.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		pvc.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}

		if pvc.Spec.Resources.Requests == nil {
			pvc.Spec.Resources.Requests = corev1.ResourceList{}
		}

		pvc.Spec.Resources.Requests[corev1.ResourceStorage] = resource.MustParse(string(obj.Spec.Resources.Storage.Size))
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	return check.Completed()
}

func (r *ServiceReconciler) createAccessCredentials(req *rApi.Request[*redisMsvcv1.StandaloneService]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(createAccessCredentials, req)

	secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: obj.Output.CredentialsRef.Name, Namespace: obj.Namespace}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, secret, func() error {
		for k, v := range fn.MapFilter(obj.GetLabels(), "kloudlite.io/") {
			fn.MapSet(&secret.Labels, k, v)
		}
		secret.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})

		if secret.Data == nil {
			password := fn.CleanerNanoid(40)

			clusterLocalHost := fmt.Sprintf("%s.%s.svc.%s", obj.Name, obj.Namespace, r.Env.ClusterInternalDNS)
			globalVPNHost := fmt.Sprintf("%s.%s.svc.%s", obj.Name, obj.Namespace, r.Env.GlobalVpnDNS)
			kloudliteDNSHostFQDN := fmt.Sprintf("%s.%s", getKloudliteDNSHostname(obj), r.Env.KloudliteDNSSuffix)
			port := "6379"

			out := types.StandaloneServiceOutput{
				RootPassword: password,
				Port:         port,

				Host: kloudliteDNSHostFQDN,
				Addr: fmt.Sprintf("%s:%s", kloudliteDNSHostFQDN, port),
				URI:  fmt.Sprintf("redis://:%s@%s?allowUsernameInURI=true", password, kloudliteDNSHostFQDN),

				ClusterLocalHost: clusterLocalHost,
				ClusterLocalAddr: fmt.Sprintf("%s:%s", clusterLocalHost, port),
				ClusterLocalURI:  fmt.Sprintf("redis://:%s@%s?allowUsernameInURI=true", password, globalVPNHost),

				GlobalVpnHost: globalVPNHost,
				GlobalVpnAddr: fmt.Sprintf("%s:%s", globalVPNHost, port),
				GlobalVpnURI:  fmt.Sprintf("redis://:%s@%s?allowUsernameInURI=true", password, globalVPNHost),
			}

			m, err := out.ToMap()
			if err != nil {
				return err
			}

			secret.StringData = m
		}

		return nil
	}); err != nil {
		return check.Failed(err)
	}

	return check.Completed()
}

func (r *ServiceReconciler) createStatefulSet(req *rApi.Request[*redisMsvcv1.StandaloneService]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(createStatefulSet, req)

	pvcName := obj.Name

	sts := &appsv1.StatefulSet{ObjectMeta: metav1.ObjectMeta{Name: obj.Name, Namespace: obj.Namespace}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, sts, func() error {
		sts.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})

		selectorLabels := fn.MapMerge(
			fn.MapFilter(obj.GetLabels(), "kloudlite.io/"),
			map[string]string{kloudliteMsvcComponent: "statefulset"},
		)

		for k, v := range selectorLabels {
			fn.MapSet(&sts.Labels, k, v)
		}

		spec := appsv1.StatefulSetSpec{
			Replicas: fn.New(int32(1)),
			Selector: &metav1.LabelSelector{
				MatchLabels: selectorLabels,
			},
			ServiceName: obj.Name,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: selectorLabels,
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: pvcName,
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: obj.Name,
									ReadOnly:  false,
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "redis",
							Image: "chainguard/redis@sha256:060a0d2c8000f1dc81d1eb3accef50eafc35d9e2584b08f997c532e0fc6178ee",
							Args: []string{
								"--requirepass",
								"$(REDIS_PASSWORD)",
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      pvcName,
									MountPath: "/data",
								},
							},
							Env: []corev1.EnvVar{
								{
									Name: "REDIS_PASSWORD",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: obj.Output.CredentialsRef.Name,
											},
											Key: "ROOT_PASSWORD",
										},
									},
								},
							},
						},
					},
				},
			},
		}

		if sts.GetGeneration() > 0 {
			// resource exists, and is being updated now
			// INFO: k8s statefulsets forbids update to spec fields, other than "replicas", "template", "ordinals", "updateStrategy",  "persistentVolumeClaimRetentionPolicy" and "minReadySeconds",

			sts.Spec.Replicas = spec.Replicas
			sts.Spec.Template = spec.Template
		} else {
			sts.Spec = spec
		}
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	if sts.Status.Replicas > 0 && sts.Status.ReadyReplicas == sts.Status.Replicas {
		return check.Completed()
	}

	return check.StillRunning(fmt.Errorf("waiting for statefulset pods to start"))
}

func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig(), kubectl.YAMLClientOpts{Logger: r.logger})

	builder := ctrl.NewControllerManagedBy(mgr).For(&redisMsvcv1.StandaloneService{})
	builder.Owns(&corev1.Secret{})
	builder.Owns(&corev1.Service{})
	builder.Owns(&corev1.PersistentVolumeClaim{})
	builder.Owns(&appsv1.StatefulSet{})

	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
