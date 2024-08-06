package standalone_service

import (
	"context"
	"fmt"

	mongodbMsvcv1 "github.com/kloudlite/operator/apis/mongodb.msvc/v1"
	"github.com/kloudlite/operator/operators/msvc-mongo/internal/env"
	"github.com/kloudlite/operator/operators/msvc-mongo/internal/types"
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

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	logger     logging.Logger
	Name       string
	Env        *env.Env
	yamlClient kubectl.YAMLClient
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	patchDefaults           string = "patch-defaults"
	createPVC               string = "create-pvc"
	createService           string = "create-service"
	createAccessCredentials string = "create-access-credentials"
	createStatefulSet       string = "create-statefulset"

	cleanupOwnedResources string = "cleanupOwnedResources"
)

const (
	kloudliteMsvcComponent string = "kloudlite.io/msvc.component"
)

var DeleteCheckList = []rApi.CheckMeta{
	{Name: cleanupOwnedResources, Title: "Cleaning owned resources"},
}

// +kubebuilder:rbac:groups=mongodb.msvc.kloudlite.io,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mongodb.msvc.kloudlite.io,resources=services/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mongodb.msvc.kloudlite.io,resources=services/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &mongodbMsvcv1.StandaloneService{})
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

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureCheckList([]rApi.CheckMeta{
		{Name: patchDefaults, Title: "Defaults Patched", Debug: true},
		{Name: createService, Title: "Access Credentials Generated"},
		{Name: createPVC, Title: "MongoDB Helm Applied"},
		{Name: createAccessCredentials, Title: "MongoDB Helm Ready"},
		{Name: createStatefulSet, Title: "MongoDB StatefulSets Ready"},
	}); !step.ShouldProceed() {
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

func (r *Reconciler) finalize(req *rApi.Request[*mongodbMsvcv1.StandaloneService]) stepResult.Result {
	check := "finalizing"

	req.LogPreCheck(check)
	defer req.LogPostCheck(check)

	if result := req.CleanupOwnedResources(); !result.ShouldProceed() {
		return result
	}

	return req.Finalize()
}

func (r *Reconciler) patchDefaults(req *rApi.Request[*mongodbMsvcv1.StandaloneService]) stepResult.Result {
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

func getKloudliteDNSHostname(obj *mongodbMsvcv1.StandaloneService) string {
	return fmt.Sprintf("%s.svc", obj.Name)
}

func (r *Reconciler) createService(req *rApi.Request[*mongodbMsvcv1.StandaloneService]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(createService, req)

	svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: obj.Name, Namespace: obj.Namespace}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, svc, func() error {
		fn.MapSet(&svc.Labels, constants.KloudliteDNSHostname, getKloudliteDNSHostname(obj))
		for k, v := range obj.GetLabels() {
			fn.MapSet(&svc.Labels, k, v)
		}

		svc.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		svc.Spec.Ports = []corev1.ServicePort{
			{
				Name:     "mongo",
				Protocol: corev1.ProtocolTCP,
				Port:     27017,
			},
		}
		svc.Spec.Selector = fn.MapMerge(fn.MapFilter(obj.GetLabels(), "kloudlite.io/"), map[string]string{"kloudlite.io/msvc.component": "statefulset"})
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	return check.Completed()
}

func (r *Reconciler) createPVC(req *rApi.Request[*mongodbMsvcv1.StandaloneService]) stepResult.Result {
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
		if obj.Spec.Resources.Storage.StorageClass != "" {
			pvc.Spec.StorageClassName = &obj.Spec.Resources.Storage.StorageClass
		}
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	return check.Completed()
}

func (r *Reconciler) createAccessCredentials(req *rApi.Request[*mongodbMsvcv1.StandaloneService]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(createAccessCredentials, req)

	secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: obj.Output.CredentialsRef.Name, Namespace: obj.Namespace}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, secret, func() error {
		for k, v := range fn.MapFilter(obj.GetLabels(), "kloudlite.io/") {
			fn.MapSet(&secret.Labels, k, v)
		}
		secret.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})

		if secret.Data == nil {
			username := "root"
			password := fn.CleanerNanoid(40)

			clusterLocalHost := fmt.Sprintf("%s.%s.svc.%s", obj.Name, obj.Namespace, r.Env.ClusterInternalDNS)
			globalVPNHost := fmt.Sprintf("%s.%s.svc.%s", obj.Name, obj.Namespace, r.Env.GlobalVpnDNS)
			kloudliteDNSHost := fmt.Sprintf("%s.%s", getKloudliteDNSHostname(obj), r.Env.KloudliteDNSSuffix)
			port := "27017"

			dbName := "admin"
			out := types.StandaloneSvcOutput{
				RootUsername: username,
				RootPassword: password,
				DBName:       dbName,
				AuthSource:   dbName,

				Port: port,

				Host: kloudliteDNSHost,
				Addr: fmt.Sprintf("%s:%s", kloudliteDNSHost, port),
				URI:  fmt.Sprintf("mongodb://%s:%s@%s:%s/%s?authSource=%s", username, password, kloudliteDNSHost, port, dbName, dbName),

				ClusterLocalHost: clusterLocalHost,
				ClusterLocalAddr: fmt.Sprintf("%s:%s", clusterLocalHost, port),
				ClusterLocalURI:  fmt.Sprintf("mongodb://%s:%s@%s:%s/%s?authSource=%s", username, password, clusterLocalHost, port, dbName, dbName),

				GlobalVpnHost: globalVPNHost,
				GlobalVpnAddr: fmt.Sprintf("%s:%s", globalVPNHost, port),
				GlobalVpnURI:  fmt.Sprintf("mongodb://%s:%s@%s:%s/%s?authSource=%s", username, password, globalVPNHost, port, dbName, dbName),
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

func (r *Reconciler) createStatefulSet(req *rApi.Request[*mongodbMsvcv1.StandaloneService]) stepResult.Result {
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

		sts.Spec.Replicas = fn.New(int32(1))

		sts.Spec = appsv1.StatefulSetSpec{
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
							Name:  "mongodb",
							Image: "mongo:latest",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      pvcName,
									MountPath: "/data/db",
								},
							},
							Env: []corev1.EnvVar{
								{
									Name: "MONGO_INITDB_ROOT_USERNAME",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: obj.Output.CredentialsRef.Name,
											},
											Key: "ROOT_USERNAME",
										},
									},
								},
								{
									Name: "MONGO_INITDB_ROOT_PASSWORD",
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
		return nil
	}); err != nil {
		r.logger.Infof("Failed to create statefulset: err=%v, resource:\n%+v\n", err, sts)
		return check.Failed(err)
	}

	if sts.Status.Replicas > 0 && sts.Status.ReadyReplicas == sts.Status.Replicas {
		return check.Completed()
	}

	return check.StillRunning(fmt.Errorf("waiting for statefulset pods to start"))
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig(), kubectl.YAMLClientOpts{Logger: r.logger})

	builder := ctrl.NewControllerManagedBy(mgr).For(&mongodbMsvcv1.StandaloneService{})
	builder.Owns(&corev1.Secret{})
	builder.Owns(&corev1.PersistentVolumeClaim{})
	builder.Owns(&corev1.Service{})
	builder.Owns(&appsv1.StatefulSet{})

	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
