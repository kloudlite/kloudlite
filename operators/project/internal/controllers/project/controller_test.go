package project

import (
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/pkg/constants"
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
			AccountName:     "sample",
			ClusterName:     "sample",
			DisplayName:     "Sample Website",
			TargetNamespace: "sample",
		},
	}

	BeforeEach(func() {
		CreateResource(proj)
	})

	It("creates/updates target namespace, with labels, and owner references", func(ctx SpecContext) {
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
			g.Expect(ns.Labels[constants.AccountNameKey]).To(Equal(proj.Spec.AccountName))
			g.Expect(ns.Labels[constants.ClusterNameKey]).To(Equal(proj.Spec.ClusterName))
		}, "2s")
	})

	It("target namespace has owner references to this project", func(ctx SpecContext) {
		Promise(func(g Gomega) {
			Reconcile(reconciler, client.ObjectKeyFromObject(proj))
			var ns corev1.Namespace
			err := Suite.K8sClient.Get(ctx, fn.NN("", proj.Name), &ns)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(len(ns.GetOwnerReferences())).To(Equal(1))
			g.Expect(ns.GetOwnerReferences()[0].UID).To(Equal(proj.UID))
		}, "2s")
	})

	It("project resource has .status.isReady set to true", func(ctx SpecContext) {
		Promise(func(g Gomega) {
			Reconcile(reconciler, client.ObjectKeyFromObject(proj))
			var p crdsv1.Project
			err := Suite.K8sClient.Get(ctx, client.ObjectKeyFromObject(proj), &p)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(p.Status.IsReady).To(BeTrue())
		}, "2s")
	})
})
