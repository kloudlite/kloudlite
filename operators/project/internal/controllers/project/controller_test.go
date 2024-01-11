package project

import (
	_ "fmt"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	fn "github.com/kloudlite/operator/pkg/functions"
	. "github.com/kloudlite/operator/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("project controller [CREATE] says", func() {
	proj := &crdsv1.Project{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "crds.kloudite.io/v1",
			Kind:       "Project",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "sample",
		},
		Spec: crdsv1.ProjectSpec{
			TargetNamespace: "sample",
		},
	}

	BeforeEach(func() {
		CreateResource(proj)
	})

	It("creates/updates target namespace, with labels for account and cluster", func(ctx SpecContext) {
		Promise(func(g Gomega) {
			Reconcile(reconciler, client.ObjectKeyFromObject(proj))
			var ns corev1.Namespace
			err := Suite.K8sClient.Get(ctx, fn.NN("", proj.Name), &ns)
			g.Expect(err).NotTo(HaveOccurred())
		}, "2s")
	})

	It("target namespace has labels for account, and cluster", func(ctx SpecContext) {
		Promise(func(g Gomega) {
			Reconcile(reconciler, client.ObjectKeyFromObject(proj))
			var ns corev1.Namespace
			err := Suite.K8sClient.Get(ctx, fn.NN("", proj.Name), &ns)
			g.Expect(err).NotTo(HaveOccurred())
		}, "2s")
	})

	It("creates service account in target namespace, and owns it", func(ctx SpecContext) {
		Promise(func(g Gomega) {
			Reconcile(reconciler, client.ObjectKeyFromObject(proj))
			var sa corev1.ServiceAccount
			err := Suite.K8sClient.Get(ctx, fn.NN(proj.Spec.TargetNamespace, reconciler.Env.SvcAccountName), &sa)
			g.Expect(err).NotTo(HaveOccurred())
			proj.EnsureGVK()
			g.Expect(sa.GetOwnerReferences()).To(Equal([]metav1.OwnerReference{fn.AsOwner(proj, true)}))
		}, "2s")
	})

	// It("project resource has .status.isReady set to true", func(ctx SpecContext) {
	// 	Promise(func(g Gomega) {
	// 		Reconcile(reconciler, client.ObjectKeyFromObject(proj))
	// 		var p crdsv1.Project
	// 		err := Suite.K8sClient.Get(ctx, client.ObjectKeyFromObject(proj), &p)
	// 		g.Expect(err).NotTo(HaveOccurred())
	// 		g.Expect(p.Status.IsReady).To(BeTrue())
	// 	}, "2s")
	// })
})
