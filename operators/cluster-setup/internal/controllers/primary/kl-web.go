package primary

import (
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "operators.kloudlite.io/apis/cluster-setup/v1"
	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	lc "operators.kloudlite.io/operators/cluster-setup/internal/constants"
	"operators.kloudlite.io/operators/cluster-setup/internal/templates"
	fn "operators.kloudlite.io/pkg/functions"
	rApi "operators.kloudlite.io/pkg/operator"
	stepResult "operators.kloudlite.io/pkg/operator/step-result"
)

const (
	AuthWebReady       string = "auth-web-ready"
	ConsoleWebCreated  string = "console-web-created"
	AccountsWebCreated string = "accounts-web-created"
	SocketWebCreated   string = "socket-web-created"
)

func (r *Reconciler) ensureAuthWeb(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	authWeb, err := rApi.Get(ctx, r.Client, fn.NN(lc.NsCore, obj.Spec.SharedConstants.AppAuthWeb), &crdsv1.App{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.CheckFailed(AuthWebReady, check, err.Error())
		}
		req.Logger.Infof("%s does not exist, will be creating it", obj.Spec.SharedConstants.AppAuthWeb)
	}

	if authWeb == nil || check.Generation > checks[AuthWebReady].Generation {
		b, err := templates.Parse(templates.AuthWeb, map[string]any{
			"namespace":        lc.NsCore,
			"shared-constants": obj.Spec.SharedConstants,
			"owner-refs":       []metav1.OwnerReference{fn.AsOwner(obj, true)},
		})
		if err != nil {
			return req.CheckFailed(AuthWebReady, check, err.Error()).Err(nil)
		}

		if err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
			return req.CheckFailed(AuthWebReady, check, err.Error()).Err(nil)
		}
	}

	check.Status = true
	if check != checks[AuthWebReady] {
		checks[AuthWebReady] = check
		return req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureConsoleWeb(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(ConsoleWebCreated)
	defer req.LogPostCheck(ConsoleWebCreated)

	consoleWeb, err := rApi.Get(ctx, r.Client, fn.NN(lc.NsCore, obj.Spec.SharedConstants.AppConsoleWeb), &crdsv1.App{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.CheckFailed(ConsoleWebCreated, check, err.Error())
		}
		req.Logger.Infof("%s does not exist, will be creating it", obj.Spec.SharedConstants.AppConsoleWeb)
	}

	if consoleWeb == nil || check.Generation > checks[ConsoleWebCreated].Generation {
		b, err := templates.Parse(templates.ConsoleWeb, map[string]any{
			"namespace":        lc.NsCore,
			"shared-constants": obj.Spec.SharedConstants,
			"owner-refs":       []metav1.OwnerReference{fn.AsOwner(obj, true)},
		})
		if err != nil {
			return req.CheckFailed(ConsoleWebCreated, check, err.Error()).Err(nil)
		}

		if err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
			return req.CheckFailed(ConsoleWebCreated, check, err.Error()).Err(nil)
		}
	}

	check.Status = true
	if check != checks[ConsoleWebCreated] {
		checks[ConsoleWebCreated] = check
		return req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureAccountsWeb(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(AccountsWebCreated)
	defer req.LogPostCheck(AccountsWebCreated)

	accountsWeb, err := rApi.Get(ctx, r.Client, fn.NN(lc.NsCore, obj.Spec.SharedConstants.AppAccountsWeb), &crdsv1.App{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.CheckFailed(AccountsWebCreated, check, err.Error())
		}
		req.Logger.Infof("%s does not exist, will be creating it", obj.Spec.SharedConstants.AppAccountsWeb)
	}

	if accountsWeb == nil || check.Generation > checks[AccountsWebCreated].Generation {
		b, err := templates.Parse(templates.AccountsWeb, map[string]any{
			"namespace":        lc.NsCore,
			"shared-constants": obj.Spec.SharedConstants,
			"owner-refs":       []metav1.OwnerReference{fn.AsOwner(obj, true)},
		})
		if err != nil {
			return req.CheckFailed(AccountsWebCreated, check, err.Error()).Err(nil)
		}

		if err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
			return req.CheckFailed(AccountsWebCreated, check, err.Error()).Err(nil)
		}
	}

	check.Status = true
	if check != checks[AccountsWebCreated] {
		checks[AccountsWebCreated] = check
		return req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureSocketWeb(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(SocketWebCreated)
	defer req.LogPostCheck(SocketWebCreated)

	socketWeb, err := rApi.Get(ctx, r.Client, fn.NN(lc.NsCore, obj.Spec.SharedConstants.AppSocketWeb), &crdsv1.App{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.CheckFailed(SocketWebCreated, check, err.Error())
		}
		req.Logger.Infof("%s does not exist, will be creating it", obj.Spec.SharedConstants.AppSocketWeb)
	}

	if socketWeb == nil || check.Generation > checks[SocketWebCreated].Generation {
		b, err := templates.Parse(templates.SocketWeb, map[string]any{
			"namespace":        lc.NsCore,
			"shared-constants": obj.Spec.SharedConstants,
			"owner-refs":       []metav1.OwnerReference{fn.AsOwner(obj, true)},
		})
		if err != nil {
			return req.CheckFailed(SocketWebCreated, check, err.Error()).Err(nil)
		}

		if err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
			return req.CheckFailed(SocketWebCreated, check, err.Error()).Err(nil)
		}
	}

	check.Status = true
	if check != checks[SocketWebCreated] {
		checks[SocketWebCreated] = check
		return req.UpdateStatus()
	}
	return req.Next()
}
