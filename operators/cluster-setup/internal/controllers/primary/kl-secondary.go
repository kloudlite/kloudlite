package primary

import (
	v1 "operators.kloudlite.io/apis/cluster-setup/v1"
	lc "operators.kloudlite.io/operators/cluster-setup/internal/constants"
	"operators.kloudlite.io/operators/cluster-setup/internal/templates"
	rApi "operators.kloudlite.io/pkg/operator"
	stepResult "operators.kloudlite.io/pkg/operator/step-result"
)

const (
	KloudliteAgentReady string = "kloudlite-agent-ready"
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
