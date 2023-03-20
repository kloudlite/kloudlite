package primary

import (
	"fmt"
	v1 "github.com/kloudlite/operator/apis/cluster-setup/v1"
	lc "github.com/kloudlite/operator/operators/cluster-setup/internal/constants"
	"github.com/kloudlite/operator/operators/cluster-setup/internal/templates"
	fn "github.com/kloudlite/operator/pkg/functions"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strings"
)

const (
	DnsApiCreated             string = "dns-api-created"
	AuthApiCreated            string = "auth-api-created"
	ConsoleApiCreated         string = "console-api-created"
	CiApiCreated              string = "ci-api-created"
	FinanceApiCreated         string = "finance-api-created"
	IAMApiCreated             string = "iAMA-api-created"
	CommsApiCreated           string = "comms-api-created"
	GatewayApiCreated         string = "gateway-api-created"
	WebhooksApiCreated        string = "webhooks-api-created"
	JsEvalApiCreated          string = "js-eval-api-created"
	WebhookAuthzSecretsReady  string = "webhook-authz-secrets-ready"
	StripeCredsReady          string = "stripe-creds-ready"
	AuditLoggingWorkerCreated string = "audit-logging-worker-created"
	InfraApiCreated           string = "infra-api-created"
)

type ReconFn func(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result

func (r *Reconciler) ensureWebhookAuthzSecrets(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(WebhookAuthzSecretsReady)
	defer req.LogPostCheck(WebhookAuthzSecretsReady)

	authzSecret, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.WebhookAuthzCreds.Namespace, obj.Spec.WebhookAuthzCreds.Name), &corev1.Secret{})
	if err != nil {
		return req.CheckFailed(WebhookAuthzSecretsReady, check, err.Error()).Err(nil)
	}

	scrt := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.WebhookAuthzCreds.Name, Namespace: lc.NsCore}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, scrt, func() error {
		scrt.Data = authzSecret.Data
		return nil
	}); err != nil {
		return req.CheckFailed(WebhookAuthzSecretsReady, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != checks[WebhookAuthzSecretsReady] {
		checks[WebhookAuthzSecretsReady] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureStripeCreds(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	screds, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.StripeCreds.Namespace, obj.Spec.StripeCreds.Name), &corev1.Secret{})
	if err != nil {
		return req.CheckFailed(StripeCredsReady, check, err.Error()).Err(nil)
	}

	localCreds := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.StripeCreds.Name, Namespace: lc.NsCore}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, localCreds, func() error {
		if !fn.IsOwner(localCreds, fn.AsOwner(obj)) {
			localCreds.SetOwnerReferences(append(localCreds.GetOwnerReferences(), fn.AsOwner(obj, true)))
		}
		localCreds.Labels = screds.Labels
		localCreds.Data = screds.Data
		return nil
	}); err != nil {
		return req.CheckFailed(StripeCredsReady, check, err.Error())
	}

	check.Status = true
	if check != checks[StripeCredsReady] {
		checks[StripeCredsReady] = check
		req.UpdateStatus()
	}
	return req.Next()

}

