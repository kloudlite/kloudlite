package controller

import (
	"context"
	"encoding/json"

	"github.com/kloudlite/operator/operators/byoc-client-operator/internal/env"
	"github.com/kloudlite/operator/operators/byoc-client-operator/types"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	"github.com/kloudlite/operator/pkg/redpanda"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	clusterv1 "github.com/kloudlite/operator/apis/clusters/v1"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	storagev1 "k8s.io/api/storage/v1"
	apiLabels "k8s.io/apimachinery/pkg/labels"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

type Reconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Logger   logging.Logger
	Name     string
	Env      *env.Env
	Producer redpanda.Producer
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	DefaultsPatched     string = "defaults-patched"
	KafkaTopicExists    string = "kafka-topic-exists"
	HarborProjectExists string = "harbor-project-exists"
)

const byocClientFinalizer = "kloudlite.io/byoc-client-finalizer"

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	var obj clusterv1.BYOC
	if err := r.Get(ctx, request.NamespacedName, &obj); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	r.processStorageClasses(ctx, request, &obj)
	r.processIngressClasses(ctx, request, &obj)
	r.processNodes(ctx, request, &obj)
	r.processHelmDeployments(ctx, request, &obj)

	if err := r.Status().Update(ctx, &obj); err != nil {
		return ctrl.Result{}, nil
	}

	if err := r.Update(ctx, &obj); err != nil {
		return ctrl.Result{}, nil
	}

	r.sendStatusUpdate(ctx, &obj)

	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, nil
}

func (r *Reconciler) processStorageClasses(ctx context.Context, req ctrl.Request, obj *clusterv1.BYOC) error {
	var scl storagev1.StorageClassList
	if err := r.List(ctx, &scl); err != nil {
		return client.IgnoreNotFound(err)
	}
	results := make([]string, len(scl.Items))
	for i := range scl.Items {
		results[i] = scl.Items[i].Name
		if scl.Items[i].Annotations["storageclass.kubernetes.io/is-default-class"] == "true" {
			results[0], results[i] = results[i], results[0]
		}
	}

	obj.Spec.StorageClasses = results
	return nil
}

func (r *Reconciler) processIngressClasses(ctx context.Context, req ctrl.Request, obj *clusterv1.BYOC) error {
	var icl networkingv1.IngressClassList
	if err := r.List(ctx, &icl); err != nil {
		return client.IgnoreNotFound(err)
	}
	results := make([]string, len(icl.Items))
	for i := range icl.Items {
		results[i] = icl.Items[i].Name
	}

	obj.Spec.IngressClasses = results
	return nil
}

func (r *Reconciler) processNodes(ctx context.Context, req ctrl.Request, obj *clusterv1.BYOC) error {
	var nl corev1.NodeList
	if err := r.List(ctx, &nl); err != nil {
		return client.IgnoreNotFound(err)
	}
	results := make([]string, len(nl.Items))
	for i := range nl.Items {
		for _, address := range nl.Items[i].Status.Addresses {
			if address.Type == "ExternalIP" {
				results[i] = address.Address
			}
		}
	}

	obj.Spec.PublicIPs = results
	return nil
}

func (r *Reconciler) processHelmDeployments(ctx context.Context, req ctrl.Request, obj *clusterv1.BYOC) error {
	var dl appsv1.DeploymentList
	if err := r.List(ctx, &dl, &client.ListOptions{
		Namespace: r.Env.HelmReleaseNamespace,
	}); err != nil {
		return client.IgnoreNotFound(err)
	}

	if obj.Status.Checks == nil {
		obj.Status.Checks = make(map[string]rApi.Check, len(dl.Items))
	}

	for i := range dl.Items {
		check := rApi.Check{
			Status: dl.Items[i].Status.Replicas == dl.Items[i].Status.ReadyReplicas,
		}

		if !check.Status {
			var podList corev1.PodList
			if err := r.List(
				ctx, &podList, &client.ListOptions{
					LabelSelector: apiLabels.SelectorFromValidatedSet(map[string]string{"app": dl.Items[i].Name}),
					Namespace:     obj.Namespace,
				},
			); err != nil {
				check.Message = err.Error()
			}

			pMessages := rApi.GetMessagesFromPods(podList.Items...)
			bMsg, err := json.Marshal(pMessages)
			if err != nil {
				check.Message = err.Error()
			}
			if bMsg != nil {
				check.Message = string(bMsg)
			}
		}

		obj.Status.Checks[dl.Items[i].Name] = check
	}

	return nil
}

func (r *Reconciler) sendStatusUpdate(ctx context.Context, obj *clusterv1.BYOC) (ctrl.Result, error) {
	obj.SetManagedFields(nil)
	b, err := json.Marshal(obj)
	if err != nil {
		return ctrl.Result{}, err
	}

	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return ctrl.Result{}, err
	}

	b, err = json.Marshal(types.StatusUpdate{
		ClusterName: obj.Name,
		AccountName: obj.Spec.AccountName,
		Object:      m,
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	pm, err := r.Producer.Produce(ctx, r.Env.KafkaTopicBYOCClientUpdates, obj.Name, b)
	if err != nil {
		return ctrl.Result{}, err
	}

	r.Logger.Infof("dispatched update to %s @ %s", pm.Topic, pm.Timestamp.String())
	if obj.GetDeletionTimestamp() != nil {
		if controllerutil.ContainsFinalizer(obj, byocClientFinalizer) {
			controllerutil.RemoveFinalizer(obj, byocClientFinalizer)
			return ctrl.Result{}, r.Update(ctx, obj)
		}
		return ctrl.Result{}, nil
	}

	if !controllerutil.ContainsFinalizer(obj, byocClientFinalizer) {
		controllerutil.AddFinalizer(obj, byocClientFinalizer)
		return ctrl.Result{}, r.Update(ctx, obj)
	}
	return ctrl.Result{}, nil
}

func (r *Reconciler) finalize(req *rApi.Request[*clusterv1.BYOC]) stepResult.Result {
	return nil
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.Logger = logger.WithName(r.Name)

	builder := ctrl.NewControllerManagedBy(mgr).For(&clusterv1.BYOC{})
	watchList := []client.Object{
		&storagev1.StorageClass{},
		&networkingv1.IngressClass{},
		&corev1.Node{},
		&appsv1.Deployment{},
	}

	for i := range watchList {
		builder.Watches(&source.Kind{Type: watchList[i]}, handler.EnqueueRequestsFromMapFunc(func(client.Object) []reconcile.Request {
			var byocList clusterv1.BYOCList
			if err := r.List(context.TODO(), &byocList); err != nil {
				return nil
			}

			reqs := make([]reconcile.Request, len(byocList.Items))
			for i := range byocList.Items {
				reqs[i] = reconcile.Request{NamespacedName: client.ObjectKeyFromObject(&byocList.Items[i])}
			}
			return reqs
		}))
	}

	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
