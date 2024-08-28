package standalone

import (
	"context"
	"fmt"

	postgresv1 "github.com/kloudlite/operator/apis/postgres.msvc/v1"
	"github.com/kloudlite/operator/operators/msvc-postgres/internal/env"
	"github.com/kloudlite/operator/operators/msvc-postgres/internal/types"
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
)

const (
	kloudliteMsvcComponent string = "kloudlite.io/msvc.component"
)

// +kubebuilder:rbac:groups=mysql.msvc.kloudlite.io,resources=standalone,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mysql.msvc.kloudlite.io,resources=standalone/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mysql.msvc.kloudlite.io,resources=standalone/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(
		rApi.NewReconcilerCtx(ctx, r.logger),
		r.Client,
		request.NamespacedName,
		&postgresv1.Standalone{},
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

	if step := req.EnsureCheckList([]rApi.CheckMeta{
		{Name: createPVC, Title: "create persistent volume claim"},
		{Name: createService, Title: "create service"},
		{Name: createAccessCredentials, Title: "create access credentials"},
		{Name: createStatefulSet, Title: "create statefulset"},
	}); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.patchDefaults(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.createService(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.createAccessCredentials(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.createPVC(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.createStatefulSet(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *Reconciler) finalize(req *rApi.Request[*postgresv1.Standalone]) stepResult.Result {
	check := "finalizing"

	req.LogPreCheck(check)
	defer req.LogPostCheck(check)

	if result := req.CleanupOwnedResources(); !result.ShouldProceed() {
		return result
	}

	return req.Finalize()
}

func (r *Reconciler) patchDefaults(req *rApi.Request[*postgresv1.Standalone]) stepResult.Result {
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

func getKloudliteDNSHostname(obj *postgresv1.Standalone) string {
	return fmt.Sprintf("%s.svc", obj.Name)
}

func (r *Reconciler) createService(req *rApi.Request[*postgresv1.Standalone]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(createService, req)

	svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: obj.Name, Namespace: obj.Namespace}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, svc, func() error {
		fn.MapSet(&svc.Labels, constants.KloudliteDNSHostname, getKloudliteDNSHostname(obj))
		for k, v := range fn.MapFilter(obj.GetLabels(), "kloudlite.io/") {
			fn.MapSet(&svc.Labels, k, v)
		}
		svc.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		svc.Spec = corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:     "postgres",
					Protocol: corev1.ProtocolTCP,
					Port:     5432,
				},
			},
			Selector: fn.MapMerge(fn.MapFilter(obj.GetLabels(), "kloudlite.io/"), map[string]string{kloudliteMsvcComponent: "statefulset"}),
			Type:     corev1.ServiceTypeClusterIP,
		}
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	return check.Completed()
}

func (r *Reconciler) createPVC(req *rApi.Request[*postgresv1.Standalone]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(createPVC, req)

	pvc := &corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: obj.Name, Namespace: obj.Namespace}}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, pvc, func() error {
		pvc.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})

		for k, v := range fn.MapFilter(obj.GetLabels(), "kloudlite.io/") {
			fn.MapSet(&pvc.Labels, k, v)
		}
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

func (r *Reconciler) createAccessCredentials(req *rApi.Request[*postgresv1.Standalone]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(createAccessCredentials, req)

	secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: obj.Output.CredentialsRef.Name, Namespace: obj.Namespace}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, secret, func() error {
		secret.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})

		for k, v := range fn.MapFilter(obj.GetLabels(), "kloudlite.io/") {
			fn.MapSet(&secret.Labels, k, v)
		}

		if secret.Data == nil {
			username := "postgres"
			password := fn.CleanerNanoid(40)

			dbName := "postgres"
			port := "5432"

			gvpnHost := fmt.Sprintf("%s.%s.svc.%s", obj.Name, obj.Namespace, r.Env.GlobalVpnDNS)
			clusterLocalHost := fmt.Sprintf("%s.%s.svc.%s", obj.Name, obj.Namespace, r.Env.ClusterInternalDNS)
			kloudliteDNSHost := fmt.Sprintf("%s.%s", getKloudliteDNSHostname(obj), r.Env.KloudliteDNSSuffix)

			creds := types.StandaloneOutput{
				Username: username,
				Password: password,
				DbName:   dbName,
				Port:     port,

				Host: kloudliteDNSHost,
				URI:  fmt.Sprintf("postgres://%s:%s@%s:%s/%s", username, password, kloudliteDNSHost, port, dbName),

				ClusterLocalHost: clusterLocalHost,
				ClusterLocalURI:  fmt.Sprintf("postgres://%s:%s@%s:%s/%s", username, password, clusterLocalHost, port, dbName),

				GlobalVPNHost: gvpnHost,
				GlobalVPNURI:  fmt.Sprintf("postgres://%s:%s@%s:%s/%s", username, password, gvpnHost, port, dbName),
			}

			m, err := creds.ToMap()
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

func (r *Reconciler) createStatefulSet(req *rApi.Request[*postgresv1.Standalone]) stepResult.Result {
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
							Name:  "postgres",
							Image: "chainguard/postgres@sha256:dd4b0fe468b76db1afe4851acc253379b9a5ba2914222e0d83156de9b126b5db",
							// Image: "postgres:latest",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      pvcName,
									MountPath: "/data",
								},
							},
							Env: []corev1.EnvVar{
								{
									Name: "POSTGRES_USER",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: obj.Output.CredentialsRef.Name,
											},
											Key: "USERNAME",
										},
									},
								},
								{
									Name:  "POSTGRES_HOST_AUTH_METHOD",
									Value: "scram-sha-256",
								},
								{
									Name: "POSTGRES_PASSWORD",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: obj.Output.CredentialsRef.Name,
											},
											Key: "PASSWORD",
										},
									},
								},
								{
									Name:  "PGDATA",
									Value: "/data/pgdata",
								},
								{
									Name:  "POSTGRES_INITDB_ARGS",
									Value: "--auth-host=scram-sha-256",
								},
							},
						},
					},
				},
			},
		}

		if obj.GetGeneration() > 0 {
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

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig(), kubectl.YAMLClientOpts{Logger: r.logger})

	builder := ctrl.NewControllerManagedBy(mgr)
	builder.For(&postgresv1.Standalone{})
	builder.Owns(&corev1.Secret{})
	builder.Owns(&appsv1.StatefulSet{})
	builder.Owns(&corev1.Service{})

	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
