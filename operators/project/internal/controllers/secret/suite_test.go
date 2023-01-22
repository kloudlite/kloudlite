package secret

import (
	"testing"
	"time"

	"github.com/kloudlite/operator/operators/project/internal/env"
	"github.com/kloudlite/operator/pkg/logging"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	. "github.com/kloudlite/operator/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}

var reconciler *Reconciler

var _ = BeforeSuite(
	func() {
		SetupKubernetes(AddToSchemes(crdsv1.AddToScheme), DefaultEnvTest)

		reconciler = &Reconciler{
			Client: Suite.K8sClient,
			Scheme: Suite.Scheme,
			Env: &env.Env{
				ReconcilePeriod:         1 * time.Second,
				MaxConcurrentReconciles: 10,
			},
			logger:     logging.NewOrDie(&logging.Options{Name: "secrets", Dev: true}),
			Name:       "secret",
			yamlClient: Suite.K8sYamlClient,
		}
	},
)
