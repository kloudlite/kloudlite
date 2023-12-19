package router

import (
	"regexp"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	. "github.com/kloudlite/operator/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	networkingv1 "k8s.io/api/networking/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func newRouter() crdsv1.Router {
	return crdsv1.Router{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sample-router-" + rand.String(10),
			Namespace: testNamespace,
		},
		Spec: crdsv1.RouterSpec{
			Region: "master",
			Https: &crdsv1.Https{
				Enabled: false,
			},
			// RateLimit:       crdsv1.RateLimit{},
			// MaxBodySizeInMB: 0,
			Domains: []string{"sample.example.com"},
			Routes: []crdsv1.Route{
				{
					App:     "example",
					Path:    "/",
					Port:    80,
					Rewrite: false,
				},
			},
			// BasicAuth:       crdsv1.BasicAuth{},
			// Cors:            &crdsv1.Cors{},
		},
	}
}

var _ = Describe("router controller [CREATE] says", Ordered, func() {
	routerCr := newRouter()

	It("should be created", func() {
		Expect(Suite.K8sClient.Create(Suite.Context, &routerCr)).NotTo(HaveOccurred())
	})

	It("should add some finalizers to the resource", func() {
		Promise(func(g Gomega) {
			Reconcile(reconciler, client.ObjectKeyFromObject(&routerCr))
			var obj crdsv1.Router
			err := Suite.K8sClient.Get(Suite.Context, fn.NN(routerCr.Namespace, routerCr.Name), &obj)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(len(obj.Finalizers)).To(BeNumerically(">=", 1))
			g.Expect(obj.Finalizers).To(ContainElement(constants.ForegroundFinalizer))
		})
	})

	It("should create a k8s ingress resource, owned by Router", func() {
		Promise(func(g Gomega) {
			Reconcile(reconciler, fn.NN(routerCr.Namespace, routerCr.Name))
			var ing networkingv1.Ingress
			err := Suite.K8sClient.Get(Suite.Context, fn.NN(routerCr.Namespace, routerCr.Name), &ing)
			g.Expect(err).NotTo(HaveOccurred())
			ownerRefs := ing.GetOwnerReferences()
			g.Expect(len(ownerRefs)).To(BeNumerically(">=", 1))
			g.Expect(ownerRefs[0].UID).To(Equal(routerCr.UID))
		})
	})
})

var _ = Describe("router controller [DELETE] says", Ordered, func() {
	routerCr := newRouter()

	BeforeAll(func() {
		Expect(Suite.K8sClient.Create(Suite.Context, &routerCr)).NotTo(HaveOccurred())
		Promise(func(g Gomega) {
			Reconcile(reconciler, fn.NN(routerCr.Namespace, routerCr.Name))
			var obj crdsv1.Router
			err := Suite.K8sClient.Get(Suite.Context, fn.NN(routerCr.Namespace, routerCr.Name), &obj)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(obj.Status.IsReady).To(BeTrue())

			var ing networkingv1.Ingress
			err = Suite.K8sClient.Get(Suite.Context, fn.NN(routerCr.Namespace, routerCr.Name), &ing)
			g.Expect(err).NotTo(HaveOccurred())
		})
	})

	It("k8s ingress resources owned by router, does not exist", func() {
		Expect(Suite.K8sClient.Delete(Suite.Context, &routerCr))
		Promise(func(g Gomega) {
			Reconcile(reconciler, client.ObjectKeyFromObject(&routerCr))

			var obj crdsv1.Router
			err := Suite.K8sClient.Get(Suite.Context, fn.NN(routerCr.Namespace, routerCr.Name), &obj)
			g.Expect(err).NotTo(HaveOccurred())

			var ing networkingv1.Ingress
			err = Suite.K8sClient.Get(Suite.Context, fn.NN(routerCr.Namespace, routerCr.Name), &ing)
			g.Expect(apiErrors.IsNotFound(err)).To(BeTrue())
		})
	})

	It("after ingress resource is deleted, router itself does not exist", func() {
		Promise(func(g Gomega) {
			Reconcile(reconciler, client.ObjectKeyFromObject(&routerCr))

			var obj crdsv1.Router
			err := Suite.K8sClient.Get(Suite.Context, fn.NN(routerCr.Namespace, routerCr.Name), &obj)
			g.Expect(apiErrors.IsNotFound(err)).To(BeTrue())
		})
	})
})

var _ = Describe("router controller [UPDATE] says", func() {
	routerCr := newRouter()
	BeforeEach(func() {
		CreateResource(&routerCr)
	})

	It("adding a new domain, reflects in each of the owned k8s ingresses", func() {
		_, err := controllerutil.CreateOrUpdate(Suite.Context, Suite.K8sClient, &routerCr, func() error {
			routerCr.Spec.Domains = append(routerCr.Spec.Domains, "dummy.example.com")
			return nil
		})
		Expect(err).NotTo(HaveOccurred())

		dMap := make(map[string]bool, len(routerCr.Spec.Domains))
		for i := range routerCr.Spec.Domains {
			dMap[routerCr.Spec.Domains[i]] = false
		}

		Promise(func(g Gomega) {
			Reconcile(reconciler, fn.NN(routerCr.Namespace, routerCr.Name))
			var ing networkingv1.Ingress
			err := Suite.K8sClient.Get(Suite.Context, fn.NN(routerCr.Namespace, routerCr.Name), &ing)
			g.Expect(err).NotTo(HaveOccurred())

			for i := range ing.Spec.Rules {
				dMap[ing.Spec.Rules[i].Host] = true
			}

			for s := range dMap {
				g.Expect(dMap[s]).To(BeTrue())
			}
		})
	})

	It("adding a new route, reflects in each of the owned k8s ingresses", func() {
		_, err := controllerutil.CreateOrUpdate(Suite.Context, Suite.K8sClient, &routerCr, func() error {
			routerCr.Spec.Routes = append(routerCr.Spec.Routes, crdsv1.Route{
				App:  "ginkgo-test",
				Path: "/.kl/test",
				Port: 80,
			})
			return nil
		})
		Expect(err).NotTo(HaveOccurred())

		pMap := make(map[string]bool, len(routerCr.Spec.Routes))
		for i := range routerCr.Spec.Routes {
			pMap[routerCr.Spec.Routes[i].Path+".*"] = false
		}

		Promise(func(g Gomega) {
			Reconcile(reconciler, fn.NN(routerCr.Namespace, routerCr.Name))
			var ing networkingv1.Ingress
			err := Suite.K8sClient.Get(Suite.Context, fn.NN(routerCr.Namespace, routerCr.Name), &ing)
			g.Expect(err).NotTo(HaveOccurred())

			re := regexp.MustCompile(`[()]`)
			for i := range ing.Spec.Rules[0].HTTP.Paths {
				p := re.ReplaceAllString(ing.Spec.Rules[0].HTTP.Paths[i].Path, "")
				if _, ok := pMap[p]; ok {
					pMap[re.ReplaceAllString(ing.Spec.Rules[0].HTTP.Paths[i].Path, "")] = true
				}
			}

			for i := range pMap {
				g.Expect(pMap[i]).To(BeTrue())
			}
		})
	})
})
