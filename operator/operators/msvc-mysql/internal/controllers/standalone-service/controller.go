package standalone_service

import (
	"context"
	"fmt"

	mysqlMsvcv1 "github.com/kloudlite/operator/apis/mysql.msvc/v1"
	"github.com/kloudlite/operator/operators/msvc-mysql/internal/env"
	"github.com/kloudlite/operator/operators/msvc-mysql/internal/types"
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
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
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
	KeyMsvcOutput string = "msvc-output"
)

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

// +kubebuilder:rbac:groups=mysql.msvc.kloudlite.io,resources=standalone,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mysql.msvc.kloudlite.io,resources=standalone/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mysql.msvc.kloudlite.io,resources=standalone/finalizers,verbs=update

func (r *ServiceReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(
		rApi.NewReconcilerCtx(ctx, r.logger),
		r.Client,
		request.NamespacedName,
		&mysqlMsvcv1.StandaloneService{},
	)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	req.PreReconcile()
	defer req.PostReconcile()

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
		{Name: patchDefaults, Title: "patch defaults"},
		{Name: createService, Title: "create service"},
		{Name: createPVC, Title: "create pvc"},
		{Name: createAccessCredentials, Title: "create access credentials"},
		{Name: createStatefulSet, Title: "create statefulset"},
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

func (r *ServiceReconciler) finalize(req *rApi.Request[*mysqlMsvcv1.StandaloneService]) stepResult.Result {
	check := "finalizing"

	req.LogPreCheck(check)
	defer req.LogPostCheck(check)

	if result := req.CleanupOwnedResources(); !result.ShouldProceed() {
		return result
	}

	return req.Finalize()
}

func (r *ServiceReconciler) patchDefaults(req *rApi.Request[*mysqlMsvcv1.StandaloneService]) stepResult.Result {
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

func getKloudliteDNSHostname(obj *mysqlMsvcv1.StandaloneService) string {
	return fmt.Sprintf("%s.svc", obj.Name)
}

func (r *ServiceReconciler) createService(req *rApi.Request[*mysqlMsvcv1.StandaloneService]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(createService, req)

	svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: obj.Name, Namespace: obj.Namespace}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, svc, func() error {
		svc.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})

		fn.MapSet(&svc.Labels, constants.KloudliteDNSHostname, getKloudliteDNSHostname(obj))
		for k, v := range fn.MapFilter(obj.GetLabels(), "kloudlite.io/") {
			svc.Labels[k] = v
		}

		svc.Spec.Ports = []corev1.ServicePort{
			{
				Name:     "mysql",
				Protocol: corev1.ProtocolTCP,
				Port:     3306,
			},
		}
		svc.Spec.Selector = fn.MapMerge(fn.MapFilter(obj.GetLabels(), "kloudlite.io/"), map[string]string{kloudliteMsvcComponent: "statefulset"})
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	return check.Completed()
}

func (r *ServiceReconciler) createPVC(req *rApi.Request[*mysqlMsvcv1.StandaloneService]) stepResult.Result {
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

func (r *ServiceReconciler) createAccessCredentials(req *rApi.Request[*mysqlMsvcv1.StandaloneService]) stepResult.Result {
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
			port := "3306"

			dbName := "mysql"

			out := types.StandaloneServiceOutput{
				Username: username,
				Password: password,
				Port:     port,
				DbName:   dbName,

				Host: kloudliteDNSHost,
				URI:  fmt.Sprintf("mysql://%s:%s@%s:%s/%s", username, password, kloudliteDNSHost, port, dbName),
				DSN:  fmt.Sprintf("mysql://%s:%s@tcp(%s:%s)/%s", username, password, kloudliteDNSHost, port, dbName),

				ClusterLocalHost: clusterLocalHost,
				ClusterLocalURI:  fmt.Sprintf("mysql://%s:%s@%s:%s/%s", username, password, clusterLocalHost, port, dbName),
				ClusterLocalDSN:  fmt.Sprintf("mysql://%s:%s@tcp(%s:%s)/%s", username, password, clusterLocalHost, port, dbName),

				GlobalVPNHost: globalVPNHost,
				GlobalVPNURI:  fmt.Sprintf("mysql://%s:%s@%s:%s/%s", username, password, globalVPNHost, port, dbName),
				GlobalVPNDSN:  fmt.Sprintf("mysql://%s:%s@tcp(%s:%s)/%s", username, password, globalVPNHost, port, dbName),
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

func (r *ServiceReconciler) createStatefulSet(req *rApi.Request[*mysqlMsvcv1.StandaloneService]) stepResult.Result {
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
							Name: "mariadb",
							// Image: "chainguard/mariadb@sha256:ca14ebcf9196ecfbea06afb6f46128b48f962bf6372b13b349275fad11c47954",
							Image: "mariadb",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      pvcName,
									MountPath: "/var/lib/mysql",
								},
							},
							Env: []corev1.EnvVar{
								{
									Name: "MARIADB_ROOT_PASSWORD",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: obj.Output.CredentialsRef.Name,
											},
											Key: "PASSWORD",
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

	builder := ctrl.NewControllerManagedBy(mgr)
	builder.For(&mysqlMsvcv1.StandaloneService{})
	builder.Owns(&corev1.Secret{})
	builder.Owns(&appsv1.StatefulSet{})
	builder.Owns(&corev1.PersistentVolumeClaim{})
	builder.Owns(&corev1.Service{})

	builder.Watches(
		&appsv1.StatefulSet{}, handler.EnqueueRequestsFromMapFunc(
			func(ctx context.Context, obj client.Object) []reconcile.Request {
				v, ok := obj.GetLabels()[constants.MsvcNameKey]
				if !ok {
					return nil
				}
				return []reconcile.Request{{NamespacedName: fn.NN(obj.GetNamespace(), v)}}
			},
		),
	)

	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
