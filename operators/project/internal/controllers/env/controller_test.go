package env

import (
	"github.com/kloudlite/operator/operators/project/internal/env"
	"github.com/kloudlite/operator/pkg/logging"
	. "github.com/kloudlite/operator/testing"
	. "github.com/onsi/ginkgo/v2"
	"k8s.io/apimachinery/pkg/util/rand"
	"time"
)

var _ = Describe("env controller says", func() {
	var test struct {
		Reconciler Reconciler
		Namespace  string
	}

	BeforeEach(func() {
		test.Namespace = "ginkgo-test-env" + rand.String(5)
		test.Reconciler = Reconciler{
			Client: Suite.K8sClient,
			Scheme: Suite.Scheme,
			Env: &env.Env{
				ReconcilePeriod:         30 * time.Second,
				MaxConcurrentReconciles: 1,

				ProjectCfgName:    "project-config",
				DockerSecretName:  "harbor-docker-secret",
				AdminRoleName:     "harbor-admin-role",
				SvcAccountName:    "kloudlite-svc-account",
				AccountRouterName: "account-router",
			},
			logger: logging.NewOrDie(&logging.Options{
				Name: "project",
				Dev:  true,
			}),
			Name:       "project",
			yamlClient: Suite.K8sYamlClient,
			IsDev:      true,
		}
	})

	BeforeEach(func() {
	})

	It("creates namespace", func() {
	})
})
