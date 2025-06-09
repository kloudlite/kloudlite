package router_controller

import (
	"fmt"
	"strings"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
)

func GenNginxIngressAnnotations(obj *crdsv1.Router) map[string]string {
	annotations := make(map[string]string)
	annotations["nginx.ingress.kubernetes.io/preserve-trailing-slash"] = "true"
	annotations["nginx.ingress.kubernetes.io/rewrite-target"] = "/$1"
	annotations["nginx.ingress.kubernetes.io/from-to-www-redirect"] = "true"

	if obj.Spec.MaxBodySizeInMB != nil {
		annotations["nginx.ingress.kubernetes.io/proxy-body-size"] = fmt.Sprintf("%vm", *obj.Spec.MaxBodySizeInMB)
	}

	if obj.Spec.Https != nil && obj.Spec.Https.Enabled {
		annotations["nginx.kubernetes.io/ssl-redirect"] = "true"
		annotations["nginx.ingress.kubernetes.io/force-ssl-redirect"] = fmt.Sprintf("%v", obj.Spec.Https.ForceRedirect)

		// cert-manager.io annotations, [read more](https://cert-manager.io/docs/usage/ingress/#supported-annotations)
		// annotations["cert-manager.io/cluster-issuer"] = r.Env.DefaultClusterIssuer
		// annotations["cert-manager.io/renew-before"] = "168h" // renew certificates a week before expiry
		// annotations["acme.cert-manager.io/http01-ingress-class"] = r.Env.DefaultIngressClass
	}

	if obj.Spec.RateLimit != nil && obj.Spec.RateLimit.Enabled {
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

	if obj.Spec.Cors != nil && obj.Spec.Cors.Enabled {
		annotations["nginx.ingress.kubernetes.io/enable-cors"] = "true"
		annotations["nginx.ingress.kubernetes.io/cors-allow-origin"] = strings.Join(obj.Spec.Cors.Origins, ",")
		annotations["nginx.ingress.kubernetes.io/cors-allow-credentials"] = fmt.Sprintf("%v", obj.Spec.Cors.AllowCredentials)
	}

	if obj.Spec.BackendProtocol != nil {
		annotations["nginx.ingress.kubernetes.io/backend-protocol"] = *obj.Spec.BackendProtocol
	}

	if obj.Spec.BasicAuth != nil && obj.Spec.BasicAuth.Enabled {
		annotations["nginx.ingress.kubernetes.io/auth-type"] = "basic"
		annotations["nginx.ingress.kubernetes.io/auth-secret"] = obj.Spec.BasicAuth.SecretName
		annotations["nginx.ingress.kubernetes.io/auth-realm"] = "route is protected by basic auth"
	}

	return annotations
}
