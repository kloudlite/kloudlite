package driver

import (
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/csi-drivers/internal/env"
	"github.com/kloudlite/operator/pkg/logging"
	. "github.com/kloudlite/operator/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"testing"
	"time"
)

func TestController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}

var reconciler *Reconciler

var testNamespace = "sample-" + rand.String(10)

var _ = BeforeSuite(func() {
	DefaultEnvTest.CRDDirectoryPaths = append(DefaultEnvTest.CRDDirectoryPaths,
		"./testdata/helm-operator-crds.yml",
	)

	SetupKubernetes(AddToSchemes(crdsv1.AddToScheme, certmanagerv1.AddToScheme), DefaultEnvTest)

	ns := corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: testNamespace}}
	Expect(Suite.K8sClient.Create(Suite.Context, &ns)).NotTo(HaveOccurred())

	reconciler = &Reconciler{
		Client: Suite.K8sClient,
		Scheme: Suite.Scheme,
		logger: logging.NewOrDie(&logging.Options{Name: "edge-router", Dev: true}),
		Name:   "edge-router",
		Env: &env.Env{
			ReconcilePeriod:         30 * time.Second,
			MaxConcurrentReconciles: 10,
		},
		yamlClient: Suite.K8sYamlClient,
	}
})
