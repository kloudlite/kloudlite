package app_test

import (
	"context"
	"fmt"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	fn "github.com/kloudlite/operator/pkg/functions"
	. "github.com/kloudlite/operator/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

func newNamespace() v1.Namespace {
	return v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ginkgo-testing-" + rand.String(10),
		},
	}
}

func newApp(namespace string) crdsv1.App {
	return crdsv1.App{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rand.String(10),
			Namespace: namespace,
			Labels: map[string]string{
				"k1": "v1",
			},
		},
		Spec: crdsv1.AppSpec{
			AccountName:    "kloudlite-dev",
			Region:         "kloudlite-dev",
			ServiceAccount: "kloudlite-svc-account",
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
	ctx := context.TODO()

	var ns v1.Namespace
	var app crdsv1.App

	BeforeEach(func() {
		ns = newNamespace()
		CreateResource(&ns)

		app = newApp(ns.Name)
		CreateResource(&app)
	})

	It("has been created", func() {
		Promise(func(g Gomega) {
			ReconcileForObject(reconciler, &app)
			var obj crdsv1.App
			err := Suite.K8sClient.Get(ctx, fn.NN(app.Namespace, app.Name), &obj)
			g.Expect(err).To(BeNil())
		})
	})

	It("has finalizers attached to it", func() {
		Promise(func(g Gomega) {
			ReconcileForObject(reconciler, &app)
			var obj crdsv1.App
			err := Suite.K8sClient.Get(ctx, fn.NN(app.Namespace, app.Name), &obj)
			g.Expect(err).To(BeNil())
			g.Expect(len(obj.Finalizers)).To(BeNumerically(">=", 1))
		})
	})

	AfterAll(func() {
		DeleteResource(&app)
	})
})

var _ = Describe("app controller [DELETE] says", Ordered, func() {
	ctx := context.TODO()

	var ns v1.Namespace
	var app crdsv1.App

	BeforeAll(func() {
		ns = newNamespace()
		CreateResource(&ns)

		app = newApp(ns.Name)
		CreateResource(&app)
	})

	It("has many resources created", func() {
		Promise(func(g Gomega) {
			ReconcileForObject(reconciler, &app)
			var obj crdsv1.App
			err := Suite.K8sClient.Get(ctx, fn.NN(app.Namespace, app.Name), &obj)
			g.Expect(err).To(BeNil())
			fmt.Printf("%+v\n", obj.Status.Resources)
			g.Expect(len(obj.Status.Resources)).To(BeNumerically(">=", 1))
		})
	})
})
