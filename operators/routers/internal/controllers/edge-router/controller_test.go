package edgeRouter

import (
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/routers/internal/controllers"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	. "github.com/kloudlite/operator/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

func newEdgeRouterCR() *crdsv1.EdgeRouter {
	return &crdsv1.EdgeRouter{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sample-" + rand.String(10),
			Namespace: testNamespace,
		},
		Spec: crdsv1.EdgeRouterSpec{
			EdgeName:   "sample-en-" + rand.String(10),
			AccountRef: "kl-core",
			//DefaultSSLCert:  crdsv1.SSLCertRef{},
			NodeSelector:    nil,
			Tolerations:     nil,
			WildcardDomains: nil,
		},
	}
}

var _ = Describe("edge router controller [CREATE] says", Ordered, func() {
	er := newEdgeRouterCR()
	BeforeAll(func() {
		CreateResource(er)
		DeferCleanup(func() {
			DeleteResource(er)
		})
	})

	It("ensures finalizers are added to the resource", func() {
		Promise(func(g Gomega) {
			Reconcile(reconciler, fn.NN(er.Namespace, er.Name))
			var obj crdsv1.EdgeRouter
			err := Suite.K8sClient.Get(Suite.Context, fn.NN(er.Namespace, er.Name), &obj)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(len(obj.Finalizers)).To(BeNumerically(">=", 1))
			g.Expect(obj.Finalizers).To(ContainElement(constants.CommonFinalizer))
		})
	})

	It("certmanager cluster issuer has been created, and owned by edgeRouter", func() {
		Promise(func(g Gomega) {
			Reconcile(reconciler, fn.NN(er.Namespace, er.Name))

			var obj crdsv1.EdgeRouter
			err := Suite.K8sClient.Get(Suite.Context, fn.NN(er.Namespace, er.Name), &obj)
			g.Expect(err).NotTo(HaveOccurred())

			var ci certmanagerv1.ClusterIssuer
			err = Suite.K8sClient.Get(Suite.Context, fn.NN(er.Namespace, controllers.GetClusterIssuerName(er.Spec.EdgeName)), &ci)
			g.Expect(err).NotTo(HaveOccurred())
		})
	})

	It("ingress nginx controller should be created, and owned by edgeRouter", func() {
		Promise(func(g Gomega) {
			Reconcile(reconciler, fn.NN(er.Namespace, er.Name))

			var obj crdsv1.EdgeRouter
			err := GetResource(fn.NN(er.Namespace, er.Name), &obj)
			g.Expect(err).NotTo(HaveOccurred())

			nginxC := fn.NewUnstructured(constants.HelmIngressNginx)
			err = GetResource(fn.NN(er.Namespace, er.Name), nginxC)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(fn.IsOwner(nginxC, fn.AsOwner(&obj))).To(BeTrue())
		})
	})

	It("edge router `.status.isReady` should be true", func() {
		Promise(func(g Gomega) {
			Reconcile(reconciler, fn.NN(er.Namespace, er.Name))

			var obj crdsv1.EdgeRouter
			err := GetResource(fn.NN(er.Namespace, er.Name), &obj)
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(obj.Status.IsReady).To(BeTrue())
		})
	})
})

var _ = Describe("edge router controller [UPDATE] says", Ordered, func() {
})

var _ = Describe("edge router controller [DELETE] says", Ordered, func() {
	When("edge router `status.isReady` is true", func() {
		var er *crdsv1.EdgeRouter
		//er := newEdgeRouterCR()

		BeforeEach(func() {
			er = newEdgeRouterCR()
			CreateResource(er)
			Promise(func(g Gomega) {
				Reconcile(reconciler, fn.NN(er.Namespace, er.Name))

				var obj crdsv1.EdgeRouter
				err := GetResource(fn.NN(er.Namespace, er.Name), &obj)
				g.Expect(err).NotTo(HaveOccurred())

				g.Expect(obj.Status.IsReady).To(BeTrue())
			})
		})

		It("ingress nginx controller has been deleted prior to edge router", func() {
			Expect(Suite.K8sClient.Delete(Suite.Context, er))
			Promise(func(g Gomega) {
				Reconcile(reconciler, fn.NN(er.Namespace, er.Name))

				var obj crdsv1.EdgeRouter
				err := GetResource(fn.NN(er.Namespace, er.Name), &obj)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(obj.GetDeletionTimestamp()).NotTo(BeNil())

				nginx := fn.NewUnstructured(constants.HelmIngressNginx)
				err = GetResource(fn.NN(er.Namespace, er.Name), nginx)
				g.Expect(apiErrors.IsNotFound(err)).To(BeTrue())
			}, "50s")
		})

		It("cluster issuer should be deleted prior to edge router", func() {
			//er = newEdgeRouterCR()
			Expect(Suite.K8sClient.Delete(Suite.Context, er))
			Promise(func(g Gomega) {
				Reconcile(reconciler, fn.NN(er.Namespace, er.Name))

				var obj crdsv1.EdgeRouter
				err := GetResource(fn.NN(er.Namespace, er.Name), &obj)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(obj.GetDeletionTimestamp()).NotTo(BeNil())

				var ci certmanagerv1.ClusterIssuer
				err = Suite.K8sClient.Get(Suite.Context, fn.NN("", controllers.GetClusterIssuerName(er.Spec.EdgeName)), &ci)
				g.Expect(apiErrors.IsNotFound(err)).To(BeTrue())
			})
		})
	})
})
