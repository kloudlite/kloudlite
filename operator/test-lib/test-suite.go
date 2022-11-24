package test_lib

import (
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	k8sScheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"operators.kloudlite.io/pkg/kubectl"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	K8sClient     client.Client
	testEnv       *envtest.Environment
	K8sYamlClient *kubectl.YAMLClient
)

type PreFn func() error
type PostFn func() error

func AddToSchemes(fns ...func(s *runtime.Scheme) error) *runtime.Scheme {
	scheme := runtime.NewScheme()
	utilruntime.Must(k8sScheme.AddToScheme(scheme))
	for i := range fns {
		utilruntime.Must(fns[i](scheme))
	}
	return scheme
}

func PreSuite(scheme *runtime.Scheme, fns ...PreFn) bool {
	return BeforeSuite(
		func() {
			Expect(os.Setenv("USE_EXISTING_CLUSTER", "true")).To(Succeed())
			logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

			By("bootstrapping test environment")
			testEnv = &envtest.Environment{
				Config: &rest.Config{Host: "localhost:8080"},
				// CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
				// ErrorIfCRDPathMissing: true,
			}

			cfg, err := testEnv.Start()
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg).NotTo(BeNil())

			K8sClient, err = client.New(cfg, client.Options{Scheme: scheme})
			Expect(err).NotTo(HaveOccurred())
			Expect(K8sClient).NotTo(BeNil())

			K8sYamlClient, err = kubectl.NewYAMLClient(cfg)
			Expect(err).NotTo(HaveOccurred())
			Expect(K8sYamlClient).NotTo(BeNil())

			for i := range fns {
				err = fns[i]()
				Expect(err).NotTo(HaveOccurred())
			}
		},
	)
}

func PostSuite(fns ...PostFn) bool {
	var _ = AfterSuite(
		func() {
			for i := range fns {
				err := fns[i]()
				Expect(err).NotTo(HaveOccurred())
			}
			By("tearing down the test environment")
			err := testEnv.Stop()
			Expect(err).NotTo(HaveOccurred())
		},
	)
	return true
}
