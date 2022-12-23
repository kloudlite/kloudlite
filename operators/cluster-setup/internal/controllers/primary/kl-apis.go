package primary

import (
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "operators.kloudlite.io/apis/cluster-setup/v1"
	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	lc "operators.kloudlite.io/operators/cluster-setup/internal/constants"
	"operators.kloudlite.io/operators/cluster-setup/internal/templates"
	fn "operators.kloudlite.io/pkg/functions"
	rApi "operators.kloudlite.io/pkg/operator"
	stepResult "operators.kloudlite.io/pkg/operator/step-result"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strings"
)

const (
	DnsApiCreated            string = "dns-api-created"
	AuthApiCreated           string = "auth-api-created"
	ConsoleApiCreated        string = "console-api-created"
	CiApiCreated             string = "ci-api-created"
	FinanceApiCreated        string = "finance-api-created"
	IAMApiCreated            string = "iAMA-api-created"
	CommsApiCreated          string = "comms-api-created"
	GatewayApiCreated        string = "gateway-api-created"
	WebhooksApiCreated       string = "webhooks-api-created"
	JsEvalApiCreated         string = "js-eval-api-created"
	WebhookAuthzSecretsReady string = "webhook-authz-secrets-ready"
	StripeCredsReady         string = "stripe-creds-ready"
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
		return req.UpdateStatus()
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
		return req.UpdateStatus()
	}
	return req.Next()

}

