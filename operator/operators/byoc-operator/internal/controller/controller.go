package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/kloudlite/operator/operators/byoc-operator/internal/env"
	"github.com/kloudlite/operator/pkg/constants"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	clusterv1 "github.com/kloudlite/operator/apis/clusters/v1"
	fn "github.com/kloudlite/operator/pkg/functions"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	redpandaMsvcv1 "github.com/kloudlite/operator/apis/redpanda.msvc/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Logger     logging.Logger
	Name       string
	Env        *env.Env
	YamlClient *kubectl.YAMLClient
	recorder   record.EventRecorder
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	DefaultsPatched     string = "defaults-patched"
	KafkaTopicExists    string = "kafka-topic-exists"
	HarborProjectExists string = "harbor-project-exists"
)

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.Logger), r.Client, request.NamespacedName, &clusterv1.BYOC{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	req.LogPreReconcile()
	defer req.LogPostReconcile()

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

	if step := r.ensureKafkaTopic(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	req.Object.Status.LastReconcileTime = &metav1.Time{Time: time.Now()}

	req.Object.Status.Resources = req.GetOwnedResources()
	if err := r.Status().Update(ctx, req.Object); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, nil
}

func (r *Reconciler) finalize(req *rApi.Request[*clusterv1.BYOC]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	finalizing := "finalizing"

	req.LogPreCheck(finalizing)
	defer req.LogPostCheck(finalizing)

	rr := req.GetOwnedResources()
	for i := range rr {
		res := &unstructured.Unstructured{Object: map[string]any{"apiVersion": rr[i].APIVersion, "kind": rr[i].Kind}}

		if err := r.Get(ctx, fn.NN(rr[i].Namespace, rr[i].Name), res); err != nil {
			if !apiErrors.IsNotFound(err) {
				if res.GetDeletionTimestamp() != nil {
					if err := r.Delete(ctx, res); err != nil {
						return req.CheckFailed(finalizing, check, err.Error())
					}
				}
				req.Logger.Infof("waiting for child resource '[%s, %s] %s/%s' to be deleted", rr[i].APIVersion, rr[i].Kind, rr[i].Namespace, rr[i].Name)
				return req.CheckFailed(finalizing, check, err.Error())
			}
		}
	}

	//  --->
	controllerutil.RemoveFinalizer(obj, constants.CommonFinalizer)
	if err := r.Update(ctx, obj); err != nil {
		return req.CheckFailed(finalizing, check, err.Error())
	}

	return req.Next()
}

func (r *Reconciler) ensureKafkaTopic(req *rApi.Request[*clusterv1.BYOC]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(KafkaTopicExists)
	defer req.LogPostCheck(KafkaTopicExists)

	kt := &redpandaMsvcv1.Topic{ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("kl-%s-%s-incoming", obj.Spec.AccountName, obj.Name), Namespace: r.Env.KafkaTopicNamespace}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, kt, func() error {
		if !fn.IsOwner(kt, fn.AsOwner(obj)) {
			kt.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
			if kt.Labels == nil {
				kt.Labels = make(map[string]string, 1)
			}
			kt.Labels["kloudlite.io/byoc.name"] = obj.Name
		}
		return nil
	}); err != nil {
		return req.CheckFailed(KafkaTopicExists, check, err.Error()).Err(nil)
	}

	req.AddToOwnedResources(rApi.ResourceRef{TypeMeta: kt.TypeMeta, Name: kt.Name, Namespace: kt.Namespace})
	req.UpdateStatus()

	var kt2 redpandaMsvcv1.Topic
	if err := r.Get(ctx, client.ObjectKeyFromObject(kt), &kt2); err != nil {
		return req.CheckFailed(KafkaTopicExists, check, err.Error()).Err(nil)
	}

	if !kt2.Status.IsReady {
		return req.CheckFailed(KafkaTopicExists, check, fmt.Sprintf("waiting for kafka topic resource %q to get ready", kt.Name))
	}

	check.Status = true
	if check != obj.Status.Checks[KafkaTopicExists] {
		obj.Status.Checks[KafkaTopicExists] = check
		req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.Logger = logger.WithName(r.Name)
	// r.YamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())
	// r.recorder = mgr.GetEventRecorderFor(r.GetName())

	builder := ctrl.NewControllerManagedBy(mgr).For(&clusterv1.BYOC{})
	builder.Owns(&redpandaMsvcv1.Topic{})

	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
