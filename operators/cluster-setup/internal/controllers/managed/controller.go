package managed

import (
	"context"
	"encoding/base64"
	"fmt"
	"golang.org/x/sync/errgroup"
	"io"
	"net/http"
	"strings"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/kloudlite/operator/pkg/constants"
	"github.com/kloudlite/operator/pkg/errors"
	"github.com/kloudlite/operator/pkg/helm"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kloudlite/operator/apis/cluster-setup/v1"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
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
	Env        *env.Env

	TemplateWgOperator     []byte
	TemplateCsiOperator    []byte
	TemplateRouterOperator []byte
}

func (r *Reconciler) GetName() string {
	return r.Name
}

const (
	WgOperatorReady         string = "internal-operator-installed"
	DefaultsPatched         string = "defaults-patched"
	KloudliteCredsValidated string = "kloudlite-creds-validated"
	UserKubeConfigCreated   string = "user-kubeconfig-created"
	CsiOperatorReady        string = "csi-operator-ready"
	RoutersOperatorReady    string = "routers-operator-ready"
	LokiReady               string = "loki-ready"
	PrometheusReady         string = "prometheus-ready"
	Finalizing              string = "finalizing"
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

	if step := r.ensureRoutersOperator(req); !step.ShouldProceed() {
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
	return ctrl.Result{}, r.Status().Update(ctx, req.Object)
}

func (r *Reconciler) finalize(req *rApi.Request[*v1.ManagedCluster]) stepResult.Result {
	//ctx, obj := req.Context(), req.Object
	//check := rApi.Check{Generation: obj.Generation}
	//
	//check.Status = true
	//if check != obj.Status.Checks[Finalizing] {
	//	obj.Status.Checks[Finalizing] = check
	//	req.UpdateStatus()
	//}
	return req.Next()
}

func (r *Reconciler) checkKloudliteCreds(req *rApi.Request[*v1.ManagedCluster]) stepResult.Result {
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{Generation: obj.Generation}

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
	if check != checks[KloudliteCredsValidated] {
		checks[KloudliteCredsValidated] = check
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

	if err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
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

	if err := r.yamlClient.ApplyYAML(ctx, b4); err != nil {
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

	if err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
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

	if err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
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

	b, err := templates.ParseBytes(r.TemplateRouterOperator, map[string]any{
		"Namespace":       lc.NsOperators,
		"EnvName":         "development",
		"ImageTag":        "v1.0.5",
		"ImagePullPolicy": "Always",
		"SvcAccountName":  lc.ClusterSvcAccount,
	})

	if err != nil {
		return req.CheckFailed(RoutersOperatorReady, check, err.Error()).Err(nil)
	}

	if err := r.yamlClient.ApplyYAML(ctx, b); err != nil {
		return req.CheckFailed(RoutersOperatorReady, check, err.Error()).Err(nil)
	}

	check.Status = true
	if check != obj.Status.Checks[RoutersOperatorReady] {
		obj.Status.Checks[RoutersOperatorReady] = check
		req.UpdateStatus()
	}

	return req.Next()
}

func (r *Reconciler) ensureLoki(req *rApi.Request[*v1.ManagedCluster]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(LokiReady)
	defer req.LogPostCheck(LokiReady)

	helmCli, err := helm.NewHelmClient(r.restConfig, helm.ClientOptions{
		Namespace: lc.NsCore,
	})
	if err != nil {
		return req.CheckFailed(LokiReady, check, err.Error()).Err(nil)
	}

	releaseName := obj.Spec.LokiValues.ServiceName
	helmValues, err := helmCli.GetReleaseValues(ctx, releaseName)
	if err != nil {
		req.Logger.Infof("helm release (%s) not found, will be creating it", releaseName)
	}

	b, err := templates.Parse(templates.LokiValues, map[string]any{"loki-values": obj.Spec.LokiValues})
	if err != nil {
		return req.CheckFailed(LokiReady, check, err.Error()).Err(nil)
	}

	if !helm.AreHelmValuesEqual(helmValues, b) {
		if err := helmCli.AddOrUpdateChartRepo(ctx, helm.RepoEntry{
			Name: "grafana",
			Url:  "https://grafana.github.io/helm-charts",
		}); err != nil {
			return req.CheckFailed(LokiReady, check, err.Error()).Err(nil)
		}

		if _, err := helmCli.InstallOrUpgradeChart(ctx, helm.ChartSpec{
			ReleaseName: releaseName,
			ChartName:   "grafana/loki-stack",
			Namespace:   lc.NsCore,
			ValuesYaml:  string(b),
		}); err != nil {
			return req.CheckFailed(LokiReady, check, err.Error()).Err(nil)
		}
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

	req.LogPreCheck(PrometheusReady)
	defer req.LogPostCheck(PrometheusReady)

	helmCli, err := helm.NewHelmClient(r.restConfig, helm.ClientOptions{Namespace: lc.NsCore})
	if err != nil {
		return req.CheckFailed(PrometheusReady, check, err.Error()).Err(nil)
	}

	releaseName := obj.Spec.PrometheusValues.ServiceName
	helmValues, err := helmCli.GetReleaseValues(ctx, releaseName)
	if err != nil {
		req.Logger.Infof("helm release (%s) not found, will be creating it", releaseName)
	}

	b, err := templates.Parse(templates.PrometheusValues, map[string]any{"prometheus-values": obj.Spec.PrometheusValues, "name": releaseName})
	if err != nil {
		return req.CheckFailed(PrometheusReady, check, err.Error()).Err(nil)
	}

	if !helm.AreHelmValuesEqual(helmValues, b) {
		if err := helmCli.AddOrUpdateChartRepo(ctx, helm.RepoEntry{
			Name: "bitnami",
			Url:  "https://charts.bitnami.com/bitnami",
		}); err != nil {
			return req.CheckFailed(PrometheusReady, check, err.Error()).Err(nil)
		}

		if _, err := helmCli.InstallOrUpgradeChart(ctx, helm.ChartSpec{
			ReleaseName: releaseName,
			ChartName:   "bitnami/kube-prometheus",
			Namespace:   lc.NsCore,
			ValuesYaml:  string(b),
		}); err != nil {
			req.Logger.Error(err)
			return req.CheckFailed(PrometheusReady, check, err.Error()).Err(nil)
		}
	}

	check.Status = true
	if check != obj.Status.Checks[PrometheusReady] {
		obj.Status.Checks[PrometheusReady] = check
		req.UpdateStatus()
	}
	return req.Next()
}

func (r *Reconciler) downloadOperatorTemplate(name string) ([]byte, error) {
	req, err := http.NewRequestWithContext(
		context.TODO(),
		http.MethodGet,
		fmt.Sprintf("https://%s/%s/operators/download/%s", r.Env.ReleasesApiEndpoint, r.Env.ReleaseVersion, name),
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

	var err error
	var g errgroup.Group

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
	if err := g.Wait(); err != nil {
		return err
	}
	r.logger.Infof("template downloading took %.2fs", time.Since(ts).Seconds())

	builder := ctrl.NewControllerManagedBy(mgr).For(&v1.ManagedCluster{})
	builder.Owns(&crdsv1.App{})
	builder.WithEventFilter(rApi.ReconcileFilter())
	return builder.Complete(r)
}
