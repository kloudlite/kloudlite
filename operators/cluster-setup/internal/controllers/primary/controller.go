package primary

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	certmanagerMetav1 "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"io"
	schedulingv1 "k8s.io/api/scheduling/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"net/http"
	ct "operators.kloudlite.io/apis/common-types"
	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	redpandaMsvcv1 "operators.kloudlite.io/apis/redpanda.msvc/v1"
	"os"
	"os/exec"
	"sigs.k8s.io/yaml"
	"time"

	"github.com/mittwald/go-helm-client"
	"helm.sh/helm/v3/pkg/repo"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"operators.kloudlite.io/apis/cluster-setup/v1"
	lc "operators.kloudlite.io/operators/cluster-setup/internal/constants"
	"operators.kloudlite.io/operators/cluster-setup/internal/env"
	"operators.kloudlite.io/operators/cluster-setup/internal/templates"
	"operators.kloudlite.io/pkg/constants"
	fn "operators.kloudlite.io/pkg/functions"
	"operators.kloudlite.io/pkg/harbor"
	kHttp "operators.kloudlite.io/pkg/http"
	"operators.kloudlite.io/pkg/kubectl"
	"operators.kloudlite.io/pkg/logging"
	rApi "operators.kloudlite.io/pkg/operator"
	stepResult "operators.kloudlite.io/pkg/operator/step-result"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	harborCli  *harbor.Client
	logger     logging.Logger
	Name       string
	yamlClient *kubectl.YAMLClient
	restConfig *rest.Config
	Env        *env.Env
}

func (r *Reconciler) GetName() string {
	return r.Name
}

func newHelmClient(config *rest.Config, namespace string) (helmclient.Client, error) {
	return helmclient.NewClientFromRestConf(&helmclient.RestConfClientOptions{
		Options: &helmclient.Options{
			Namespace: namespace,
		},
		RestConfig: config,
	})
}

func areHelmValuesEqual(releaseValues map[string]any, templateValues []byte) bool {
	b, err := json.Marshal(releaseValues)
	if err != nil {
		return false
	}

	tv, err := yaml.YAMLToJSON(templateValues)
	if err != nil {
		return false
	}

	if len(b) != len(tv) || bytes.Compare(b, tv) != 0 {
		return false
	}
	return true
}

const (
	NamespacesReady       string = "namespaces-ready"
	SvcAccountsReady      string = "svc-accounts-ready"
	LokiReady             string = "loki-ready"
	GrafanaReady          string = "grafana-ready"
	RedpandaReady         string = "redpanda-ready"
	PrometheusReady       string = "prometheus-ready"
	CertManagerReady      string = "cert-manager-ready"
	CertIssuerReady       string = "cert-issuer-ready"
	IngressReady          string = "ingress-ready"
	OperatorCRDsReady     string = "operator-crds-ready"
	MsvcAndMresReady      string = "msvc-and-mres-ready"
	OperatorsEnvReady     string = "operator-env-ready"
	KloudliteAPIsReady    string = "kloudlite-apis-ready"
	KloudliteWebReady     string = "kloudlite-web-ready"
	DefaultsPatched       string = "defaults-patched"
	HarborAdminCredsReady string = "harbor-admin-creds"
	RedpandaTopicsCreated string = "redpanda-topic-created"
)

var (
	namespacesList = []string{lc.NsCore, lc.NsRedpanda, lc.NsCertManager, lc.NsMonitoring, lc.NsIngress, lc.NsOperators}
)

type githubRelease struct {
	Assets []struct {
		Id   int64  `json:"id"`
		Name string `json:"name"`
	} `json:"assets"`
}

const ImageRegistryHost = "registry.kloudlite.io"

