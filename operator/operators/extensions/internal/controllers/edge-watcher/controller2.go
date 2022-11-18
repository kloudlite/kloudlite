package edgeWatcher

// import (
// 	"context"
// 	"encoding/json"
//
// 	apiErrors "k8s.io/apimachinery/pkg/api/errors"
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// 	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
// 	"k8s.io/apimachinery/pkg/runtime"
// 	crdsv1 "operators.kloudlite.io/apis/crds/v1"
// 	csiv1 "operators.kloudlite.io/apis/csi/v1"
// 	extensionsv1 "operators.kloudlite.io/apis/extensions/v1"
// 	"operators.kloudlite.io/lib/constants"
// 	fn "operators.kloudlite.io/lib/functions"
// 	"operators.kloudlite.io/lib/logging"
// 	stepResult "operators.kloudlite.io/lib/operator/step-result"
// 	"operators.kloudlite.io/operators/extensions/internal/env"
// 	ctrl "sigs.k8s.io/controller-runtime"
// 	"sigs.k8s.io/controller-runtime/pkg/client"
// 	"sigs.k8s.io/controller-runtime/pkg/handler"
// 	"sigs.k8s.io/controller-runtime/pkg/reconcile"
// 	"sigs.k8s.io/controller-runtime/pkg/source"
// )
//
// type Reconciler struct {
// 	client.Client
// 	Scheme *runtime.Scheme
// 	logger logging.Logger
// 	Name   string
// 	Env    *env.Env
// }
//
// func (r *Reconciler) GetName() string {
// 	return r.Name
// }
//
// const SSLSecretName = "kl-cert-issuer-tls"
// const SSLSecretNamespace = "kl-init-cert-manager"
//
// type Edge struct {
// 	metav1.TypeMeta   `json:",inline"`
// 	metav1.ObjectMeta `json:"metadata,omitempty"`
// 	Spec              struct {
// 		AccountId      string `json:"accountId"`
// 		CredentialsRef struct {
// 			Namespace  string `json:"namespace"`
// 			SecretName string `json:"secretName"`
// 		} `json:"credentialsRef"`
// 		Provider string `json:"provider"`
// 		Region   string `json:"region"`
// 	} `json:"spec"`
// }
//
// func parseEdge(edge *unstructured.Unstructured) (*Edge, error) {
// 	b, err := json.Marshal(edge.Object)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	var j Edge
//
// 	if err := json.Unmarshal(b, &j); err != nil {
// 		return nil, err
// 	}
//
// 	return &j, nil
// }
//
// // +kubebuilder:rbac:groups=extensions.kloudlite.io,resources=edgeWatchers,verbs=get;list;watch;create;update;patch;delete
// // +kubebuilder:rbac:groups=extensions.kloudlite.io,resources=edgeWatchers/status,verbs=get;update;patch
// // +kubebuilder:rbac:groups=extensions.kloudlite.io,resources=edgeWatchers/finalizers,verbs=update
//
// func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
// 	// check if request is for an Infra Edge or ignore otherwise
// 	edgeRes := fn.NewUnstructured(constants.EdgeInfraType)
// 	if err := r.Get(ctx, request.NamespacedName, edgeRes); err != nil {
// 		return ctrl.Result{}, client.IgnoreNotFound(err)
// 	}
//
// 	edge, err := parseEdge(edgeRes)
// 	if err != nil {
// 		return ctrl.Result{}, err
// 	}
//
// 	logger := r.logger.WithName(request.NamespacedName.String())
// 	logger.Infof("NEW RECONCILATION")
// 	defer func() {
// 		logger.Infof("RECONCILATION COMPLETE")
// 	}()
//
// 	// check if csi driver is present
// 	if step := r.ensureCSIDriver(ctx, edge, logger); !step.ShouldProceed() {
// 		return step.ReconcilerResponse()
// 	}
//
// 	// check if ingress is installed on that edge
// 	if step := r.ensureEdgeRouters(ctx, edge, logger); !step.ShouldProceed() {
// 		return step.ReconcilerResponse()
// 	}
//
// 	logger.Infof("RECONCILATION COMPLETE")
// 	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, nil
// }
//
// func (r *Reconciler) ensureEdgeNamespace(ctx context.Context, edge *Edge, logger logging.Logger) stepResult.Result {
// 	return nil
// }
//
// func (r *Reconciler) ensureCSIDriver(ctx context.Context, edge *Edge, logger logging.Logger) stepResult.Result {
// 	if err := r.Get(ctx, fn.NN("", edge.Spec.CredentialsRef.SecretName), &csiv1.Driver{}); err != nil {
// 		if apiErrors.IsNotFound(err) {
// 			logger.Infof("creating CSI Driver for (edge=%s, provider=%s), as it does not exist", edge.Name, edge.Spec.Provider)
//
// 			if err := r.Create(
// 				ctx, &csiv1.Driver{
// 					ObjectMeta: metav1.ObjectMeta{
// 						Name:            edge.Spec.CredentialsRef.SecretName,
// 						OwnerReferences: edge.GetOwnerReferences(),
// 						Labels: map[string]string{
// 							"kloudlite.io/created-by-edge-watcher": edge.Name,
// 						},
// 					},
// 					Spec: csiv1.DriverSpec{
// 						Provider:  edge.Spec.Provider,
// 						SecretRef: edge.Spec.CredentialsRef.SecretName,
// 					},
// 				},
// 			); err != nil {
// 				return stepResult.New().Err(err)
// 			}
// 		}
// 		return stepResult.New().Err(err)
// 	}
//
// 	return stepResult.New().Continue(true)
// }
//
// func (r *Reconciler) ensureEdgeRouters(ctx context.Context, edge *Edge, logger logging.Logger) stepResult.Result {
// 	var edgeRouter crdsv1.EdgeRouter
// 	if err := r.Get(ctx, fn.NN("", edge.Name), &edgeRouter); err != nil {
// 		if apiErrors.IsNotFound(err) {
// 			logger.Infof("creating EdgeRouter for (edge=%s, provider=%s), as it does not exist", edge.Name, edge.Spec.Provider)
// 			if err := r.Create(
// 				ctx, &crdsv1.EdgeRouter{
// 					ObjectMeta: metav1.ObjectMeta{
// 						Name:            edge.Name,
// 						OwnerReferences: edge.GetOwnerReferences(),
// 						Labels: map[string]string{
// 							"kloudlite.io/created-by-edge-watcher": edge.Name,
// 						},
// 					},
// 					Spec: crdsv1.EdgeRouterSpec{
// 						Region:     edge.Spec.Region,
// 						AccountRef: edge.Spec.AccountId,
// 						DefaultSSLCert: crdsv1.SSLCertRef{
// 							SecretName: SSLSecretName,
// 							Namespace:  SSLSecretNamespace,
// 						},
// 					},
// 				},
// 			); err != nil {
// 				return stepResult.New().Err(err)
// 			}
// 		}
// 		return stepResult.New().Err(err)
// 	}
// 	return stepResult.New().Continue(true)
// }
//
// func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
// 	r.Client = mgr.GetClient()
// 	r.Scheme = mgr.GetScheme()
// 	r.logger = logger.WithName(r.Name)
//
// 	builder := ctrl.NewControllerManagedBy(mgr).For(&extensionsv1.Cluster{})
// 	for _, k := range []client.Object{&csiv1.Driver{}, &crdsv1.EdgeRouter{}} {
// 		builder.Watches(
// 			&source.Kind{Type: k}, handler.EnqueueRequestsFromMapFunc(
// 				func(obj client.Object) []reconcile.Request {
// 					if v, ok := obj.GetLabels()["kloudlite.io/created-by-edge-watcher"]; ok {
// 						return []reconcile.Request{{NamespacedName: fn.NN("", v)}}
// 					}
// 					return nil
// 				},
// 			),
// 		)
// 	}
//
// 	builder.Watches(
// 		&source.Kind{Type: fn.NewUnstructured(constants.EdgeInfraType)}, handler.EnqueueRequestsFromMapFunc(
// 			func(obj client.Object) []reconcile.Request {
// 				return []reconcile.Request{{NamespacedName: fn.NN(obj.GetNamespace(), obj.GetName())}}
// 			},
// 		),
// 	)
// 	return builder.Complete(r)
// }
