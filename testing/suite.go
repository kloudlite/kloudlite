package testing

import (
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	k8sScheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"operators.kloudlite.io/pkg/kubectl"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
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
	K8sYamlClient *kubectl.YAMLClient
	GetManager    func(options manager.Options) manager.Manager
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
	Config: &rest.Config{Host: "localhost:8080"},
}

func SetupKubernetes(scheme *runtime.Scheme, envTest *envtest.Environment) {
	Expect(os.Setenv("USE_EXISTING_CLUSTER", "true")).To(Succeed())
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	//TestEnv = &envtest.Environment{
	//	Config: &rest.Config{Host: "localhost:8080"},
	//	// CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
	//	// ErrorIfCRDPathMissing: true,
	//}

	cfg, err := envTest.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())
	Suite.Config = cfg

	DeferCleanup(func() {
		err := envTest.Stop()
		Expect(err).NotTo(HaveOccurred())
	})

	Suite.Scheme = scheme

	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())
	Suite.K8sClient = k8sClient

	k8sYamlClient, err := kubectl.NewYAMLClient(cfg)
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sYamlClient).NotTo(BeNil())
	Suite.K8sYamlClient = k8sYamlClient

	Suite.GetManager = func(opts manager.Options) manager.Manager {
		opts.Scheme = scheme
		opts.MetricsBindAddress = "0"
		mgr, err := ctrl.NewManager(cfg, opts)
		Expect(err).NotTo(HaveOccurred())
		return mgr
	}
}