// +kubebuilder:rbac:groups=cluster-setup.kloudlite.io,resources=primaryclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster-setup.kloudlite.io,resources=primaryclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cluster-setup.kloudlite.io,resources=primaryclusters/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(context.WithValue(ctx, "logger", r.logger), r.Client, request.NamespacedName, &v1.PrimaryCluster{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}
		return ctrl.Result{}, nil
	}

	req.Logger.Infof("NEW RECONCILATION")
	defer func() {
		req.Logger.Infof("RECONCILATION COMPLETE (isReady=%v)", req.Object.Status.IsReady)
	}()

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.RestartIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureChecks(NamespacesReady, DefaultsPatched, SvcAccountsReady, LokiReady, GrafanaReady, PrometheusReady, CertManagerReady, CertIssuerReady, IngressReady, OperatorCRDsReady, MsvcAndMresReady, OperatorsEnvReady, KloudliteAPIsReady, KloudliteWebReady); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.patchDefaults(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureNamespaces(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureSvcAccounts(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureHarborAdminCreds(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureLoki(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensurePrometheus(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureGrafana(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureCertManager(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureRedpanda(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureCertIssuer(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureIngressNginx(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureOperatorCRDs(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureOperators(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureRedpandaTopics(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureMsvcAndMres(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureKloudliteApis(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.ensureKloudliteWebs(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if req.Object.Spec.ShouldInstallSecondary {
		if step := r.installKloudliteAgent(req); !step.ShouldProceed() {
			return step.ReconcilerResponse()
		}
	}

	req.Object.Status.IsReady = true
	req.Object.Status.LastReconcileTime = metav1.Time{Time: time.Now()}
	return ctrl.Result{RequeueAfter: r.Env.ReconcilePeriod}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) patchDefaults(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(DefaultsPatched)
	defer req.LogPostCheck(DefaultsPatched)

	sharedC := v1.SharedConstants{
		SubDomain: obj.Spec.Domain,

		// mongo
		MongoSvcName:  "mongo-svc",
		AuthDbName:    "auth-db",
		ConsoleDbName: "console-db",
		CiDbName:      "ci-db",
		DnsDbName:     "dns-db",
		FinanceDbName: "finance-db",
		IamDbName:     "iam-db",
		CommsDbName:   "comms-db",
		EventsDbName:  "events-db",

		// redis
		RedisSvcName:     "redis-svc",
		AuthRedisName:    "auth-redis",
		ConsoleRedisName: "console-redis",
		CiRedisName:      "ci-redis",
		DnsRedisName:     "dns-redis",
		IamRedisName:     "iam-redis",
		SocketRedisName:  "socket-redis",
		FinanceRedisName: "finance-redis",

		// Apps api
		AppAuthApi:       "auth-api",
		AppConsoleApi:    "console-api",
		AppCiApi:         "ci-api",
		AppFinanceApi:    "finance-api",
		AppCommsApi:      "comms-api",
		AppDnsApi:        "dns-api",
		AppIAMApi:        "iam-api",
		AppJsEvalApi:     "js-eval-api",
		AppGqlGatewayApi: "gateway",
		AppWebhooksApi:   "webhooks",

		// Apps Agent
		AppKlAgent: "kl-agent",

		// Apps web
		AppAuthWeb:     "auth-web",
		AppAccountsWeb: "accounts-web",
		AppConsoleWeb:  "console-web",
		AppSocketWeb:   "socket-web",

		CookieDomain: obj.Spec.Domain,

		// Images
		ImageAuthApi:       fmt.Sprintf("%s/kloudlite/production/auth:v1.0.4", ImageRegistryHost),
		ImageConsoleApi:    fmt.Sprintf("%s/kloudlite/production/console:v1.0.4", ImageRegistryHost),
		ImageCiApi:         fmt.Sprintf("%s/kloudlite/production/ci:v1.0.4", ImageRegistryHost),
		ImageFinanceApi:    fmt.Sprintf("%s/kloudlite/production/finance:v1.0.4", ImageRegistryHost),
		ImageCommsApi:      fmt.Sprintf("%s/kloudlite/production/comms:v1.0.4", ImageRegistryHost),
		ImageDnsApi:        fmt.Sprintf("%s/kloudlite/production/dns:v1.0.4", ImageRegistryHost),
		ImageIAMApi:        fmt.Sprintf("%s/kloudlite/production/iam:v1.0.4", ImageRegistryHost),
		ImageJsEvalApi:     fmt.Sprintf("%s/kloudlite/production/js-eval:v1.0.4", ImageRegistryHost),
		ImageGqlGatewayApi: fmt.Sprintf("%s/kloudlite/production/gateway:v1.0.4", ImageRegistryHost),
		ImageWebhooksApi:   fmt.Sprintf("%s/kloudlite/production/webhooks:v1.0.4", ImageRegistryHost),
		ImageKlAgent:       fmt.Sprintf("%s/kloudlite/production/kl-agent:v1.0.4", ImageRegistryHost),

		// Images Web
		ImageAuthWeb:     fmt.Sprintf("%s/kloudlite/production/web-auth:v1.0.3", ImageRegistryHost),
		ImageAccountsWeb: fmt.Sprintf("%s/kloudlite/production/web-accounts:v1.0.3", ImageRegistryHost),
		ImageConsoleWeb:  fmt.Sprintf("%s/kloudlite/production/web-console:v1.0.3", ImageRegistryHost),
		ImageSocketWeb:   fmt.Sprintf("%s/kloudlite/production/web-socket:v1.0.3", ImageRegistryHost),

		// constants
		RedpandaAdminSecretName:    "msvc-redpanda-admin-creds",
		OAuthSecretName:            obj.Spec.OAuthCreds.Name,
		HarborAdminCredsSecretName: "harbor-admin-creds",
		StripeSecretName:           obj.Spec.StripeCreds.Name,

		// KafkaTopics
		KafkaTopicGitWebhooks:        "kl-git-webhooks",
		KafkaTopicHarborWebhooks:     "kl-harbor-webhooks",
		KafkaTopicPipelineRunUpdates: "kl-pipeline-run-updates",
		KafkaTopicsStatusUpdates:     "kl-status-updates",
		KafkaTopicBillingUpdates:     "kl-billing-updates",
		KafkaTopicEvents:             "kl-events",

		StatefulPriorityClass:  "stateful",
		WebhookAuthzSecretName: obj.Spec.WebhookAuthzCreds.Name,

		// Routers
		AuthWebDomain:     fmt.Sprintf("auth.%s", obj.Spec.Domain),
		ConsoleWebDomain:  fmt.Sprintf("console.%s", obj.Spec.Domain),
		AccountsWebDomain: fmt.Sprintf("accounts.%s", obj.Spec.Domain),
		SocketWebDomain:   fmt.Sprintf("socket.%s", obj.Spec.Domain),
		WebhookApiDomain:  fmt.Sprintf("webhook-api.%s", obj.Spec.Domain),
	}

	hasUpdated := false

	if obj.Spec.SharedConstants == nil || *obj.Spec.SharedConstants != sharedC {
		hasUpdated = true
		obj.Spec.SharedConstants = &sharedC
	}

	// redpanda
	if obj.Spec.RedpandaValues.Resources.Storage == nil {
		hasUpdated = true
		obj.Spec.RedpandaValues.Resources.Storage = &ct.Storage{
			Size:         "1Gi",
			StorageClass: obj.Spec.StorageClass,
		}
	}

	if obj.Spec.RedpandaValues.Resources.Storage.StorageClass == "" {
		hasUpdated = true
		obj.Spec.RedpandaValues.Resources.Storage.StorageClass = obj.Spec.StorageClass
	}

	if hasUpdated {
		if err := r.Update(ctx, obj); err != nil {
			return req.CheckFailed(DefaultsPatched, check, err.Error()).Err(nil)
		}
		return req.Done().RequeueAfter(1 * time.Second)
	}

	check.Status = true
	if check != checks[DefaultsPatched] {
		checks[DefaultsPatched] = check
		return req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) finalize(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	return req.Finalize()
}

func (r *Reconciler) ensureNamespaces(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(NamespacesReady)
	defer req.LogPostCheck(NamespacesReady)

	for i := range namespacesList {
		ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{
			Name:            namespacesList[i],
			OwnerReferences: []metav1.OwnerReference{fn.AsOwner(obj, true)},
		}}
		if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, ns, func() error {
			if ns.Labels == nil {
				ns.Labels = make(map[string]string, 1)
			}
			ns.Labels["kloudlite.io/cluster-installation"] = obj.Name
			return nil
		}); err != nil {
			return req.CheckFailed(NamespacesReady, check, err.Error()).Err(nil)
		}
	}

	pc := &schedulingv1.PriorityClass{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.SharedConstants.StatefulPriorityClass}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, pc, func() error {
		pc.PreemptionPolicy = fn.New(corev1.PreemptLowerPriority)
		pc.Value = 1000000
		return nil
	}); err != nil {
		return req.CheckFailed(NamespacesReady, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != checks[NamespacesReady] {
		checks[NamespacesReady] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureHarborAdminCreds(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	harborCreds, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.HarborAdminCreds.Namespace, obj.Spec.HarborAdminCreds.Name), &corev1.Secret{})
	if err != nil {
		return req.CheckFailed(HarborAdminCredsReady, check, err.Error()).Err(nil)
	}

	for _, ns := range []string{lc.NsOperators, lc.NsCore} {
		localHarborCreds := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.HarborAdminCreds.Name, Namespace: ns}}
		if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, localHarborCreds, func() error {
			if fn.IsOwner(localHarborCreds, fn.AsOwner(obj)) {
				localHarborCreds.OwnerReferences = append(localHarborCreds.OwnerReferences, fn.AsOwner(obj, true))
			}
			localHarborCreds.Labels = harborCreds.Labels
			localHarborCreds.Data = harborCreds.Data
			return nil
		}); err != nil {
			return req.CheckFailed(HarborAdminCredsReady, check, err.Error()).Err(nil)
		}
	}

	check.Status = true
	if check != checks[HarborAdminCredsReady] {
		checks[HarborAdminCredsReady] = check
		req.UpdateStatus()
	}
	return req.Next()

}

func (r *Reconciler) ensureSvcAccounts(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(SvcAccountsReady)
	defer req.LogPostCheck(SvcAccountsReady)

	for _, ps := range obj.Spec.ImgPullSecrets {
		pullScrt, err := rApi.Get(ctx, r.Client, fn.NN(ps.Namespace, ps.Name), &corev1.Secret{})
		if err != nil {
			return req.CheckFailed(SvcAccountsReady, check, err.Error()).Err(nil)
		}

		for _, ns := range namespacesList {
			newPullSecret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: ps.Name, Namespace: ns}, Type: "kubernetes.io/dockerconfigjson"}
			if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, newPullSecret, func() error {
				newPullSecret.Data = pullScrt.Data
				return nil
			}); err != nil {
				return req.CheckFailed(SvcAccountsReady, check, err.Error()).Err(nil)
			}

			normalSvcAccount := &corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: lc.DefaultSvcAccount, Namespace: ns}}
			if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, normalSvcAccount, func() error {
				normalSvcAccount.OwnerReferences = []metav1.OwnerReference{fn.AsOwner(obj, true)}
				if !fn.ContainsAll(normalSvcAccount.ImagePullSecrets, []corev1.LocalObjectReference{{Name: newPullSecret.Name}}) {
					normalSvcAccount.ImagePullSecrets = append(normalSvcAccount.ImagePullSecrets, corev1.LocalObjectReference{Name: newPullSecret.Name})
				}
				return nil
			}); err != nil {
				return req.CheckFailed(SvcAccountsReady, check, err.Error()).Err(nil)
			}

			clusterSvcAccount := &corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: lc.ClusterSvcAccount, Namespace: ns}}

			if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, clusterSvcAccount, func() error {
				clusterSvcAccount.OwnerReferences = []metav1.OwnerReference{fn.AsOwner(obj, true)}
				if !fn.ContainsAll(clusterSvcAccount.ImagePullSecrets, []corev1.LocalObjectReference{{Name: newPullSecret.Name}}) {
					clusterSvcAccount.ImagePullSecrets = append(clusterSvcAccount.ImagePullSecrets, corev1.LocalObjectReference{Name: newPullSecret.Name})
				}
				return nil
			}); err != nil {
				return req.CheckFailed(SvcAccountsReady, check, err.Error()).Err(nil)
			}
		}
	}

	// cluster role binding
	clusterRb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: lc.ClusterSvcAccount + "-rb",
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, clusterRb, func() error {
		subjects := make(map[string]bool, len(clusterRb.Subjects))
		for i := range clusterRb.Subjects {
			subjects[clusterRb.Subjects[i].Namespace] = true
		}

		for i := range namespacesList {
			if !subjects[namespacesList[i]] {
				clusterRb.Subjects = append(clusterRb.Subjects, rbacv1.Subject{
					Kind:      "ServiceAccount",
					APIGroup:  "",
					Name:      lc.ClusterSvcAccount,
					Namespace: namespacesList[i],
				})
			}

			clusterRb.RoleRef = rbacv1.RoleRef{
				APIGroup: "",
				Kind:     "ClusterRole",
				Name:     "cluster-admin",
			}
		}
		return nil
	}); err != nil {
		return req.CheckFailed(SvcAccountsReady, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != checks[SvcAccountsReady] {
		checks[SvcAccountsReady] = check
		req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) ensureLoki(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(LokiReady)
	defer req.LogPostCheck(LokiReady)

	const releaseName = "loki"

	helmCli, err := newHelmClient(r.restConfig, lc.NsMonitoring)
	if err != nil {
		return req.CheckFailed(LokiReady, check, err.Error()).Err(nil)
	}

	helmValues, err := helmCli.GetReleaseValues(releaseName, false)
	if err != nil {
		req.Logger.Infof("helm release (%s) not found, will be creating it", releaseName)
	}

	b, err := templates.Parse(templates.LokiValues, map[string]any{"loki-values": obj.Spec.LokiValues})
	if err != nil {
		return req.CheckFailed(LokiReady, check, err.Error()).Err(nil)
	}

	if !areHelmValuesEqual(helmValues, b) {
		if err := helmCli.AddOrUpdateChartRepo(repo.Entry{
			Name: "grafana",
			URL:  "https://grafana.github.io/helm-charts",
		}); err != nil {
			return req.CheckFailed(LokiReady, check, err.Error()).Err(nil)
		}

		if _, err := helmCli.InstallOrUpgradeChart(ctx, &helmclient.ChartSpec{
			ReleaseName: releaseName,
			ChartName:   "grafana/loki-stack",
			Namespace:   lc.NsMonitoring,
			ValuesYaml:  string(b),
		}, &helmclient.GenericHelmOptions{}); err != nil {
			return req.CheckFailed(LokiReady, check, err.Error()).Err(nil)
		}
	}

	check.Status = true
	if check != checks[LokiReady] {
		checks[LokiReady] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensurePrometheus(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(PrometheusReady)
	defer req.LogPostCheck(PrometheusReady)

	const releaseName = "prometheus"

	helmCli, err := newHelmClient(r.restConfig, lc.NsMonitoring)
	if err != nil {
		return req.CheckFailed(PrometheusReady, check, err.Error()).Err(nil)
	}

	helmValues, err := helmCli.GetReleaseValues(releaseName, false)
	if err != nil {
		req.Logger.Infof("helm release (%s) not found, will be creating it", releaseName)
	}

	b, err := templates.Parse(templates.PrometheusValues, map[string]any{"prometheus-values": obj.Spec.PrometheusValues, "name": releaseName})
	if err != nil {
		return req.CheckFailed(PrometheusReady, check, err.Error()).Err(nil)
	}

	if !areHelmValuesEqual(helmValues, b) {
		if err := helmCli.AddOrUpdateChartRepo(repo.Entry{
			Name: "bitnami",
			URL:  "https://charts.bitnami.com/bitnami",
		}); err != nil {
			return req.CheckFailed(PrometheusReady, check, err.Error()).Err(nil)
		}

		if _, err := helmCli.InstallOrUpgradeChart(ctx, &helmclient.ChartSpec{
			ReleaseName: releaseName,
			ChartName:   "bitnami/kube-prometheus",
			Namespace:   lc.NsMonitoring,
			ValuesYaml:  string(b),
		}, &helmclient.GenericHelmOptions{}); err != nil {
			req.Logger.Error(err)
			return req.CheckFailed(PrometheusReady, check, err.Error()).Err(nil)
		}
	}

	check.Status = true
	if check != checks[PrometheusReady] {
		checks[PrometheusReady] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureGrafana(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(GrafanaReady)
	defer req.LogPostCheck(GrafanaReady)

	const releaseName = "grafana"

	helmCli, err := newHelmClient(r.restConfig, lc.NsMonitoring)
	if err != nil {
		return req.CheckFailed(GrafanaReady, check, err.Error()).Err(nil)
	}

	helmValues, err := helmCli.GetReleaseValues(releaseName, false)
	if err != nil {
		req.Logger.Infof("helm release (%s) not found, will be creating it", releaseName)
	}

	b, err := templates.Parse(templates.PrometheusValues, map[string]any{"prometheus-values": obj.Spec.PrometheusValues, "name": releaseName})
	if err != nil {
		return req.CheckFailed(GrafanaReady, check, err.Error()).Err(nil)
	}

	if !areHelmValuesEqual(helmValues, b) {
		if err := helmCli.AddOrUpdateChartRepo(repo.Entry{
			Name: "bitnami",
			URL:  "https://charts.bitnami.com/bitnami",
		}); err != nil {
			return req.CheckFailed(GrafanaReady, check, err.Error()).Err(nil)
		}

		if _, err := helmCli.InstallOrUpgradeChart(ctx, &helmclient.ChartSpec{
			ReleaseName: releaseName,
			ChartName:   "bitnami/grafana",
			Namespace:   lc.NsMonitoring,
			ValuesYaml:  string(b),
		}, &helmclient.GenericHelmOptions{}); err != nil {
			return nil
		}
	}

	check.Status = true
	if check != checks[GrafanaReady] {
		checks[GrafanaReady] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureRedpanda(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(GrafanaReady)
	defer req.LogPostCheck(GrafanaReady)

	const releaseName = "redpanda"

	helmCli, err := newHelmClient(r.restConfig, lc.NsRedpanda)
	if err != nil {
		return req.CheckFailed(RedpandaReady, check, err.Error()).Err(nil)
	}

	release, err := helmCli.GetRelease(releaseName)
	if err != nil {
		release = nil
	}

	if release != nil && release.Info != nil && release.Info.Status == "uninstalled" {
		if err := helmCli.UninstallReleaseByName(releaseName); err != nil {
			return req.CheckFailed(RedpandaReady, check, err.Error()).Err(nil)
		}
		return req.Done().RequeueAfter(1 * time.Second)
	}

	helmValues, err := helmCli.GetReleaseValues(releaseName, false)
	if err != nil {
		req.Logger.Infof("helm release (%s) not found, will be creating it", releaseName)
	}

	b, err := templates.Parse(templates.RedpandaValues, map[string]any{})
	if err != nil {
		return req.CheckFailed(RedpandaReady, check, err.Error()).Err(nil)
	}

	if !areHelmValuesEqual(helmValues, b) {
		// installing Redpanda CRDs
		cmd := exec.Command("kubectl", "apply", "-k", fmt.Sprintf("https://github.com/redpanda-data/redpanda/src/go/k8s/config/crd?ref=%s", obj.Spec.RedpandaValues.Version))
		cmd.Stdout = os.Stdout
		if err := cmd.Run(); err != nil {
			return req.CheckFailed(RedpandaReady, check, err.Error()).Err(nil)
		}

		if err := helmCli.AddOrUpdateChartRepo(repo.Entry{
			Name: "redpanda",
			URL:  "https://charts.vectorized.io",
		}); err != nil {
			return req.CheckFailed(RedpandaReady, check, err.Error()).Err(nil)
		}

		if _, err := helmCli.InstallOrUpgradeChart(ctx, &helmclient.ChartSpec{
			ReleaseName: releaseName,
			ChartName:   "redpanda/redpanda-operator",
			Namespace:   lc.NsRedpanda,
			ValuesYaml:  string(b),
			Version:     "v22.1.6",
		}, &helmclient.GenericHelmOptions{}); err != nil {
			return nil
		}
	}

	oneNodeCluster, err := rApi.Get(ctx, r.Client, fn.NN(lc.NsRedpanda, releaseName), fn.NewUnstructured(metav1.TypeMeta{APIVersion: "redpanda.vectorized.io/v1alpha1", Kind: "Cluster"}))
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.CheckFailed(RedpandaReady, check, err.Error())
		}
		oneNodeCluster = nil
		req.Logger.Infof("creating single node cluster for redpanda")
	}

	if oneNodeCluster == nil {
		clusterBytes, err := templates.Parse(templates.RedpandaSingleNodeCluster, map[string]any{
			"namespace":     lc.NsRedpanda,
			"cluster-name":  releaseName,
			"storage-size":  obj.Spec.RedpandaValues.Resources.Storage.Size,
			"storage-class": obj.Spec.RedpandaValues.Resources.Storage.StorageClass,
			"node-selector": obj.Spec.NodeSelector,
			"tolerations":   obj.Spec.Tolerations,
		})
		if err != nil {
			return req.CheckFailed(RedpandaReady, check, err.Error()).Err(nil)
		}

		if err := r.yamlClient.ApplyYAML(ctx, clusterBytes); err != nil {
			return req.CheckFailed(RedpandaReady, check, err.Error()).Err(nil)
		}
	}

	// RPK admin
	redpandaAdmin := &redpandaMsvcv1.Admin{ObjectMeta: metav1.ObjectMeta{Name: "admin", Namespace: lc.NsCore}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, redpandaAdmin, func() error {
		redpandaAdmin.Spec.AdminEndpoint = fmt.Sprintf("%s.%s.svc.cluster.local:9644", releaseName, lc.NsRedpanda)
		redpandaAdmin.Spec.KafkaBrokers = fmt.Sprintf("%s.%s.svc.cluster.local:9092", releaseName, lc.NsRedpanda)
		redpandaAdmin.Spec.Output = &ct.Output{
			SecretRef: &ct.SecretRef{
				Name:      obj.Spec.SharedConstants.RedpandaAdminSecretName,
				Namespace: lc.NsCore,
			},
		}
		return nil
	}); err != nil {
		return req.CheckFailed(RedpandaReady, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != checks[RedpandaReady] {
		checks[RedpandaReady] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureCertManager(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(CertManagerReady)
	defer req.LogPostCheck(CertManagerReady)

	const releaseName = "cert-manager"

	helmCli, err := newHelmClient(r.restConfig, lc.NsCertManager)
	if err != nil {
		return req.CheckFailed(CertManagerReady, check, err.Error()).Err(nil)
	}

	helmValues, err := helmCli.GetReleaseValues(releaseName, false)
	if err != nil {
		req.Logger.Infof("helm release (%s) not found, will be creating it", releaseName)
	}

	b, err := templates.Parse(templates.CertManagerValues, map[string]any{"cert-manager-values": obj.Spec.CertManagerValues})
	if err != nil {
		return req.CheckFailed(CertManagerReady, check, err.Error()).Err(nil)
	}

	if !areHelmValuesEqual(helmValues, b) {
		if err := helmCli.AddOrUpdateChartRepo(repo.Entry{
			Name: "jetstack",
			URL:  "https://charts.jetstack.io",
		}); err != nil {
			return req.CheckFailed(CertManagerReady, check, err.Error()).Err(nil)
		}

		if _, err := helmCli.InstallOrUpgradeChart(ctx, &helmclient.ChartSpec{
			ReleaseName: releaseName,
			ChartName:   "jetstack/cert-manager",
			Namespace:   lc.NsCertManager,
			ValuesYaml:  string(b),
		}, &helmclient.GenericHelmOptions{}); err != nil {
			return req.CheckFailed(CertManagerReady, check, err.Error()).Err(nil)
		}
	}

	check.Status = true
	if check != checks[CertManagerReady] {
		checks[CertManagerReady] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureIngressNginx(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(IngressReady)
	defer req.LogPostCheck(IngressReady)

	const releaseName = "ingress-nginx"

	helmCli, err := newHelmClient(r.restConfig, lc.NsIngress)
	if err != nil {
		return req.CheckFailed(IngressReady, check, err.Error()).Err(nil)
	}

	helmValues, err := helmCli.GetReleaseValues(releaseName, false)
	if err != nil {
		req.Logger.Infof("helm release (%s) not found, will be creating it", releaseName)
	}

	b, err := templates.Parse(templates.IngressNginxValues, map[string]any{
		"ingress-values":          obj.Spec.IngressValues,
		"wildcard-cert-namespace": lc.NsCertManager,
		"wildcard-cert-name":      obj.Spec.CertManagerValues.ClusterIssuer.Name + "-tls",
	})
	if err != nil {
		return req.CheckFailed(IngressReady, check, err.Error()).Err(nil)
	}

	if !areHelmValuesEqual(helmValues, b) {
		if err := helmCli.AddOrUpdateChartRepo(repo.Entry{
			Name: "ingress-nginx",
			URL:  "https://kubernetes.github.io/ingress-nginx",
		}); err != nil {
			return req.CheckFailed(IngressReady, check, err.Error()).Err(nil)
		}

		if _, err := helmCli.InstallOrUpgradeChart(ctx, &helmclient.ChartSpec{
			ReleaseName: releaseName,
			ChartName:   "ingress-nginx/ingress-nginx",
			Namespace:   lc.NsIngress,
			ValuesYaml:  string(b),
		}, &helmclient.GenericHelmOptions{}); err != nil {
			return req.CheckFailed(IngressReady, check, err.Error()).Err(nil)
		}
	}

	check.Status = true
	if check != checks[IngressReady] {
		checks[IngressReady] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureCertIssuer(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(CertIssuerReady)
	defer req.LogPostCheck(CertIssuerReady)

	// issuer secret ref
	cloudflareScrt, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.CertManagerValues.ClusterIssuer.Cloudflare.SecretKeyRef.Namespace, obj.Spec.CertManagerValues.ClusterIssuer.Cloudflare.SecretKeyRef.Name), &corev1.Secret{})
	if err != nil {
		return req.CheckFailed(CertIssuerReady, check, err.Error()).Err(nil)
	}

	localcfScrt := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Namespace: lc.NsCertManager, Name: obj.Spec.CertManagerValues.ClusterIssuer.Cloudflare.SecretKeyRef.Name}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, localcfScrt, func() error {
		if !fn.IsOwner(localcfScrt, fn.AsOwner(obj)) {
			localcfScrt.SetOwnerReferences(append(localcfScrt.GetOwnerReferences(), fn.AsOwner(obj)))
		}
		localcfScrt.Labels = cloudflareScrt.Labels
		localcfScrt.Data = cloudflareScrt.Data
		return nil
	}); err != nil {
		return req.CheckFailed(CertIssuerReady, check, err.Error()).Err(nil)
	}

	// issuer
	issuerVals := map[string]any{
		"cluster-issuer": obj.Spec.CertManagerValues.ClusterIssuer,
	}

	//dnsNames := make([]string, 0, (len(obj.Spec.CloudflareCreds.DnsNames)+1)*2)
	dnsNames := make([]string, 0, len(obj.Spec.CloudflareCreds.DnsNames)+1)
	for i := range obj.Spec.CloudflareCreds.DnsNames {
		dnsNames = append(dnsNames, obj.Spec.CloudflareCreds.DnsNames[i])
		//dnsNames = append(dnsNames, fmt.Sprintf("www.%s", obj.Spec.CloudflareCreds.DnsNames[i]))
	}

	if obj.Spec.ShouldInstallSecondary {
		dnsNames = append(dnsNames, fmt.Sprintf("*.%s.clusters.%s", obj.Spec.SecondaryClusterId, obj.Spec.Domain))
		//dnsNames = append(dnsNames, fmt.Sprintf("www.*.%s.clusters.%s", obj.Spec.SecondaryClusterId, obj.Spec.Domain))
	}

	issuerVals["dns-names"] = dnsNames
	b, err := templates.Parse(templates.ClusterIssuer, issuerVals)
	if err != nil {
		return req.CheckFailed(CertIssuerReady, check, err.Error()).Err(nil)
	}

	if err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
		return req.CheckFailed(CertIssuerReady, check, err.Error()).Err(nil)
	}

	// creating wildcard certificates for specified cloudflare domains
	wcert := &certmanagerv1.Certificate{ObjectMeta: metav1.ObjectMeta{
		Name:      obj.Spec.CertManagerValues.ClusterIssuer.Name,
		Namespace: lc.NsCertManager,
	}}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, wcert, func() error {
		wcert.Spec.DNSNames = dnsNames

		wcert.Spec.SecretName = obj.Spec.CertManagerValues.ClusterIssuer.Name + "-tls"
		wcert.Spec.IssuerRef = certmanagerMetav1.ObjectReference{
			Kind: "ClusterIssuer",
			Name: obj.Spec.CertManagerValues.ClusterIssuer.Name,
		}
		return nil
	}); err != nil {
		return req.CheckFailed(CertIssuerReady, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != checks[CertIssuerReady] {
		checks[CertIssuerReady] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureOperatorCRDs(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(OperatorCRDsReady)
	defer req.LogPostCheck(OperatorCRDsReady)

	if check.Generation > checks[OperatorCRDsReady].Generation || !checks[OperatorCRDsReady].Status {
		for _, ghSource := range obj.Spec.Operators.Manifests {
			artifactsMap := make(map[string]bool, len(ghSource.Artifacts))
			for _, a := range ghSource.Artifacts {
				artifactsMap[a] = true
			}

			artifactIds := make([]int64, 0, len(artifactsMap))

			ghTokenScrt, err := rApi.Get(ctx, r.Client, fn.NN(ghSource.TokenSecret.Namespace, ghSource.TokenSecret.Name), &corev1.Secret{})
			if err != nil {
				return req.CheckFailed(OperatorCRDsReady, check, err.Error()).Err(nil)
			}

			ghToken := string(ghTokenScrt.Data[ghSource.TokenSecret.Key])

			httpReq, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://api.github.com/repos/%s/releases/tags/%s", ghSource.Repo, ghSource.Tag), nil)
			httpReq.Header.Set("Authorization", fmt.Sprintf("token %s", ghToken))
			if err != nil {
				return req.CheckFailed(OperatorCRDsReady, check, err.Error()).Err(nil)
			}

			ghRelease, _, err := kHttp.Get[githubRelease](httpReq)
			if err != nil {
				return req.CheckFailed(OperatorsEnvReady, check, err.Error()).Err(nil)
			}

			if ghRelease == nil {
				return req.CheckFailed(OperatorCRDsReady, check, "github release not found").Err(nil)
			}

			for i := range ghRelease.Assets {
				if artifactsMap[ghRelease.Assets[i].Name] {
					artifactIds = append(artifactIds, ghRelease.Assets[i].Id)
				}
			}

			for i := range artifactIds {
				dReq, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://api.github.com/repos/%s/releases/assets/%d", ghSource.Repo, artifactIds[i]), nil)
				dReq.Header.Set("Authorization", fmt.Sprintf("token %s", ghToken))
				if err != nil {
					return req.CheckFailed(OperatorCRDsReady, check, err.Error())
				}
				dReq.Header.Set("Accept", "application/octet-stream")
				resp, err := http.DefaultClient.Do(dReq)
				if err != nil {
					return req.CheckFailed(OperatorCRDsReady, check, err.Error())
				}
				output, err := io.ReadAll(resp.Body)
				if err != nil {
					return req.CheckFailed(OperatorCRDsReady, check, err.Error())
				}

				b, err := templates.ParseBytes(output, map[string]any{
					"Namespace":       lc.NsOperators,
					"SvcAccountName":  lc.ClusterSvcAccount,
					"ImagePullPolicy": "Always",
					"EnvName":         "production",
					"ImageTag":        "v1.0.5",
				})

				if err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
					return req.CheckFailed(OperatorCRDsReady, check, err.Error())
				}
			}
		}
	}

	check.Status = true
	if check != checks[OperatorCRDsReady] {
		checks[OperatorCRDsReady] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureOperators(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	// operator, helm-operator and internal-operator
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(OperatorsEnvReady)
	defer req.LogPostCheck(OperatorsEnvReady)

	b, err := templates.Parse(templates.InternalOperatorEnv, map[string]any{
		"namespace":       "kl-core",
		"cluster-id":      obj.Spec.ClusterID,
		"wildcard-domain": fmt.Sprintf("*.%s", obj.Spec.Domain),
	})
	if err != nil {
		return req.CheckFailed(OperatorsEnvReady, check, err.Error()).Err(nil)
	}

	if err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
		return req.CheckFailed(OperatorsEnvReady, check, err.Error()).Err(nil)
	}

	cfSecret, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.CloudflareCreds.SecretKeyRef.Namespace, obj.Spec.CloudflareCreds.SecretKeyRef.Name), &corev1.Secret{})
	if err != nil {
		return req.CheckFailed(OperatorsEnvReady, check, err.Error()).Err(nil)
	}

	cfCertManagerScrt := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.CloudflareCreds.SecretKeyRef.Name, Namespace: lc.NsCertManager}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, cfCertManagerScrt, func() error {
		if cfCertManagerScrt.Type != "" {
			cfCertManagerScrt.Type = cfSecret.Type
		}
		cfCertManagerScrt.Data = cfSecret.Data
		return nil
	}); err != nil {
		return req.CheckFailed(OperatorsEnvReady, check, err.Error()).Err(nil)
	}

	b, err = templates.Parse(templates.RouterOperatorEnv, map[string]any{
		"namespace":                   lc.NsOperators,
		"default-cluster-issuer-name": obj.Spec.CertManagerValues.ClusterIssuer.Name,
	})
	if err != nil {
		return req.CheckFailed(OperatorsEnvReady, check, err.Error()).Err(nil)
	}

	if err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
		return req.CheckFailed(OperatorsEnvReady, check, err.Error())
	}

	if obj.Spec.ShouldInstallSecondary {
		b, err = templates.Parse(templates.InternalOperatorEnv, map[string]any{
			"namespace":           lc.NsOperators,
			"cluster-id":          obj.Spec.SecondaryClusterId,
			"wildcard-domain":     obj.Spec.Domain,
			"nameserver-endpoint": fmt.Sprintf("https://%s.%s", obj.Spec.SharedConstants.AppDnsApi, obj.Spec.Domain),
		})
		if err != nil {
			return req.CheckFailed(OperatorsEnvReady, check, err.Error()).Err(nil)
		}

		if err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
			return req.CheckFailed(OperatorsEnvReady, check, err.Error())
		}
	}

	check.Status = true
	if check != checks[OperatorsEnvReady] {
		checks[OperatorsEnvReady] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureRedpandaTopics(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	topics := []string{
		obj.Spec.SharedConstants.KafkaTopicGitWebhooks,
		obj.Spec.SharedConstants.KafkaTopicEvents,
		obj.Spec.SharedConstants.KafkaTopicHarborWebhooks,
		obj.Spec.SharedConstants.KafkaTopicsStatusUpdates,
		obj.Spec.SharedConstants.KafkaTopicPipelineRunUpdates,
		obj.Spec.SharedConstants.KafkaTopicBillingUpdates,
	}

	for i := range topics {
		kt := &redpandaMsvcv1.Topic{ObjectMeta: metav1.ObjectMeta{Name: topics[i], Namespace: lc.NsCore}}
		if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, kt, func() error {
			if !fn.IsOwner(kt, fn.AsOwner(obj)) {
				kt.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
			}
			kt.Spec = redpandaMsvcv1.TopicSpec{
				AdminSecretRef: ct.SecretRef{
					Name:      obj.Spec.SharedConstants.RedpandaAdminSecretName,
					Namespace: lc.NsCore,
				},
			}
			return nil
		}); err != nil {
			return req.CheckFailed(RedpandaTopicsCreated, check, err.Error()).Err(nil)
		}
	}

	check.Status = true
	if check != checks[RedpandaTopicsCreated] {
		checks[RedpandaTopicsCreated] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureMsvcAndMres(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(MsvcAndMresReady)
	defer req.LogPostCheck(MsvcAndMresReady)

	mongoSvc, err := rApi.Get(ctx, r.Client, fn.NN(lc.NsCore, obj.Spec.SharedConstants.MongoSvcName), &crdsv1.ManagedService{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.CheckFailed(MsvcAndMresReady, check, err.Error())
		}
		req.Logger.Infof("%s does not exist, will be creating it", obj.Spec.SharedConstants.MongoSvcName)
	}

	if mongoSvc == nil || check.Generation > checks[MsvcAndMresReady].Generation {
		b, err := templates.Parse(templates.MongoMsvcAndMres, map[string]any{
			"namespace":           lc.NsCore,
			"local-storage-class": obj.Spec.StorageClass,
			"region":              "master",
			"shared-constants":    obj.Spec.SharedConstants,
		})
		if err != nil {
			return req.CheckFailed(MsvcAndMresReady, check, err.Error()).Err(nil)
		}

		if err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
			return req.CheckFailed(MsvcAndMresReady, check, err.Error()).Err(nil)
		}
	}

	redisSvc, err := rApi.Get(ctx, r.Client, fn.NN(lc.NsCore, obj.Spec.SharedConstants.RedisSvcName), &crdsv1.ManagedService{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return req.CheckFailed(MsvcAndMresReady, check, err.Error())
		}
		req.Logger.Infof("%s does not exist, will be creating it", obj.Spec.SharedConstants.RedisSvcName)
	}

	if redisSvc == nil || check.Generation > checks[MsvcAndMresReady].Generation {
		b, err := templates.Parse(templates.RedisMsvcAndMres, map[string]any{
			"namespace":           lc.NsCore,
			"local-storage-class": obj.Spec.StorageClass,
			"region":              "master",
			"shared-constants":    obj.Spec.SharedConstants,
		})
		if err != nil {
			return req.CheckFailed(MsvcAndMresReady, check, err.Error()).Err(nil)
		}

		if err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
			return req.CheckFailed(MsvcAndMresReady, check, err.Error()).Err(nil)
		}
	}

	check.Status = true
	if check != checks[MsvcAndMresReady] {
		checks[MsvcAndMresReady] = check
		req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) ensureKloudliteApis(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(KloudliteAPIsReady)
	defer req.LogPostCheck(KloudliteAPIsReady)

	// patch oauth-secrets
	oauthSecret, err := rApi.Get(ctx, r.Client, fn.NN(obj.Spec.OAuthCreds.Namespace, obj.Spec.OAuthCreds.Name), &corev1.Secret{})
	if err != nil {
		return req.CheckFailed(KloudliteAPIsReady, check, err.Error()).Err(nil)
	}

	localOAuthScrt := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.OAuthCreds.Name, Namespace: lc.NsCore}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, localOAuthScrt, func() error {
		d := make(map[string]string, len(oauthSecret.Data))
		for k, v := range oauthSecret.Data {
			d[k] = string(v)
		}
		b, err := json.Marshal(d)
		if err != nil {
			return err
		}

		parsedBytes, err := templates.ParseBytes(b, obj.Spec.SharedConstants)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(parsedBytes, &localOAuthScrt.StringData); err != nil {
			return err
		}

		//localOAuthScrt.Data = oauthSecret.Data
		//localOAuthScrt.Data["GITHUB_CALLBACK_URL"] = []byte(strings.Replace(string(localOAuthScrt.Data["GITHUB_CALLBACK_URL"]), "AUTH_WEB_DOMAIN", fmt.Sprintf("auth.%s", obj.Spec.Domain), 1))
		//localOAuthScrt.Data["GITLAB_CALLBACK_URL"] = []byte(strings.Replace(string(localOAuthScrt.Data["GITLAB_CALLBACK_URL"]), "AUTH_WEB_DOMAIN", fmt.Sprintf("auth.%s", obj.Spec.Domain), 1))
		//localOAuthScrt.Data["GOOGLE_CALLBACK_URL"] = []byte(strings.Replace(string(localOAuthScrt.Data["GOOGLE_CALLBACK_URL"]), "AUTH_WEB_DOMAIN", fmt.Sprintf("auth.%s", obj.Spec.Domain), 1))
		return nil
	}); err != nil {
		return req.CheckFailed(KloudliteAPIsReady, check, err.Error()).Err(nil)
	}

	apis := []ReconFn{
		r.ensureWebhookAuthzSecrets,
		r.ensureStripeCreds,
		r.ensureAuthApi,
		r.ensureConsoleApi,
		r.ensureCIApi,
		r.ensureFinanceApi,
		r.ensureCommsApi,
		r.ensureDnsApi,
		r.ensureIAMApi,
		r.ensureWebhooksApi,
		r.ensureJsEvalApi,
		r.ensureAuditLoggingWorker,
		r.ensureGatewayApi,
	}

	for i := range apis {
		if step := apis[i](req); !step.ShouldProceed() {
			return step
		}
	}

	check.Status = true
	if check != checks[KloudliteAPIsReady] {
		checks[KloudliteAPIsReady] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) ensureKloudliteWebs(req *rApi.Request[*v1.PrimaryCluster]) stepResult.Result {
	_, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(KloudliteWebReady)
	defer req.LogPostCheck(KloudliteWebReady)

	webs := []ReconFn{
		r.ensureAuthWeb,
		r.ensureAccountsWeb,
		r.ensureConsoleWeb,
		r.ensureSocketWeb,
	}

	for i := range webs {
		if step := webs[i](req); !step.ShouldProceed() {
			return step
		}
	}

	check.Status = true
	if check != checks[KloudliteWebReady] {
		checks[KloudliteWebReady] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig())
	r.restConfig = mgr.GetConfig()

	builder := ctrl.NewControllerManagedBy(mgr).For(&v1.PrimaryCluster{})
	builder.Owns(&crdsv1.App{})
	builder.Owns(&crdsv1.ManagedService{})
	builder.Owns(&crdsv1.ManagedResource{})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