func (r *Reconciler) ensureAuthApi(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(AuthApiCreated)
	defer req.LogPostCheck(AuthApiCreated)

	authApi, err := rApi.Get(ctx, r.Client, fn.NN(lc.NsCore, obj.Spec.SharedConstants.AppAuthApi), &crdsv1.App{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.CheckFailed(AuthApiCreated, check, err.Error()).Err(nil)
		}
		req.Logger.Infof("auth-api does not exist, will be creating it")
	}

	if authApi == nil || check.Generation > checks[AuthApiCreated].Generation {
		b, err := templates.Parse(templates.AuthApi, map[string]any{
			"namespace":        lc.NsCore,
			"image-auth-api":   obj.Spec.SharedConstants.ImageAuthApi,
			"owner-refs":       []metav1.OwnerReference{fn.AsOwner(obj, true)},
			"shared-constants": obj.Spec.SharedConstants,
		})
		if err != nil {
			return req.CheckFailed(AuthApiCreated, check, err.Error()).Err(nil)
		}

		if err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
			return req.CheckFailed(AuthApiCreated, check, err.Error()).Err(nil)
		}
	}

	check.Status = true
	if check != checks[AuthApiCreated] {
		checks[AuthApiCreated] = check
		return req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureConsoleApi(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(ConsoleApiCreated)
	defer req.LogPostCheck(ConsoleApiCreated)

	consoleApi, err := rApi.Get(ctx, r.Client, fn.NN(lc.NsCore, obj.Spec.SharedConstants.AppConsoleApi), &crdsv1.App{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.CheckFailed(ConsoleApiCreated, check, err.Error()).Err(nil)
		}
		req.Logger.Infof("console-api does not exist, will be creating it")
	}

	if consoleApi == nil || check.Generation > checks[ConsoleApiCreated].Generation {
		b, err := templates.Parse(templates.ConsoleApi, map[string]any{
			"namespace":        lc.NsCore,
			"svc-account":      lc.ClusterSvcAccount,
			"shared-constants": obj.Spec.SharedConstants,
			"owner-refs":       []metav1.OwnerReference{fn.AsOwner(obj, true)},
		})
		if err != nil {
			return req.CheckFailed(ConsoleApiCreated, check, err.Error()).Err(nil)
		}

		if err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
			return req.CheckFailed(ConsoleApiCreated, check, err.Error()).Err(nil)
		}
	}

	check.Status = true
	if check != checks[ConsoleApiCreated] {
		checks[ConsoleApiCreated] = check
		return req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureCIApi(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(CiApiCreated)
	defer req.LogPostCheck(CiApiCreated)

	ciApi, err := rApi.Get(ctx, r.Client, fn.NN(lc.NsCore, obj.Spec.SharedConstants.AppCiApi), &crdsv1.App{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.CheckFailed(CiApiCreated, check, err.Error()).Err(nil)
		}
		req.Logger.Infof("ci-api does not exist, will be creating it")
	}

	if ciApi == nil || check.Generation > checks[CiApiCreated].Generation {
		b, err := templates.Parse(templates.CiApi, map[string]any{
			"namespace":        lc.NsCore,
			"svc-account":      lc.DefaultSvcAccount,
			"shared-constants": obj.Spec.SharedConstants,
			"owner-refs":       []metav1.OwnerReference{fn.AsOwner(obj, true)},
		})
		if err != nil {
			return req.CheckFailed(CiApiCreated, check, err.Error()).Err(nil)
		}

		if err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
			return req.CheckFailed(CiApiCreated, check, err.Error()).Err(nil)
		}
	}

	check.Status = true
	if check != checks[CiApiCreated] {
		checks[CiApiCreated] = check
		return req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureDnsApi(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(DnsApiCreated)
	defer req.LogPostCheck(DnsApiCreated)

	dnsApi, err := rApi.Get(ctx, r.Client, fn.NN(lc.NsCore, obj.Spec.SharedConstants.AppDnsApi), &crdsv1.App{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.CheckFailed(DnsApiCreated, check, err.Error())
		}
		req.Logger.Infof("dns-api does not exist, will be creating it")
	}

	if dnsApi == nil || check.Generation > checks[DnsApiCreated].Generation {
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

		if err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
			return req.CheckFailed(DnsApiCreated, check, err.Error()).Err(nil)
		}
	}

	check.Status = true
	if check != checks[DnsApiCreated] {
		checks[DnsApiCreated] = check
		return req.UpdateStatus()
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

	if err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
		return req.CheckFailed(FinanceApiCreated, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != checks[FinanceApiCreated] {
		checks[FinanceApiCreated] = check
		return req.UpdateStatus()
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

	if err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
		return req.CheckFailed(IAMApiCreated, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != checks[IAMApiCreated] {
		checks[IAMApiCreated] = check
		return req.UpdateStatus()
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

	if err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
		return req.CheckFailed(CommsApiCreated, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != checks[CommsApiCreated] {
		checks[CommsApiCreated] = check
		return req.UpdateStatus()
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

	if err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
		return req.CheckFailed(WebhooksApiCreated, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != checks[WebhooksApiCreated] {
		checks[WebhooksApiCreated] = check
		return req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureJsEvalApi(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(JsEvalApiCreated)
	defer req.LogPostCheck(JsEvalApiCreated)

	commsApi, err := rApi.Get(ctx, r.Client, fn.NN(lc.NsCore, obj.Spec.SharedConstants.AppCommsApi), &crdsv1.App{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.CheckFailed(JsEvalApiCreated, check, err.Error())
		}
		req.Logger.Infof("comms-api does not exist, will be creating it")
	}

	if commsApi == nil || check.Generation > checks[JsEvalApiCreated].Generation {
		b, err := templates.Parse(templates.JsEvalApi, map[string]any{
			"namespace":        lc.NsCore,
			"svc-account":      lc.DefaultSvcAccount,
			"shared-constants": obj.Spec.SharedConstants,
			"owner-refs":       []metav1.OwnerReference{fn.AsOwner(obj, true)},
		})
		if err != nil {
			return req.CheckFailed(JsEvalApiCreated, check, err.Error()).Err(nil)
		}

		if err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
			return req.CheckFailed(JsEvalApiCreated, check, err.Error()).Err(nil)
		}
	}

	check.Status = true
	if check != checks[JsEvalApiCreated] {
		checks[JsEvalApiCreated] = check
		return req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureGatewayApi(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(GatewayApiCreated)
	defer req.LogPostCheck(GatewayApiCreated)

	gatewayApi, err := rApi.Get(ctx, r.Client, fn.NN(lc.NsCore, obj.Spec.SharedConstants.AppGqlGatewayApi), &crdsv1.App{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.CheckFailed(GatewayApiCreated, check, err.Error())
		}
		req.Logger.Infof("comms-api does not exist, will be creating it")
	}

	if gatewayApi == nil || check.Generation > checks[GatewayApiCreated].Generation {
		b, err := templates.Parse(templates.GatewayApi, map[string]any{
			"namespace":        lc.NsCore,
			"svc-account":      lc.DefaultSvcAccount,
			"shared-constants": obj.Spec.SharedConstants,
			"owner-refs":       []metav1.OwnerReference{fn.AsOwner(obj, true)},
		})
		if err != nil {
			return req.CheckFailed(GatewayApiCreated, check, err.Error()).Err(nil)
		}

		if err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
			return req.CheckFailed(GatewayApiCreated, check, err.Error()).Err(nil)
		}
	}

	check.Status = true
	if check != checks[GatewayApiCreated] {
		checks[GatewayApiCreated] = check
		return req.UpdateStatus()
	}
	return req.Next()
}
