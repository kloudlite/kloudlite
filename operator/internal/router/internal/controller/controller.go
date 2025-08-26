package controller

import (
	"context"
	"fmt"
	"maps"
	"strings"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"golang.org/x/crypto/bcrypt"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/yaml"

	"github.com/codingconcepts/env"
	fn "github.com/kloudlite/kloudlite/operator/toolkit/functions"
	"github.com/kloudlite/kloudlite/operator/toolkit/reconciler"
	v1 "github.com/kloudlite/operator/api/v1"
	"github.com/kloudlite/operator/internal/router/internal/templates"
)

type envVars struct {
	MaxConcurrentReconciles int `env:"MAX_CONCURRENT_RECONCILES" default:"5"`
}

type Reconciler struct {
	client.Client
	Scheme          *runtime.Scheme
	env             envVars
	templateIngress []byte
}

func (r *Reconciler) GetName() string {
	return v1.ProjectDomain + "/router-controller"
}

// +kubebuilder:rbac:groups=kloudlite.io,resources=routers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kloudlite.io,resources=routers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kloudlite.io,resources=routers/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := reconciler.NewRequest(ctx, r.Client, request.NamespacedName, &v1.Router{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	req.PreReconcile()
	defer req.PostReconcile()

	return reconciler.ReconcileSteps(req, []reconciler.Step[*v1.Router]{
		{
			Name:     "setup-basic-auth",
			Title:    "Setup Basic Auth",
			OnCreate: r.setupBasicAuth,
			OnDelete: r.cleanupBasicAuth,
		},
		{
			Name:     "setup-ingress",
			Title:    "Setup Ingress",
			OnCreate: r.setupIngress,
			OnDelete: r.cleanupIngress,
		},
	})
}

func (r *Reconciler) findClusterIssuer(ctx context.Context, obj *v1.Router) (*certmanagerv1.ClusterIssuer, error) {
	https := obj.Spec.Https

	if https != nil && https.ClusterIssuer != "" {
		var issuer certmanagerv1.ClusterIssuer
		if err := r.Get(ctx, fn.NN("", https.ClusterIssuer), &issuer, &client.GetOptions{}); err != nil {
			return nil, err
		}

		return &issuer, nil
	}

	var issuerList certmanagerv1.ClusterIssuerList
	if err := r.List(ctx, &issuerList, &client.ListOptions{Limit: 1}); err != nil {
		return nil, err
	}

	if len(issuerList.Items) != 1 {
		return nil, fmt.Errorf("no cluster issuer found")
	}

	return &issuerList.Items[0], nil
}

func (r *Reconciler) findIngressClass(ctx context.Context, obj *v1.Router) (string, error) {
	if obj.Spec.IngressClass != "" {
		return obj.Spec.IngressClass, nil
	}

	var ingressClassList networkingv1.IngressClassList
	if err := r.List(ctx, &ingressClassList, &client.ListOptions{Limit: 1}); err != nil {
		return "", err
	}

	if len(ingressClassList.Items) != 1 {
		return "", fmt.Errorf("no/multiple ingress classes found")
	}

	return ingressClassList.Items[0].Name, nil
}

func isHttpsEnabled(obj *v1.Router) bool {
	return obj.Spec.Https != nil && (obj.Spec.Https.Enabled == nil || *obj.Spec.Https.Enabled)
}

func GenNginxIngressAnnotations(obj *v1.Router) map[string]string {
	annotations := make(map[string]string)
	annotations["nginx.ingress.kubernetes.io/preserve-trailing-slash"] = "true"
	annotations["nginx.ingress.kubernetes.io/rewrite-target"] = "/$1"
	annotations["nginx.ingress.kubernetes.io/from-to-www-redirect"] = "true"

	if obj.Spec.MaxBodySizeInMB != nil {
		annotations["nginx.ingress.kubernetes.io/proxy-body-size"] = fmt.Sprintf("%vm", *obj.Spec.MaxBodySizeInMB)
	}

	if obj.Spec.Https.IsEnabled() {
		annotations["nginx.kubernetes.io/ssl-redirect"] = "true"
		annotations["nginx.ingress.kubernetes.io/force-ssl-redirect"] = fmt.Sprintf("%v", obj.Spec.Https.ForceRedirect)
	}

	if obj.Spec.RateLimit.IsEnabled() {
		if obj.Spec.RateLimit.Rps > 0 {
			annotations["nginx.ingress.kubernetes.io/limit-rps"] = fmt.Sprintf("%v", obj.Spec.RateLimit.Rps)
		}
		if obj.Spec.RateLimit.Rpm > 0 {
			annotations["nginx.ingress.kubernetes.io/limit-rpm"] = fmt.Sprintf("%v", obj.Spec.RateLimit.Rpm)
		}
		if obj.Spec.RateLimit.Connections > 0 {
			annotations["nginx.ingress.kubernetes.io/limit-connections"] = fmt.Sprintf("%v", obj.Spec.RateLimit.Connections)
		}
	}

	if obj.Spec.Cors.IsEnabled() {
		annotations["nginx.ingress.kubernetes.io/enable-cors"] = "true"
		annotations["nginx.ingress.kubernetes.io/cors-allow-origin"] = strings.Join(obj.Spec.Cors.Origins, ",")
		annotations["nginx.ingress.kubernetes.io/cors-allow-credentials"] = fmt.Sprintf("%v", obj.Spec.Cors.AllowCredentials)
	}

	if obj.Spec.BackendProtocol != nil {
		annotations["nginx.ingress.kubernetes.io/backend-protocol"] = *obj.Spec.BackendProtocol
	}

	if obj.Spec.BasicAuth.IsEnabled() {
		annotations["nginx.ingress.kubernetes.io/auth-type"] = "basic"
		annotations["nginx.ingress.kubernetes.io/auth-secret"] = obj.Spec.BasicAuth.SecretName
		annotations["nginx.ingress.kubernetes.io/auth-realm"] = "route is protected by basic auth"
	}

	maps.Copy(annotations, obj.Spec.NginxIngressAnnotations)

	return annotations
}

