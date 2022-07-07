package templates

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/Masterminds/sprig/v3"

	"operators.kloudlite.io/lib/errors"
)

func useTemplate(file templateFile) (*kt, error) {
	tFiles := []string{file.Path()}
	tFiles = append(tFiles, helperFiles...)

	var klFuncs template.FuncMap = map[string]any{}

	// SOURCE: https://github.com/helm/helm/blob/8648ccf5d35d682dcd5f7a9c2082f0aaf071e817/pkg/engine/engine.go#L147-L154
	t := template.New(filepath.Base(file.Path()))
	funcMap := sprig.TxtFuncMap()

	klFuncs["include"] = func(templateName string, templateData any) (string, error) {
		buf := bytes.NewBuffer(nil)
		if err := t.ExecuteTemplate(buf, templateName, templateData); err != nil {
			return "", err
		}
		return buf.String(), nil
	}

	klFuncs["toYAML"] = func(txt any) (string, error) {
		a, ok := funcMap["toPrettyJson"].(func(any) string)
		if !ok {
			panic("could not convert sprig.TxtFuncMap[toPrettyJson] into func(any) string")
		}
		ys, err := yaml.JSONToYAML([]byte(a(txt)))
		if err != nil {
			return "", err
		}
		return string(ys), nil
	}

	_, err := t.Funcs(funcMap).Funcs(klFuncs).ParseFiles(tFiles...)
	if err != nil {
		return nil, err
	}
	return &kt{Template: t}, nil
}

type kt struct {
	*template.Template
}

func (kt *kt) withValues(v interface{}) ([]byte, error) {
	w := new(bytes.Buffer)
	if err := kt.ExecuteTemplate(w, kt.Name(), v); err != nil {
		return nil, errors.NewEf(err, "could not execute template")
	}
	return w.Bytes(), nil
}

func Parse(f templateFile, values interface{}) ([]byte, error) {
	t, err := useTemplate(f)
	if err != nil {
		return nil, err
	}
	return t.withValues(values)
}

func ParseObject(f templateFile, values interface{}) (client.Object, error) {
	t, err := useTemplate(f)
	if err != nil {
		return nil, err
	}
	b, err := t.withValues(values)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := yaml.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return &unstructured.Unstructured{Object: m}, nil
}

var templateDir = filepath.Join(os.Getenv("PWD"), "lib/templates")

var helperFiles []string

func init() {
	if v, ok := os.LookupEnv("TEMPLATES_DIR"); ok {
		templateDir = v
	}

	helpersDir := filepath.Join(templateDir, "_helpers")
	dir, err := os.ReadDir(helpersDir)
	if err != nil {
		fmt.Printf("ERR listing templateDir.... %+v\n", err)
		panic(err)
	}
	for _, entry := range dir {
		if !entry.IsDir() {
			helperFiles = append(helperFiles, filepath.Join(helpersDir, entry.Name()))
		}
	}
}

type templateFile string

func (tf templateFile) Path() string {
	return filepath.Join(templateDir, string(tf))
}

const (
	MongoDBStandalone templateFile = "./msvc/mongodb/helm-standalone.tpl.yml"
	MySqlStandalone   templateFile = "./msvc/mysql/helm-standalone.tpl.yml"
	RedisStandalone   templateFile = "./msvc/redis/helm-standalone.tpl.yml"

	// ---

	MongoDBCluster   templateFile = "mongodb-helm-cluster.tmpl.yml"
	MongoDBWatcher   templateFile = "mongo-msvc-watcher.tmpl.yml"
	Deployment       templateFile = "app.tpl.yml"
	Service          templateFile = "service.tmpl.yml"
	Secret           templateFile = "secret.tmpl.yml"
	AccountWireguard templateFile = "account-deploy.tmpl.yml"
	CommonMsvc       templateFile = "msvc-common-service.tmpl.yml"
	CommonMres       templateFile = "mres-common.tmpl.yml"
	ConfigMap        templateFile = "configmap.tmpl.yml"
	Ingress          templateFile = "./ingress.tmpl.yml"

	IngressLambda templateFile = "./ingress-lambda.tmpl.yml"

	ServerlessLambda templateFile = "./serverless/lambda.yml.tpl"

	ElasticSearch templateFile = "./msvc/elasticsearch.tpl.yml"
	OpenSearch    templateFile = "./msvc/opensearch/helm.tpl.yml"
	InfluxDB      templateFile = "./msvc/influx/helm.tpl.yml"

	// ---

	Project templateFile = "./project.tpl.yml"
)

var CoreV1 = struct {
	ExternalNameSvc    templateFile
	Ingress            templateFile
	DockerConfigSecret templateFile
}{
	ExternalNameSvc:    "./corev1/external-name-service.tpl.yml",
	Ingress:            "./corev1/ingress.tpl.yml",
	DockerConfigSecret: "./corev1/docker-config-secret.tpl.yml",
}

var CrdsV1 = struct {
	App templateFile
}{
	App: "./app.tpl.yml",
}
