package crds

import (
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func GetLogger(nn types.NamespacedName) *zap.SugaredLogger {
	cfg := zap.NewDevelopmentConfig()
	cfg.EncoderConfig.LineEnding = "\n\n"
	cfg.EncoderConfig.TimeKey = ""
	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	// logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	sugar := logger.Sugar()
	return sugar.With("REF", nn.String())
}

const ImagePullSecretName = "kloudlite-docker-registry"

const NamespaceAdminRole = "kloudlite-ns-admin"
const NamespaceAdminRoleBinding = "kloudlite-ns-admin"
const SvcAccountName = "kloudlite-svc-account"

var TypeSecret = metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"}
var TypeRole = metav1.TypeMeta{APIVersion: "rbac.authorization.k8s.io/v1", Kind: "Role"}
var TypeRoleBinding = metav1.TypeMeta{APIVersion: "rbac.authorization.k8s.io/v1", Kind: "RoleBinding"}
var TypeSvcAccount = metav1.TypeMeta{APIVersion: "v1", Kind: "ServiceAccount"}

var IngressAnnotations = map[string]string{
	"kubernetes.io/ingress.class":    "nginx",
	"cert-manager.io/cluster-issuer": "prod-cert-issuer",
}

type ManagedServiceType string

const (
	MongoDBStandalone ManagedServiceType = "MongoDBStandalone"
	MongoDBCluster    ManagedServiceType = "MongoDBCluster"
	ElasticSearch     ManagedServiceType = "ElasticSearch"
	MySqlStandalone   ManagedServiceType = "MySqlStandalone"
	MySqlCluster      ManagedServiceType = "MySqlCluster"
)

func (m ManagedServiceType) String() string {
	return string(m)
}
