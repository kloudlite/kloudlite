package clusterService

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kloudlite/operator/operators/msvc-mongo/internal/env"
	"github.com/kloudlite/operator/operators/msvc-mongo/internal/templates"
	"github.com/kloudlite/operator/operators/msvc-mongo/internal/types"

	mongodbMsvcv1 "github.com/kloudlite/operator/apis/mongodb.msvc/v1"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/harbor"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type Reconciler struct {
	client.Client
	Scheme                     *runtime.Scheme
	Env                        *env.Env
	harborCli                  *harbor.Client
	logger                     logging.Logger
	Name                       string
	yamlClient                 kubectl.YAMLClient
	templateHelmMongoDBCluster []byte
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	HelmReady            string = "helm-ready"
	StsReady             string = "sts-ready"
	AccessCredsReady     string = "access-creds-ready"
	ReconcileCredentials string = "reconcile-credentials"
	CheckPatchDefaults   string = "patch-defaults"
)

const (
	KeyRootPassword string = "root-password"
)

// +kubebuilder:rbac:groups=mongodb.msvc.kloudlite.io,resources=clusterServices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mongodb.msvc.kloudlite.io,resources=clusterServices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mongodb.msvc.kloudlite.io,resources=clusterServices/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &mongodbMsvcv1.ClusterService{})
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

	if step := req.EnsureChecks(HelmReady, StsReady, AccessCredsReady); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.patchDefaults(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconCredentials(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconHelm(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconSts(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *Reconciler) finalize(req *rApi.Request[*mongodbMsvcv1.ClusterService]) stepResult.Result {
	checkName := "finalizing"
	req.LogPreCheck(checkName)
	defer req.LogPostCheck(checkName)

	if step := req.CleanupOwnedResources(); !step.ShouldProceed() {
		return step
	}

	return req.Finalize()
}

func (r *Reconciler) patchDefaults(req *rApi.Request[*mongodbMsvcv1.ClusterService]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(CheckPatchDefaults)
	defer req.LogPostCheck(CheckPatchDefaults)

	hasPatched := false

	if obj.Spec.Output.Credentials.Name == "" {
		hasPatched = true
		obj.Spec.Output.Credentials.Name = fmt.Sprintf("msvc-%s-creds", obj.Name)
	}

	if obj.Spec.Output.Credentials.Namespace == "" {
		hasPatched = true
		obj.Spec.Output.Credentials.Namespace = obj.Namespace
	}

	if obj.Spec.Output.HelmSecret.Name == "" {
		hasPatched = true
		obj.Spec.Output.HelmSecret.Name = fmt.Sprintf("helm-%s-creds", obj.Name)
	}

	if obj.Spec.Output.HelmSecret.Namespace == "" {
		hasPatched = true
		obj.Spec.Output.HelmSecret.Namespace = obj.Namespace
	}

	if hasPatched {
		if err := r.Update(ctx, obj); err != nil {
			return req.CheckFailed(CheckPatchDefaults, check, err.Error())
		}
	}

	check.Status = true
	if check != obj.Status.Checks[CheckPatchDefaults] {
		obj.Status.Checks[CheckPatchDefaults] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

func (r *Reconciler) reconCredentials(req *rApi.Request[*mongodbMsvcv1.ClusterService]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(ReconcileCredentials)
	defer req.LogPostCheck(ReconcileCredentials)

	rootPassword := fn.CleanerNanoid(40)
	replicasetKey := fn.CleanerNanoid(10) // should not be more than 10, as it crashes our mongodb process

	msvcOutput := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.Output.Credentials.Name, Namespace: obj.Namespace}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, msvcOutput, func() error {
		msvcOutput.SetLabels(obj.GetLabels())
		msvcOutput.SetFinalizers([]string{constants.GenericFinalizer})

		msvcOutput.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})

		if msvcOutput.Data == nil {
			hosts := make([]string, obj.Spec.Replicas)
			for i := 0; i < obj.Spec.Replicas; i++ {
				hosts[i] = fmt.Sprintf("%s-%d.%s-headless.%s.svc.%s:27017", obj.Name, i, obj.Name, obj.Namespace, r.Env.ClusterInternalDNS)
			}

			rootUsername := "root"
			replicaSetName := "rs"

			authSource := "admin"

			output := types.ClusterSvcOutput{
				RootUsername:    rootUsername,
				RootPassword:    rootPassword,
				Hosts:           strings.Join(hosts, ","),
				URI:             fmt.Sprintf("mongodb://%s:%s@%s/%s?authSource=%s&replicaSet=%s", rootUsername, rootPassword, strings.Join(hosts, ","), "admin", authSource, replicaSetName),
				AuthSource:      authSource,
				ReplicasSetName: replicaSetName,
				ReplicaSetKey:   replicasetKey,
			}

			var err error
			msvcOutput.StringData, err = output.ToMap()
			return err
		}
		return nil
	}); err != nil {
		return req.CheckFailed(ReconcileCredentials, check, err.Error())
	}

	req.AddToOwnedResources(rApi.ParseResourceRef(msvcOutput))

	helmSecret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.Output.HelmSecret.Name, Namespace: obj.Namespace}}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, helmSecret, func() error {
		helmSecret.SetLabels(obj.GetLabels())
		helmSecret.SetFinalizers([]string{constants.GenericFinalizer})

		helmSecret.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})

		if helmSecret.Data == nil {
			helmSecret.StringData = map[string]string{
				"mongodb-root-password":   rootPassword,
				"mongodb-replica-set-key": replicasetKey,
			}
		}

		return nil
	}); err != nil {
		return req.CheckFailed(ReconcileCredentials, check, err.Error())
	}

	req.AddToOwnedResources(rApi.ParseResourceRef(helmSecret))
	rApi.SetLocal(req, "creds", msvcOutput.Data)

	check.Status = true
	if check != obj.Status.Checks[ReconcileCredentials] {
		obj.Status.Checks[ReconcileCredentials] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

func (r *Reconciler) reconHelm(req *rApi.Request[*mongodbMsvcv1.ClusterService]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(HelmReady)
	defer req.LogPostCheck(HelmReady)

	creds, ok := rApi.GetLocal[map[string][]byte](req, "creds")
	if !ok {
		return req.CheckFailed(HelmReady, check, "creds not found")
	}

	b, err := templates.ParseBytes(r.templateHelmMongoDBCluster, map[string]any{
		"name":      obj.Name,
		"namespace": obj.Namespace,
		"labels": map[string]string{
			constants.MsvcNameKey: obj.Name,
		},
		"owner-refs": []metav1.OwnerReference{fn.AsOwner(obj, true)},

		"storage-class": "sc-xfs",
		"storage-size":  obj.Spec.Resources.Storage.Size,

		"replica-count":        obj.Spec.Replicas,
		"root-user":            string(creds["USERNAME"]),
		"auth-existing-secret": obj.Spec.Output.HelmSecret.Name,

		"cpu-min": obj.Spec.Resources.Cpu.Min,
		"cpu-max": obj.Spec.Resources.Cpu.Max,

		"memory-min": obj.Spec.Resources.Memory,
		"memory-max": obj.Spec.Resources.Memory,
	})
	if err != nil {
		return req.CheckFailed(HelmReady, check, err.Error())
	}

	rr, err := r.yamlClient.ApplyYAML(ctx, b)
	if err != nil {
		return req.CheckFailed(HelmReady, check, err.Error())
	}

	req.AddToOwnedResources(rr...)

	check.Status = true
	if check != checks[HelmReady] {
		checks[HelmReady] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

func (r *Reconciler) reconSts(req *rApi.Request[*mongodbMsvcv1.ClusterService]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(StsReady)
	defer req.LogPostCheck(StsReady)

	var stsList appsv1.StatefulSetList

	if err := r.List(
		ctx, &stsList, &client.ListOptions{
			LabelSelector: labels.SelectorFromValidatedSet(obj.GetEnsuredLabels()),
			Namespace:     obj.Namespace,
		},
	); err != nil {
		return req.CheckFailed(StsReady, check, err.Error())
	}

	for i := range stsList.Items {
		item := stsList.Items[i]
		if item.Status.AvailableReplicas != item.Status.Replicas {
			check.Status = false

			var podsList corev1.PodList
			if err := r.List(
				ctx, &podsList, &client.ListOptions{
					LabelSelector: labels.SelectorFromValidatedSet(obj.GetEnsuredLabels()),
				},
			); err != nil {
				return req.CheckFailed(StsReady, check, err.Error())
			}

			messages := rApi.GetMessagesFromPods(podsList.Items...)
			if len(messages) > 0 {
				b, err := json.Marshal(messages)
				if err != nil {
					return req.CheckFailed(StsReady, check, err.Error())
				}
				return req.CheckFailed(StsReady, check, string(b))
			}
		}
	}

	check.Status = true
	if check != checks[StsReady] {
		checks[StsReady] = check
		if sr := req.UpdateStatus(); !sr.ShouldProceed() {
			return sr
		}
	}

	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())

	var err error
	r.templateHelmMongoDBCluster, err = templates.Read(templates.HelmMongoDBCluster)
	if err != nil {
		return err
	}

	builder := ctrl.NewControllerManagedBy(mgr).For(&mongodbMsvcv1.ClusterService{})
	builder.Owns(&corev1.Secret{})
	builder.Watches(
		&source.Kind{Type: &appsv1.StatefulSet{}}, handler.EnqueueRequestsFromMapFunc(
			func(obj client.Object) []reconcile.Request {
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
