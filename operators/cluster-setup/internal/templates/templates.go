package templates

import (
	"bytes"
	"embed"
	"path/filepath"
	"text/template"

	"operators.kloudlite.io/pkg/errors"
	libTemplates "operators.kloudlite.io/pkg/templates"
)

var (
	//go:embed templates
	FS embed.FS
)

func Parse(f templateFile, values any) ([]byte, error) {
	name := filepath.Base(string(f))
	t := template.New(name)
	t = libTemplates.WithFunctions(t)

	if _, err := t.ParseFS(FS, string(f)); err != nil {
		return nil, err
	}

	out := new(bytes.Buffer)
	if err := t.ExecuteTemplate(out, name, values); err != nil {
		return nil, errors.NewEf(err, "could not execute template")
	}

	return out.Bytes(), nil
}

func ParseBytes(b []byte, values any) ([]byte, error) {
	t := template.New("parse-bytes")
	t = libTemplates.WithFunctions(t)
	if _, err := t.Parse(string(b)); err != nil {
		return nil, err
	}
	out := new(bytes.Buffer)
	if err := t.ExecuteTemplate(out, "parse-bytes", values); err != nil {
		return nil, errors.NewEf(err, "could not execute template")
	}
	return out.Bytes(), nil
}

type templateFile string

const (
	LokiValues                templateFile = "templates/loki-values.yml.tpl"
	PrometheusValues          templateFile = "templates/prometheus-values.yml.tpl"
	CertManagerValues         templateFile = "templates/cert-manager.yml.tpl"
	ClusterIssuer             templateFile = "templates/cluster-issuer.yml.tpl"
	IngressNginxValues        templateFile = "templates/ingress-nginx.yml.tpl"
	RedpandaValues            templateFile = "templates/helm/redpanda-values.yml.tpl"
	RedpandaSingleNodeCluster templateFile = "templates/helm/redpanda-single-node-cluster.yml.tpl"
)

const (
	MongoMsvcAndMres templateFile = "templates/kl-mongo-svc-n-res.yml.tpl"
	RedisMsvcAndMres templateFile = "templates/kl-redis-svc-n-res.yml.tpl"
)

const (
	RouterOperatorEnv   templateFile = "templates/secrets/router-operator-env.yml.tpl"
	InternalOperatorEnv templateFile = "templates/secrets/internal-operator-env.yml.tpl"
	ProjectOperatorEnv  templateFile = "templates/secrets/project-operator-env.yml.tpl"
)

const (
	AuthApi            templateFile = "templates/apps/auth-api.yml.tpl"
	ConsoleApi         templateFile = "templates/apps/console-api.yml.tpl"
	CiApi              templateFile = "templates/apps/ci-api.yml.tpl"
	DnsApi             templateFile = "templates/apps/dns-api.yml.tpl"
	FinanceApi         templateFile = "templates/apps/finance-api.yml.tpl"
	CommsApi           templateFile = "templates/apps/comms-api.yml.tpl"
	GatewayApi         templateFile = "templates/apps/gateway-api.yml.tpl"
	IamApi             templateFile = "templates/apps/iam-api.yml.tpl"
	JsEvalApi          templateFile = "templates/apps/js-eval-api.yml.tpl"
	WebhooksApi        templateFile = "templates/apps/webhooks-api.yml.tpl"
	AuditLoggingWorker templateFile = "templates/apps/audit-logging-worker.yml.tpl"
)

const (
	AuthWeb     templateFile = "templates/apps/web/auth-web.yml.tpl"
	ConsoleWeb  templateFile = "templates/apps/web/console-web.yml.tpl"
	AccountsWeb templateFile = "templates/apps/web/accounts-web.yml.tpl"
	SocketWeb   templateFile = "templates/apps/web/socket-web.yml.tpl"
)

const (
	KloudliteAgent templateFile = "templates/apps/kloudlite-agent.yml.tpl"
)
