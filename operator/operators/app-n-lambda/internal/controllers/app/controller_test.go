package app

import (
	"context"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	v1 "operators.kloudlite.io/apis/crds/v1"
	fn "operators.kloudlite.io/pkg/functions"
	. "operators.kloudlite.io/testing"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"
)

var testNamespace = "ginkgo-testing-" + rand.String(5)
var testName = "test-app-" + rand.String(5)

var yamlApp = fmt.Sprintf(`
---
apiVersion: "crds.kloudlite.io/v1"
kind: App
metadata:
  name: %s
  namespace: %s
  labels:
    kloudlite.io/cluster: "cluster"
    kloudlite.io/region: "region"
    kloudlite.io/account-ref: "account-ref"
spec:
  region: master
  services:
    - port: 80
      targetPort: 3000
      type: tcp
  containers:
    - name: samplex
      image: nginx:latest
      resourceCpu:
        min: "20m"
        max: "30m"
      resourceMemory:
        min: "30Mi"
        max: "50Mi"
`, testName, testNamespace)

func setupNs() {
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: testNamespace}}
	_, err := controllerutil.CreateOrUpdate(context.TODO(), Suite.K8sClient, ns, func() error {
		return nil
	})
	Expect(err).NotTo(HaveOccurred())

	DeferCleanup(func() {
		err = Suite.K8sClient.Delete(context.TODO(), ns)
		Expect(err).NotTo(HaveOccurred())
	})
}

func setupApp() {
	err := Suite.K8sYamlClient.ApplyYAML(context.TODO(), []byte(yamlApp))
	Expect(err).NotTo(HaveOccurred())

	DeferCleanup(func() {
		err := Suite.K8sYamlClient.DeleteYAML(context.TODO(), []byte(yamlApp))
		Expect(err).NotTo(HaveOccurred())
	})
}

type test struct {
	Reconciler   *Reconciler
	AppNamespace string
	AppName      string
	App          *v1.App
}

var _ = Describe("app controller says", func() {
	var t test
	BeforeEach(func() {
		t = test{
			Reconciler:   reconciler,
			AppNamespace: testNamespace,
			AppName:      testName,
		}
	})

	BeforeEach(func() {
		app := &v1.App{}
		err := reconciler.Get(context.TODO(), fn.NN(t.AppNamespace, t.AppName), app)
		Expect(err).NotTo(HaveOccurred())
		t.App = app
	})

	Context("deployment created with same name as App", func() {
		deployment := &appsv1.Deployment{}
		BeforeEach(func() {
			Eventually(func(g Gomega) {
				err := t.Reconciler.Get(context.TODO(), fn.NN(t.AppNamespace, t.AppName), deployment)
				Expect(err).NotTo(HaveOccurred())
			}).WithPolling(1 * time.Second).WithTimeout(20 * time.Second).Should(Succeed())
		})

		It("pods must have label 'app: <app-name>', for loki and prometheus to scrape", func() {
			Expect(deployment.Spec.Template.Labels["app"]).To(Equal(t.AppName))
		})

		It("All App Labels has been forwarded", func() {
			Expect(deployment.Labels["kloudlite.io/cluster"]).To(Equal("cluster"))
			Expect(deployment.Labels["kloudlite.io/region"]).To(Equal("region"))
			Expect(deployment.Labels["kloudlite.io/account-ref"]).To(Equal("account-ref"))
		})
	})

	Context("internal and an external service, named after app itself", func() {
		externalSvc := &corev1.Service{}
		internalSvc := &corev1.Service{}

		BeforeEach(func() {
			Eventually(func(g Gomega) {
				err := t.Reconciler.Get(context.TODO(), fn.NN(t.AppNamespace, t.AppName), externalSvc)
				g.Expect(err).NotTo(HaveOccurred())
				err = t.Reconciler.Get(context.TODO(), fn.NN(t.AppNamespace, t.AppName+"-internal"), internalSvc)
				g.Expect(err).NotTo(HaveOccurred())
			}).WithPolling(1 * time.Second).WithTimeout(10 * time.Second).Should(Succeed())
		})

		It("internal service should be pointing to pods created by deployment", func() {
			deployment := &appsv1.Deployment{}
			err := t.Reconciler.Get(context.TODO(), fn.NN(t.AppNamespace, t.AppName), deployment)
			Expect(err).NotTo(HaveOccurred())
		})

		It("internal service should have ports specified by app", func() {
			Expect(len(internalSvc.Spec.Ports)).To(Equal(len(t.App.Spec.Services)))
		})
	})

	Context("Horizontal Pod Autoscaler to be created, if hpa has been enabled on app", func() {
		It("", func() {
			Fail("empty test")
		})
	})

	Context("Freezing and Interception", func() {
		It("If App has been frozen, deployment should scale down to 0", func() {
			Fail("empty test")
		})
		It("If App has been intercepted, deployment should also scale down to 0", func() {
			Fail("empty test")
		})
		It("If App has been intercepted, external name service should be pointing to wireguard device", func() {
			Fail("empty test")
		})
	})

	It("if Deployment is ready, app status field should be set", func() {
		Fail("empty test")
	})
})
