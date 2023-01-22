package project

import (
	"context"
	"github.com/go-logr/logr"
	artifactsv1 "github.com/kloudlite/operator/apis/artifacts/v1"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/project/internal/env"
	"github.com/kloudlite/operator/pkg/logging"
	. "github.com/kloudlite/operator/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/config/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"testing"
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}

var testProjectName = "ginkgo-project-test-" + rand.String(5)
var schemes = AddToSchemes(crdsv1.AddToScheme, artifactsv1.AddToScheme)
var reconciler *Reconciler

var _ = BeforeSuite(
	func() {
		SetupKubernetes(AddToSchemes(crdsv1.AddToScheme, artifactsv1.AddToScheme), LocalProxyEnvTest)
		//setupNs()
		//setupApp()
		mgr := Suite.NewManager(manager.Options{
			Scheme:                        nil,
			MapperProvider:                nil,
			SyncPeriod:                    nil,
			Logger:                        logr.Logger{},
			LeaderElection:                false,
			LeaderElectionResourceLock:    "",
			LeaderElectionNamespace:       "",
			LeaderElectionID:              "",
			LeaderElectionConfig:          nil,
			LeaderElectionReleaseOnCancel: false,
			LeaseDuration:                 nil,
			RenewDeadline:                 nil,
			RetryPeriod:                   nil,
			Namespace:                     "",
			MetricsBindAddress:            "",
			HealthProbeBindAddress:        "",
			ReadinessEndpointName:         "",
			LivenessEndpointName:          "",
			Port:                          0,
			Host:                          "",
			CertDir:                       "",
			WebhookServer:                 nil,
			NewCache:                      nil,
			NewClient:                     nil,
			BaseContext:                   nil,
			ClientDisableCacheFor:         nil,
			DryRunClient:                  false,
			EventBroadcaster:              nil,
			GracefulShutdownTimeout:       nil,
			Controller:                    v1alpha1.ControllerConfigurationSpec{},
		})

		reconciler = &Reconciler{
			Client: Suite.K8sClient,
			Scheme: Suite.Scheme,
			Env:    &env.Env{},
			logger: logging.NewOrDie(&logging.Options{
				Name: "project",
				Dev:  true,
			}),
			Name:       "project",
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
