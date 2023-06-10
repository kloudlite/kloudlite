package edgeWatcher

import (
	"context"
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	csiv1 "github.com/kloudlite/operator/apis/csi/v1"
	extensionsv1 "github.com/kloudlite/operator/apis/extensions/v1"
	"github.com/kloudlite/operator/operators/cluster-setup/internal/env"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
)

type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
	logger logging.Logger
	Name   string
	Env    *env.Env
}

func (r *Reconciler) GetName() string {
	return r.Name
}

type Edge struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              struct {
		AccountId      string `json:"accountId"`
		CredentialsRef struct {
			Namespace  string `json:"namespace"`
			SecretName string `json:"secretName"`
		} `json:"credentialsRef"`
		Provider string `json:"provider"`
		Region   string `json:"region"`
	} `json:"spec"`
}

func parseEdge(edge *unstructured.Unstructured) (*Edge, error) {
	b, err := json.Marshal(edge.Object)
	if err != nil {
		return nil, err
	}

	var j Edge

	if err := json.Unmarshal(b, &j); err != nil {
		return nil, err
	}

	return &j, nil
}

func (r *Reconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	edgeRes := fn.NewUnstructured(constants.EdgeInfraType)
	if err := r.Get(ctx, request.NamespacedName, edgeRes); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	edge, err := parseEdge(edgeRes)
	if err != nil {
		return ctrl.Result{}, err
	}

	logger := r.logger.WithName(request.NamespacedName.String())
	logger.Infof("NEW RECONCILATION")
	defer func() {
		logger.Infof("RECONCILATION COMPLETE")
	}()

	edgeWorker := &extensionsv1.EdgeWorker{ObjectMeta: metav1.ObjectMeta{Name: edge.Name}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, edgeWorker, func() error {
		if edgeWorker.Labels == nil {
			edgeWorker.Labels = make(map[string]string, 1)
		}
		edgeWorker.Labels["kloudlite.io/created-by-edge-watcher"] = "true"
		edgeWorker.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(edgeRes, true)})

		edgeWorker.Spec = extensionsv1.EdgeWorkerSpec{
			AccountName: edge.Spec.AccountId,
			Creds:       edge.Spec.CredentialsRef,
			Provider:    edge.Spec.Provider,
			Region:      edge.Spec.Region,
		}
		return nil
	}); err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)

	builder := ctrl.NewControllerManagedBy(mgr).For(&extensionsv1.Cluster{})
	for _, k := range []client.Object{&csiv1.Driver{}, &crdsv1.EdgeRouter{}} {
		builder.Watches(
			&source.Kind{Type: k}, handler.EnqueueRequestsFromMapFunc(
				func(obj client.Object) []reconcile.Request {
					if v, ok := obj.GetLabels()["kloudlite.io/created-by-edge-watcher"]; ok {
						return []reconcile.Request{{NamespacedName: fn.NN("", v)}}
					}
					return nil
				},
			),
		)
	}

	builder.Watches(
		&source.Kind{Type: fn.NewUnstructured(constants.EdgeInfraType)}, handler.EnqueueRequestsFromMapFunc(
			func(obj client.Object) []reconcile.Request {
				return []reconcile.Request{{NamespacedName: fn.NN(obj.GetNamespace(), obj.GetName())}}
			},
		),
	)
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
