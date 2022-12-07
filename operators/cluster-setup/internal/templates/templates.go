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
	LokiValues         templateFile = "templates/loki-values.yml.tpl"
	PrometheusValues   templateFile = "templates/prometheus-values.yml.tpl"
	CertManagerValues  templateFile = "templates/cert-manager.yml.tpl"
	CertIssuer         templateFile = "templates/cert-issuer.yml.tpl"
	IngressNginxValues templateFile = "templates/ingress-nginx.yml.tpl"
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
	AuthApi templateFile = "templates/apps/auth-api.yml.tpl"
)
