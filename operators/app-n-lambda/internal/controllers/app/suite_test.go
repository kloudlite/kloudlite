package app_test

import (
	"testing"
	"time"

	"github.com/kloudlite/operator/operators/app-n-lambda/internal/controllers/app"
	"github.com/kloudlite/operator/operators/app-n-lambda/internal/env"
	"github.com/kloudlite/operator/pkg/logging"

	// artifactsv1 "github.com/kloudlite/operator/apis/artifacts/v1"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	. "github.com/kloudlite/operator/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}

var schemes = AddToSchemes(crdsv1.AddToScheme)
var reconciler *app.Reconciler

var _ = BeforeSuite(
	func() {
		SetupKubernetes(AddToSchemes(crdsv1.AddToScheme), DefaultEnvTest)
		reconciler = &app.Reconciler{
			Client: Suite.K8sClient,
			Scheme: Suite.Scheme,
			Env: &env.Env{
				ReconcilePeriod:         30 * time.Second,
				MaxConcurrentReconciles: 1,
			},
			Logger: logging.NewOrDie(&logging.Options{
				Name: "app",
				Dev:  true,
			}),
			Name:       "app",
			YamlClient: *Suite.K8sYamlClient,
		}
	},
)
