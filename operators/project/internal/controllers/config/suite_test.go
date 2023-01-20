package config

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/kloudlite/operator/operators/project/internal/env"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/logging"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	. "github.com/kloudlite/operator/testing"
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}

var reconciler *Reconciler

func doReconcile(config *crdsv1.Config) reconcile.Result {
	fmt.Println("Hey, i have been called")
	result, err := reconciler.Reconcile(Suite.Context, reconcile.Request{
		NamespacedName: fn.NN(config.Namespace, config.Name),
	})

	Expect(err).ToNot(HaveOccurred(), func() string {
		var t interface{ StackTrace() errors.StackTrace }
		if errors.As(err, &t) {
			return fmt.Sprintf("[partial] error trace:%+v\n", t.StackTrace()[:1])
		}
		return ""
	})

	return result
}

var _ = BeforeSuite(
	func() {
		SetupKubernetes(AddToSchemes(crdsv1.AddToScheme), LocalProxyEnvTest)

		reconciler = &Reconciler{
			Client: Suite.K8sClient,
			Scheme: Suite.Scheme,
			Env: &env.Env{
				ReconcilePeriod:         1 * time.Second,
				MaxConcurrentReconciles: 10,
			},
			logger:     logging.NewOrDie(&logging.Options{Name: "app", Dev: true}),
			Name:       "app",
			yamlClient: Suite.K8sYamlClient,
		}

		//err := reconciler.SetupWithManager(Suite.Manager, reconciler.logger)
		//Expect(err).NotTo(HaveOccurred())
		//
		//ctx, cancel := context.WithCancel(context.TODO())
		//go func() {
		//	defer GinkgoRecover()
		//	err = Suite.Manager.Start(ctx)
		//	Expect(err).ToNot(HaveOccurred(), "failed to run manager")
		//}()
		//
		//DeferCleanup(func() {
		//	cancel()
		//})
	},
)
