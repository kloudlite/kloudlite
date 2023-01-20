package _old

import (
	"context"
	"fmt"
	"time"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	fn "github.com/kloudlite/operator/pkg/functions"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const TestNamespace = "ginkgo-test-1"

var YamlApp = fmt.Sprintf(`
---
apiVersion: "crds.kloudlite.io/v1"
kind: App
metadata:
  name: sample
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
`, TestNamespace)

var _ = Describe(
	"App controller", func() {
		Context(
			"Initialize an App CR", func() {
				It(
					"Should Succeed", func() {
						err := k8sYamlClient.ApplyYAML(
							context.TODO(), []byte(YamlApp),
						)
						Expect(err).NotTo(HaveOccurred())
					},
				)
			},
		)

		Context("Check app resource", func() {
			It(
				"App should be patched with a default svc account", func() {
					Eventually(func() bool {
						app, err := GetObject(context.TODO(), fn.NN(TestNamespace, "sample"), &crdsv1.App{})
						if err != nil {
							return false
						}

						return app.Spec.ServiceAccount != ""
					}).WithPolling(1 * time.Second).WithTimeout(3 * time.Second).Should(Equal(true))
				},
			)

			It("App has been labelled with images-hash correctly, for restarting purposes", func() {
				Eventually(func() bool {
					app, err := GetObject[*crdsv1.App](context.TODO(), fn.NN(TestNamespace, "sample"), &crdsv1.App{})
					if err != nil {
						return false
					}

					for i := range app.Spec.Containers {
						imageSha := fn.Sha1Sum([]byte(app.Spec.Containers[i].Image))
						if app.Labels[fmt.Sprintf("kloudlite.io/image-%s", imageSha)] != "true" {
							return false
						}
					}

					return true
				}).WithPolling(1 * time.Second).WithTimeout(3 * time.Second).Should(Equal(true))
			})

			It("App pods have all the labels defined in the app metadata", func() {
				Eventually(func() bool {
					app, err := GetObject[*crdsv1.App](context.TODO(), fn.NN(TestNamespace, "sample"), &crdsv1.App{})
					if err != nil {
						return false
					}

					if app.GetLabels()["kloudlite.io/cluster"] == "cluster" &&
						app.GetLabels()["kloudlite.io/region"] == "region" &&
						app.GetLabels()["kloudlite.io/account-ref"] == "account-ref" {
						return true
					}

					return false
				}).WithPolling(1 * time.Second).WithTimeout(3 * time.Second).Should(Equal(true))
			})
		})
	},
)
