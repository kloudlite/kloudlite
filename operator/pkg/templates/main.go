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

	"github.com/Masterminds/sprig/v3"
	"sigs.k8s.io/yaml"

	"github.com/kloudlite/operator/pkg/errors"
)

//go:embed templates
var templateFS embed.FS

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

func ParseBytes(b []byte, values any) ([]byte, error) {
	t, err := newTemplate("parse-bytes")
	if err != nil {
		return nil, err
	}
	//t := template.New("parse-bytes")
	//t.Funcs(txtFuncs(t))
	if _, err := t.Parse(string(b)); err != nil {
		return nil, err
	}
	out := new(bytes.Buffer)
	if err := t.ExecuteTemplate(out, "parse-bytes", values); err != nil {
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
		if x == "null" || x == "" || x == "[]" || x == "{}" {
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
	funcs["endl"] = func() (string, error) {
		return "\n", nil
	}

	return funcs
}

func WithFunctions(t *template.Template) *template.Template {
	return t.Funcs(txtFuncs(t))
}

type templateFile string

const (
	MongoDBStandalone templateFile = "templates/msvc/mongodb/helm-mongodb-standalone.yml.tpl"

	MySqlStandalone templateFile = "templates/msvc/mysql/helm-mongodb-standalone.yml.tpl"
	MysqlCluster    templateFile = "templates/msvc/mysql/helm-cluster.yml.tpl"

	RedisStandalone   templateFile = "templates/msvc/redis/helm-mongodb-standalone.yml.tpl"
	RedisACLConfigMap templateFile = "templates/msvc/redis/acl-configmap.yml.tpl"

	// ---

	MongoDBCluster   templateFile = "templates/mongodb-helm-one-node-cluster.yml.tpl"
	Deployment       templateFile = "templates/app.yml.tpl"
	Service          templateFile = "templates/service.yml.tpl"
	Secret           templateFile = "templates/corev1/secret.yml.tpl"
	AccountWireguard templateFile = "templates/account-deploy.yml.tpl"
	CommonMsvc       templateFile = "templates/msvc-common-service.yml.tpl"
	CommonMres       templateFile = "templates/mres-common.yml.tpl"

	ServerlessLambda templateFile = "templates/serverless/lambda.yml.tpl"

	ElasticSearch templateFile = "templates/msvc/elasticsearch/elastic-helm.yml.tpl"
	Kibana        templateFile = "templates/msvc/elasticsearch/kibana-helm.yml.tpl"
	OpenSearch    templateFile = "templates/msvc/opensearch/helm.yml.tpl"
	InfluxDB      templateFile = "templates/msvc/influx/helm.yml.tpl"

	// ---

	Project templateFile = "templates/project.yml.tpl"

	RedpandaOneNodeCluster templateFile = "templates/msvc/redpanda/one-node-cluster.yml.tpl"

	HelmIngressNginx     templateFile = "templates/ingress-nginx/helm.yml.tpl"
	AccountIngressBridge templateFile = "templates/ingress-nginx/ingress-bridge.tpl.yml"

	ProjectRBAC   templateFile = "templates/project-rbac.yml.tpl"
	ProjectHarbor templateFile = "templates/project-harbor.yml.tpl"

	MsvcHelmZookeeper templateFile = "templates/msvc/zookeeper/helm.yml.tpl"

	MsvcHelmNeo4jStandalone templateFile = "templates/msvc/neo4j/helm-mongodb-standalone.yml.tpl"

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
	ExternalNameSvc:    "templates/corev1/external-name-service.yml.tpl",
	Ingress:            "templates/corev1/ingress.yml.tpl",
	DockerConfigSecret: "templates/corev1/docker-config-secret.yml.tpl",
	Secret:             "templates/corev1/secret.yml.tpl",
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

var Clusters = struct {
	Job templateFile
}{
	Job: "templates/clustersv1/job.yml.tpl",
}

var Wireguard = struct {
	Config        templateFile
	Deploy        templateFile
	CoreDns       templateFile
	DeviceConfig  templateFile
	DeviceService templateFile
	DnsConfig     templateFile
}{
	Config:        "templates/wireguardv1/config.tmpl.conf",
	Deploy:        "templates/wireguardv1/deploy.yml.tpl",
	CoreDns:       "templates/wireguardv1/coredns.yml.tpl",
	DeviceConfig:  "templates/wireguardv1/device-config.tmpl.conf",
	DeviceService: "templates/wireguardv1/device-service.yml.tpl",
	DnsConfig:     "templates/wireguardv1/dns-config.yml.tpl",
}

var Distribution = struct {
	BuildJob templateFile
}{
	BuildJob: "templates/distribution/build-job.yml.tpl",
}
