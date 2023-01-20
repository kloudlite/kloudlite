package primary

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "github.com/kloudlite/operator/apis/cluster-setup/v1"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	lc "github.com/kloudlite/operator/operators/cluster-setup/internal/constants"
	"github.com/kloudlite/operator/operators/cluster-setup/internal/templates"
	fn "github.com/kloudlite/operator/pkg/functions"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	KloudliteAgentReady     string = "kloudlite-agent-ready"
	LokiRouterCreated       string = "loki-router-created"
	PrometheusRouterCreated string = "prometheus-router-created"
)

func (r *Reconciler) installKloudliteAgent(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	b, err := templates.Parse(templates.KloudliteAgent, map[string]any{
		"name":              obj.Spec.SharedConstants.AppKlAgent,
		"namespace":         lc.NsCore,
		"image":             obj.Spec.SharedConstants.ImageKlAgent,
		"svc-account":       lc.ClusterSvcAccount,
		"cluster-id":        obj.Spec.SecondaryClusterId,
		"kafka-secret-name": obj.Spec.SharedConstants.RedpandaAdminSecretName,
	})

	if err != nil {
		return req.CheckFailed(KloudliteAgentReady, check, err.Error()).Err(nil)
	}

	if err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
		return req.CheckFailed(KloudliteAgentReady, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != checks[KloudliteAgentReady] {
		checks[KloudliteAgentReady] = check
		return req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureLokiRouter(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(LokiRouterCreated)
	defer req.LogPostCheck(LokiRouterCreated)

	lokiRouter := &crdsv1.Router{ObjectMeta: metav1.ObjectMeta{Name: "loki-router", Namespace: lc.NsMonitoring}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, lokiRouter, func() error {
		if !fn.IsOwner(lokiRouter, fn.AsOwner(obj)) {
			lokiRouter.SetOwnerReferences(append(lokiRouter.GetOwnerReferences(), fn.AsOwner(obj, true)))
		}
		lokiRouter.Spec.Domains = []string{fmt.Sprintf("loki-external.%s.clusters.%s", obj.Spec.SecondaryClusterId, obj.Spec.SharedConstants.SubDomain)}
		lokiRouter.Spec.BasicAuth = crdsv1.BasicAuth{
			Enabled:    true,
			Username:   obj.Spec.SecondaryClusterId,
			SecretName: fmt.Sprintf("%s-basic-auth", obj.Spec.LokiValues.ServiceName),
		}
		//lokiRouter.Spec.Region = "master"
		lokiRouter.Spec.Https.Enabled = true
		lokiRouter.Spec.Https.ForceRedirect = true
		lokiRouter.Spec.Routes = []crdsv1.Route{
			{
				App:     obj.Spec.LokiValues.ServiceName,
				Path:    "/",
				Port:    3100,
				Rewrite: false,
			},
		}
		return nil
	}); err != nil {
		return req.CheckFailed(LokiRouterCreated, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != checks[LokiRouterCreated] {
		checks[LokiRouterCreated] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensurePrometheusRouter(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(PrometheusRouterCreated)
	defer req.LogPostCheck(PrometheusRouterCreated)

	promRouter := &crdsv1.Router{ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-router", obj.Spec.PrometheusValues.ServiceName), Namespace: lc.NsMonitoring}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, promRouter, func() error {
		if !fn.IsOwner(promRouter, fn.AsOwner(obj)) {
			promRouter.SetOwnerReferences(append(promRouter.GetOwnerReferences(), fn.AsOwner(obj, true)))
		}

		promRouter.Spec.Domains = []string{fmt.Sprintf("prom-external.%s.clusters.%s", obj.Spec.SecondaryClusterId, obj.Spec.SharedConstants.SubDomain)}
		promRouter.Spec.BasicAuth = crdsv1.BasicAuth{
			Enabled:    true,
			Username:   obj.Spec.SecondaryClusterId,
			SecretName: fmt.Sprintf("%s-basic-auth", obj.Spec.PrometheusValues.ServiceName),
		}
		promRouter.Spec.Https.Enabled = true
		promRouter.Spec.Https.ForceRedirect = true
		promRouter.Spec.Routes = []crdsv1.Route{
			{
				App:     fmt.Sprintf("prometheus-%s", obj.Spec.PrometheusValues.ServiceName),
				Path:    "/",
				Port:    9090,
				Rewrite: false,
			},
		}
		return nil
	}); err != nil {
		return req.CheckFailed(PrometheusRouterCreated, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != checks[PrometheusRouterCreated] {
		checks[PrometheusRouterCreated] = check
		req.UpdateStatus()
	}
	return req.Next()
}
