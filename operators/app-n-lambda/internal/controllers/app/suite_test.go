package app

import (
	"context"
	"operators.kloudlite.io/operators/app-n-lambda/internal/env"
	"operators.kloudlite.io/pkg/logging"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	artifactsv1 "operators.kloudlite.io/apis/artifacts/v1"
	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	. "operators.kloudlite.io/testing"
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}

var schemes = AddToSchemes(crdsv1.AddToScheme, artifactsv1.AddToScheme)
var reconciler *Reconciler

var _ = BeforeSuite(
	func() {
		SetupKubernetes(AddToSchemes(crdsv1.AddToScheme, artifactsv1.AddToScheme), LocalProxyEnvTest)
		setupNs()
		setupApp()
		mgr := Suite.GetManager(manager.Options{Namespace: testNamespace})

		reconciler = &Reconciler{
			Client: Suite.K8sClient,
			Scheme: Suite.Scheme,
			Env: &env.Env{
				ReconcilePeriod:         30 * time.Second,
				MaxConcurrentReconciles: 1,
			},
			logger: logging.NewOrDie(&logging.Options{
				Name: "app",
				Dev:  true,
			}),
			Name:       "app",
			yamlClient: Suite.K8sYamlClient,
		}

		err := reconciler.SetupWithManager(mgr, reconciler.logger)
		Expect(err).NotTo(HaveOccurred())

		ctx, cancel := context.WithCancel(context.TODO())
		go func() {
			defer GinkgoRecover()
			err = mgr.Start(ctx)
			Expect(err).ToNot(HaveOccurred(), "failed to run manager")
		}()

		DeferCleanup(func() {
			cancel()
		})
	},
)
