package config

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

func newConfigCR() crdsv1.Config {
	return crdsv1.Config{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Config",
			APIVersion: "crds.kloudlite.i/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sample-" + rand.String(12),
			Namespace: "default",
			Labels: map[string]string{
				"kloudlite.io/label1": "value1",
			},
		},
		Spec: crdsv1.ConfigSpec{
			Data: map[string]string{
				"k1": "v1",
			},
		},
		//Enabled:    false,
		//Overrides:  nil,
		//Status:     rApi.Status{},
	}
}

var _ = Describe("config controller [CREATE] says", func() {
	ctx := context.TODO()

	It("ensures resource has finalizers attached to it", func() {
		config := newConfigCR()
		CreateResource(&config)

		Promise(func(g Gomega) {
			ReconcileForObject(reconciler, &config)
			var obj crdsv1.Config
			err := Suite.K8sClient.Get(ctx, fn.NN(config.Namespace, config.Name), &obj)
			g.Expect(err).To(BeNil())
			g.Expect(len(obj.GetFinalizers())).To(BeNumerically(">=", 1))

			DeleteResource(&config)
		})
	})

	It("labels should be copied to real k8s config", func() {
		config := newConfigCR()
		CreateResource(&config)

		Promise(func(g Gomega) {
			ReconcileForObject(reconciler, &config)
			var obj crdsv1.Config
			err := Suite.K8sClient.Get(ctx, fn.NN(config.Namespace, config.Name), &obj)
			g.Expect(err).To(BeNil())

			var cfg corev1.ConfigMap
			err = Suite.K8sClient.Get(ctx, fn.NN(config.Namespace, config.Name), &cfg)
			g.Expect(err).To(BeNil())

			g.Expect(fn.MapContains(cfg.Labels, obj.Labels)).To(BeTrue())
			DeleteResource(&config)
		})
	})

	It("data should be copied to real k8s config", func() {
		config := newConfigCR()
		CreateResource(&config)

		Promise(func(g Gomega) {
			ReconcileForObject(reconciler, &config)
			var obj crdsv1.Config
			err := Suite.K8sClient.Get(ctx, fn.NN(config.Namespace, config.Name), &obj)
			g.Expect(err).To(BeNil())

			var cfg corev1.ConfigMap
			err = Suite.K8sClient.Get(ctx, fn.NN(config.Namespace, config.Name), &cfg)
			g.Expect(err).To(BeNil())
			g.Expect(cfg.Data).To(Equal(obj.Data))

			DeleteResource(&config)
		})
	})

	It("config.status.isReady should be true", func() {
		config := newConfigCR()
		CreateResource(&config)
		Promise(func(g Gomega) {
			ReconcileForObject(reconciler, &config)
			var obj crdsv1.Config
			err := Suite.K8sClient.Get(ctx, fn.NN(config.Namespace, config.Name), &obj)
			g.Expect(err).To(BeNil())
			g.Expect(obj.Status.IsReady).To(BeTrue())

			DeleteResource(&config)
		})
	})
})

var _ = Describe("config controller [UPDATE] says", func() {
	ctx := context.TODO()

	Context("updating data, should reflect in k8s config", func() {
		It("updating data to nil", func() {
			config := newConfigCR()
			CreateResource(&config)
			DeferCleanup(func() {
				DeleteResource(&config)
			})

			_, err := controllerutil.CreateOrUpdate(ctx, Suite.K8sClient, &config, func() error {
				config.Data = nil
				return nil
			})
			if err != nil {
				if !apiErrors.IsConflict(err) {
					Expect(err).NotTo(HaveOccurred())
				}
			}

			Promise(func(g Gomega) {
				ReconcileForObject(reconciler, &config)
				var cfg corev1.ConfigMap
				err = Suite.K8sClient.Get(ctx, fn.NN(config.Namespace, config.Name), &cfg)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(cfg.ResourceVersion).ToNot(BeEmpty())
				g.Expect(cfg.Data).To(BeNil())
			})
		})

		It("updating data by adding new KV", func() {
			config := newConfigCR()
			CreateResource(&config)

			DeferCleanup(func() {
				DeleteResource(&config)
			})

			_, err := controllerutil.CreateOrUpdate(ctx, Suite.K8sClient, &config, func() error {
				config.Data["xk1"] = "xv2"
				return nil
			})
			if err != nil {
				if !apiErrors.IsConflict(err) {
					Expect(err).NotTo(HaveOccurred())
				}
			}

			Promise(func(g Gomega) {
				ReconcileForObject(reconciler, &config)
				var cfg corev1.ConfigMap
				err = Suite.K8sClient.Get(ctx, fn.NN(config.Namespace, config.Name), &cfg)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(cfg.Data).To(Equal(config.Data))
			}, "30s")
		})
	})

	It(fmt.Sprintf("clears status if annotation (`%s`) is on the resource", constants.ClearStatusKey), func() {
		config := newConfigCR()
		_, err := controllerutil.CreateOrUpdate(ctx, Suite.K8sClient, &config, func() error {
			config.Annotations = map[string]string{constants.ClearStatusKey: "true"}
			return nil
		})
		Expect(err).NotTo(HaveOccurred())

		Promise(func(g Gomega) {
			ReconcileForObject(reconciler, &config)
			var obj crdsv1.Config
			err := Suite.K8sClient.Get(ctx, fn.NN(config.Namespace, config.Name), &obj)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(obj.Annotations).NotTo(HaveKey(constants.ClearStatusKey))
		})
	})
})

var _ = Describe("config controller [DELETE] says", func() {
	ctx := context.TODO()

	Context("Deleting a resource", func() {
		It("finalizers get cleaned", func() {
			config := newConfigCR()
			config.ObjectMeta.Finalizers = append(config.ObjectMeta.Finalizers, "ginkgo-test")

			CreateResource(&config)
			Promise(func(g Gomega) {
				ReconcileForObject(reconciler, &config)
				var obj crdsv1.Config
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
			config := newConfigCR()
			config.ObjectMeta.Finalizers = append(config.ObjectMeta.Finalizers, "ginkgo-test")

			CreateResource(&config)
			Promise(func(g Gomega) {
				ReconcileForObject(reconciler, &config)
				var cfg corev1.ConfigMap
				var obj crdsv1.Config
				err := Suite.K8sClient.Get(ctx, fn.NN(config.Namespace, config.Name), &obj)
				g.Expect(err).NotTo(HaveOccurred())

				g.Expect(Suite.K8sClient.Delete(ctx, &obj)).NotTo(HaveOccurred())

				err = Suite.K8sClient.Get(ctx, fn.NN(config.Namespace, config.Name), &cfg)
				g.Expect(apiErrors.IsNotFound(err)).To(BeTrue())
			})
		})
	})
})
