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
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, nil
}

func (r *ServiceReconciler) finalize(req *rApi.Request[*mysqlMsvcv1.StandaloneService]) stepResult.Result {
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

func (r *ServiceReconciler) createService(req *rApi.Request[*mysqlMsvcv1.StandaloneService]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(createService, req)

	svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: obj.Name, Namespace: obj.Namespace}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, svc, func() error {
		svc.SetLabels(obj.GetLabels())
		svc.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		svc.Spec.Ports = []corev1.ServicePort{
			{
				Name:     "mysql",
				Protocol: corev1.ProtocolTCP,
				Port:     3306,
			},
		}
		svc.Spec.Selector = fn.MapFilter(obj.GetLabels(), "kloudlite.io/")
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
		pvc.SetLabels(obj.GetLabels())
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
		secret.SetLabels(obj.GetLabels())
		secret.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})

		if secret.Data == nil {
			username := "root"
			password := fn.CleanerNanoid(40)

			clusterLocalHost := fmt.Sprintf("%s.%s.svc.%s", obj.Name, obj.Namespace, r.Env.ClusterInternalDNS)
			globalVPNHost := fmt.Sprintf("%s.%s.svc.%s", obj.Name, obj.Namespace, r.Env.GlobalVpnDNS)
			port := "3306"

			dbName := "mysql"

			out := types.StandaloneServiceOutput{
				StandaloneDatabaseOutput: types.StandaloneDatabaseOutput{
					Username: username,
					Password: password,
					Port:     port,
					DbName:   dbName,

					Host: globalVPNHost,
					URI:  fmt.Sprintf("mysql://%s:%s@%s:%s/%s", username, password, globalVPNHost, port, dbName),
					DSN:  fmt.Sprintf("mysql://%s:%s@tcp(%s:%s)/%s", username, password, globalVPNHost, port, dbName),

					ClusterLocalHost: clusterLocalHost,
					ClusterLocalURI:  fmt.Sprintf("mysql://%s:%s@%s:%s/%s", username, password, clusterLocalHost, port, dbName),
					ClusterLocalDSN:  fmt.Sprintf("mysql://%s:%s@tcp(%s:%s)/%s", username, password, clusterLocalHost, port, dbName),
				},
			}

			secret.Data = out.ToSecretData()
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
		sts.SetLabels(obj.GetLabels())

		sts.Spec = appsv1.StatefulSetSpec{
			Replicas: fn.New(int32(1)),
			Selector: &metav1.LabelSelector{
				MatchLabels: fn.MapFilter(obj.GetLabels(), "kloudlite.io/"),
			},
			ServiceName: obj.Name,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: fn.MapFilter(obj.GetLabels(), "kloudlite.io/"),
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
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	return check.Completed()
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
