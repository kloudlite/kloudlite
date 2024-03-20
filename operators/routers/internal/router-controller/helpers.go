package router_controller

import (
	"fmt"
	"strings"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	certmanagermetav1 "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	fn "github.com/kloudlite/operator/pkg/functions"
	rApi "github.com/kloudlite/operator/pkg/operator"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
)

func (r *Reconciler) parseAndExtractDomains(req *rApi.Request[*crdsv1.Router]) ([]string, []string, error) {
	ctx, obj := req.Context(), req.Object

	var wildcardPatterns []string

	if obj.Spec.Https != nil && obj.Spec.Https.Enabled {
		issuerName := r.getRouterClusterIssuer(obj)

		if issuerName == "" {
			return nil, nil, fmt.Errorf("no cluster issuer found, could not proceed, when https is enabled")
		}

		clusterIssuer, err := rApi.Get(ctx, r.Client, fn.NN("", issuerName), &certmanagerv1.ClusterIssuer{})
		if err != nil {
			if !apiErrors.IsNotFound(err) {
				return nil, nil, err
			}
			clusterIssuer = nil
		}

		if clusterIssuer != nil {
			for _, solver := range clusterIssuer.Spec.ACME.Solvers {
				if solver.DNS01 != nil {
					wildcardPatterns = solver.Selector.DNSNames
				}
			}
		}
	}

	wildcardDomains, nonWildcardDomains := FilterDomains(wildcardPatterns, obj.Spec.Domains)
	return wildcardDomains, nonWildcardDomains, nil
}

func FilterDomains(wildcardPatterns []string, domains []string) (wildcardDomains, nonWildcardDomains []string) {
	wildcardBases := map[string]struct{}{}
	for _, pattern := range wildcardPatterns {
		if strings.HasPrefix(pattern, "*.") {
			wildcardBases[pattern[2:]] = struct{}{}
			continue
		}
		wildcardBases[pattern] = struct{}{}
	}

	for _, domain := range domains {
		if _, ok := wildcardBases[domain]; ok {
			wildcardDomains = append(wildcardDomains, domain)
			continue
		}

		sp := strings.SplitN(domain, ".", 2)
		if len(sp) != 2 {
			nonWildcardDomains = append(nonWildcardDomains, domain)
			continue
		}

		if _, ok := wildcardBases[sp[1]]; ok {
			wildcardDomains = append(wildcardDomains, domain)
			continue
		}

		nonWildcardDomains = append(nonWildcardDomains, domain)
	}

	return
}

func IsHttpsCertReady(cert *certmanagerv1.Certificate) (bool, error) {
	var errmsg string
	var isReady bool

	if cert == nil {
		return false, fmt.Errorf("certificate, does not exist")
	}

	for _, cond := range cert.Status.Conditions {
		if cond.Type == certmanagerv1.CertificateConditionReady {
			isReady = cond.Status == certmanagermetav1.ConditionTrue
			if !isReady {
				errmsg = fmt.Sprintf("HTTPS Certificate is not ready yet. cert-manager says '%s'", cond.Message)
			}
		}
	}

	if !isReady {
		for _, cond := range cert.Status.Conditions {
			if cond.Type == certmanagerv1.CertificateConditionIssuing {
				errmsg = fmt.Sprintf("%s. It also says '%s'.", errmsg, cond.Message)
			}
		}

		return false, fmt.Errorf(errmsg)
	}

	if !isReady {
		return false, fmt.Errorf("waiting for cert-manager to reconcile certificate")
	}

	return true, nil
}