func (r *Reconciler) ensureAuthApi(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(AuthApiCreated)
	defer req.LogPostCheck(AuthApiCreated)

	b, err := templates.Parse(templates.AuthApi, map[string]any{
		"namespace":        lc.NsCore,
		"image-auth-api":   obj.Spec.SharedConstants.ImageAuthApi,
		"owner-refs":       []metav1.OwnerReference{fn.AsOwner(obj, true)},
		"shared-constants": obj.Spec.SharedConstants,
	})
	if err != nil {
		return req.CheckFailed(AuthApiCreated, check, err.Error()).Err(nil)
	}

	if _, err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
		return req.CheckFailed(AuthApiCreated, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != checks[AuthApiCreated] {
		checks[AuthApiCreated] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureConsoleApi(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(ConsoleApiCreated)
	defer req.LogPostCheck(ConsoleApiCreated)

	lokiAuthScrt, err := rApi.Get(ctx, r.Client, fn.NN(lc.NsMonitoring, fmt.Sprintf("%s-basic-auth", obj.Spec.LokiValues.ServiceName)), &corev1.Secret{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.CheckFailed(ConsoleApiCreated, check, err.Error()).Err(nil)
		}
		lokiAuthScrt = &corev1.Secret{}
	}

	promAuthScrt, err := rApi.Get(ctx, r.Client, fn.NN(lc.NsMonitoring, fmt.Sprintf("%s-basic-auth", obj.Spec.PrometheusValues.ServiceName)), &corev1.Secret{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.CheckFailed(ConsoleApiCreated, check, err.Error()).Err(nil)
		}
		promAuthScrt = &corev1.Secret{}
	}

	b, err := templates.Parse(templates.ConsoleApi, map[string]any{
		"namespace":                lc.NsCore,
		"svc-account":              lc.ClusterSvcAccount,
		"shared-constants":         obj.Spec.SharedConstants,
		"owner-refs":               []metav1.OwnerReference{fn.AsOwner(obj, true)},
		"loki-basic-auth-password": string(lokiAuthScrt.Data["password"]),
		"prom-basic-auth-password": string(promAuthScrt.Data["password"]),
	})
	if err != nil {
		return req.CheckFailed(ConsoleApiCreated, check, err.Error()).Err(nil)
	}

	if _, err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
		return req.CheckFailed(ConsoleApiCreated, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != checks[ConsoleApiCreated] {
		checks[ConsoleApiCreated] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureCIApi(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(CiApiCreated)
	defer req.LogPostCheck(CiApiCreated)

	b, err := templates.Parse(templates.CiApi, map[string]any{
		"namespace":        lc.NsCore,
		"svc-account":      lc.DefaultSvcAccount,
		"shared-constants": obj.Spec.SharedConstants,
		"owner-refs":       []metav1.OwnerReference{fn.AsOwner(obj, true)},
	})
	if err != nil {
		return req.CheckFailed(CiApiCreated, check, err.Error()).Err(nil)
	}

	if _, err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
		return req.CheckFailed(CiApiCreated, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != checks[CiApiCreated] {
		checks[CiApiCreated] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureDnsApi(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(DnsApiCreated)
	defer req.LogPostCheck(DnsApiCreated)

	b, err := templates.Parse(templates.DnsApi, map[string]any{
		"namespace":         lc.NsCore,
		"svc-account":       lc.DefaultSvcAccount,
		"shared-constants":  obj.Spec.SharedConstants,
		"owner-refs":        []metav1.OwnerReference{fn.AsOwner(obj, true)},
		"dns-names":         strings.Join(obj.Spec.Networking.DnsNames, ","),
		"cname-base-domain": obj.Spec.Networking.EdgeCNAME,
	})
	if err != nil {
		return req.CheckFailed(DnsApiCreated, check, err.Error()).Err(nil)
	}

	if _, err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
		return req.CheckFailed(DnsApiCreated, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != checks[DnsApiCreated] {
		checks[DnsApiCreated] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureFinanceApi(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(FinanceApiCreated)
	defer req.LogPostCheck(FinanceApiCreated)

	b, err := templates.Parse(templates.FinanceApi, map[string]any{
		"namespace":        lc.NsCore,
		"svc-account":      lc.ClusterSvcAccount,
		"shared-constants": obj.Spec.SharedConstants,
		"owner-refs":       []metav1.OwnerReference{fn.AsOwner(obj, true)},
	})
	if err != nil {
		return req.CheckFailed(FinanceApiCreated, check, err.Error()).Err(nil)
	}

	if _, err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
		return req.CheckFailed(FinanceApiCreated, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != checks[FinanceApiCreated] {
		checks[FinanceApiCreated] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureIAMApi(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(IAMApiCreated)
	defer req.LogPostCheck(IAMApiCreated)

	b, err := templates.Parse(templates.IamApi, map[string]any{
		"namespace":        lc.NsCore,
		"svc-account":      lc.DefaultSvcAccount,
		"shared-constants": obj.Spec.SharedConstants,
		"owner-refs":       []metav1.OwnerReference{fn.AsOwner(obj, true)},
	})
	if err != nil {
		return req.CheckFailed(IAMApiCreated, check, err.Error()).Err(nil)
	}

	if _, err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
		return req.CheckFailed(IAMApiCreated, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != checks[IAMApiCreated] {
		checks[IAMApiCreated] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureCommsApi(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(CommsApiCreated)
	defer req.LogPostCheck(CommsApiCreated)

	b, err := templates.Parse(templates.CommsApi, map[string]any{
		"namespace":        lc.NsCore,
		"svc-account":      lc.DefaultSvcAccount,
		"shared-constants": obj.Spec.SharedConstants,
		"owner-refs":       []metav1.OwnerReference{fn.AsOwner(obj, true)},
	})
	if err != nil {
		return req.CheckFailed(CommsApiCreated, check, err.Error()).Err(nil)
	}

	if _, err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
		return req.CheckFailed(CommsApiCreated, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != checks[CommsApiCreated] {
		checks[CommsApiCreated] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureWebhooksApi(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(WebhooksApiCreated)
	defer req.LogPostCheck(WebhooksApiCreated)

	b, err := templates.Parse(templates.WebhooksApi, map[string]any{
		"namespace":        lc.NsCore,
		"svc-account":      lc.DefaultSvcAccount,
		"shared-constants": obj.Spec.SharedConstants,
		"owner-refs":       []metav1.OwnerReference{fn.AsOwner(obj, true)},
	})
	if err != nil {
		return req.CheckFailed(WebhooksApiCreated, check, err.Error()).Err(nil)
	}

	if _, err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
		return req.CheckFailed(WebhooksApiCreated, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != checks[WebhooksApiCreated] {
		checks[WebhooksApiCreated] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureJsEvalApi(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(JsEvalApiCreated)
	defer req.LogPostCheck(JsEvalApiCreated)

	b, err := templates.Parse(templates.JsEvalApi, map[string]any{
		"namespace":        lc.NsCore,
		"svc-account":      lc.DefaultSvcAccount,
		"shared-constants": obj.Spec.SharedConstants,
		"owner-refs":       []metav1.OwnerReference{fn.AsOwner(obj, true)},
	})
	if err != nil {
		return req.CheckFailed(JsEvalApiCreated, check, err.Error()).Err(nil)
	}

	if _, err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
		return req.CheckFailed(JsEvalApiCreated, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != checks[JsEvalApiCreated] {
		checks[JsEvalApiCreated] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureAuditLoggingWorker(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	b, err := templates.Parse(templates.AuditLoggingWorker, map[string]any{
		"namespace":        lc.NsCore,
		"svc-account":      lc.DefaultSvcAccount,
		"shared-constants": obj.Spec.SharedConstants,
		"owner-refs":       []metav1.OwnerReference{fn.AsOwner(obj, true)},
	})
	if err != nil {
		return req.CheckFailed(AuditLoggingWorkerCreated, check, err.Error()).Err(nil)
	}

	if _, err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
		return req.CheckFailed(AuditLoggingWorkerCreated, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != checks[AuditLoggingWorkerCreated] {
		checks[AuditLoggingWorkerCreated] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureInfraApi(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(InfraApiCreated)
	defer req.LogPostCheck(InfraApiCreated)

	b, err := templates.Parse(templates.InfraApi, map[string]any{
		"namespace":        lc.NsCore,
		"svc-account":      lc.DefaultSvcAccount,
		"shared-constants": obj.Spec.SharedConstants,
		"owner-refs":       []metav1.OwnerReference{fn.AsOwner(obj, true)},
	})
	if err != nil {
		return req.CheckFailed(InfraApiCreated, check, err.Error()).Err(nil)
	}

	if _, err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
		return req.CheckFailed(InfraApiCreated, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != checks[InfraApiCreated] {
		checks[InfraApiCreated] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureGatewayApi(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(GatewayApiCreated)
	defer req.LogPostCheck(GatewayApiCreated)

	b, err := templates.Parse(templates.GatewayApi, map[string]any{
		"namespace":        lc.NsCore,
		"svc-account":      lc.DefaultSvcAccount,
		"shared-constants": obj.Spec.SharedConstants,
		"owner-refs":       []metav1.OwnerReference{fn.AsOwner(obj, true)},
	})
	if err != nil {
		return req.CheckFailed(GatewayApiCreated, check, err.Error()).Err(nil)
	}

	if _, err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
		return req.CheckFailed(GatewayApiCreated, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != checks[GatewayApiCreated] {
		checks[GatewayApiCreated] = check
		req.UpdateStatus()
	}
	return req.Next()
}
