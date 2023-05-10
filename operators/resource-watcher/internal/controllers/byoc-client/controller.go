package byoc_client

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	storagev1 "k8s.io/api/storage/v1"
	apiLabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	clusterv1 "github.com/kloudlite/operator/apis/clusters/v1"
	"github.com/kloudlite/operator/operators/resource-watcher/internal/env"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
)

type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Logger logging.Logger
	Name   string
	Env    *env.Env
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	StorageClassProcessed    string = "storage-class-processed"
	IngressClassProcessed    string = "ingress-class-processed"
	NodesProcessed           string = "nodes-processed"
	HelmDeploymentsProcessed string = "helm-deployments-processed"
)

const (
	DefaultStorageClassAnnotation = "storageclass.kubernetes.io/is-default-class"
)

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=apps/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.Logger), r.Client, request.NamespacedName, &clusterv1.BYOC{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	obj := req.Object

	req.LogPreReconcile()
	defer req.LogPostReconcile()

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	r.processStorageClasses(req)
	r.processIngressClasses(req)
	r.processNodes(req)
	r.processHelmDeployments(req)

	obj.Status.IsReady = true
	if step := req.UpdateStatus(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, nil
}

func (r *Reconciler) processStorageClasses(req *rApi.Request[*clusterv1.BYOC]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(StorageClassProcessed)
	defer req.LogPostCheck(StorageClassProcessed)

	var scl storagev1.StorageClassList
	if err := r.List(ctx, &scl); err != nil {
		return req.Done().Err(client.IgnoreNotFound(err))
	}

	results := make([]string, len(scl.Items))
	for i := range scl.Items {
		results[i] = scl.Items[i].Name
		if scl.Items[i].Annotations[DefaultStorageClassAnnotation] == "true" {
			results[0], results[i] = results[i], results[0]
		}
	}

	if !reflect.DeepEqual(obj.Spec.StorageClasses, results) {
		obj.Spec.StorageClasses = results
		if err := r.Update(ctx, obj); err != nil {
			return req.CheckFailed(StorageClassProcessed, check, err.Error())
		}
		return req.Done().RequeueAfter(100 * time.Millisecond)
	}

	check.Status = true
	if check != obj.Status.Checks[StorageClassProcessed] {
		obj.Status.Checks[StorageClassProcessed] = check
		req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) processIngressClasses(req *rApi.Request[*clusterv1.BYOC]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(IngressClassProcessed)
	defer req.LogPostCheck(IngressClassProcessed)

	var icl networkingv1.IngressClassList
	if err := r.List(ctx, &icl); err != nil {
		return req.Done().Err(client.IgnoreNotFound(err))
	}

	results := make([]string, len(icl.Items))
	for i := range icl.Items {
		results[i] = icl.Items[i].Name
	}

	if !reflect.DeepEqual(obj.Spec.IngressClasses, results) {
		obj.Spec.IngressClasses = results
		if err := r.Update(ctx, obj); err != nil {
			return req.CheckFailed(IngressClassProcessed, check, err.Error())
		}
		return req.Done().RequeueAfter(100 * time.Millisecond)
	}

	check.Status = true
	if check != obj.Status.Checks[IngressClassProcessed] {
		obj.Status.Checks[IngressClassProcessed] = check
		req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) processNodes(req *rApi.Request[*clusterv1.BYOC]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(NodesProcessed)
	defer req.LogPostCheck(NodesProcessed)

	var nl corev1.NodeList
	if err := r.List(ctx, &nl); err != nil {
		return req.Done().Err(client.IgnoreNotFound(err))
	}

	results := make([]string, len(nl.Items))
	for i := range nl.Items {
		for _, address := range nl.Items[i].Status.Addresses {
			if address.Type == "ExternalIP" {
				results[i] = address.Address
			}
		}
	}

	if !reflect.DeepEqual(obj.Spec.PublicIPs, results) {
		obj.Spec.PublicIPs = results
		if err := r.Update(ctx, obj); err != nil {
			return req.CheckFailed(NodesProcessed, check, err.Error())
		}
		return req.Done().RequeueAfter(100 * time.Millisecond)
	}

	check.Status = true
	if check != obj.Status.Checks[NodesProcessed] {
		obj.Status.Checks[NodesProcessed] = check
		req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) processHelmDeployments(req *rApi.Request[*clusterv1.BYOC]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(HelmDeploymentsProcessed)
	defer req.LogPostCheck(HelmDeploymentsProcessed)

	var dl appsv1.DeploymentList
	if err := r.List(ctx, &dl, &client.ListOptions{
		Namespace: r.Env.OperatorsNamespace,
	}); err != nil {
		return req.Done().Err(client.IgnoreNotFound(err))
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

	check.Status = true
	if check != obj.Status.Checks[HelmDeploymentsProcessed] {
		obj.Status.Checks[HelmDeploymentsProcessed] = check
		req.UpdateStatus()
	}

	return req.Next()
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
		builder.Watches(
			&source.Kind{Type: watchList[i]},
			handler.EnqueueRequestsFromMapFunc(func(client.Object) []reconcile.Request {
				var byocList clusterv1.BYOCList
				if err := r.List(context.TODO(), &byocList); err != nil {
					return nil
				}

				reqs := make([]reconcile.Request, len(byocList.Items))
				for i := range byocList.Items {
					reqs[i] = reconcile.Request{
						NamespacedName: client.ObjectKeyFromObject(&byocList.Items[i]),
					}
				}
				return reqs
			}),
		)
	}

	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
