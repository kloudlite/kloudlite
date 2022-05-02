package controllers

import (
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func GetLogger(nn types.NamespacedName) *zap.SugaredLogger {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	sugar := logger.Sugar()
	return sugar.With("REF", nn.String())
}

const ImagePullSecretName = "kloudlite-docker-registry"

const NamespaceAdminRole = "kloudlite-ns-admin"
const NamespaceAdminRoleBinding = "kloudlite-ns-admin"
const SvcAccountName = "kloudlite-svc-account"

var TypeNamespace = metav1.TypeMeta{APIVersion: "v1", Kind: "Namespace"}
var TypeSecret = metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"}
var TypeConfigMap = metav1.TypeMeta{APIVersion: "v1", Kind: "ConfigMap"}
var TypeRole = metav1.TypeMeta{APIVersion: "rbac.authorization.k8s.io/v1", Kind: "Role"}
var TypeIngress = metav1.TypeMeta{APIVersion: "networking.k8s.io", Kind: "Ingress"}
var TypeRoleBinding = metav1.TypeMeta{APIVersion: "rbac.authorization.k8s.io/v1", Kind: "RoleBinding"}
var TypeSvcAccount = metav1.TypeMeta{APIVersion: "v1", Kind: "ServiceAccount"}

var IngressAnnotations = map[string]string{
	"kubernetes.io/ingress.class":    "nginx",
	"cert-manager.io/cluster-issuer": "prod-cert-issuer",
}
