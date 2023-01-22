package secret

import (
	"context"
	"fmt"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	. "github.com/kloudlite/operator/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func newSecretCR() crdsv1.Secret {
	return crdsv1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "crds.kloudlite.i/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sample-" + rand.String(12),
			Namespace: "default",
			Labels: map[string]string{
				"kloudlite.io/label1": "value1",
			},
		},
		Data: map[string][]byte{
			"k1": []byte("v1"),
		},
	}
}

var _ = Describe("secret controller [CREATE] says", func() {
	ctx := context.TODO()

	It("ensures resource has finalizers attached to it", func() {
		secret := newSecretCR()
		CreateResource(&secret)

		Promise(func(g Gomega) {
			ReconcileForObject(reconciler, &secret)
			var obj crdsv1.Secret
			err := Suite.K8sClient.Get(ctx, fn.NN(secret.Namespace, secret.Name), &obj)
			g.Expect(err).To(BeNil())
			g.Expect(len(obj.GetFinalizers())).To(BeNumerically(">=", 1))

			DeleteResource(&secret)
		})
	})

	It("labels should be copied to real k8s secret", func() {
		secret := newSecretCR()
		CreateResource(&secret)

		Promise(func(g Gomega) {
			ReconcileForObject(reconciler, &secret)
			var obj crdsv1.Secret
			err := Suite.K8sClient.Get(ctx, fn.NN(secret.Namespace, secret.Name), &obj)
			g.Expect(err).To(BeNil())

			var scrt corev1.Secret
			err = Suite.K8sClient.Get(ctx, fn.NN(secret.Namespace, secret.Name), &scrt)
			g.Expect(err).To(BeNil())

			g.Expect(fn.MapContains(scrt.Labels, obj.Labels)).To(BeTrue())
			DeleteResource(&secret)
		})
	})

	It("data should be copied to real k8s secret", func() {
		secret := newSecretCR()
		CreateResource(&secret)

		Promise(func(g Gomega) {
			ReconcileForObject(reconciler, &secret)
			var obj crdsv1.Secret
			err := Suite.K8sClient.Get(ctx, fn.NN(secret.Namespace, secret.Name), &obj)
			g.Expect(err).To(BeNil())

			var scrt corev1.Secret
			err = Suite.K8sClient.Get(ctx, fn.NN(secret.Namespace, secret.Name), &scrt)
			g.Expect(err).To(BeNil())
			g.Expect(scrt.Data).To(Equal(obj.Data))

			DeleteResource(&secret)
		})
	})

	It("secret.status.isReady should be true", func() {
		secret := newSecretCR()
		CreateResource(&secret)

		Promise(func(g Gomega) {
			ReconcileForObject(reconciler, &secret)
			var obj crdsv1.Secret
			err := Suite.K8sClient.Get(ctx, fn.NN(secret.Namespace, secret.Name), &obj)
			g.Expect(err).To(BeNil())
			g.Expect(obj.Status.IsReady).To(BeTrue())

			DeleteResource(&secret)
		})
	})
})

var _ = Describe("secret controller [UPDATE] says", func() {
	ctx := context.TODO()

	Context("updating data, should reflect in k8s secret", func() {
		It("updating data to nil", func() {
			secret := newSecretCR()
			CreateResource(&secret)
			DeferCleanup(func() {
				DeleteResource(&secret)
			})

			_, err := controllerutil.CreateOrUpdate(ctx, Suite.K8sClient, &secret, func() error {
				secret.Data = nil
				return nil
			})
			if err != nil {
				if !apiErrors.IsConflict(err) {
					Expect(err).NotTo(HaveOccurred())
				}
			}

			Promise(func(g Gomega) {
				ReconcileForObject(reconciler, &secret)
				var obj corev1.Secret
				err = Suite.K8sClient.Get(ctx, fn.NN(secret.Namespace, secret.Name), &obj)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(obj.ResourceVersion).ToNot(BeEmpty())
				g.Expect(obj.Data).To(BeNil())
			})
		})

		It("updating data by adding new KV", func() {
			secret := newSecretCR()
			CreateResource(&secret)

			DeferCleanup(func() {
				DeleteResource(&secret)
			})

			_, err := controllerutil.CreateOrUpdate(ctx, Suite.K8sClient, &secret, func() error {
				secret.Data["xk1"] = []byte("xv2")
				return nil
			})
			if err != nil {
				if !apiErrors.IsConflict(err) {
					Expect(err).NotTo(HaveOccurred())
				}
			}

			Promise(func(g Gomega) {
				ReconcileForObject(reconciler, &secret)
				var obj corev1.Secret
				err = Suite.K8sClient.Get(ctx, fn.NN(secret.Namespace, secret.Name), &obj)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(obj.Data).To(Equal(secret.Data))
			}, "30s")
		})
	})

	It(fmt.Sprintf("clears status if annotation (`%s`) is on the resource", constants.ClearStatusKey), func() {
		config := newSecretCR()
		_, err := controllerutil.CreateOrUpdate(ctx, Suite.K8sClient, &config, func() error {
			config.Annotations = map[string]string{constants.ClearStatusKey: "true"}
			return nil
		})
		Expect(err).NotTo(HaveOccurred())

		Promise(func(g Gomega) {
			ReconcileForObject(reconciler, &config)
			var obj crdsv1.Secret
			err := Suite.K8sClient.Get(ctx, fn.NN(config.Namespace, config.Name), &obj)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(obj.Annotations).NotTo(HaveKey(constants.ClearStatusKey))
		})
	})
})

var _ = Describe("secret controller [DELETE] says", func() {
	ctx := context.TODO()

	Context("Deleting a resource", func() {
		It("finalizers get cleaned", func() {
			config := newSecretCR()
			config.ObjectMeta.Finalizers = append(config.ObjectMeta.Finalizers, "ginkgo-test")

			CreateResource(&config)
			Promise(func(g Gomega) {
				ReconcileForObject(reconciler, &config)
				var obj crdsv1.Secret
				err := Suite.K8sClient.Get(ctx, fn.NN(config.Namespace, config.Name), &obj)
				g.Expect(err).NotTo(HaveOccurred())

				g.Expect(Suite.K8sClient.Delete(ctx, &obj)).NotTo(HaveOccurred())

				err = Suite.K8sClient.Get(ctx, fn.NN(config.Namespace, config.Name), &obj)
				g.Expect(err).NotTo(HaveOccurred())

				g.Expect(obj.GetDeletionTimestamp()).NotTo(BeNil())
				g.Expect(len(obj.Finalizers)).To(Equal(1))
			})
		})

		It("k8s configmap is deleted", func() {
			config := newSecretCR()
			config.ObjectMeta.Finalizers = append(config.ObjectMeta.Finalizers, "ginkgo-test")

			CreateResource(&config)
			Promise(func(g Gomega) {
				ReconcileForObject(reconciler, &config)
				var obj crdsv1.Secret
				err := Suite.K8sClient.Get(ctx, fn.NN(config.Namespace, config.Name), &obj)
				g.Expect(err).NotTo(HaveOccurred())

				g.Expect(Suite.K8sClient.Delete(ctx, &obj)).NotTo(HaveOccurred())

				var scrt corev1.Secret
				err = Suite.K8sClient.Get(ctx, fn.NN(config.Namespace, config.Name), &scrt)
				g.Expect(apiErrors.IsNotFound(err)).To(BeTrue())
			})
		})
	})
})
