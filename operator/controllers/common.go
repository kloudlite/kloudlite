package controllers

import (
	"fmt"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetLogger(name types.NamespacedName) *zap.SugaredLogger {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	sugar := logger.Sugar()
	return sugar.With(
		"NAME", name.String(),
	)
}

const maxCoolingTime = 5
const minCoolingTime = 2
const semiCoolingTime = 2

const NamespaceAdminRole = "kloudlite-ns-admin"
const NamespaceAdminRoleBinding = "kloudlite-ns-admin"
const SvcAccountName = "kloudlite-svc-account"

const foregroundFinalizer = "foreground"

var TypeNamespace = metav1.TypeMeta{APIVersion: "v1", Kind: "Namespace"}
var TypeSecret = metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"}
var TypeConfigMap = metav1.TypeMeta{APIVersion: "v1", Kind: "ConfigMap"}
var TypeRole = metav1.TypeMeta{APIVersion: "rbac.authorization.k8s.io/v1", Kind: "Role"}
var TypeRoleBinding = metav1.TypeMeta{APIVersion: "rbac.authorization.k8s.io/v1", Kind: "RoleBinding"}
var TypeSvcAccount = metav1.TypeMeta{APIVersion: "v1", Kind: "ServiceAccount"}

func toRefString(k client.Object) string {
	return fmt.Sprintf("%s/%s", k.GetNamespace(), k.GetName())
}
