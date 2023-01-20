package env

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	artifactsv1 "github.com/kloudlite/operator/apis/artifacts/v1"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	. "github.com/kloudlite/operator/testing"
	"testing"
)

func TestController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	SetupKubernetes(AddToSchemes(crdsv1.AddToScheme, artifactsv1.AddToScheme), LocalProxyEnvTest)
})
