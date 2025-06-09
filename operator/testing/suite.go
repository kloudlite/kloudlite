package testing

import (
	"context"
	"os"
	"path/filepath"

	"github.com/kloudlite/operator/pkg/kubectl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/discovery"
	k8sScheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var Suite struct {
	K8sClient     client.Client
	Config        *rest.Config
	Scheme        *runtime.Scheme
	K8sYamlClient kubectl.YAMLClient
	NewManager    func(options manager.Options) manager.Manager
	Manager       manager.Manager
	Context       context.Context
	CancelFunc    context.CancelFunc
}

func AddToSchemes(fns ...func(s *runtime.Scheme) error) *runtime.Scheme {
	scheme := runtime.NewScheme()
	utilruntime.Must(k8sScheme.AddToScheme(scheme))
	for i := range fns {
		utilruntime.Must(fns[i](scheme))
	}
	return scheme
}

var LocalProxyEnvTest = &envtest.Environment{
	Config:                &rest.Config{Host: "localhost:8080"},
	BinaryAssetsDirectory: filepath.Join(os.Getenv("PROJECT_ROOT"), "bin", "k8s", "1.26.0-linux-amd64"),
	CRDDirectoryPaths:     []string{filepath.Join(os.Getenv("PROJECT_ROOT"), "config", "crd", "bases")},
}

var DefaultEnvTest = &envtest.Environment{
	CRDDirectoryPaths:     []string{filepath.Join(os.Getenv("PROJECT_ROOT"), "config", "crd", "bases")},
	BinaryAssetsDirectory: filepath.Join(os.Getenv("PROJECT_ROOT"), "bin", "k8s", "1.26.0-linux-amd64"),
}

func withoutManager() {
	c, err := client.New(Suite.Config, client.Options{
		Scheme: Suite.Scheme,
		Mapper: nil,
		WarningHandler: client.WarningHandlerOptions{
			SuppressWarnings:   true,
			AllowDuplicateLogs: false,
		},
	})
	Expect(err).NotTo(HaveOccurred())

	Suite.K8sClient = c

	k8sYamlClient, err := kubectl.NewYAMLClient(Suite.Config, kubectl.YAMLClientOpts{})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sYamlClient).NotTo(BeNil())
	Suite.K8sYamlClient = k8sYamlClient

	dc, err := discovery.NewDiscoveryClientForConfig(Suite.Config)
	Expect(err).ToNot(HaveOccurred())

	_, err = dc.ServerVersion()
	Expect(err).ToNot(HaveOccurred())
}

func SetupKubernetes(scheme *runtime.Scheme, envTest *envtest.Environment) {
	if envTest == LocalProxyEnvTest {
		Expect(os.Setenv("USE_EXISTING_CLUSTER", "true")).To(Succeed())
	}
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
	rest.SetDefaultWarningHandler(rest.NoWarnings{})

	envTest.Scheme = scheme

	By("bootstrapping test environment")
	_, err := envTest.Start()
	Expect(err).NotTo(HaveOccurred())
	Suite.Config = envTest.Config

	DeferCleanup(func() {
		err := envTest.Stop()
		Expect(err).NotTo(HaveOccurred())
	})

	Suite.Scheme = scheme
	Suite.Context, Suite.CancelFunc = context.WithCancel(context.TODO())

	withoutManager()
}
