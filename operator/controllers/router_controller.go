package controllers

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	networkingv1 "k8s.io/api/networking/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	crdsv1 "operators.kloudlite.io/api/v1"
	"operators.kloudlite.io/lib/errors"
	"operators.kloudlite.io/lib/finalizers"
	reconcileResult "operators.kloudlite.io/lib/reconcile-result"
)

// RouterReconciler reconciles a Router object
type RouterReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	logger   *zap.SugaredLogger
	resource *crdsv1.Router
}

//+kubebuilder:rbac:groups=crds.kloudlite.io,resources=routers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=crds.kloudlite.io,resources=routers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=crds.kloudlite.io,resources=routers/finalizers,verbs=update

func (r *RouterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.logger = GetLogger(req.NamespacedName)
	logger := r.logger.With("RECONCILE", "true")

	logger.Infof("Request received for: %v\n", req.String())

	router := &crdsv1.Router{}
	if err := r.Get(ctx, req.NamespacedName, router); err != nil {
		if apiErrors.IsNotFound(err) {
			return reconcileResult.OK()
		}
		return reconcileResult.FailedE(err)
	}
	r.resource = router

	if router.GetDeletionTimestamp() != nil {
		r.finalizeRouter(ctx, router)
	}

	var ingRules []networkingv1.IngressRule
	for _, domain := range router.Spec.Domains {
		rule := networkingv1.IngressRule{
			Host: domain,
			IngressRuleValue: networkingv1.IngressRuleValue{
				HTTP: &networkingv1.HTTPIngressRuleValue{},
			},
		}

		rule.IngressRuleValue.HTTP.Paths = []networkingv1.HTTPIngressPath{}
		pathType := networkingv1.PathTypePrefix
		for _, route := range router.Spec.Routes {
			p := networkingv1.HTTPIngressPath{
				Path:     route.Path,
				PathType: &pathType,
				Backend: networkingv1.IngressBackend{
					Service: &networkingv1.IngressServiceBackend{
						Name: route.App,
						Port: networkingv1.ServiceBackendPort{
							Number: int32(route.Port),
						},
					},
				},
			}
			rule.HTTP.Paths = append(rule.HTTP.Paths, p)
		}
		ingRules = append(ingRules, rule)
	}

	ing := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: router.Namespace,
			Name:      router.Name,
		},
	}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, ing, func() error {
		ing.ObjectMeta.Annotations = IngressAnnotations
		ing.Spec.TLS = []networkingv1.IngressTLS{
			{
				Hosts:      router.Spec.Domains,
				SecretName: fmt.Sprintf("%s-router-tls", router.Name),
			},
		}
		ing.Spec.Rules = ingRules
		// ingress := networkingv1.Ingress{
		// 	TypeMeta: TypeIngress,
		// 	ObjectMeta: metav1.ObjectMeta{
		// 		Namespace:   router.Namespace,
		// 		Name:        router.Name,
		// 		Annotations: IngressAnnotations,
		// 	},
		// 	Spec: networkingv1.IngressSpec{
		// 		TLS: []networkingv1.IngressTLS{
		// 			{
		// 				Hosts:      router.Spec.Domains,
		// 				SecretName: fmt.Sprintf("%s-router-tls", router.Name),
		// 			},
		// 		},
		// 		Rules: ingRules,
		// 	},
		// 	Status: networkingv1.IngressStatus{},
		// }
		if err := controllerutil.SetControllerReference(router, ing, r.Scheme); err != nil {
			logger.Info("could not set controller references")
			return err
		}
		return nil
	}); err != nil {
		e := errors.NewEf(err, "could not apply ingress")
		return reconcileResult.FailedE(e)
	}

	router.Status.IPs = []string{}
	if len(ing.Status.LoadBalancer.Ingress) > 0 {
		for _, item := range ing.Status.LoadBalancer.Ingress {
			router.Status.IPs = append(router.Status.IPs, item.IP)
		}
	}

	if err := r.Status().Update(ctx, router); err != nil {
		return reconcileResult.FailedE(errors.StatusUpdate(err))
	}
	// ingress := networkingv1.Ingress{}

	// logger.Info("Ingress: %+\n")
	// if err := r.apply(ctx, &ingress, func() error {
	// 	logger.Infof("INgress: %v", ingress.Spec.Rules)
	// 	return nil
	// }); err != nil {
	// 	e := errors.NewEf(err, "could not apply ingress resource")
	// 	logger.Info(e.Error())
	// 	return reconcileResult.FailedE(e)
	// }

	logger.Info("Reconcile Completed ...")

	return ctrl.Result{}, nil
}

func (r *RouterReconciler) apply(ctx context.Context, obj client.Object, fn ...controllerutil.MutateFn) error {
	x, err := controllerutil.CreateOrUpdate(ctx, r.Client, obj, func() error {
		if err := ctrl.SetControllerReference(r.resource, obj, r.Scheme); err != nil {
			r.logger.Infof("could not update controller reference")
			return err
		}

		if len(fn) > 0 {
			return fn[0]()
		}
		return nil
	})
	r.logger.Info(x, err)
	return err
}

func (r *RouterReconciler) finalizeRouter(ctx context.Context, router *crdsv1.Router) (ctrl.Result, error) {
	logger := r.logger.With("FINALIZER", "true")
	logger.Debug("finalizing ...")

	if controllerutil.ContainsFinalizer(router, finalizers.Router.String()) {
		controllerutil.RemoveFinalizer(router, finalizers.Router.String())
		if err := r.Update(ctx, router); err != nil {
			return reconcileResult.FailedE(err)
		}
		return reconcileResult.Retry()
	}

	if controllerutil.ContainsFinalizer(router, finalizers.Foreground.String()) {
		var tr networkingv1.Ingress
		if err := r.Get(ctx, types.NamespacedName{Name: router.Name, Namespace: router.Namespace}, &tr); err != nil {
			if apiErrors.IsNotFound(err) {
				controllerutil.RemoveFinalizer(router, finalizers.Foreground.String())
				if err := r.Update(ctx, router); err != nil {
					return reconcileResult.FailedE(err)
				}
				return reconcileResult.Retry()
			}
		}
	}

	return reconcileResult.OK()
}

// SetupWithManager sets up the controller with the Manager.
func (r *RouterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crdsv1.Router{}).
		Owns(&networkingv1.Ingress{}).
		Complete(r)
}
