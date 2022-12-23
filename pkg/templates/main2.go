package templates

// import (
// 	"bytes"
// 	"crypto/md5"
// 	"embed"
// 	"encoding/hex"
// 	"fmt"
// 	"os"
// 	"path/filepath"
// 	"reflect"
// 	"text/template"
//
// 	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
// 	"sigs.k8s.io/controller-runtime/pkg/client"
// 	"sigs.k8s.io/yaml"
//
// 	"github.com/Masterminds/sprig/v3"
// 	"operators.kloudlite.io/types/errors"
// )
//
// func txtFuncs(t *template.Template) template.FuncMap {
// 	funcs := sprig.TxtFuncMap()
//
// 	// inspired by helm include
// 	funcs["include"] = func(templateName string, templateData any) (string, error) {
// 		buf := bytes.NewBuffer(nil)
// 		if err := t.ExecuteTemplate(buf, templateName, templateData); err != nil {
// 			return "", err
// 		}
// 		return buf.String(), nil
// 	}
//
// 	funcs["toYAML"] = func(txt any) (string, error) {
// 		if txt == nil {
// 			return "", nil
// 		}
//
// 		a, ok := funcs["toPrettyJson"].(func(any) string)
// 		if !ok {
// 			panic("could not convert sprig.TxtFuncMap[toPrettyJson] into func(any) string")
// 		}
//
// 		x := a(txt)
// 		if x == "null" {
// 			return "", nil
// 		}
//
// 		ys, err := yaml.JSONToYAML([]byte(x))
// 		if err != nil {
// 			return "", err
// 		}
// 		return string(ys), nil
// 	}
//
// 	funcs["md5"] = func(txt string) string {
// 		hash := md5.New()
// 		hash.Write([]byte(txt))
// 		return hex.EncodeToString(hash.Sum(nil))
// 	}
//
// 	return funcs
// }
//
// func WithFunctions(t *template.Template) *template.Template {
// 	return t.Funcs(txtFuncs(t))
// }
//
// func useTemplate(file templateFile) (*kt, error) {
// 	tFiles := []string{file.Path()}
// 	tFiles = append(tFiles, helperFiles...)
//
// 	var klFuncs template.FuncMap = map[string]any{}
//
// 	// SOURCE: https://github.com/helm/helm/blob/8648ccf5d35d682dcd5f7a9c2082f0aaf071e817/pkg/engine/engine.go#L147-L154
// 	t := template.New(filepath.Base(file.Path())).Option("missingkey=error")
// 	funcMap := sprig.TxtFuncMap()
//
// 	klFuncs["include"] = func(templateName string, templateData any) (string, error) {
// 		buf := bytes.NewBuffer(nil)
// 		if err := t.ExecuteTemplate(buf, templateName, templateData); err != nil {
// 			return "", err
// 		}
// 		return buf.String(), nil
// 	}
//
// 	klFuncs["toYAML"] = func(txt any) (string, error) {
// 		if txt == nil {
// 			return "", nil
// 		}
//
// 		a, ok := funcMap["toPrettyJson"].(func(any) string)
// 		if !ok {
// 			panic("could not convert sprig.TxtFuncMap[toPrettyJson] into func(any) string")
// 		}
//
// 		x := a(txt)
// 		if x == "null" {
// 			return "", nil
// 		}
//
// 		ys, err := yaml.JSONToYAML([]byte(x))
// 		if err != nil {
// 			return "", err
// 		}
// 		return string(ys), nil
// 	}
//
// 	klFuncs["ENDL"] = func() string {
// 		return "\n"
// 	}
//
// 	klFuncs["K8sAnnotation"] = func(cond any, key string, value any) string {
// 		if cond == reflect.Zero(reflect.TypeOf(cond)).Interface() {
// 			return ""
// 		}
// 		return fmt.Sprintf("%s: \"%v\"", key, value)
// 	}
// 	klFuncs["K8sLabel"] = klFuncs["K8sAnnotation"]
//
// 	klFuncs["md5"] = func(txt string) string {
// 		hash := md5.New()
// 		hash.Write([]byte(txt))
// 		return hex.EncodeToString(hash.Sum(nil))
// 	}
//
// 	_, err := t.Funcs(funcMap).Funcs(klFuncs).ParseFiles(tFiles...)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &kt{Template: t}, nil
// }
//
// type kt struct {
// 	*template.Template
// }
//
// func (kt *kt) withValues(v interface{}) ([]byte, error) {
// 	w := new(bytes.Buffer)
// 	if err := kt.ExecuteTemplate(w, kt.Name(), v); err != nil {
// 		return nil, errors.NewEf(err, "could not execute template")
// 	}
// 	return w.Bytes(), nil
// }
//
// func Parse(f templateFile, values interface{}) ([]byte, error) {
// 	t, err := useTemplate(f)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return t.withValues(values)
// }
//
// func ParseObject(f templateFile, values interface{}) (client.Object, error) {
// 	t, err := useTemplate(f)
// 	if err != nil {
// 		return nil, err
// 	}
// 	b, err := t.withValues(values)
// 	if err != nil {
// 		return nil, err
// 	}
// 	var m map[string]any
// 	if err := yaml.Unmarshal(b, &m); err != nil {
// 		return nil, err
// 	}
// 	return &unstructured.Unstructured{Object: m}, nil
// }
//
// var templateDir = filepath.Join(os.Getenv("PWD"), "types/controller-templates")
//
// var helperFiles []string
//
// var (
// 	//go:embed controller-templates
// 	templateFS embed.FS
// )
//
// func init() {
// 	if v, ok := os.LookupEnv("TEMPLATES_DIR"); ok {
// 		templateDir = v
// 	}
//
// 	helpersDir := filepath.Join(templateDir, "helpers")
// 	dir, err := os.ReadDir(helpersDir)
// 	if err != nil {
// 		fmt.Printf("ERR listing templateDir.... %+v\n", err)
// 		panic(err)
// 	}
// 	for _, entry := range dir {
// 		if !entry.IsDir() {
// 			helperFiles = append(helperFiles, filepath.Join(helpersDir, entry.Name()))
// 		}
// 	}
// }
//
// type templateFile string
//
// func (tf templateFile) Path() string {
// 	return filepath.Join(templateDir, string(tf))
// }
//
// const (
// 	MongoDBStandalone templateFile = "./msvc/mongodb/helm-standalone.yml.tpl"
// 	MySqlStandalone   templateFile = "./msvc/mysql/helm-standalone.yml.tpl"
// 	RedisStandalone   templateFile = "./msvc/redis/helm-standalone.yml.tpl"
//
// 	// ---
//
// 	MongoDBCluster   templateFile = "mongodb-helm-one-node-cluster.tpl.yml"
// 	MongoDBWatcher   templateFile = "mongo-msvc-watcher.tmpl.yml"
// 	Deployment       templateFile = "app-n-lambda.tpl.yml"
// 	Service          templateFile = "service.tmpl.yml"
// 	Secret           templateFile = "./corev1/secret.tpl.yml"
// 	AccountWireguard templateFile = "account-deploy.tmpl.yml"
// 	CommonMsvc       templateFile = "msvc-common-service.tpl.yml"
// 	CommonMres       templateFile = "mres-common.tmpl.yml"
// 	ConfigMap        templateFile = "configmap.tmpl.yml"
// 	Ingress          templateFile = "./ingress.tmpl.yml"
//
// 	IngressLambda templateFile = "./ingress-lambda.tmpl.yml"
//
// 	ServerlessLambda templateFile = "./serverless/lambda.tpl.yml"
//
// 	ElasticSearch templateFile = "./msvc/elasticsearch/helm.yml.tpl"
// 	OpenSearch    templateFile = "./msvc/opensearch/helm.yml.tpl"
// 	InfluxDB      templateFile = "./msvc/influx/helm.yml.tpl"
//
// 	// ---
//
// 	Project templateFile = "./project.tpl.yml"
//
// 	RedpandaOneNodeCluster templateFile = "./msvc/redpanda/one-node-cluster.tpl.yml"
//
// 	HelmIngressNginx     templateFile = "./ingress-nginx/helm.yml.tpl"
// 	AccountIngressBridge templateFile = "./ingress-nginx/ingress-bridge.tpl.yml"
//
// 	ProjectRBAC   templateFile = "./project-rbac.yml.tpl"
// 	ProjectHarbor templateFile = "./project-artifacts-harbor.yml.tpl"
// )
//
// var CoreV1 = struct {
// 	ExternalNameSvc    templateFile
// 	Ingress            templateFile
// 	DockerConfigSecret templateFile
// 	Secret             templateFile
// 	Namespace          templateFile
// 	ConfigMap          templateFile
// }{
// 	ExternalNameSvc:    "./corev1/external-name-service.tpl.yml",
// 	Ingress:            "./corev1/ingress.yml.tpl",
// 	DockerConfigSecret: "./corev1/docker-config-secret.tpl.yml",
// 	Secret:             "./corev1/secret.tpl.yml",
// 	Namespace:          "./corev1/namespace.yml.tpl",
// 	ConfigMap:          "./corev1/configmap.yml.tpl",
// }
//
// var CrdsV1 = struct {
// 	App           templateFile
// 	AccountRouter templateFile
// }{
// 	App:           "./app-n-lambda.tpl.yml",
// 	AccountRouter: "./crdsv1/account-router.yml.tpl",
// }
