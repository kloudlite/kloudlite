package testing

//import (
//	"context"
//	"k8s.io/apimachinery/pkg/runtime"
//	"k8s.io/client-go/rest"
//	"operators.kloudlite.io/pkg/kubectl"
//	"os"
//	"sigs.k8s.io/controller-runtime/pkg/client"
//	"sigs.k8s.io/controller-runtime/pkg/envtest"
//	logf "sigs.k8s.io/controller-runtime/pkg/log"
//	"sigs.k8s.io/controller-runtime/pkg/log/zap"
//	"sigs.k8s.io/controller-runtime/pkg/manager"
//
//	. "github.com/onsi/ginkgo/v2"
//	. "github.com/onsi/gomega"
//)
//
//var (
//	K8sClient     client.Client
//	TestEnv       *envtest.Environment
//	K8sYamlClient *kubectl.YAMLClient
//)
//
//type PreFn func() error
//type PostFn func() error
//
//func PreSuite(scheme *runtime.Scheme, fns ...PreFn) bool {
//	return BeforeSuite(
//		func() {
//			Expect(os.Setenv("USE_EXISTING_CLUSTER", "true")).To(Succeed())
//			logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
//
//			By("bootstrapping test environment")
//			TestEnv = &envtest.Environment{
//				Config: &rest.Config{Host: "localhost:8080"},
//				// CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
//				// ErrorIfCRDPathMissing: true,
//			}
//
//			cfg, err := TestEnv.Start()
//			Expect(err).NotTo(HaveOccurred())
//			Expect(cfg).NotTo(BeNil())
//
//			K8sClient, err = client.New(cfg, client.Options{Scheme: scheme})
//			Expect(err).NotTo(HaveOccurred())
//			Expect(K8sClient).NotTo(BeNil())
//
//			K8sYamlClient, err = kubectl.NewYAMLClient(cfg)
//			Expect(err).NotTo(HaveOccurred())
//			Expect(K8sYamlClient).NotTo(BeNil())
//
//			for i := range fns {
//				err = fns[i]()
//				Expect(err).NotTo(HaveOccurred())
//			}
//		},
//	)
//}
//
//func PostSuite(fns ...PostFn) bool {
//	AfterSuite(
//		func() {
//			for i := range fns {
//				err := fns[i]()
//				Expect(err).NotTo(HaveOccurred())
//			}
//			By("tearing down the test environment")
//			err := TestEnv.Stop()
//			Expect(err).NotTo(HaveOccurred())
//		},
//	)
//	return true
//}
//
//func SetupManager(config *rest.Config, scheme *runtime.Scheme, setupFn func(mgr manager.Manager) error) (context.Context, context.CancelFunc) {
//	mgr, err := manager.New(config, manager.Options{Scheme: scheme, MetricsBindAddress: "0"})
//	if err != nil {
//		panic(err)
//	}
//	if err := setupFn(mgr); err != nil {
//		panic(err)
//	}
//
//	ctx, cancel := context.WithCancel(context.TODO())
//	go func() {
//		if err := mgr.Start(ctx); err != nil {
//			panic(err)
//		}
//	}()
//	return ctx, cancel
//}
//
//func TearManager(cancel context.CancelFunc) {
//	cancel()
//}
//
//func SetupWithDefaults(scheme *runtime.Scheme) bool {
//	PreSuite(scheme)
//	PostSuite()
//	return true
//}
