package domain

import (
	"bytes"
	"fmt"
	"net/url"
	"text/template"

	dbv1 "github.com/kloudlite/operator/apis/distribution/v1"

	common_types "github.com/kloudlite/operator/apis/common-types"
	"kloudlite.io/apps/container-registry/templates"
	text_templates "kloudlite.io/pkg/text-templates"
)

func BuildUrl(repo, pullToken string) (string, error) {
	parsedURL, err := url.Parse(repo)
	if err != nil {
		fmt.Println("Error parsing Repo URL:", err)
		return "", err
	}

	parsedURL.User = url.User(pullToken)

	return parsedURL.String(), nil
}

type BuildJobTemplateData struct {
	AccountName string
	Name        string
	Namespace   string
	Labels      map[string]string
	Annotations map[string]string

	Registry     dbv1.Registry
	CacheKeyName *string
	Resource     dbv1.Resource
	GitRepo      dbv1.GitRepo
	BuildOptions *dbv1.BuildOptions

	CredentialsRef common_types.SecretRef
}

func getTemplate(obj BuildJobTemplateData) ([]byte, error) {
	b, err := templates.ReadBuildJobTemplate()
	if err != nil {
		return nil, err
	}

	tmpl := text_templates.WithFunctions(template.New("build-job-template"))

	tmpl, err = tmpl.Parse(string(b))
	if err != nil {
		return nil, err
	}

	out := new(bytes.Buffer)
	err = tmpl.Execute(out, obj)
	if err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}

func (*Impl) GetBuildTemplate(obj BuildJobTemplateData) ([]byte, error) {
	return getTemplate(obj)
}
