package managed

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/kloudlite/operator/pkg/constants"
	"github.com/kloudlite/operator/pkg/errors"
	"github.com/kloudlite/operator/pkg/helm"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/kloudlite/operator/apis/cluster-setup/v1"
	lc "github.com/kloudlite/operator/operators/cluster-setup/internal/constants"
	"github.com/kloudlite/operator/operators/cluster-setup/internal/env"
	"github.com/kloudlite/operator/operators/cluster-setup/internal/templates"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-playground/validator/v10"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	logger     logging.Logger
	Name       string
	yamlClient *kubectl.YAMLClient
	restConfig *rest.Config
	helmClient helm.Client
	Env        *env.Env

	TemplateWgOperator         []byte
	TemplateCsiOperator        []byte
	TemplateRouterOperator     []byte
	TemplateAppNLambdaOperator []byte
	TemplateMsvcNMresOperator  []byte
	TemplateMsvcRedisOperator  []byte
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	WgOperatorReady            string = "internal-operator-installed"
	DefaultsPatched            string = "defaults-patched"
	KloudliteCredsValidated    string = "kloudlite-creds-validated"
	UserKubeConfigCreated      string = "user-kubeconfig-created"
	CsiOperatorReady           string = "csi-operator-ready"
	RoutersOperatorReady       string = "routers-operator-ready"
	LokiReady                  string = "loki-ready"
	PrometheusReady            string = "prometheus-ready"
	Finalizing                 string = "finalizing"
	GitlabRunnerInstalled      string = "gitlab-runner-installed"
	CertManagerInstalled       string = "cert-manager-installed"
	AppOperatorInstalled       string = "app-operator-installed"
	MsvcNMresOperatorInstalled string = "msvc-n-mres-operator-installed"
	RedisOperatorInstalled      string = "redis-operator-installed"
)

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &v1.ManagedCluster{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	req.LogPreReconcile()
	defer req.LogPostReconcile()

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.checkKloudliteCreds(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.patchDefaults(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureWgOperator(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureUserKubeConfig(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureCsiDriversOperator(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureCertManager(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureRoutersOperator(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureAppOperator(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureMsvcNMresOperator(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureMsvcRedisOperator(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureGitlabRunner(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureLoki(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensurePrometheus(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	req.Object.Status.LastReconcileTime = &metav1.Time{Time: time.Now()}
	// return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
	if err := r.Status().Update(ctx, req.Object); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
}

func (r *Reconciler) finalize(req *rApi.Request[*v1.ManagedCluster]) stepResult.Result {
	return req.Next()
}

func (r *Reconciler) checkKloudliteCreds(req *rApi.Request[*v1.ManagedCluster]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(KloudliteCredsValidated)
	defer req.LogPostCheck(KloudliteCredsValidated)

	scrt, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.KloudliteCreds.Namespace, obj.Spec.KloudliteCreds.Name), &corev1.Secret{})
	if err != nil {
		return req.CheckFailed(KloudliteCredsValidated, check, err.Error()).Err(nil)
	}

	klCreds, err := fn.ParseFromSecret[v1.KloudliteCreds](scrt)
	if err != nil {
		return req.CheckFailed(KloudliteCredsValidated, check, err.Error()).Err(nil)
	}
	if err := validator.New().Struct(klCreds); err != nil {
		return req.CheckFailed(KloudliteCredsValidated, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != obj.Status.Checks[KloudliteCredsValidated] {
		obj.Status.Checks[KloudliteCredsValidated] = check
		req.UpdateStatus()
	}

	rApi.SetLocal(req, "kl-creds", klCreds)
	return req.Next()
}

func (r *Reconciler) patchDefaults(req *rApi.Request[*v1.ManagedCluster]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(DefaultsPatched)
	defer req.LogPostCheck(DefaultsPatched)

	hasUpdated := false

	if obj.Spec.Domain == nil {
		hasUpdated = true
		obj.Spec.Domain = fn.New(fmt.Sprintf("%s.clusters.kloudlite.io", obj.Name))
	}

	if hasUpdated {
		if err := r.Update(ctx, obj); err != nil {
			return req.CheckFailed(DefaultsPatched, check, err.Error())
		}

		return req.Done().RequeueAfter(100 * time.Millisecond)
	}

	check.Status = true
	if check != obj.Status.Checks[DefaultsPatched] {
		obj.Status.Checks[DefaultsPatched] = check
		req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) ensureWgOperator(req *rApi.Request[*v1.ManagedCluster]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(WgOperatorReady)
	defer req.LogPostCheck(WgOperatorReady)

	klCreds, ok := rApi.GetLocal[*v1.KloudliteCreds](req, "kl-creds")
	if !ok {
		return req.CheckFailed(KloudliteCredsValidated, check, errors.NotInLocals("kl-creds").Error()).Err(nil)
	}

	b, err := templates.Parse(templates.WgOperatorEnv, map[string]any{
		"namespace":           lc.NsOperators,
		"owner-refs":          []metav1.OwnerReference{fn.AsOwner(obj, true)},
		"wildcard-domain":     obj.Spec.Domain,
		"nameserver-endpoint": klCreds.DnsApiEndpoint,
		"nameserver-username": klCreds.DnsApiUsername,
		"nameserver-password": klCreds.DnsApiPassword,
	})

	if err != nil {
		return req.CheckFailed(WgOperatorReady, check, err.Error()).Err(nil)
	}

	if _, err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
		return req.CheckFailed(WgOperatorReady, check, err.Error()).Err(nil)
	}

	b4, err := templates.ParseBytes(r.TemplateWgOperator, map[string]any{
		"Namespace":       lc.NsOperators,
		"EnvName":         "development",
		"ImageTag":        "v1.0.5",
		"ImagePullPolicy": "Always",
		"SvcAccountName":  lc.ClusterSvcAccount,
	})

	if err != nil {
		return req.CheckFailed(WgOperatorReady, check, err.Error()).Err(nil)
	}

	if _, err := r.yamlClient.ApplyYAML(ctx, b4); err != nil {
		return req.CheckFailed(WgOperatorReady, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != obj.Status.Checks[WgOperatorReady] {
		obj.Status.Checks[WgOperatorReady] = check
		req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) ensureUserKubeConfig(req *rApi.Request[*v1.ManagedCluster]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	svcAccountName := obj.Name + "-admin"
	svcAccountNs := "kube-system"

	// 1. create service account for user
	// 2. create cluster role for this user
	// 3. create service account token for that user
	b, err := templates.Parse(templates.UserAccountRbac, map[string]any{
		"svc-account-name":      svcAccountName,
		"svc-account-namespace": svcAccountNs,
	})

	if err != nil {
		return req.CheckFailed(UserKubeConfigCreated, check, err.Error()).Err(nil)
	}

	if _, err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
		return req.CheckFailed(UserKubeConfigCreated, check, err.Error()).Err(nil)
	}

	// 4. then read that service account secret for `.data.token` field
	s, err := rApi.Get(ctx, r.Client, fn.NN(svcAccountNs, svcAccountName), &corev1.Secret{})
	if err != nil {
		return req.CheckFailed(UserKubeConfigCreated, check, err.Error()).Err(nil)
	}

	b64CaCrt := strings.TrimSpace(base64.StdEncoding.EncodeToString(s.Data["ca.crt"]))

	b, err = templates.Parse(templates.Kubeconfig, map[string]any{
		"ca-data":    b64CaCrt,
		"user-token": string(s.Data["token"]),

		"cluster-endpoint": fmt.Sprintf("https://k8s.%s:6443", *obj.Spec.Domain),
		"user-name":        svcAccountName,
	})

	if err != nil {
		return req.CheckFailed(UserKubeConfigCreated, check, err.Error()).Err(nil)
	}

	kubeConfig := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: svcAccountName + "-kubeconfig", Namespace: "kube-system"}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, kubeConfig, func() error {
		if kubeConfig.Data == nil {
			kubeConfig.Data = make(map[string][]byte, 1)
		}
		kubeConfig.Data["kubeconfig"] = b
		return nil
	}); err != nil {
		return req.CheckFailed(UserKubeConfigCreated, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != obj.Status.Checks[UserKubeConfigCreated] {
		obj.Status.Checks[UserKubeConfigCreated] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureCsiDriversOperator(req *rApi.Request[*v1.ManagedCluster]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(CsiOperatorReady)
	defer req.LogPostCheck(CsiOperatorReady)

	b, err := templates.ParseBytes(r.TemplateCsiOperator, map[string]any{
		"Namespace":       lc.NsOperators,
		"EnvName":         "development",
		"ImageTag":        "v1.0.5",
		"ImagePullPolicy": "Always",
		"SvcAccountName":  "kloudlite-cluster-svc-account",
	})

	if err != nil {
		return req.CheckFailed(CsiOperatorReady, check, err.Error()).Err(nil)
	}

	if _, err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
		return req.CheckFailed(CsiOperatorReady, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != obj.Status.Checks[CsiOperatorReady] {
		obj.Status.Checks[CsiOperatorReady] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureRoutersOperator(req *rApi.Request[*v1.ManagedCluster]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(RoutersOperatorReady)
	defer req.LogPostCheck(RoutersOperatorReady)

	routerOperatorEnv := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "router-operator-env", Namespace: obj.Spec.KlOperators.Namespace}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, routerOperatorEnv, func() error {
		if routerOperatorEnv.StringData == nil {
			routerOperatorEnv.StringData = map[string]string{}
		}
		routerOperatorEnv.StringData["ACME_EMAIL"] = obj.Spec.KlOperators.ACMEEmail
		routerOperatorEnv.StringData["DEFAULT_CLUSTER_ISSUER_NAME"] = obj.Spec.KlOperators.ClusterIssuerName
		return nil
	}); err != nil {
		return req.CheckFailed(RoutersOperatorReady, check, err.Error()).Err(nil)
	}

	b, err := templates.ParseBytes(r.TemplateRouterOperator, map[string]any{
		"Namespace":       obj.Spec.KlOperators.Namespace,
		"EnvName":         obj.Spec.KlOperators.InstallationMode,
		"ImageTag":        obj.Spec.KlOperators.ImageTag,
		"ImagePullPolicy": obj.Spec.KlOperators.ImagePullPolicy,
		"SvcAccountName":  obj.Spec.KlOperators.ClusterSvcAccount,
	})

	if err != nil {
		return req.CheckFailed(RoutersOperatorReady, check, err.Error()).Err(nil)
	}

	if _, err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
		return req.CheckFailed(RoutersOperatorReady, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != obj.Status.Checks[RoutersOperatorReady] {
		obj.Status.Checks[RoutersOperatorReady] = check
		req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) ensureAppOperator(req *rApi.Request[*v1.ManagedCluster]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(AppOperatorInstalled)
	defer req.LogPostCheck(AppOperatorInstalled)

	b, err := templates.ParseBytes(r.TemplateAppNLambdaOperator, map[string]any{
		"Namespace":       obj.Spec.KlOperators.Namespace,
		"EnvName":         obj.Spec.KlOperators.InstallationMode,
		"ImageTag":        obj.Spec.KlOperators.ImageTag,
		"ImagePullPolicy": obj.Spec.KlOperators.ImagePullPolicy,
		"SvcAccountName":  obj.Spec.KlOperators.ClusterSvcAccount,
	})

	if err != nil {
		return req.CheckFailed(AppOperatorInstalled, check, err.Error()).Err(nil)
	}

	if _, err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
		return req.CheckFailed(AppOperatorInstalled, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != obj.Status.Checks[AppOperatorInstalled] {
		obj.Status.Checks[AppOperatorInstalled] = check
		req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) ensureMsvcNMresOperator(req *rApi.Request[*v1.ManagedCluster]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(MsvcNMresOperatorInstalled)
	defer req.LogPostCheck(MsvcNMresOperatorInstalled)

	b, err := templates.ParseBytes(r.TemplateMsvcNMresOperator, map[string]any{
		"Namespace":       obj.Spec.KlOperators.Namespace,
		"EnvName":         obj.Spec.KlOperators.InstallationMode,
		"ImageTag":        obj.Spec.KlOperators.ImageTag,
		"ImagePullPolicy": obj.Spec.KlOperators.ImagePullPolicy,
		"SvcAccountName":  obj.Spec.KlOperators.ClusterSvcAccount,
	})

	if err != nil {
		return req.CheckFailed(MsvcNMresOperatorInstalled, check, err.Error()).Err(nil)
	}

	if _, err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
		return req.CheckFailed(MsvcNMresOperatorInstalled, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != obj.Status.Checks[MsvcNMresOperatorInstalled] {
		obj.Status.Checks[MsvcNMresOperatorInstalled] = check
		req.UpdateStatus()
	}

	return req.Next()
}
func (r *Reconciler) ensureMsvcRedisOperator(req *rApi.Request[*v1.ManagedCluster]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(RedisOperatorInstalled) 
	defer req.LogPostCheck(RedisOperatorInstalled) 

	b, err := templates.ParseBytes(r.TemplateMsvcRedisOperator, map[string]any{
		"Namespace":       obj.Spec.KlOperators.Namespace,
		"EnvName":         obj.Spec.KlOperators.InstallationMode,
		"ImageTag":        obj.Spec.KlOperators.ImageTag,
		"ImagePullPolicy": obj.Spec.KlOperators.ImagePullPolicy,
		"SvcAccountName":  obj.Spec.KlOperators.ClusterSvcAccount,
	})

	if err != nil {
		return req.CheckFailed(RedisOperatorInstalled, check, err.Error()).Err(nil)
	}

	if _, err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
		return req.CheckFailed(RedisOperatorInstalled, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != obj.Status.Checks[RedisOperatorInstalled] {
		obj.Status.Checks[RedisOperatorInstalled] = check
		req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) ensureGitlabRunner(req *rApi.Request[*v1.ManagedCluster]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	if obj.Spec.GitlabRunner == nil || !obj.Spec.GitlabRunner.Enabled {
		req.Logger.Infof("skipping gitlab runner reconcilation")
		return req.Next()
	}

	req.LogPreCheck(GitlabRunnerInstalled)
	defer req.LogPostCheck(GitlabRunnerInstalled)

	gitlabNs := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.GitlabRunner.Namespace}}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, gitlabNs, func() error {
		if !fn.IsOwner(gitlabNs, fn.AsOwner(obj)) {
			gitlabNs.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		}
		return nil
	}); err != nil {
		return req.CheckFailed(GitlabRunnerInstalled, check, err.Error()).Err(nil)
	}

	b, err := templates.Parse(templates.GitlabRunnerValues, map[string]any{"runner-token": obj.Spec.GitlabRunner.RunnerToken})
	if err != nil {
		return req.CheckFailed(GitlabRunnerInstalled, check, err.Error()).Err(nil)
	}

	if _, err := r.helmClient.EnsureRelease(ctx, helm.ChartSpec{
		ReleaseName: obj.Spec.GitlabRunner.ReleaseName,
		// ChartName:   "gitlab/gitlab-runner",
		ChartName:  fmt.Sprintf("%s/gitlab-runner-0.50.1.tgz", r.Env.HelmChartsDir),
		Namespace:  obj.Spec.GitlabRunner.Namespace,
		ValuesYaml: string(b),
	}); err != nil {
		return req.CheckFailed(GitlabRunnerInstalled, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != obj.Status.Checks[GitlabRunnerInstalled] {
		obj.Status.Checks[GitlabRunnerInstalled] = check
		req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) ensureCertManager(req *rApi.Request[*v1.ManagedCluster]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(CertManagerInstalled)
	defer req.LogPostCheck(CertManagerInstalled)

	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.CertManager.Namespace}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, ns, func() error {
		if ns.Labels == nil {
			ns.Labels = make(map[string]string, 1)
		}
		ns.Labels["kloudlite.io/created-by"] = obj.Name
		return nil
	}); err != nil {
		return req.CheckFailed(CertManagerInstalled, check, err.Error()).Err(nil)
	}

	b, err := templates.Parse(templates.HelmCertManagerValues, map[string]any{})
	if err != nil {
		return req.CheckFailed(CertManagerInstalled, check, err.Error()).Err(nil)
	}

	if _, err := r.helmClient.EnsureRelease(ctx, helm.ChartSpec{
		ReleaseName: obj.Spec.CertManager.ReleaseName,
		ChartName:   fmt.Sprintf("%s/cert-manager-v1.11.0.tgz", r.Env.HelmChartsDir),
		Namespace:   obj.Spec.CertManager.Namespace,
		ValuesYaml:  string(b),
	}); err != nil {
		r.logger.Error(err)
		return req.CheckFailed(CertManagerInstalled, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != obj.Status.Checks[CertManagerInstalled] {
		obj.Status.Checks[CertManagerInstalled] = check
		req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) ensureLoki(req *rApi.Request[*v1.ManagedCluster]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(LokiReady)
	defer req.LogPostCheck(LokiReady)

	if obj.Spec.Loki == nil || !obj.Spec.Loki.Enabled {
		req.Logger.Infof("skipping loki installation")
		return req.Next()
	}

	b, err := templates.Parse(templates.LokiValues, map[string]any{"loki-values": obj.Spec.Loki})
	if err != nil {
		return req.CheckFailed(LokiReady, check, err.Error()).Err(nil)
	}

	if _, err := r.helmClient.EnsureRelease(ctx, helm.ChartSpec{
		ReleaseName: obj.Spec.Loki.ReleaseName,
		// ChartName:   "grafana/loki-stack",
		ChartName:  fmt.Sprintf("%s/loki-stack-2.8.7.tgz", r.Env.HelmChartsDir),
		Namespace:  lc.NsCore,
		ValuesYaml: string(b),
	}); err != nil {
		return req.CheckFailed(LokiReady, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != obj.Status.Checks[LokiReady] {
		obj.Status.Checks[LokiReady] = check
		req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) ensurePrometheus(req *rApi.Request[*v1.ManagedCluster]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	if obj.Spec.Prometheus == nil || !obj.Spec.Prometheus.Enabled {
		req.Logger.Infof("skipping prometheus installation")
		return req.Next()
	}

	req.LogPreCheck(PrometheusReady)
	defer req.LogPostCheck(PrometheusReady)

	b, err := templates.Parse(templates.PrometheusValues, map[string]any{"prometheus-values": obj.Spec.Prometheus, "name": obj.Spec.Prometheus.ReleaseName})
	if err != nil {
		return req.CheckFailed(PrometheusReady, check, err.Error()).Err(nil)
	}

	if _, err := r.helmClient.EnsureRelease(ctx, helm.ChartSpec{
		ReleaseName: obj.Spec.Prometheus.ReleaseName,
		// ChartName:   "bitnami/kube-prometheus",
		ChartName:  fmt.Sprintf("%s/kube-prometheus-8.2.2.tgz", r.Env.HelmChartsDir),
		Namespace:  lc.NsCore,
		ValuesYaml: string(b),
	}); err != nil {
		return req.CheckFailed(PrometheusReady, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != obj.Status.Checks[PrometheusReady] {
		obj.Status.Checks[PrometheusReady] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) downloadOperatorTemplate(name string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/%s/operators/download/%s", r.Env.ReleasesApiEndpoint, r.Env.ReleaseVersion, name),
		nil,
	)

	req.SetBasicAuth(r.Env.ReleasesApiUsername, r.Env.ReleasesApiPassword)

	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("bad status code, expected %d got %d", 200, resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())
	r.restConfig = mgr.GetConfig()

	chartsDir, err := filepath.Abs(r.Env.HelmChartsDir)
	if err != nil {
		return err
	}
	r.Env.HelmChartsDir = chartsDir

	r.helmClient = helm.NewHelmClientOrDie(mgr.GetConfig(), helm.ClientOptions{Logger: r.logger})

	var g errgroup.Group

	r.logger.Infof("trying to download operator templates from releases-api")
	ts := time.Now()
	g.Go(func() error {
		r.TemplateCsiOperator, err = r.downloadOperatorTemplate("csi-drivers")
		return err
	})
	g.Go(func() error {
		r.TemplateRouterOperator, err = r.downloadOperatorTemplate("routers")
		return err
	})
	g.Go(func() error {
		r.TemplateWgOperator, err = r.downloadOperatorTemplate("wg-operator")
		return err
	})
	g.Go(func() error {
		r.TemplateAppNLambdaOperator, err = r.downloadOperatorTemplate("app-n-lambda")
		return err
	})
	g.Go(func() error {
		r.TemplateMsvcNMresOperator, err = r.downloadOperatorTemplate("msvc-n-mres")
		return err
	})
	g.Go(func() error {
		r.TemplateMsvcRedisOperator, err = r.downloadOperatorTemplate("msvc-redis")
		return err
	})
	g.Go(func() error {
		r.TemplateMsvcRedisOperator, err = r.downloadOperatorTemplate("msvc-redis")
		return err
	})

	if err := g.Wait(); err != nil {
		return err
	}
	r.logger.Infof("template downloading took %.2fs", time.Since(ts).Seconds())

	builder := ctrl.NewControllerManagedBy(mgr).For(&v1.ManagedCluster{})
	builder.WithOptions(controller.Options{MaxConcurrentReconciles: r.Env.MaxConcurrentReconciles})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
