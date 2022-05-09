package crds

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/api/meta"
	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	"operators.kloudlite.io/lib"

	"go.uber.org/zap"
	networkingv1 "k8s.io/api/networking/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"operators.kloudlite.io/lib/errors"
	"operators.kloudlite.io/lib/finalizers"
	reconcileResult "operators.kloudlite.io/lib/reconcile-result"
)

// RouterReconciler reconciles a Router object
type RouterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	logger *zap.SugaredLogger
	router *crdsv1.Router
	lib.MessageSender
	lt metav1.Time
}

func (r *RouterReconciler) notifyAndDie(ctx context.Context, err error) (ctrl.Result, error) {
	r.buildConditions("", metav1.Condition{
		Type:    "Ready",
		Status:  "False",
		Reason:  "ErrWhileReconcilation",
		Message: err.Error(),
	})
	return r.notify(ctx)
}

func (r *RouterReconciler) notify(ctx context.Context) (ctrl.Result, error) {
	err := r.SendMessage(r.router.LogRef(), lib.MessageReply{
		Key:        r.router.LogRef(),
		Conditions: r.router.Status.Conditions,
		Status:     meta.IsStatusConditionTrue(r.router.Status.Conditions, "Ready"),
	})
	if err != nil {
		return reconcileResult.FailedE(errors.NewEf(err, "could not send message into kafka"))
	}

	if err := r.Status().Update(ctx, r.router); err != nil {
		return reconcileResult.FailedE(errors.NewEf(err, "could not update status for (%s)", r.router.LogRef()))
	}
	return reconcileResult.OK()
}

func (r *RouterReconciler) buildConditions(source string, conditions ...metav1.Condition) {
	meta.SetStatusCondition(&r.router.Status.Conditions, metav1.Condition{
		Type:               "Ready",
		Status:             "False",
		Reason:             "ChecksNotCompleted",
		LastTransitionTime: r.lt,
		Message:            "Not All Checks completed",
	})
	for _, c := range conditions {
		if c.Reason == "" {
			c.Reason = "NotSpecified"
		}
		if !c.LastTransitionTime.IsZero() {
			if c.LastTransitionTime.Time.Sub(r.lt.Time).Seconds() > 0 {
				r.lt = c.LastTransitionTime
			}
		}
		if c.LastTransitionTime.IsZero() {
			c.LastTransitionTime = r.lt
		}
		if source != "" {
			c.Reason = fmt.Sprintf("%s:%s", source, c.Reason)
			c.Type = fmt.Sprintf("%s%s", source, c.Type)
		}
		meta.SetStatusCondition(&r.router.Status.Conditions, c)
	}
}

// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=routers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=routers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crds.kloudlite.io,resources=routers/finalizers,verbs=update

func (r *RouterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.logger = GetLogger(req.NamespacedName)
	logger := r.logger.With("RECONCILE", "true")

	logger.Infof("Request received for: %v\n", req.String())

	router := &crdsv1.Router{}
	if err := r.Get(ctx, req.NamespacedName, router); err != nil {
		if apiErrors.IsNotFound(err) {
			return reconcileResult.OK()
		}
		return reconcileResult.Failed()
	}
	r.router = router

	if router.GetDeletionTimestamp() != nil {
		return r.finalizeRouter(ctx, router)
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
		if err := controllerutil.SetControllerReference(router, ing, r.Scheme); err != nil {
			return errors.NewEf(err, "could not set controller reference")
		}
		return nil
	}); err != nil {
		return r.notifyAndDie(ctx, errors.NewEf(err, "could not apply ingress"))
	}

	router.Status.IPs = []string{}
	if len(ing.Status.LoadBalancer.Ingress) > 0 {
		for _, item := range ing.Status.LoadBalancer.Ingress {
			router.Status.IPs = append(router.Status.IPs, item.IP)
		}
	}
	r.buildConditions("", metav1.Condition{
		Type:    "Ready",
		Status:  metav1.ConditionTrue,
		Reason:  "IngressIsLive",
		Message: "Ingress resource has acquired IP",
	})
	logger.Info("Reconcile Completed ...")
	return r.notify(ctx)
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
