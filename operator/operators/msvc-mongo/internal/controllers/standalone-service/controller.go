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

	templateHelmMongoDB     []byte
	templateHelmMongoDBAuth []byte
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	Cleanup       string = "cleanup"
	KeyMsvcOutput string = "msvc-output"

	AnnotationCurrentStorageSize string = "kloudlite.io/msvc.storage-size"
)

const (
	patchDefaults           string = "patch-defaults"
	createPVC               string = "create-pvc"
	createService           string = "create-service"
	createAccessCredentials string = "create-access-credentials"
	createStatefulSet       string = "create-statefulset"
)

// DefaultsPatched string = "defaults-patched"
var DeleteCheckList = []rApi.CheckMeta{}

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

func (r *Reconciler) createService(req *rApi.Request[*mongodbMsvcv1.StandaloneService]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(createService, req)

	svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: obj.Name, Namespace: obj.Namespace}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, svc, func() error {
		svc.SetLabels(obj.GetLabels())
		svc.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		svc.Spec.Ports = []corev1.ServicePort{
			{
				Name:     "mongo",
				Protocol: corev1.ProtocolTCP,
				Port:     27017,
			},
		}
		svc.Spec.Selector = fn.MapFilter(obj.GetLabels(), "kloudlite.io/")
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

func (r *Reconciler) createAccessCredentials(req *rApi.Request[*mongodbMsvcv1.StandaloneService]) stepResult.Result {
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
			port := "27017"

			out := types.StandaloneSvcOutput{
				RootUsername: username,
				RootPassword: password,
				DBName:       "admin",
				AuthSource:   "admin",

				Port: port,

				Hosts: globalVPNHost,
				Addr:  fmt.Sprintf("%s:%s", globalVPNHost, port),
				URI:   fmt.Sprintf("mongodb://%s:%s@%s/%s?authSource=%s", username, password, globalVPNHost, "admin", "admin"),

				ClusterLocalHosts: clusterLocalHost,
				ClusterLocalAddr:  fmt.Sprintf("%s:%s", clusterLocalHost, port),
				ClusterLocalURI:   fmt.Sprintf("mongodb://%s:%s@%s/%s?authSource=%s", username, password, clusterLocalHost, "admin", "admin"),
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
							Name: "mongodb",
							// Image: "chainguard/mongodb@sha256:393fef34550e77d390da54855c81f5665a2a51af3a8cf24c13563d22eda0c8ae",
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
		return check.Failed(err)
	}

	if sts.Status.Replicas > 0 && sts.Status.ReadyReplicas != sts.Status.Replicas {
		return check.Failed(fmt.Errorf("waiting for statefulset pods to start"))
	}

	return check.Completed()
}

// func (r *Reconciler) applyMongoDBStandaloneHelm(req *rApi.Request[*mongodbMsvcv1.StandaloneService]) stepResult.Result {
// 	ctx, obj := req.Context(), req.Object
// 	check := rApi.NewRunningCheck(MongoDBHelmApplied, req)
//
// 	if _, ok := obj.GetAnnotations()[AnnotationCurrentStorageSize]; !ok {
// 		fn.MapSet(&obj.Annotations, AnnotationCurrentStorageSize, string(obj.Spec.Resources.Storage.Size))
// 		if err := r.Update(ctx, obj); err != nil {
// 			return check.Failed(err)
// 		}
// 	}
//
// 	hc, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Name), &crdsv1.HelmChart{})
// 	if err != nil {
// 		if !apiErrors.IsNotFound(err) {
// 			return check.Failed(err)
// 		}
// 		hc = nil
// 	}
//
// 	if hc != nil {
// 		req.AddToOwnedResources(rApi.ParseResourceRef(hc))
// 	}
//
// 	hasPVCUpdates := false
//
// 	if hc != nil {
// 		oldSize, ok := obj.GetAnnotations()[AnnotationCurrentStorageSize]
// 		if ok {
// 			oldSizeNum, err := ct.StorageSize(oldSize).ToInt()
// 			if err != nil {
// 				return check.Failed(err).Err(nil)
// 			}
//
// 			newSizeNum, err := ct.StorageSize(obj.Spec.Resources.Storage.Size).ToInt()
// 			if err != nil {
// 				return check.Failed(err).Err(nil)
// 			}
//
// 			if oldSizeNum > newSizeNum {
// 				return check.Failed(fmt.Errorf("new storage size (%s), must be higher than or equal to old size (%s)", obj.Spec.Resources.Storage.Size, oldSize)).Err(nil)
// 			}
//
// 			hasPVCUpdates = newSizeNum > oldSizeNum
// 		}
//
// 		if !ok {
// 			hasPVCUpdates = true
// 		}
//
// 		if hasPVCUpdates {
// 			// need to do something
// 			// 1. Patch the PVC directly
// 			// 2. Rollout the Statefulsets
//
// 			ss, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Name), &appsv1.StatefulSet{})
// 			if err != nil {
// 				if !apiErrors.IsNotFound(err) {
// 					return check.Failed(err).Err(nil)
// 				}
// 				ss = nil
// 			}
//
// 			if ss != nil {
// 				m := types.ExtractPVCLabelsFromStatefulSetLabels(ss.GetLabels())
//
// 				var pvclist corev1.PersistentVolumeClaimList
// 				if err := r.List(ctx, &pvclist, &client.ListOptions{
// 					LabelSelector: apiLabels.SelectorFromValidatedSet(m),
// 					Namespace:     obj.Namespace,
// 				}); err != nil {
// 					return check.Failed(err)
// 				}
//
// 				for i := range pvclist.Items {
// 					pvclist.Items[i].Spec.Resources.Requests[corev1.ResourceStorage] = resource.MustParse(string(obj.Spec.Resources.Storage.Size))
// 					if err := r.Update(ctx, &pvclist.Items[i]); err != nil {
// 						return check.Failed(err)
// 					}
// 				}
//
// 				// STEP 2: rollout statefulset
// 				if err := fn.RolloutRestart(r.Client, fn.StatefulSet, obj.Namespace, ss.GetLabels()); err != nil {
// 					return check.Failed(err)
// 				}
//
// 				fn.MapSet(&obj.Annotations, AnnotationCurrentStorageSize, string(obj.Spec.Resources.Storage.Size))
// 				if err := r.Update(ctx, obj); err != nil {
// 					return check.Failed(err)
// 				}
// 			}
// 		}
// 	}
//
// 	if !hasPVCUpdates {
// 		b, err := templates.ParseBytes(r.templateHelmMongoDB, map[string]any{
// 			"name":          obj.Name,
// 			"namespace":     obj.Namespace,
// 			"labels":        obj.GetLabels(),
// 			"owner-refs":    []metav1.OwnerReference{fn.AsOwner(obj, true)},
// 			"node-selector": obj.Spec.NodeSelector,
// 			"tolerations":   obj.Spec.Tolerations,
//
// 			"pod-labels":      obj.GetLabels(),
// 			"pod-annotations": fn.FilterObservabilityAnnotations(obj.GetAnnotations()),
//
// 			"storage-class": obj.Spec.Resources.Storage.StorageClass,
// 			"storage-size":  obj.Spec.Resources.Storage.Size,
//
// 			"requests-cpu": obj.Spec.Resources.Cpu.Min,
// 			"requests-mem": obj.Spec.Resources.Memory.Min,
//
// 			"limits-cpu": obj.Spec.Resources.Cpu.Max,
// 			"limits-mem": obj.Spec.Resources.Memory.Max,
//
// 			"existing-secret": getHelmSecretName(obj.Name),
// 		})
// 		if err != nil {
// 			return check.Failed(err).Err(nil)
// 		}
//
// 		if _, err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
// 			return check.Failed(err)
// 		}
// 	}
//
// 	return check.Completed()
// }
//
// func (r *Reconciler) checkHelmReady(req *rApi.Request[*mongodbMsvcv1.StandaloneService]) stepResult.Result {
// 	ctx, obj := req.Context(), req.Object
// 	check := rApi.NewRunningCheck(MongoDBHelmReady, req)
//
// 	hc, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, obj.Name), &crdsv1.HelmChart{})
// 	if err != nil {
// 		return check.Failed(err)
// 	}
//
// 	if !hc.Status.IsReady {
// 		return check.Failed(fmt.Errorf("waiting for helm installation to complete"))
// 	}
//
// 	return check.Completed()
// }
//
// func (r *Reconciler) reconSts(req *rApi.Request[*mongodbMsvcv1.StandaloneService]) stepResult.Result {
// 	ctx, obj := req.Context(), req.Object
// 	check := rApi.NewRunningCheck(MongoDBStatefulSetsReady, req)
//
// 	var stsList appsv1.StatefulSetList
// 	if err := r.List(
// 		ctx, &stsList, &client.ListOptions{
// 			LabelSelector: apiLabels.SelectorFromValidatedSet(
// 				map[string]string{constants.MsvcNameKey: obj.Name},
// 			),
// 			Namespace: obj.Namespace,
// 		},
// 	); err != nil {
// 		return check.Failed(err)
// 	}
//
// 	if len(stsList.Items) == 0 {
// 		return check.Failed(fmt.Errorf("no statefulset pods found, waiting for helm controller to reconcile the resource"))
// 	}
//
// 	for i := range stsList.Items {
// 		item := stsList.Items[i]
// 		if item.Status.AvailableReplicas != item.Status.Replicas {
// 			check.Status = false
//
// 			var podsList corev1.PodList
// 			if err := r.List(
// 				ctx, &podsList, &client.ListOptions{
// 					LabelSelector: apiLabels.SelectorFromValidatedSet(
// 						map[string]string{
// 							constants.MsvcNameKey: obj.Name,
// 						},
// 					),
// 				},
// 			); err != nil {
// 				return check.Failed(err)
// 			}
//
// 			messages := rApi.GetMessagesFromPods(podsList.Items...)
// 			if len(messages) > 0 {
// 				b, err := json.Marshal(messages)
// 				if err != nil {
// 					return check.Failed(err).Err(nil)
// 				}
// 				return check.Failed(fmt.Errorf("%s", b)).Err(nil)
// 			}
//
// 			return check.StillRunning(fmt.Errorf("waiting for statefulset pods to start"))
// 		}
// 	}
//
// 	return check.Completed()
// }

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
