package app_test

import (
	"strings"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	fn "github.com/kloudlite/operator/pkg/functions"
	. "github.com/kloudlite/operator/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func newNamespace() v1.Namespace {
	return v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ginkgo-testing-" + strings.ToLower(rand.String(10)),
		},
	}
}

func newApp(namespace string) crdsv1.App {
	return crdsv1.App{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ginkgo-test-app-" + strings.ToLower(rand.String(10)),
			Namespace: namespace,
			Labels: map[string]string{
				"k1": "v1",
			},
		},
		Spec: crdsv1.AppSpec{
			Region:         "kloudlite-dev",
			ServiceAccount: "default",
			Replicas:       1,
			Services: []crdsv1.AppSvc{
				{
					Port:       80,
					TargetPort: 3000,
					Type:       "tcp",
					Name:       "http",
				},
			},
			Containers: []crdsv1.AppContainer{
				{
					Name:  "main",
					Image: "nginx",
				},
			},
		},
	}
}

var _ = Describe("app controller [CREATE] says", Ordered, func() {
	ns := newNamespace()
	app := newApp(ns.Name)

	BeforeEach(func() {
		CreateResource(&ns)
		CreateResource(&app)
	})

	AfterAll(func() {
		DeleteResource(&app)
		DeleteResource(&ns)
	})

	It("deployment, internal and external services are created", func(ctx SpecContext) {
		Promise(func(g Gomega) {
			By("deployment has been created")
			Reconcile(reconciler, client.ObjectKeyFromObject(&app))
			var depl appsv1.Deployment
			err := Suite.K8sClient.Get(ctx, fn.NN(app.Namespace, app.Name), &depl)
			g.Expect(err).NotTo(HaveOccurred())

			By("deployment has labels from app")
			for k, v := range app.Labels {
				g.Expect(depl.Labels[k]).To(Equal(v))
			}

			By("deployment has ownership that points to app")
			g.Expect(len(depl.OwnerReferences)).To(Equal(1))
			g.Expect(depl.OwnerReferences[0].UID).To(Equal(app.UID))
		}, "2s")
	})

	It("deployment replicas are equal to app.replicas", func(ctx SpecContext) {
		Promise(func(g Gomega) {
			Reconcile(reconciler, client.ObjectKeyFromObject(&app))
			var depl appsv1.Deployment
			err := Suite.K8sClient.Get(ctx, fn.NN(app.Namespace, app.Name), &depl)
			g.Expect(err).NotTo(HaveOccurred())

			var tApp crdsv1.App
			err = Suite.K8sClient.Get(ctx, fn.NN(app.Namespace, app.Name), &tApp)
			g.Expect(err).NotTo(HaveOccurred())

			reconciler.Logger.Infof("depl replicas: %+v\n", depl.Status.Replicas)
			reconciler.Logger.Infof("app.StatusChecks %+v\n", tApp.Status.Checks)
			reconciler.Logger.Infof("depl.Status %+v\n", depl.Status)
			g.Expect(depl.Status.Replicas).To(Equal(app.Spec.Replicas))
		}, "20s")
	})
})

// var _ = Describe("app controller [DELETE] says", Ordered, func() {
// 	ctx := context.TODO()
//
// 	var ns v1.Namespace
// 	var app crdsv1.App
//
// 	BeforeAll(func() {
// 		ns = newNamespace()
// 		CreateResource(&ns)
//
// 		app = newApp(ns.Name)
// 		CreateResource(&app)
// 	})
//
// 	It("has many resources created", func() {
// 		Promise(func(g Gomega) {
// 			ReconcileForObject(reconciler, &app)
// 			var obj crdsv1.App
// 			err := Suite.K8sClient.Get(ctx, fn.NN(app.Namespace, app.Name), &obj)
// 			g.Expect(err).To(BeNil())
// 			fmt.Printf("%+v\n", obj.Status.Resources)
// 			g.Expect(len(obj.Status.Resources)).To(BeNumerically(">=", 1))
// 		})
// 	})
// })
