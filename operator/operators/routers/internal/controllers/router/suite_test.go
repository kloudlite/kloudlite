package router

import (
	"os"
	"testing"
	"time"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"github.com/kloudlite/operator/pkg/logging"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/routers/internal/env"
	. "github.com/kloudlite/operator/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

func TestController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}

var reconciler *Reconciler

var testNamespace = "sample-" + rand.String(10)

var _ = BeforeSuite(func() {
	SetupKubernetes(AddToSchemes(crdsv1.AddToScheme, certmanagerv1.AddToScheme), DefaultEnvTest)

	b, err := os.ReadFile("./testdata/certmanager-crds.yml")
	Expect(err).NotTo(HaveOccurred())

	Expect(Suite.K8sYamlClient.ApplyYAML(Suite.Context, b)).NotTo(HaveOccurred())

	ns := corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: testNamespace}}
	Expect(Suite.K8sClient.Create(Suite.Context, &ns)).NotTo(HaveOccurred())

	reconciler = &Reconciler{
		Client: Suite.K8sClient,
		Scheme: Suite.Scheme,
		logger: logging.NewOrDie(&logging.Options{Name: "router", Dev: true}),
		Name:   testNamespace,
		Env: &env.Env{
			ReconcilePeriod:          30 * time.Second,
			MaxConcurrentReconciles:  10,
			DefaultClusterIssuerName: "kl-cert-issuer",
			AcmeEmail:                "sample@gmail.com",
		},
		yamlClient: Suite.K8sYamlClient,
	}
})
