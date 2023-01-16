package templates

import (
	"bytes"
	"crypto/md5"
	"embed"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"reflect"
	"text/template"

	"sigs.k8s.io/yaml"

	"github.com/Masterminds/sprig/v3"
	"operators.kloudlite.io/pkg/errors"
)

var (
	//go:embed templates
	templateFS embed.FS
)

func newTemplate(name string) (*template.Template, error) {
	t := template.New(name).Option("missingkey=error")
	t.Funcs(txtFuncs(t))

	dirs, err := templateFS.ReadDir("templates/helpers")
	if err != nil {
		return nil, err
	}
	for i := range dirs {
		if !dirs[i].IsDir() {
			_, err := t.ParseFS(templateFS, fmt.Sprintf("templates/helpers/%s", dirs[i].Name()))
			if err != nil {
				return nil, err
			}
		}
	}
	return t, nil
}

func Parse(f templateFile, values any) ([]byte, error) {
	name := filepath.Base(string(f))
	t, err := newTemplate(name)
	if err != nil {
		return nil, err
	}

	_, err = t.ParseFS(templateFS, string(f))
	if err != nil {
		return nil, err
	}

	out := new(bytes.Buffer)
	if err := t.ExecuteTemplate(out, name, values); err != nil {
		return nil, errors.NewEf(err, "could not execute template")
	}

	return out.Bytes(), nil
}

func txtFuncs(t *template.Template) template.FuncMap {
	funcs := sprig.TxtFuncMap()

	// inspired by helm include
	funcs["include"] = func(templateName string, templateData any) (string, error) {
		buf := bytes.NewBuffer(nil)
		if err := t.ExecuteTemplate(buf, templateName, templateData); err != nil {
			return "", err
		}
		return buf.String(), nil
	}

	funcs["toYAML"] = func(txt any) (string, error) {
		if txt == nil {
			return "", nil
		}

		a, ok := funcs["toPrettyJson"].(func(any) string)
		if !ok {
			panic("could not convert sprig.TxtFuncMap[toPrettyJson] into func(any) string")
		}

		x := a(txt)
		if x == "null" {
			return "", nil
		}

		ys, err := yaml.JSONToYAML([]byte(x))
		if err != nil {
			return "", err
		}
		return string(ys), nil
	}

	funcs["md5"] = func(txt string) string {
		hash := md5.New()
		hash.Write([]byte(txt))
		return hex.EncodeToString(hash.Sum(nil))
	}

	funcs["K8sAnnotation"] = func(cond any, key string, value any) string {
		if cond == reflect.Zero(reflect.TypeOf(cond)).Interface() {
			return ""
		}
		return fmt.Sprintf("%s: '%v'", key, value)
	}

	funcs["K8sLabel"] = funcs["K8sAnnotation"]

	funcs["Iterate"] = func(count int) []int {
		var i int
		var Items []int
		for i = 0; i < count; i++ {
			Items = append(Items, i)
		}
		return Items
	}

	return funcs
}

func WithFunctions(t *template.Template) *template.Template {
	return t.Funcs(txtFuncs(t))
}

type templateFile string

const (
	MongoDBStandalone templateFile = "templates/msvc/mongodb/helm-standalone.yml.tpl"

	MySqlStandalone templateFile = "templates/msvc/mysql/helm-standalone.yml.tpl"
	MysqlCluster    templateFile = "templates/msvc/mysql/helm-cluster.tpl.yml"

	RedisStandalone   templateFile = "templates/msvc/redis/helm-standalone.yml.tpl"
	RedisACLConfigMap templateFile = "templates/msvc/redis/acl-configmap.tpl.yml"

	// ---

	MongoDBCluster   templateFile = "templates/mongodb-helm-one-node-cluster.tpl.yml"
	MongoDBWatcher   templateFile = "templates/mongo-msvc-watcher.tmpl.yml"
	Deployment       templateFile = "templates/app.yml.tpl"
	Service          templateFile = "templates/service.tmpl.yml"
	Secret           templateFile = "templates/corev1/secret.tpl.yml"
	AccountWireguard templateFile = "templates/account-deploy.tmpl.yml"
	CommonMsvc       templateFile = "templates/msvc-common-service.tpl.yml"
	CommonMres       templateFile = "templates/mres-common.yml.tpl"
	Ingress          templateFile = "templates/ingress.tmpl.yml"

	IngressLambda templateFile = "templates/ingress-lambda.tmpl.yml"

	ServerlessLambda templateFile = "templates/serverless/lambda.tpl.yml"

	ElasticSearch templateFile = "templates/msvc/elasticsearch/elastic-helm.yml.tpl"
	Kibana        templateFile = "templates/msvc/elasticsearch/kibana-helm.yml.tpl"
	OpenSearch    templateFile = "templates/msvc/opensearch/helm.yml.tpl"
	InfluxDB      templateFile = "templates/msvc/influx/helm.yml.tpl"

	// ---

	Project templateFile = "templates/project.tpl.yml"

	RedpandaOneNodeCluster templateFile = "templates/msvc/redpanda/one-node-cluster.tpl.yml"

	HelmIngressNginx     templateFile = "templates/ingress-nginx/helm.yml.tpl"
	AccountIngressBridge templateFile = "templates/ingress-nginx/ingress-bridge.tpl.yml"

	ProjectRBAC   templateFile = "templates/project-rbac.yml.tpl"
	ProjectHarbor templateFile = "templates/project-harbor.yml.tpl"

	MsvcHelmZookeeper templateFile = "templates/msvc/zookeeper/helm.yml.tpl"

	MsvcHelmNeo4jStandalone templateFile = "templates/msvc/neo4j/helm-standalone.yaml.tpl"

	AwsEbsCsiDriver    templateFile = "templates/csi/aws-ebs-csi-driver.yml.tpl"
	AwsEbsStorageClass templateFile = "templates/csi/aws-storage-class.yml.tpl"

	DigitaloceanCSIDriver    templateFile = "templates/csi/digitalocean/csi-driver.yml.tpl"
	DigitaloceanStorageClass templateFile = "templates/csi/digitalocean/storage-class.yml.tpl"

	ClusterIssuer templateFile = "templates/cluster-issuer.yml.tpl"
)

var CoreV1 = struct {
	ExternalNameSvc    templateFile
	Ingress            templateFile
	DockerConfigSecret templateFile
	Secret             templateFile
	Namespace          templateFile
	ConfigMap          templateFile
	Deployment         templateFile
}{
	ExternalNameSvc:    "templates/corev1/external-name-service.tpl.yml",
	Ingress:            "templates/corev1/ingress.yml.tpl",
	DockerConfigSecret: "templates/corev1/docker-config-secret.tpl.yml",
	Secret:             "templates/corev1/secret.tpl.yml",
	Namespace:          "templates/corev1/namespace.yml.tpl",
	ConfigMap:          "templates/corev1/configmap.yml.tpl",
}

var CrdsV1 = struct {
	App           templateFile
	Secret        templateFile
	AccountRouter templateFile
}{
	App:           "templates/app.yml.tpl",
	Secret:        "templates/crdsv1/secret.yml.tpl",
	AccountRouter: "templates/crdsv1/account-router.yml.tpl",
}