func (r *Reconciler) setupBasicAuth(check *reconciler.Check[*v1.Router], obj *v1.Router) reconciler.StepResult {
	if !obj.IsBasicAuthEnabled() {
		return r.cleanupBasicAuth(check, obj)
	}

	hasUpdate := false
	if obj.Spec.BasicAuth.Username == "" {
		hasUpdate = true
		obj.Spec.BasicAuth.Username = obj.Name
	}

	if obj.Spec.BasicAuth.SecretName == "" {
		hasUpdate = true
		obj.Spec.BasicAuth.SecretName = obj.Name + "-basic-auth"
	}

	if hasUpdate {
		if err := r.Update(check.Context(), obj); err != nil {
			return check.Failed(err)
		}
		return check.Abort("waiting for resource reconciliation")
	}

	basicAuthSecret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.BasicAuth.SecretName, Namespace: obj.Namespace}, Type: "Opaque"}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, basicAuthSecret, func() error {
		basicAuthSecret.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		if _, ok := basicAuthSecret.Data["password"]; ok {
			return nil
		}

		password := fn.CleanerNanoid(48)
		passwdHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		basicAuthSecret.StringData = map[string]string{
			"auth":     fmt.Sprintf("%s:%s", obj.Spec.BasicAuth.Username, passwdHash),
			"username": obj.Spec.BasicAuth.Username,
			"password": password,
		}
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	return check.Passed()
}

func (r *Reconciler) cleanupBasicAuth(check *reconciler.Check[*v1.Router], obj *v1.Router) reconciler.StepResult {
	if obj.Spec.BasicAuth == nil {
		return check.Passed()
	}

	basicAuthSecret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.BasicAuth.SecretName, Namespace: obj.Namespace}}
	if err := fn.DeleteAndWait(check.Context(), r.Client, basicAuthSecret); err != nil {
		return check.Errored(err)
	}

	return check.Passed()
}

func (r *Reconciler) setupIngress(check *reconciler.Check[*v1.Router], obj *v1.Router) reconciler.StepResult {
	ctx := check.Context()

	hasUpdate := false

	// Set default ingress class if not specified
	if obj.Spec.IngressClass == "" {
		ingClass, err := r.findIngressClass(ctx, obj)
		if err != nil {
			return check.Failed(err)
		}
		obj.Spec.IngressClass = ingClass
		hasUpdate = true
	}

	// Set default cluster issuer for HTTPS if not specified
	if obj.Spec.Https != nil && obj.Spec.Https.ClusterIssuer == "" {
		issuer, err := r.findClusterIssuer(check.Context(), obj)
		if err != nil {
			return check.Failed(err)
		}
		obj.Spec.Https.ClusterIssuer = issuer.Name
		hasUpdate = true
	}

	if hasUpdate {
		if err := r.Update(ctx, obj); err != nil {
			return check.Failed(err)
		}
		return check.Abort("waiting for resource reconciliation")
	}

	// Skip ingress creation if no routes defined
	if len(obj.Spec.Routes) == 0 {
		return check.Passed()
	}

	// Collect all hosts from routes
	var allHosts []string
	for _, route := range obj.Spec.Routes {
		allHosts = append(allHosts, route.Host)
	}

	nginxIngressAnnotations := GenNginxIngressAnnotations(obj)

	b, err := templates.ParseBytes(r.templateIngress, templates.IngressTemplateArgs{
		CertSecretNamePrefix: obj.Name,
		IngressClassName:     obj.Spec.IngressClass,
		HttpsEnabled:         isHttpsEnabled(obj),
		Hosts:                allHosts,
		Routes:               obj.Spec.Routes,
	})
	if err != nil {
		return check.Failed(fmt.Errorf("failed to parse ingress template spec: %w", err))
	}

	ing := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      obj.Name,
			Namespace: obj.Namespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, ing, func() error {
		ing.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		ing.SetLabels(fn.MapMerge(ing.Labels, obj.GetLabels()))
		ing.SetAnnotations(fn.MapMerge(ing.Annotations, nginxIngressAnnotations))

		if err := yaml.Unmarshal(b, &ing.Spec); err != nil {
			return fmt.Errorf("failed to unmarshal ingress spec: %w", err)
		}

		return nil
	}); err != nil {
		return check.Failed(err)
	}

	return check.Passed()
}

func (r *Reconciler) cleanupIngress(check *reconciler.Check[*v1.Router], obj *v1.Router) reconciler.StepResult {
	ctx := check.Context()

	// Delete the ingress resource
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      obj.Name,
			Namespace: obj.Namespace,
		},
	}

	if err := fn.DeleteAndWait(ctx, r.Client, ingress); err != nil {
		return check.Errored(err)
	}

	return check.Passed()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()

	if err := env.Set(&r.env); err != nil {
		return fmt.Errorf("failed to set env: %w", err)
	}

	var err error
	r.templateIngress, err = templates.Read(templates.IngressTemplate)
	if err != nil {
		return err
	}

	builder := ctrl.NewControllerManagedBy(mgr).For(&v1.Router{}).Named(r.GetName())
	builder.Owns(&networkingv1.Ingress{})

	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.env.MaxConcurrentReconciles})
	builder.WithEventFilter(reconciler.ReconcileFilter(mgr.GetEventRecorderFor(r.GetName())))
	return builder.Complete(r)
}
